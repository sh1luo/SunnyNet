package Interface

import (
	"github.com/qtgolang/SunnyNet/src/http"
	"io"
)

// ConnHTTPScriptCall 脚本中使用
type ConnHTTPScriptCall interface {
	connHTTP
	/*
		SetDisplay
		设置是否显示当前请求(默认为true)
		如果设置为false 将不会通知回调进行处理消息
		仅在发起请求时有效
	*/
	SetDisplay(Display bool)
	/*
		SetBreak
		设置是否需要通知回调拦截该请求(默认为false)
		如果设置为true 回调函数中的err为Debug字符串,表示请求需要拦截
	*/
	SetBreak(Break bool)
}

// ConnHTTPCall go中使用
type ConnHTTPCall interface {
	connHTTP
	/*
		SetHTTP2Config
		设置HTTP 2.0 请求指纹配置 (若服务器支持则使用,若服务器不支持,设置了也不会使用)
	*/
	SetHTTP2Config(config string) bool
}

// go和脚本都使用
type connHTTP interface {
	general
	proxy
	/*
		Type

		返回当前消息事件类型

		请使用public.Http...

		1=发送请求 2=接收到响应 3=请求失败
	*/
	Type() int
	/*
		ClientIP
		返回请求的客户端IP地址
	*/
	ClientIP() string

	/*
		ServerIP
		返回请求的服务器IP地址
	*/
	ServerIP() string

	/*
		StopRequest
		阻止请求,仅支持在发起请求时使用
		StatusCode要响应的状态码
		Data=要响应的数据 可以是string 也可以是[]byte
		Header=要响应的Header 可以忽略
	*/
	StopRequest(StatusCode int, Data any, Header ...http.Header)

	/*
		Error
		获取当前请求错误信息
	*/
	Error() string

	/*
		URL
		返回请求的URL
	*/
	URL() string
	/*
		UpdateURL
		修改请求的URL,仅支持在发起请求时使用
	*/
	UpdateURL(NewUrl string) bool

	/*
		Proto
		返回请求的协议版本 例如 HTTP/1.1
	*/
	Proto() string

	/*
		Method
		返回请求的方法 例如 GET POST
	*/
	Method() string

	/*
		GetRequestHeader
		返回请求的Header
	*/
	GetRequestHeader() http.Header

	/*
		GetRequestBody
		返回POST请求提交的数据
		当请求提交数据超过一定大小时,请使用 SaveRawRequestData 命令
		(这个大小由自己设置默认:10240000 字节)
	*/
	GetRequestBody() []byte

	/*
		ResponseHeader
		返回响应的Header
	*/
	GetResponseHeader() http.Header
	/*
		GetResponseProto
		返回响应协议版本
		即服务器的版本 HTTP/1.1 或 HTTP/2.0
	*/
	GetResponseProto() string
	/*
		GetResponseBody
		返回响应的数据
	*/
	GetResponseBody() []byte

	/*
		SetResponseBody
		设置响应的数据
	*/
	SetResponseBody(data []byte) bool

	/*
		SetResponseBodyIO
		设置响应的数据
	*/
	SetResponseBodyIO(data io.ReadCloser) bool
	/*
		GetResponseCode
		返回响应的状态码
	*/
	GetResponseCode() int
	/*
		SetResponseCode
		修改响应的状态码
	*/
	SetResponseCode(code int) bool
	/*
		SetRequestBody
		设置POST请求提交的数据
	*/
	SetRequestBody(data []byte) bool

	/*
		SetRequestBodyIO
		设置POST请求提交的数据
	*/
	SetRequestBodyIO(data io.ReadCloser) bool

	/*
		SaveRawRequestData
		保存原始数据到本地文件
	*/
	SaveRawRequestData(SaveFilePath string) bool

	/*
		IsRawRequestBody
		判断当前请求是否原始数据(是否超过了超过一定大小)
		(这个大小由自己设置默认:10240000 字节)
	*/
	IsRawRequestBody() bool
}

