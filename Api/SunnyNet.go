package Api

import "C"
import (
	"bytes"
	"fmt"
	"github.com/qtgolang/SunnyNet/SunnyNet"
	"github.com/qtgolang/SunnyNet/src/Call"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"io/ioutil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetSunnyVersion 获取SunnyNet版本
func GetSunnyVersion() uintptr {
	return public.PointerPtr(public.SunnyVersion)
}

// SetRequestHeader 设置HTTP/S请求体中的协议头
func SetRequestHeader(MessageId int, name, val string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return
	}
	if k.Request.Header == nil {
		k.Request.Header = make(http.Header)
	}
	array := strings.Split(strings.ReplaceAll(val, "\r", ""), "\n")
	var arr []string
	for _, v := range array {
		if v != "" {
			arr = append(arr, v)
		}
	}
	k.Request.Header.SetArray(name, arr)
}

// SetRequestALLHeader 设置HTTP/S请求体中的全部协议头
func SetRequestALLHeader(MessageId int, value string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return
	}
	k.Request.Header = make(http.Header)
	arr := strings.Split(strings.ReplaceAll(value, "\r", ""), "\n")
	if len(arr) > 0 {
		for _, v := range arr {
			arr2 := strings.Split(v, ":")
			if len(arr2) >= 1 {
				key := strings.TrimSpace(arr2[0])
				if key != "" {
					if len(v) >= len(arr2[0])+1 {
						data := strings.TrimSpace(v[len(key)+1:])
						if len(k.Request.Header[key]) > 0 {
							k.Request.Header[key] = append(k.Request.Header[key], data)
						} else {
							k.Request.Header[key] = []string{data}
						}
					} else {
						if len(k.Request.Header[key]) < 1 {
							k.Request.Header[key] = []string{}
						}
					}
				}
			}
		}
	}
}

// SetRequestProxy 设置HTTP/S请求代理，仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
func SetRequestProxy(MessageId int, ProxyUrl string, outTime int) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Proxy == nil {
		k.Proxy, _ = SunnyProxy.ParseProxy(ProxyUrl, outTime)
	}
	if k.Proxy == nil {
		return false
	}
	return true
}

// SetRequestHTTP2Config .版本 2
//
// .DLL命令 Sunny_SetRequestHTTP2Config, 逻辑型, "Sunny.dll", "@SetRequestHTTP2Config", , //设置HTTP 2.0 请求指纹配置 (若服务器支持则使用,若服务器不支持,设置了也不会使用)
func SetRequestHTTP2Config(MessageId int, h2Config string) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	if k.TlsConfig == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.TlsConfig.NextProtos = []string{http.H11Proto, http.H2Proto}
	if h2Config != "" {
		c, e := http.StringToH2Config(h2Config)
		if e != nil {
			k.Request.SetHTTP2Config(nil)
			return false
		}
		k.Request.SetHTTP2Config(c)
		return true
	}
	k.Request.SetHTTP2Config(nil)
	return false
}

// GetResponseStatusCode 获取HTTP/S返回的状态码
func GetResponseStatusCode(MessageId int) int {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return -1
	}
	if k == nil {
		return -1
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return -1
	}
	return k.Response.StatusCode
}

// GetRequestClientIp 获取当前HTTP/S请求由哪个IP发起
func GetRequestClientIp(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return public.PointerPtr(k.Conn.RemoteAddr().String())
}
func GetHttpServerName(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return public.PointerPtr(k.ServerName)
}

// GetResponseStatus 获取HTTP/S返回的状态文本 例如 [200 OK]
func GetResponseStatus(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	k.Response.Status = strconv.Itoa(k.Response.StatusCode) + public.Space + http.StatusText(k.Response.StatusCode)
	return public.PointerPtr(k.Response.Status)
}

// SetResponseStatus 修改HTTP/S返回的状态码
func SetResponseStatus(MessageId, code int) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		k.Response.Response = new(http.Response)
		k.Response.Header = make(http.Header)
		k.Response.Header.Set("Connection", "Close")
		k.Response.ContentLength = 0
	}
	k.Response.StatusCode = code
	k.Response.Status = strconv.Itoa(code) + public.Space + http.StatusText(code)
}

// DelResponseHeader 删除HTTP/S返回数据中指定的协议头
func DelResponseHeader(MessageId int, name string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return
	}
	if k.Response.Header == nil {
		k.Response.Header = make(http.Header)
	}
	k.Response.Header.Del(name)
}

