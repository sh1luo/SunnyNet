package SunnyNet

import (
	"context"
	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"os"
	"sync"
	"time"
)

type objHook struct {
	*ReadWriteObject.ReadWriteObject
	aheadData []byte
}

func newObjHook(obj *ReadWriteObject.ReadWriteObject, aheadData []byte) *objHook {
	var hookObj objHook
	hookObj.ReadWriteObject = obj
	hookObj.aheadData = aheadData
	return &hookObj
}
func (n *objHook) Read(p []byte) (int, error) {
	if len(n.aheadData) < 1 {
		return n.ReadWriteObject.Read(p)
	}
	// 确定可以复制的最大字节数
	copyLength := len(n.aheadData)
	if len(p) < copyLength {
		copyLength = len(p)
	}
	copy(p, n.aheadData[:copyLength])
	n.aheadData = n.aheadData[copyLength:]
	if copyLength == 1 && copyLength < len(p) {
		a, e := n.ReadWriteObject.Read(p[copyLength:])
		return a + copyLength, e
	}
	return copyLength, nil
}
func (s *proxyRequest) h2Request(aheadData []byte) {
	http.H2NewConn(newObjHook(s.RwObj, aheadData), s.httpCall)
}
func (s *proxyRequest) h1Request(aheadData []byte) {
	http.H1NewConn(newObjHook(s.RwObj, aheadData), s.httpCall)
}
func (s *proxyRequest) httpCall(rw http.ResponseWriter, req *http.Request) {
	if req == nil {
		return
	}
	r := s.clone()
	defer r.free()
	ctx, ch := context.WithCancel(context.Background())
	defer ch()
	res := req.Clone(ctx)
	if res.GetIsNullBody() {
		if res.Body != nil {
			_ = res.Body.Close()
		}
		res.Body = nil
		res.ContentLength = 0
	} else {
		res.Body = &httpBody{Body: req.Body, c: s.Conn, req: res}
	}
	IsRequestRawBody := res.GetBodyLength() >= s.Global._http_max_body_len
	Length := res.GetBodyLength()
	res.SetContext(public.SunnyNetRawRequestBodyLength, Length)
	res.IsRawBody = IsRequestRawBody
	{
		res.RequestURI = ""
		if res.URL != nil {
			if s.defaultScheme == "" || req.URL.Scheme == "https" {
				res.URL.Scheme = "https"
			} else {
				res.URL.Scheme = s.defaultScheme
			}
			res.URL.Host = res.Host
			a := IpDns(res.Host)
			if a == "" && req.Header.Get("host") != "" {
				res.URL.Host = req.Header.Get("host")
				u, _ := url.Parse(res.URL.String())
				if u != nil {
					res.URL = u
					res.Host = u.Host
				}
			} else if a == "" && r.Target.Host != "" {
				res.URL.Host = r.Target.String()
				u, _ := url.Parse(res.URL.String())
				if u != nil {
					res.URL = u
					res.Host = u.Host
				}
			}
			p := res.URL.Port()
			if (p == "443" && res.URL.Scheme == "https") || (p == "80" && res.URL.Scheme == "http") {
				host, _, _ := net.SplitHostPort(res.Host)
				res.URL.Host = host
				res.Host = host
			}
		}

	}
	r.Response.rw = rw
	if Length < 4096 {
		res.ContentLength = Length
	} else {
		res.ContentLength = -1
	}
	if res.ProtoMajor == 2 {
		reHeader := make(http.Header)
		for k, v := range res.Header {
			name := textproto.CanonicalMIMEHeaderKey(k)
			if len(reHeader[name]) < 1 {
				reHeader[name] = v
			} else {
				reHeader[name] = append(reHeader[name], v...)
			}
		}
		res.Header = reHeader
	}
	r.sendHttp(res)
}

type httpBody struct {
	Body io.ReadCloser
	c    net.Conn
	req  *http.Request
	file io.WriteCloser
	init bool
	lock *sync.Mutex
}

func (h *httpBody) Read(p []byte) (n int, err error) {
	if !h.init {
		h.init = true
		SaveFilePath, ok := h.req.Context().Value(public.SunnyNetRawBodySaveFilePath).(string)
		if ok && SaveFilePath != "" {
			//防止多个请求写入同一个文件，导致闪退等问题
			_lock.Lock()
			lo := _lockfileMap[SaveFilePath]
			if lo == nil {
				lo = &sync.Mutex{}
				_lockfileMap[SaveFilePath] = lo
			}
			h.lock = lo
			_lock.Unlock()
			file, er1 := os.OpenFile(SaveFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if er1 == nil {
				h.file = file
			}
		}
	}
	if h.lock != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	}

	_ = h.c.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, e := h.Body.Read(p)
	if h.file != nil && n > 0 {
		_, _ = h.file.Write(p[0:n])
	}
	if h.file != nil && e != nil {
		_ = h.file.Close()
		h.file = nil
	}
	return n, e
}

func (h *httpBody) Close() error {
	if h.lock != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	}
	if h.file != nil {
		_ = h.file.Close()
		h.file = nil
	}
	return h.Body.Close()
}

var _lock sync.Mutex
var _lockfileMap = make(map[string]*sync.Mutex)