type ConnWebSocketScriptCall interface {
	general

	/*
		Body

		获取消息内容
	*/
	Body() []byte

	/*
		Body

		获取消息内容长度
	*/
	BodyLen() int

	/*
		Type
		返回当前消息事件类型

		请使用public.Websocket...

		1=连接成功 2=客户端发送数据 3=客户端收到数据 4=连接关闭
	*/
	Type() int
	/*
		ClientIP
		返回请求的客户端IP地址
	*/
	ClientIP() string
	/*
		GetMessageType

		获取消息事件类型

		Text=1 Binary=2 Close=8 Ping=9 Pong=10 Invalid=-1/255
	*/
	MessageType() int

	/*
		SetBody

		修改当前消息内容
	*/
	SetBody(data []byte) bool

	/*
		SendToServer

		发送消息到服务器

		MessageType：1文本，2二进制，8关闭，9心跳
	*/
	SendToServer(MessageType int, data []byte) bool

	/*
		SendToClient

		发送消息到客户端

		MessageType：1文本，2二进制，8关闭，9心跳
	*/
	SendToClient(MessageType int, data []byte) bool

	/*
		Close

		关闭、断开当前Websocket会话
	*/
	Close() bool

	/*
		URL
		当前请求的URL
	*/
	URL() string
}
type ConnTCPScriptCall interface {
	general
	proxy
	address
	/*
		Body

		获取消息内容

		如果事件类型为连接成功时,这里返回SunnyNet发送出去的本地地址
	*/
	Body() []byte

	/*
		Body

		获取消息内容长度
	*/
	BodyLen() int

	/*
		Type

		返回当前消息事件类型

		请使用public.SunnyNetMsgTypeTCP...

		0=连接成功 1=客户端发送数据 2=客户端收到数据 3=连接关闭或连接失败 4=即将开始连接
	*/
	Type() int

	/*
		SetBody

		修改消息内容
	*/
	SetBody(data []byte) bool

	/*
		Close

		关闭、断开当前TCP会话
	*/
	Close() bool

	/*
		SetNewAddress
		设置目标连接地址 目标地址必须带端口号 例如 baidu.com:443 [仅限即将连接时使用]
	*/
	SetNewAddress(ip string) bool
}
type ConnUDPScriptCall interface {
	general
	address

	/*
		Body
		获取消息内容
	*/
	Body() []byte

	/*
		Body
		获取消息比如长度
	*/
	BodyLen() int

	/*
		Type

		返回当前消息事件类型

		请使用 public.SunnyNetUDPType...

		1=关闭 2=发送数据 3=收到数据

	*/
	Type() int

	/*
		SetBody

		修改消息内容
	*/
	SetBody(data []byte) bool
}
type proxy interface {
	/*
		SetAgent
		设置代理

		ProxyUrl:
		格式1: http://127.0.0.1:8080
		格式2: socks5://127.0.0.1:8080
		格式3: socks5://user:pass@127.0.0.1:8080
		格式4: http://user:pass@127.0.0.1:8080
		timeout: 代理超时时间(毫秒)

		返回:  是否设置成功
	*/
	SetAgent(ProxyUrl string, timeout ...int) bool
}
type address interface {
	/*
		LocalAddress
		获取客户端地址
	*/
	LocalAddress() string
	/*
		RemoteAddress
		获取远程地址
	*/
	RemoteAddress() string
	/*
		SendToServer

		主动发送消息到服务器
	*/
	SendToServer(data []byte) bool

	/*
		SendToClient

		主动发送消息到客户端
	*/
	SendToClient(data []byte) bool
}
type general interface {

	/*
		Context

		返回SunnyNet 上下文Context
	*/
	Context() int

	/*
		MessageId

		当前消息的ID
	*/
	MessageId() int
	/*
		PID

		返回当前会话由发起的进程的PID

		如果为0表示非本机的设备通过代理发起
	*/
	PID() int

	/*
		Theology

		获取当前请求唯一ID
	*/
	Theology() int

	/*
		GetProcessName

		返回当前会话由发起的进程名称

		如果非本机的设备通过代理发起,返回"代理连接"
	*/
	GetProcessName() string
}