// DelRequestHeader 删除HTTP/S请求数据中指定的协议头
func DelRequestHeader(MessageId int, name string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return
	}
	if k.Request.Header == nil {
		k.Request.Header = make(http.Header)
	}
	k.Request.Header.Del(name)
}

// SetRequestCipherSuites 设置CipherSuites
func SetRequestCipherSuites(MessageId int) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	if k.TlsConfig == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.TlsConfig.CipherSuites = k.Global.GetTLSTestValues()
	return true
}

// GetRequestCipherSuites 获取CipherSuites
func GetRequestCipherSuites(MessageId int) string {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return ""
	}
	if k == nil {
		return ""
	}
	if k.TlsConfig == nil {
		return ""
	}

	k.Lock.Lock()
	defer k.Lock.Unlock()
	s := ""
	for _, v := range k.TlsConfig.CipherSuites {
		if s == "" {
			s = strconv.Itoa(int(v))
		} else {
			s += "," + strconv.Itoa(int(v))
		}
	}
	return s
}

// SetRequestOutTime 请求设置超时-毫秒
func SetRequestOutTime(MessageId int, times int) {

	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.SendTimeout = time.Duration(times) * time.Millisecond

}

// SetRequestUrl 修改HTTP/S当前请求的URL
func SetRequestUrl(MessageId int, URI string) bool {
	f := URI
	arr := strings.Split(f, "/")
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return false
	}
	Host := k.Request.Host
	if len(arr) >= 3 {
		Host = arr[2]
	}
	_u, _ := url.Parse(f)
	if _u == nil {
		if strings.HasSuffix(f, public.HttpRequestPrefix) || strings.HasSuffix(f, public.HttpsRequestPrefix) {
			return false
		}
		_u, _ = url.Parse(public.HttpRequestPrefix + f)
		if _u == nil {
			return false
		}
	}
	k.Request.Host = Host
	k.Request.URL = _u
	k.Request.RequestURI = ""
	return true
}

// SetRequestCookie 修改、设置 HTTP/S当前请求数据中指定Cookie
func SetRequestCookie(MessageId int, name, val string) {
	Cookie := public.NULL
	books := false
	sn := name
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return
	}
	values := k.Request.Cookies()
	for i := 0; i < len(values); i++ {
		if values[i].Name == sn {
			books = true
			Cookie += values[i].Name + "=" + val + "; "
		} else {
			Cookie += values[i].Name + "=" + values[i].Value + "; "
		}
	}
	if books == false {
		Cookie += sn + "=" + val + "; "
	}

	if k.Request.Header == nil {
		k.Request.Header = make(http.Header)
	}
	k.Request.Header.Set("Cookie", Cookie)
}

// SetRequestAllCookie 修改、设置 HTTP/S当前请求数据中的全部Cookie
func SetRequestAllCookie(MessageId int, val string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return
	}
	if k.Request.Header == nil {
		k.Request.Header = make(http.Header)
	}
	k.Request.Header.Set("Cookie", val)
}

// GetRequestHeader 获取 HTTP/S当前请求数据中的指定协议头
func GetRequestHeader(MessageId int, name string) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return public.NULLPtr
	}
	if k.Request.Header == nil {
		k.Request.Header = make(http.Header)
	}
	val := k.Request.Header.GetArray(name)
	if strings.EqualFold(name, "cookie") {
		return public.PointerPtr(strings.Join(val, "; "))
	}
	if len(val) < 1 {
		return public.NULLPtr
	}
	s := ""
	for i, vv := range val {
		if i == 0 {
			s = vv
		} else {
			s += "\r\n" + vv
		}
	}
	if len(s) > 0 {
		return public.PointerPtr(s)
	}
	return public.NULLPtr
}

// SetResponseHeader 修改、设置 HTTP/S当前返回数据中的指定协议头
func SetResponseHeader(MessageId int, name string, val string) {

	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		k.Response.Response = new(http.Response)
		k.Response.Header = make(http.Header)
		k.Response.Header.Set("Connection", "Close")
		k.Response.ContentLength = 0
	}
	if k.Response.Header == nil {
		k.Response.Header = make(http.Header)
	}
	arr := strings.Split(strings.ReplaceAll(val, "\r", ""), "\n")
	var array []string
	for _, v := range arr {
		if v != "" {
			array = append(array, v)
		}
	}
	k.Response.Header.SetArray(name, array)
}

