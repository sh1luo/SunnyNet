// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests of internal functions and things with no better homes.

package http

import (
	"bytes"
	"internal/testenv"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestForeachHeaderElement(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"Foo", []string{"Foo"}},
		{" Foo", []string{"Foo"}},
		{"Foo ", []string{"Foo"}},
		{" Foo ", []string{"Foo"}},

		{"foo", []string{"foo"}},
		{"anY-cAsE", []string{"anY-cAsE"}},

		{"", nil},
		{",,,,  ,  ,,   ,,, ,", nil},

		{" Foo,Bar, Baz,lower,,Quux ", []string{"Foo", "Bar", "Baz", "lower", "Quux"}},
	}
	for _, tt := range tests {
		var got []string
		foreachHeaderElement(tt.in, func(v string) {
			got = append(got, v)
		})
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("foreachHeaderElement(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestCleanHost(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"www.google.com", "www.google.com"},
		{"www.google.com foo", "www.google.com"},
		{"www.google.com/foo", "www.google.com"},
		{" first character is a space", ""},
		{"[1::6]:8080", "[1::6]:8080"},

		// Punycode:
		{"гофер.рф/foo", "xn--c1ae0ajs.xn--p1ai"},
		{"bücher.de", "xn--bcher-kva.de"},
		{"bücher.de:8080", "xn--bcher-kva.de:8080"},
		// Verify we convert to lowercase before punycode:
		{"BÜCHER.de", "xn--bcher-kva.de"},
		{"BÜCHER.de:8080", "xn--bcher-kva.de:8080"},
		// Verify we normalize to NFC before punycode:
		{"gophér.nfc", "xn--gophr-esa.nfc"},            // NFC input; no work needed
		{"goph\u0065\u0301r.nfd", "xn--gophr-esa.nfd"}, // NFD input
	}
	for _, tt := range tests {
		got := cleanHost(tt.in)
		if tt.want != got {
			t.Errorf("cleanHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Test that cmd/go doesn't link in the HTTP server.
//
// This catches accidental dependencies between the HTTP transport and
// server code.
func TestCmdGoNoHTTPServer(t *testing.T) {
	t.Parallel()
	goBin := testenv.GoToolPath(t)
	out, err := exec.Command(goBin, "tool", "nm", goBin).CombinedOutput()
	if err != nil {
		t.Fatalf("go tool nm: %v: %s", err, out)
	}
	wantSym := map[string]bool{
		// Verify these exist: (sanity checking this test)
		"github.com/qtgolang/SunnyNet/src/http.(*Client).do":           true,
		"github.com/qtgolang/SunnyNet/src/http.(*Transport).RoundTrip": true,

		// Verify these don't exist:
		"github.com/qtgolang/SunnyNet/src/http.http2Server":           false,
		"github.com/qtgolang/SunnyNet/src/http.(*Server).Serve":       false,
		"github.com/qtgolang/SunnyNet/src/http.(*ServeMux).ServeHTTP": false,
		"github.com/qtgolang/SunnyNet/src/http.DefaultServeMux":       false,
	}
	for sym, want := range wantSym {
		got := bytes.Contains(out, []byte(sym))
		if !want && got {
			t.Errorf("cmd/go unexpectedly links in HTTP server code; found symbol %q in cmd/go", sym)
		}
		if want && !got {
			t.Errorf("expected to find symbol %q in cmd/go; not found", sym)
		}
	}
}

// Tests that the nethttpomithttp2 build tag doesn't rot too much,
// even if there's not a regular builder on it.
func TestOmitHTTP2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()
	goTool := testenv.GoToolPath(t)
	out, err := exec.Command(goTool, "test", "-short", "-tags=nethttpomithttp2", "github.com/qtgolang/SunnyNet/src/http").CombinedOutput()
	if err != nil {
		t.Fatalf("go test -short failed: %v, %s", err, out)
	}
}

// Tests that the nethttpomithttp2 build tag at least type checks
// in short mode.
// The TestOmitHTTP2 test above actually runs tests (in long mode).
func TestOmitHTTP2Vet(t *testing.T) {
	t.Parallel()
	goTool := testenv.GoToolPath(t)
	out, err := exec.Command(goTool, "vet", "-tags=nethttpomithttp2", "github.com/qtgolang/SunnyNet/src/http").CombinedOutput()
	if err != nil {
		t.Fatalf("go vet failed: %v, %s", err, out)
	}
}

var valuesCount int

func BenchmarkCopyValues(b *testing.B) {
	b.ReportAllocs()
	src := url.Values{
		"a": {"1", "2", "3", "4", "5"},
		"b": {"2", "2", "3", "4", "5"},
		"c": {"3", "2", "3", "4", "5"},
		"d": {"4", "2", "3", "4", "5"},
		"e": {"1", "1", "2", "3", "4", "5", "6", "7", "abcdef", "l", "a", "b", "c", "d", "z"},
		"j": {"1", "2"},
		"m": nil,
	}
	for i := 0; i < b.N; i++ {
		dst := url.Values{"a": {"b"}, "b": {"2"}, "c": {"3"}, "d": {"4"}, "j": nil, "m": {"x"}}
		copyValues(dst, src)
		if valuesCount = len(dst["a"]); valuesCount != 6 {
			b.Fatalf(`%d items in dst["a"] but expected 6`, valuesCount)
		}
	}
	if valuesCount == 0 {
		b.Fatal("Benchmark wasn't run")
	}
}

var forbiddenStringsFunctions = map[string]bool{
	// Functions that use Unicode-aware case folding.
	"EqualFold":      true,
	"Title":          true,
	"ToLower":        true,
	"ToLowerSpecial": true,
	"ToTitle":        true,
	"ToTitleSpecial": true,
	"ToUpper":        true,
	"ToUpperSpecial": true,

	// Functions that use Unicode-aware spaces.
	"Fields":    true,
	"TrimSpace": true,
}

// TestNoUnicodeStrings checks that nothing in net/http uses the Unicode-aware
// strings and bytes package functions. HTTP is mostly ASCII based, and doing
// Unicode-aware case folding or space stripping can introduce vulnerabilities.
func TestNoUnicodeStrings(t *testing.T) {
	if !testenv.HasSrc() {
		t.Skip("source code not available")
	}

	re := regexp.MustCompile(`(strings|bytes).([A-Za-z]+)`)
	if err := fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}

		if path == "internal/ascii" {
			return fs.SkipDir
		}
		if !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") ||
			path == "h2_bundle.go" || d.IsDir() {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for lineNum, line := range strings.Split(string(contents), "\n") {
			for _, match := range re.FindAllStringSubmatch(line, -1) {
				if !forbiddenStringsFunctions[match[2]] {
					continue
				}
				t.Errorf("disallowed call to %s at %s:%d", match[0], path, lineNum+1)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
