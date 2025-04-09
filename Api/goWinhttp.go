package Api

import (
	"bytes"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/Certificate"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/httpClient"
	"github.com/qtgolang/SunnyNet/src/public"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------
type request struct {
	resp      *http.Response
	req       *http.Request
	lock      sync.Mutex
	proxy     *SunnyProxy.Proxy
	outTime   int
	redirect  bool
	tlsConfig *tls.Config
	randomTLS bool
	respBody  []byte
}

var HTTPMap = make(map[int]*request)
var HTTPMapLock sync.Mutex

func LoadHTTPClient(Context int) *request {
	HTTPMapLock.Lock()
	s := HTTPMap[Context]
	HTTPMapLock.Unlock()
	if s == nil {
		return nil
	}
	return s
}

// 创建 HTTP 客户端
//
//export CreateHTTPClient
func CreateHTTPClient() int {
	Context := newMessageId()
	HTTPMapLock.Lock()
	HTTPMap[Context] = &request{req: &http.Request{}, tlsConfig: &tls.Config{NextProtos: public.HTTP2NextProtos}}
	HTTPMapLock.Unlock()
	return Context
}

// RemoveHTTPClient
// 释放 HTTP客户端
func RemoveHTTPClient(Context int) {
	HTTPMapLock.Lock()
	defer HTTPMapLock.Unlock()
	obj := HTTPMap[Context]
	if obj != nil {
		obj.lock.Lock()
		defer obj.lock.Unlock()
		if obj.req != nil {
			if obj.req.Body != nil {
				_ = obj.req.Body.Close()
			}
		}
		if obj.resp != nil {
			if obj.resp.Body != nil {
				_ = obj.resp.Body.Close()
			}
		}
	}
	delete(HTTPMap, Context)
}

// HTTPOpen
// HTTP 客户端 Open
func HTTPOpen(Context int, Method, URL string) {
	k := LoadHTTPClient(Context)
	if k == nil {
		return
	}

	k.lock.Lock()
	defer k.lock.Unlock()
	if k.req != nil {
		if k.req.Body != nil {
			_ = k.req.Body.Close()
		}
	}
	k.req, _ = http.NewRequest(Method, URL, nil)
}

// HTTPSetHeader
// HTTP 客户端 设置协议头
func HTTPSetHeader(Context int, name, value string) {
	k := LoadHTTPClient(Context)
	if k == nil {
		return
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	arr := strings.Split(strings.ReplaceAll(value, "\r", ""), "\n")
	for _, v := range arr {
		if v == "" {
			continue
		}
		k.req.Header.Add(name, v)
	}
}

// HTTPSetProxyIP
// HTTP 客户端 设置代理IP http://admin:pass@127.0.0.1:8888
func HTTPSetProxyIP(Context int, ProxyUrl string) bool {
	k := LoadHTTPClient(Context)
	if k == nil {
		return false
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	k.proxy, _ = SunnyProxy.ParseProxy(ProxyUrl)
	if k.outTime != 0 {
		k.proxy.SetTimeout(time.Duration(k.outTime) * time.Millisecond)
	}
	return k.proxy != nil
}

// HTTPSetTimeouts
// HTTP 客户端 设置超时 毫秒
func HTTPSetTimeouts(Context int, t1 int) {
	k := LoadHTTPClient(Context)
	if k == nil {
		return
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	k.outTime = t1
	if k.proxy != nil {
		k.proxy.SetTimeout(time.Duration(t1) * time.Millisecond)
	}
}

// HTTPSetServerIP
// HTTP 客户端 设置真实连接IP地址，
func HTTPSetServerIP(Context int, s string) {
	k := LoadHTTPClient(Context)
	if k == nil {
		return
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	k.req.SetContext(public.Connect_Raw_Address, func() string { return s })
}

// HTTPSendBin
// HTTP 客户端 发送Body
func HTTPSendBin(Context int, data []byte) {

	k := LoadHTTPClient(Context)
	if k == nil {
		return
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.req != nil {
		if k.req.Body != nil {
			_ = k.req.Body.Close()
		}
	}
	k.req.Body = io.NopCloser(bytes.NewReader(data))
	k.req.ContentLength = int64(len(data))
	if k.req.ContentLength < 1 {
		k.req.Body = nil
	}
	if k.req.ContentLength > 0 {
		k.req.Header["Content-Length"] = []string{fmt.Sprintf("%d", len(data))}
	} else {
		k.req.Header.Del("Content-Length")
	}
	var random func() []uint16
	if k.randomTLS {
		random = public.GetTLSValues
	}
	k.respBody = nil
	resp, _, _, f := httpClient.Do(k.req, k.proxy, k.redirect, k.tlsConfig, time.Duration(k.outTime)*time.Millisecond, random, nil)

	defer func() {
		if f != nil {
			f()
		}
	}()
	if k.resp != nil {
		if k.resp.Body != nil {
			_ = k.resp.Body.Close()
		}
	}
	k.resp = resp
	if k.resp != nil {
		if k.resp.Body != nil {
			i, _ := io.ReadAll(k.resp.Body)
			k.respBody = i
		}
	}
}

// HTTPGetBodyLen
// HTTP 客户端 返回响应长度
func HTTPGetBodyLen(Context int) int {
	k := LoadHTTPClient(Context)
	if k == nil {
		return 0
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.respBody == nil {
		return 0
	}
	return len(k.respBody)
}

// HTTPGetHeads
// HTTP 客户端 返回响应全部Heads
func HTTPGetHeads(Context int) string {
	k := LoadHTTPClient(Context)
	if k == nil {
		return ""
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.resp == nil {
		return ""
	}
	if k.resp.Header == nil {
		return ""
	}
	if len(k.resp.Header) < 1 {
		return ""
	}
	Head := ""
	var key []string
	for value, _ := range k.resp.Header {
		key = append(key, value)
	}
	sort.Strings(key)
	for _, kv := range key {
		for _, value := range k.resp.Header[kv] {
			if Head == "" {
				Head = kv + ": " + value
			} else {
				Head += "\r\n" + kv + ": " + value
			}
		}
	}
	return Head
}

// HTTPGetRequestHeader
// HTTP 客户端 添加的全部协议头
func HTTPGetRequestHeader(Context int) string {
	k := LoadHTTPClient(Context)
	if k == nil {
		return ""
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.req == nil {
		return ""
	}
	if k.req.Header == nil {
		return ""
	}
	if len(k.req.Header) < 1 {
		return ""
	}
	Head := ""
	var key []string
	for value, _ := range k.req.Header {
		key = append(key, value)
	}
	sort.Strings(key)
	for _, kv := range key {
		for _, value := range k.req.Header[kv] {
			if Head == "" {
				Head = kv + ": " + value
			} else {
				Head += "\r\n" + kv + ": " + value
			}
		}
	}
	return Head
}

// HTTPGetHeader
// HTTP 客户端 返回响应Header
func HTTPGetHeader(Context int, name string) string {
	k := LoadHTTPClient(Context)
	if k == nil {
		return ""
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.resp == nil {
		return ""
	}
	if k.resp.Header == nil {
		return ""
	}
	if len(k.resp.Header) < 1 {
		return ""
	}
	Head := ""
	for _, value := range k.resp.Header.GetArray(name) {
		if Head == "" {
			Head = value
		} else {
			Head += "\r\n" + value
		}
	}
	return Head
}

// HTTPGetBody
// HTTP 客户端 返回响应内容
func HTTPGetBody(Context int) []byte {
	k := LoadHTTPClient(Context)
	if k == nil {
		return nil
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.respBody == nil {
		return nil
	}
	if len(k.respBody) < 1 {
		return nil
	}
	return k.respBody
}

// HTTPGetCode
// HTTP 客户端 返回响应状态码
func HTTPGetCode(Context int) int {
	k := LoadHTTPClient(Context)
	if k == nil {
		return 0
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.resp == nil {
		return 0
	}
	return k.resp.StatusCode
}

// HTTPSetCertManager
// HTTP 客户端 设置证书管理器
func HTTPSetCertManager(Context, CertManagerContext int) bool {
	k := LoadHTTPClient(Context)
	if k == nil {
		return false
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	Certificate.Lock.Lock()
	defer Certificate.Lock.Unlock()
	c := Certificate.LoadCertificateContext(CertManagerContext)
	if c == nil {
		return false
	}
	if c.Tls == nil {
		return false
	}
	k.tlsConfig = c.Tls
	k.tlsConfig.NextProtos = public.HTTP2NextProtos
	return true
}

// HTTPSetRedirect
// HTTP 客户端 设置重定向
func HTTPSetRedirect(Context int, Redirect bool) bool {
	k := LoadHTTPClient(Context)
	if k == nil {
		return false
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	k.redirect = Redirect
	return true
}

// HTTPSetRandomTLS
// HTTP 客户端 设置随机使用TLS指纹
func HTTPSetRandomTLS(Context int, randomTLS bool) bool {
	k := LoadHTTPClient(Context)
	if k == nil {
		return false
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	k.randomTLS = randomTLS
	return true
}

// SetH2Config
// HTTP 客户端 设置HTTP2指纹
func SetH2Config(Context int, h2Config string) bool {
	k := LoadHTTPClient(Context)
	if k == nil {
		return false
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	c, e := http.StringToH2Config(h2Config)
	if e != nil {
		return false
	}
	k.req.SetHTTP2Config(c)
	return true
}