// SetResponseAllHeader 修改、设置 HTTP/S当前返回数据中的全部协议头，例如设置返回两条Cookie 使用本命令设置 使用设置、修改 单条命令无效
func SetResponseAllHeader(MessageId int, value string) {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return
	}
	if k == nil {
		return
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		k.Response.Response = new(http.Response)
		k.Response.Header = make(http.Header)
		k.Response.Header.Set("Connection", "Close")
		k.Response.ContentLength = 0
	}
	if k.Response.Header == nil {
		k.Response.Header = make(http.Header)
	}
	arr := strings.Split(strings.ReplaceAll(value, "\r", ""), "\n")
	if len(arr) > 0 {
		k.Response.Header = make(http.Header)
		for _, v := range arr {
			arr2 := strings.Split(v, ":")
			if len(arr2) >= 1 {
				name := arr2[0]
				if name == "" {
					continue
				}
				if len(v) >= len(name)+1 {
					data := strings.TrimSpace(v[len(name)+1:])
					k.Response.Header.Add(name, data)
				} else {
					k.Response.Header.SetArray(name, []string{})
				}
			}
		}
	}
}

// GetRequestCookie 获取 HTTP/S当前请求数据中指定的Cookie
func GetRequestCookie(MessageId int, name string) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return public.NULLPtr
	}
	val, E := k.Request.Cookie(name)
	if E != nil {
		return public.NULLPtr
	}
	return public.PointerPtr(val.Name + "=" + val.Value + "; ")
}

// SetResponseData 设置、修改 HTTP/S 当前请求返回数据 如果再发起请求时调用本命令，请求将不会被发送，将会直接返回 data=数据指针  dataLen=数据长度
func SetResponseData(MessageId int, data uintptr, dataLen int) bool {
	n := public.CStringToBytes(data, dataLen)
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		k.Response.Response = new(http.Response)
		k.Response.Header = make(http.Header)
		k.Response.Header.Set("Server", "Sunny")
		k.Response.Header.Set("Accept-Ranges", "bytes")
		k.Response.Header.Set("Connection", "Close")
	}
	if k.Response.Header == nil {
		k.Response.Header = make(http.Header)
	}
	k.Response.Header.Set("Content-Length", strconv.Itoa(len(n)))
	k.Response.ContentLength = int64(len(n))
	k.Response.Body = ioutil.NopCloser(bytes.NewBuffer(n))
	return true
}

// GetRequestBody 获取 HTTP/S 当前POST提交数据 返回 数据指针
func GetRequestBody(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	if k.Request == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	body := k.Request.GetData()
	if body != nil {
		return public.PointerPtr(body)
	}
	return public.NULLPtr
}

// GetRequestBodyLen 获取 HTTP/S 当前请求POST提交数据长度
func GetRequestBodyLen(MessageId int) int {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return 0
	}
	if k == nil {
		return 0
	}
	if k.Request == nil {
		return 0
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	body := k.Request.GetData()
	return len(body)
}

// GetResponseBodyLen 获取 HTTP/S 当前返回  数据长度
func GetResponseBodyLen(MessageId int) int {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return 0
	}
	if k == nil {
		return 0
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return 0
	}
	if k.Response.Body != nil {
		bodyBytes, e := ioutil.ReadAll(k.Response.Body)
		k.Response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		if e != nil {
			return 0
		}
		return len(bodyBytes)
	}
	return 0
}

// SetRequestData 设置、修改 HTTP/S 当前请求POST提交数据  data=数据指针  dataLen=数据长度
func SetRequestData(MessageId int, data uintptr, dataLen int) bool {
	n := public.CStringToBytes(data, dataLen)
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	if k.Request == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.Request.SetData(n)
	return true
}

// IsRequestRawBody 此请求是否为原始body 如果是 将无法修改提交的Body，请使用 RawRequestDataToFile 命令来储存到文件
func IsRequestRawBody(MessageId int) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	return k.IsRequestRawBody()
}

// RawRequestDataToFile 获取 HTTP/S 当前POST提交数据原始Data,传入保存文件名路径,例如"c:\1.txt"
func RawRequestDataToFile(MessageId int, saveFileName string) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	return k.RawRequestDataToFile(saveFileName)
}

// GetResponseBody 获取 HTTP/S 当前返回数据  返回 数据指针
func GetResponseBody(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	if k.Response.Body != nil {
		bodyBytes, _ := ioutil.ReadAll(k.Response.Body)
		k.Response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return public.PointerPtr(bodyBytes)
	}
	return public.NULLPtr
}

// GetRequestALLCookie 获取 HTTP/S 当前请求全部Cookie
func GetRequestALLCookie(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return public.NULLPtr
	}
	val := k.Request.Cookies()
	Cookie := public.NULL
	for i := 0; i < len(val); i++ {
		Cookie += val[i].Name + "=" + val[i].Value + "; "
	}
	return public.PointerPtr(Cookie)
}

// GetRequestProto 获取 HTTPS 请求的协议版本
func GetRequestProto(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	if k.Request == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return public.PointerPtr(k.Request.Proto)
}

// GetResponseProto 获取 HTTPS 响应的协议版本
func GetResponseProto(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	return public.PointerPtr(k.Response.Proto)
}

// GetResponseAllHeader 获取 HTTP/S 当前返回全部协议头
func GetResponseAllHeader(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	if k.Response.Header == nil {
		return public.NULLPtr
	}
	Head := public.NULL
	var key []string
	for value, _ := range k.Response.Header {
		key = append(key, value)
	}
	sort.Strings(key)
	for _, kv := range key {
		for _, value := range k.Response.Header[kv] {
			Head += kv + ": " + value + "\r\n"
		}
	}
	return public.PointerPtr(Head)
}

// GetResponseHeader 获取 HTTP/S 当前返回数据中指定的协议头
func GetResponseHeader(MessageId int, name string) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	if k.Response.Header == nil {
		return public.NULLPtr
	}
	Head := k.Response.Header.GetArray(name)
	if len(Head) < 1 {
		return public.NULLPtr
	}
	s := ""
	for i, vv := range Head {
		if i == 0 {
			s = vv
		} else {
			s += "\r\n" + vv
		}
	}
	if len(s) > 0 {
		return public.PointerPtr(s)
	}
	return public.NULLPtr
}

// GetResponseServerAddress 获取 HTTP/S 相应的服务器地址
func GetResponseServerAddress(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Response.Response == nil {
		return public.NULLPtr
	}
	return public.PointerPtr(k.Response.ServerIP)
}

// GetRequestAllHeader 获取 HTTP/S 当前请求数据全部协议头
func GetRequestAllHeader(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.Request == nil {
		return public.NULLPtr
	}
	if k.Request.Header == nil {
		return public.NULLPtr
	}
	Head := public.NULL
	var key []string
	for value, _ := range k.Request.Header {
		key = append(key, value)
	}
	sort.Strings(key)
	for _, kv := range key {
		if strings.EqualFold(kv, "cookie") {
			Head += kv + ": " + strings.Join(k.Request.Header[kv], "; ") + "\r\n"
			continue
		}
		for _, value := range k.Request.Header[kv] {
			Head += kv + ": " + value + "\r\n"
		}
	}
	return public.PointerPtr(Head)
}

// SetTcpBody 修改 TCP消息数据 MsgType=1 发送的消息 MsgType=2 接收的消息 如果 MsgType和MessageId不匹配，将不会执行操作  data=数据指针  dataLen=数据长度
func SetTcpBody(MessageId, MsgType int, data uintptr, dataLen int) bool {
	n := public.CStringToBytes(data, dataLen)
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if MsgType == 1 {
		if k.TCP.Send == nil {
			return false
		}
		k.TCP.Send.Data.Reset()
		k.TCP.Send.Data.Write(n)
	}
	if MsgType == 2 {
		if k.TCP.Receive == nil {
			return false
		}
		k.TCP.Receive.Data.Reset()
		k.TCP.Receive.Data.Write(n)
	}
	return true
}

// SetTcpAgent 给当前TCP连接设置S5代理 仅先TCP回调 即将连接时使用
func SetTcpAgent(MessageId int, ProxyUrl string) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	if k.TCP.Send == nil {
		return false
	}
	proxy, err := SunnyProxy.ParseProxy(ProxyUrl)
	if err != nil || proxy == nil {
		return false
	}
	k.TCP.Send.Proxy, _ = SunnyProxy.ParseProxy(ProxyUrl, 30000)
	return true
}

// TcpCloseClient 根据唯一ID关闭指定的TCP连接  唯一ID在回调参数中
func TcpCloseClient(theology int) bool {
	SunnyNet.TcpSceneLock.Lock()
	w := SunnyNet.TcpStorage[theology]
	SunnyNet.TcpSceneLock.Unlock()
	if w == nil {
		return false
	}
	w.L.Lock()
	if w.ConnSend != nil {
		_ = w.ConnSend.Close()
	}
	if w.ConnServer != nil {
		_ = w.ConnServer.Close()
	}
	w.L.Unlock()
	return true
}

// SetTcpConnectionIP 给指定的TCP连接 修改目标连接地址 目标地址必须带端口号 例如 baidu.com:443
func SetTcpConnectionIP(MessageId int, data string) bool {
	k, ok := SunnyNet.GetSceneProxyRequest(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	if k.TCP.Send == nil {
		return false
	}
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.TCP.Send.Data.Reset()
	k.TCP.Send.Data.WriteString(data)
	return true
}

// TcpSendMsg 指定的TCP连接 模拟客户端向服务器端主动发送数据
func TcpSendMsg(theology int, data uintptr, dataLen int) int {
	n := public.CStringToBytes(data, dataLen)
	SunnyNet.TcpSceneLock.Lock()
	w := SunnyNet.TcpStorage[theology]
	SunnyNet.TcpSceneLock.Unlock()
	if w == nil {
		return 0
	}
	if w.Send == nil {
		return 0
	}
	w.L.Lock()
	defer w.L.Unlock()
	if len(n) > 0 {
		x, e := w.ReceiveBw.Write(n)
		if e == nil {
			_ = w.ReceiveBw.Flush()
		}
		return x
	}
	return 0
}

// TcpSendMsgClient 指定的TCP连接 模拟服务器端向客户端主动发送数据
func TcpSendMsgClient(theology int, data uintptr, dataLen int) int {
	n := public.CStringToBytes(data, dataLen)
	SunnyNet.TcpSceneLock.Lock()
	w := SunnyNet.TcpStorage[theology]
	SunnyNet.TcpSceneLock.Unlock()
	if w == nil {
		return 0
	}
	if w.Receive == nil {
		return 0
	}
	if len(n) > 0 {
		w.L.Lock()
		defer w.L.Unlock()
		x, e := w.SendBw.Write(n)
		if e == nil {
			_ = w.SendBw.Flush()
		}
		return x
	}
	return 0
}

// CloseWebsocket 主动关闭Websocket
func CloseWebsocket(Theology int) bool {
	m, ok := SunnyNet.GetSceneProxyRequest(Theology)
	if ok == false {
		return false
	}
	if m == nil {
		return false
	}
	k := m.Websocket
	if k == nil {
		return false
	}
	k.Sync.Lock()
	if k.Server != nil {
		_ = k.Server.Close()
	}
	if k.Client != nil {
		_ = k.Client.Close()
	}
	k.Sync.Unlock()
	return true
}

// GetWebsocketBodyLen 获取 WebSocket消息长度
func GetWebsocketBodyLen(MessageId int) int {
	k, ok := SunnyNet.GetSceneWebSocketMsg(MessageId)
	if ok == false {
		return 0
	}
	if k == nil {
		return 0
	}
	k.Sync.Lock()
	defer k.Sync.Unlock()
	return k.Data.Len()
}

// GetWebsocketBody 获取 WebSocket消息 返回数据指针
func GetWebsocketBody(MessageId int) uintptr {
	k, ok := SunnyNet.GetSceneWebSocketMsg(MessageId)
	if ok == false {
		return public.NULLPtr
	}
	if k == nil {
		return public.NULLPtr
	}
	k.Sync.Lock()
	defer k.Sync.Unlock()
	return public.PointerPtr(k.Data.String())
}

// SetWebsocketBody 修改 WebSocket消息 data=数据指针  dataLen=数据长度
func SetWebsocketBody(MessageId int, data uintptr, dataLen int) bool {
	n := public.CStringToBytes(data, dataLen)
	k, ok := SunnyNet.GetSceneWebSocketMsg(MessageId)
	if ok == false {
		return false
	}
	if k == nil {
		return false
	}
	k.Sync.Lock()
	k.Data.Reset()
	k.Data.Write(n)
	k.Sync.Unlock()
	return true
}

// SendWebsocketBody 主动向Websocket服务器发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
func SendWebsocketBody(Theology, MessageType int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	m, ok := SunnyNet.GetSceneProxyRequest(Theology)
	if ok == false {
		return false
	}
	if m == nil {
		return false
	}
	if m.Websocket == nil {
		return false
	}
	if m.Websocket.Sync == nil {
		return false
	}
	if m.Websocket.Server == nil {
		return false
	}
	m.Websocket.Sync.Lock()
	e := m.Websocket.Server.WriteMessage(MessageType, bs)
	m.Websocket.Sync.Unlock()
	if e != nil {
		return false
	}
	return true
}

// SendWebsocketClientBody 主动向Websocket客户端发送消息 MessageType=WS消息类型 data=数据指针  dataLen=数据长度
func SendWebsocketClientBody(Theology, MessageType int, data uintptr, dataLen int) bool {
	bs := public.CStringToBytes(data, dataLen)
	m, ok := SunnyNet.GetSceneProxyRequest(Theology)
	if ok == false {
		return false
	}
	if m == nil {
		return false
	}
	if m.Websocket == nil {
		return false
	}
	if m.Websocket.Sync == nil {
		return false
	}
	if m.Websocket.Client == nil {
		return false
	}
	m.Websocket.Sync.Lock()
	e := m.Websocket.Client.WriteMessage(MessageType, bs)
	m.Websocket.Sync.Unlock()
	if e != nil {
		return false
	}
	return true
}

// CreateSunnyNet 创建Sunny中间件对象,可创建多个
func CreateSunnyNet() int {
	Sunny := SunnyNet.NewSunny()
	SunnyNet.SunnyStorageLock.Lock()
	SunnyNet.SunnyStorage[Sunny.SunnyContext] = Sunny
	SunnyNet.SunnyStorageLock.Unlock()
	return Sunny.SunnyContext
}

// ReleaseSunnyNet 释放SunnyNet
func ReleaseSunnyNet(SunnyContext int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	defer SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Close()
	delete(SunnyNet.SunnyStorage, SunnyContext)
	return true
}

// SetHTTPRequestMaxUpdateLength 设置HTTP请求,提交数据,最大的长度
func SetHTTPRequestMaxUpdateLength(SunnyContext int, i int64) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetHTTPRequestMaxUpdateLength(i)
	return true
}

// SunnyNetStart 启动Sunny中间件 成功返回true
func SunnyNetStart(SunnyContext int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Start()
	return w.Error == nil
}

// SunnyNetSetPort 设置指定端口 Sunny中间件启动之前调用
func SunnyNetSetPort(SunnyContext, Port int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetPort(Port)
	return true
}

// SunnyNetClose 关闭停止指定Sunny中间件
func SunnyNetClose(SunnyContext int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Close()
	return true
}

// SunnyNetSetCert 设置自定义证书
func SunnyNetSetCert(SunnyContext, CertificateManagerId int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetCert(CertificateManagerId)
	return true
}

// SunnyNetInstallCert 安装证书 将证书安装到Windows系统内
func SunnyNetInstallCert(SunnyContext int) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return public.PointerPtr("SunnyNet no exist")
	}
	return public.PointerPtr(w.InstallCert())
}

// SunnyNetSetCallback 是否中间件回调地址 httpCallback =HTTP、Websocket 回调地址  tcpCallback=TCP回调地址
func SunnyNetSetCallback(SunnyContext, httpCallback, tcpCallback, wsCallback, udpCallback int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetCallback(httpCallback, tcpCallback, wsCallback, udpCallback)
	return true
}

// SunnyNetVerifyUser 开启或关闭身份验证模式
func SunnyNetVerifyUser(SunnyContext int, open bool) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Socket5VerifyUser(open)
	return true
}

// SunnyNetSocket5AddUser 添加 S5代理需要验证的用户名
func SunnyNetSocket5AddUser(SunnyContext int, User, Pass string) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Socket5AddUser(User, Pass)
	return true
}

// SunnyNetSocket5DelUser 删除 S5需要验证的用户名
func SunnyNetSocket5DelUser(SunnyContext int, User string) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.Socket5DelUser(User)
	return true
}

// SunnyNetGetSocket5User 开启身份验证模式后 获取授权的S5账号,注意UDP请求无法获取到授权的s5账号
func SunnyNetGetSocket5User(Theology int) uintptr {
	return public.PointerPtr(SunnyNet.GetSocket5User(Theology))
}

// SunnyNetError 获取中间件启动时的错误信息
func SunnyNetError(SunnyContext int) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return public.NULLPtr
	}
	if w.Error == nil {
		return public.NULLPtr
	}
	return public.PointerPtr(w.Error.Error())
}

// SunnyNetMustTcp 设置中间件是否开启强制走TCP
func SunnyNetMustTcp(SunnyContext int, open bool) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return
	}
	w.MustTcp(open)
}

// CompileProxyRegexp 创建上游代理使用规则
func CompileProxyRegexp(SunnyContext int, Regexp string) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	return w.CompileProxyRegexp(Regexp) == nil
}

// SetMustTcpRegexp 设置强制走TCP规则,如果 打开了全部强制走TCP状态,本功能则无效
func SetMustTcpRegexp(SunnyContext int, Regexp string, RulesAllow bool) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	return w.SetMustTcpRegexp(Regexp, RulesAllow) == nil
}

// SetGlobalProxy 设置全局上游代理 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
func SetGlobalProxy(SunnyContext int, ProxyAddress string, outTime int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		return w.SetGlobalProxy(ProxyAddress, outTime)
	}
	return false
}

// ExportCert 导出已设置的证书
func ExportCert(SunnyContext int) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		return public.PointerPtr(w.ExportCert())
	}
	return 0
}

// SetIeProxy 设置IE代理 [Off=true 取消] [Off=false 设置] 在中间件设置端口后调用
func SetIeProxy(SunnyContext int, Off bool) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	return w.SetIeProxy(Off)
}

// StartProcess 开始进程代理 只允许一个 SunnyNet 使用
func StartProcess(SunnyContext int) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		return w.StartProcess()
	}
	return false
}

// ProcessALLName 设置是否全部进程通过  所有 SunnyNet 通用
func ProcessALLName(SunnyContext int, open bool) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessALLName(open)
	}
}

// ProcessDelName 进程代理 删除进程名
func ProcessDelName(SunnyContext int, s string) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessDelName(s)
	}
}

// ProcessAddName 进程代理 添加进程名
func ProcessAddName(SunnyContext int, s string) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessAddName(s)
	}
}

// ProcessDelPid 进程代理 删除PID
func ProcessDelPid(SunnyContext, pid int) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessDelPid(pid)
	}
}

// ProcessAddPid 进程代理 添加PID
func ProcessAddPid(SunnyContext, pid int) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessAddPid(pid)
	}
}

// ProcessCancelAll 进程代理 取消全部已设置的进程名
func ProcessCancelAll(SunnyContext int) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		w.ProcessCancelAll()
	}
}

// SetScriptCode 加载用户的脚本代码
func SetScriptCode(SunnyContext int, code string) string {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		return w.SetScriptCode(code)
	}
	return "SunnyContext Error"
}

// SetScriptCall 设置脚本代码的回调函数
func SetScriptCall(SunnyContext int, log, save uintptr) {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w != nil {
		l := func(info ...any) {
			Call.Call(int(log), fmt.Sprintf("%v", info))
		}
		s := func(code []byte) {
			Call.Call(int(save), code, int32(len(code)))
		}
		w.SetScriptCall(l, s)
	}
}

// SetScriptPage  设置脚本编辑器页面 需不少于8个字符
func SetScriptPage(SunnyContext int, Page string) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return 0
	}
	return public.PointerPtr(w.SetScriptPage(Page))
}

// DisableTCP  禁用TCP 仅对当前SunnyContext有效
func DisableTCP(SunnyContext int, Disable bool) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.DisableTCP(Disable)
	return true
}

// SetRandomFixedTLS  设置密码套件，改变ja3指纹,仅对当前SunnyContext有效
func SetRandomFixedTLS(SunnyContext int, data []byte) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetRandomFixedTLS(string(data))
	return true
}

// RandomFixedTLSGet 随机生成一个密码套件,
func RandomFixedTLSGet(SunnyContext int) uintptr {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return 0
	}
	r := w.GetTLSTestValues()
	s := ""
	for _, v := range r {
		if s == "" {
			s = strconv.Itoa(int(v))
		} else {
			s += "," + strconv.Itoa(int(v))
		}
	}
	return public.PointerPtr(s)
}

// SetRandomTLS  是否使用随机TLS指纹 仅对当前SunnyContext有效
func SetRandomTLS(SunnyContext int, open bool) bool {
	SunnyNet.SunnyStorageLock.Lock()
	w := SunnyNet.SunnyStorage[SunnyContext]
	SunnyNet.SunnyStorageLock.Unlock()
	if w == nil {
		return false
	}
	w.SetRandomTLS(open)
	return true
}
