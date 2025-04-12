package SunnyNet

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/HttpCertificate"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/public"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type _whois map[string]*_cert

var whoisLock sync.Mutex
var whois = make(_whois)

type _certType byte
type _cert struct {
	Cert     *tls.Certificate
	Type     _certType
	Expire   *time.Time
	DNSNames []string
}

const (
	netCert = _certType(iota + 1)
	localCert
)

var httpTypeMap = make(map[string]*httpTypeInfo)

const whoisUndefined = 0
const whoisNoHTTPS = 1
const whoisHTTPS1 = 2
const whoisHTTPS2 = 3

type httpTypeInfo struct {
	_type byte
	_time time.Time
	_lock sync.Mutex
}

func clean() {
	for {
		time.Sleep(time.Minute)
		whoisLock.Lock()
		for key, v := range httpTypeMap {
			if time.Now().Sub(v._time) > time.Minute*10 {
				delete(httpTypeMap, key)
			}
		}
		whoisLock.Unlock()
	}
}
func init() {
	go clean()
}
func ClientIsHttps(server string) byte {
	whoisLock.Lock()
	defer whoisLock.Unlock()
	res := httpTypeMap[server]
	if res == nil {
		return whoisUndefined
	}
	res._time = time.Now()
	return res._type
}
func ClientRequestIsHttps(Sunny *Sunny, server string, bytesData []byte) (res byte) {
	var obj *httpTypeInfo
	whoisLock.Lock()
	if httpTypeMap[server] == nil {
		obj = &httpTypeInfo{_time: time.Now()}
	} else {
		obj = httpTypeMap[server]
	}
	whoisLock.Unlock()
	obj._lock.Lock()
	defer func() {
		whoisLock.Lock()
		obj._type = res
		obj._time = time.Now()
		whoisLock.Unlock()
		obj._lock.Unlock()
	}()
	proxyHost, proxyPort, e := net.SplitHostPort(server)
	var ips []net.IP
	var first net.IP
	if e != nil {
		return whoisUndefined
	}
	var conn net.Conn
	ip := net.ParseIP(proxyHost)
	if ip == nil {
		first = dns.GetFirstIP(proxyHost, "")
		if first != nil {
			conn, _ = Sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(first, proxyPort), time.Second*3, Sunny.outRouterIP)
		}
		if conn == nil {
			ips, _ = dns.LookupIP(proxyHost, "", Sunny.outRouterIP, nil)
			//优先尝试IPV4
			for _, _ip := range ips {
				if _ip2 := _ip.To4(); _ip2 != nil {
					conn, _ = Sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(_ip, proxyPort), 2*time.Second, Sunny.outRouterIP)
					if conn != nil {
						dns.SetFirstIP(proxyHost, "", _ip)
						break
					}
				}
			}
			//最后尝试IPV6
			if conn == nil {
				for _, _ip := range ips {
					if _ip2 := _ip.To16(); _ip2 != nil {
						conn, _ = Sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(_ip, proxyPort), 2*time.Second, Sunny.outRouterIP)
						if conn != nil {
							dns.SetFirstIP(proxyHost, "", _ip)
							break
						}
					}
				}
			}
		}
	} else {
		conn, _ = Sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(ip, proxyPort), time.Second*3, Sunny.outRouterIP)
	}
	if conn == nil {
		return whoisUndefined
	}
	defer func() {
		_ = conn.Close()
	}()
	_ = conn.SetDeadline(time.Now().Add(time.Second * 3))
	_, _ = conn.Write(bytesData)
	hdr := make([]byte, 5)
	n, _ := conn.Read(hdr)
	if hdr[0] == 0 {
		return whoisUndefined
	}
	if n < 5 {
		return whoisNoHTTPS
	}
	var bs bytes.Buffer
	if hdr[0] == 21 || hdr[0] == 22 || hdr[0] == 23 {
		n = int(hdr[3])<<8 | int(hdr[4])
		if n < 40960 {
			data := make([]byte, 128)
			for {
				if bs.Len() >= n {
					break
				}
				nx, err := conn.Read(data)
				bs.Write(data[:nx])
				if err != nil {
					break
				}
			}
			if bs.Len() >= n {
				data = bs.Bytes()[:n]
				mmm := tls.UnMarshal(data)
				bs.Reset()
				if mmm != nil {
					if mmm.SupportedVersion == 772 {
						return whoisHTTPS2
					}
				}
			}
			return whoisHTTPS1
		}
	}

	return whoisNoHTTPS
}

func WhoisCache(Sunny *Sunny, serverName, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, error) {
	{
		if in := getTlsConfig(host); in != nil {
			return in, nil, nil
		}
		if in := getTlsConfig(serverName); in != nil {
			return in, nil, nil
		}
	}
	{
		c, d := getLocalCert(serverName)
		if c != nil {
			return c, d, nil
		}
		c, d = getLocalCert(host)
		if c != nil {
			return c, d, nil
		}
	}
	{
		c, d := createLocalCert(Sunny, serverName, host, parent, priv)
		if c != nil {
			return c, d, nil
		}
	}
	return nil, nil, _GetIpCertError
}
func WhoisLoopCache(Sunny *Sunny, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, error) {
	c, d := getLocalCert(host)
	if c != nil {
		return c, d, nil
	}
	c, d = createLocalCert(Sunny, host, host, parent, priv)
	if c != nil {
		return c, d, nil
	}
	return nil, nil, _GetIpCertError
}
func createLocalCert(Sunny *Sunny, serverName, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string) {
	var mHost string
	var keyName string
	var err error
	if serverName == "" || serverName == "null" {
		//是否为DNS解析服务器,如果是直接本地生成证书即可,就不需要从网络获取证书了
		if !strings.HasSuffix(host, ":853") {
			a, b, _ := createNetCert(Sunny, host, parent, priv)
			if a != nil {
				return a, b
			}
			keyName = host
			mHost, _, err = public.SplitHostPort(host)
		} else {
			keyName = host
			mHost, _, err = public.SplitHostPort(host)
		}
	} else {
		keyName = serverName
		mHost, _, err = public.SplitHostPort(serverName)
	}
	if err != nil {
		return nil, nil
	}
	certByte, priByte, not, er := generatePemTemp(mHost, parent, priv)
	if er != nil {
		return nil, nil
	}
	certificate, er := tls.X509KeyPair(certByte, priByte)
	if er != nil {
		return nil, nil
	}
	whoisLock.Lock()
	whois[keyName] = &_cert{Cert: &certificate, Type: localCert, Expire: not}
	whoisLock.Unlock()
	return &certificate, nil
}
func createNetCert(Sunny *Sunny, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, error) {
	mHost, _, err := public.SplitHostPort(host)
	if err != nil {
		return nil, nil, err
	}
	if ip := net.ParseIP(mHost); ip != nil {
		var rr *x509.Certificate
		for i := 0; i < 5; i++ {
			rr, err = GetIpAddressHost(Sunny.proxy, host, Sunny.outRouterIP)
			if rr != nil {
				break
			}
		}
		if rr == nil {
			return nil, nil, _GetIpCertError
		}
		certByte, priByte, not, er := generatePem(rr, mHost, parent, priv)
		if er != nil {
			return nil, nil, er
		}
		certificate, er := tls.X509KeyPair(certByte, priByte)
		if er != nil {
			return nil, nil, er
		}
		DNSNames := rr.DNSNames
		for _, v := range rr.IPAddresses {
			DNSNames = append(DNSNames, v.String())
		}
		whoisLock.Lock()
		whois[host] = &_cert{Cert: &certificate, Type: netCert, Expire: not, DNSNames: DNSNames}
		whoisLock.Unlock()
		return &certificate, DNSNames, nil
	}
	return nil, nil, _ParseIPError
}

var _ParseIPError = errors.New("Not an IP address ")

func getLocalCert(host string) (*tls.Certificate, []string) {
	if host == "" || host == "null" {
		return nil, nil
	}
	whoisLock.Lock()
	defer whoisLock.Unlock()
	//查询证书到期时间
	val := whois[host]
	if val != nil {
		//查询到了
		//和现在的的时间对比，如果证书即将到期则丢弃,重新获取证书
		tenMinutesAgo := time.Now().Add(-10 * time.Minute)
		if val.Expire.After(tenMinutesAgo) {
			//如果证书没有即将到期 则获取该证书,如果没有获取到则重新获取证书
			if val.Cert != nil {
				//如果是临时证书，则向后台添加一个请求,当前先使用缓存中的临时证书
				//如果不是临时证书，则直接返回该证书
				if val.Type == localCert {
					//TempNetAdd(Sunny, host)
				}
				return val.Cert, val.DNSNames
			}
		}
		delete(whois, host)
	}
	return nil, nil
}
func getTlsConfig(host string) *tls.Certificate {
	in := HttpCertificate.GetTlsConfig(host, public.CertificateRequestManagerRulesReceive)
	if in != nil {
		if len(in.Certificates) > 0 {
			return &in.Certificates[0]
		}
	}
	return nil
}
func GetIpAddressHost(proxy *SunnyProxy.Proxy, ipAddress string, outRouterIP *net.TCPAddr) (*x509.Certificate, error) {
	config := &tls.Config{InsecureSkipVerify: true}
	var x *x509.Certificate
	config.VerifyServerCertificate = func(certificate *x509.Certificate) error {
		x = certificate
		return io.EOF
	}
	var conn net.Conn
	var err error
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	conn, err = proxy.Dial("tcp", ipAddress, outRouterIP)
	if err != nil {
		return nil, err
	}
	t := tls.Client(conn, config)
	err = t.Handshake()
	if x != nil {
		err = nil
	}
	return x, err
}

var _GetIpCertError = fmt.Errorf("no success Get Certificate")

func generatePem(template *x509.Certificate, mHost string, parent *x509.Certificate, priv *rsa.PrivateKey) ([]byte, []byte, *time.Time, error) {
	template1 := x509.Certificate{
		SerialNumber:                template.SerialNumber,                // 序列号，CA 颁发的唯一序列号，通常为随机生成
		Subject:                     template.Subject,                     // 证书主题，包含持有者的信息（国家、组织等）
		NotBefore:                   template.NotBefore,                   // 证书开始生效时间
		NotAfter:                    template.NotAfter,                    // 证书到期时间
		KeyUsage:                    template.KeyUsage,                    // 密钥用法，指明证书可用于的操作（如签名、加密等）
		ExtKeyUsage:                 template.ExtKeyUsage,                 // 扩展密钥用法，指明证书的额外用途（如客户端认证、服务器认证等）
		EmailAddresses:              template.EmailAddresses,              // 证书持有者的电子邮件地址
		IPAddresses:                 template.IPAddresses,                 // 证书包含的 IP 地址列表
		DNSNames:                    template.DNSNames,                    // 证书关联的 DNS 域名列表
		Issuer:                      template.Issuer,                      // 证书颁发者的信息
		IssuingCertificateURL:       template.IssuingCertificateURL,       // 颁发者证书的 URL
		BasicConstraintsValid:       template.BasicConstraintsValid,       // 基础约束是否有效
		IsCA:                        template.IsCA,                        // 标识证书是否为证书颁发机构（CA）证书
		AuthorityKeyId:              template.AuthorityKeyId,              // CA 密钥标识符
		UnknownExtKeyUsage:          template.UnknownExtKeyUsage,          // 未知的扩展密钥用途列表
		ExtraExtensions:             template.ExtraExtensions,             // 额外的 X.509 扩展字段
		PermittedDNSDomainsCritical: template.PermittedDNSDomainsCritical, // 是否为关键的允许 DNS 域名
		PermittedDNSDomains:         template.PermittedDNSDomains,         // 允许的 DNS 域名列表
		PolicyIdentifiers:           template.PolicyIdentifiers,           // 策略标识符的列表
		MaxPathLen:                  template.MaxPathLen,                  // 最大路径长度，限制证书链的深度
		MaxPathLenZero:              template.MaxPathLenZero,              // 最大路径长度是否可以为零
		SubjectKeyId:                template.SubjectKeyId,                // 证书持有者密钥的标识符
	}
	if ip := net.ParseIP(mHost); ip != nil {
		template1.IPAddresses = append(template1.IPAddresses, ip)
	} else {
		template1.DNSNames = append(template1.DNSNames, mHost)
	}
	cer, err := x509.CreateCertificate(rand.Reader, &template1, parent, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{ // 证书
			Type:  "CERTIFICATE",
			Bytes: cer,
		}), pem.EncodeToMemory(&pem.Block{ // 私钥
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		}), &template1.NotAfter, err
}
func generatePemTemp(mHost string, parent *x509.Certificate, priv *rsa.PrivateKey) ([]byte, []byte, *time.Time, error) {
	serialNumber, _ := rand.Int(rand.Reader, public.MaxBig)
	not := time.Now().AddDate(0, 0, 365)
	template := x509.Certificate{
		SerialNumber: serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject: pkix.Name{ //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
			CommonName: mHost,
		},
		NotBefore:      time.Now().AddDate(0, 0, -1),
		NotAfter:       not,
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, //KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},               // 密钥扩展用途的序列
		EmailAddresses: []string{"forward.nice.cp@gmail.com"},
	}
	{
		if ip := net.ParseIP(mHost); ip != nil {
			template.IPAddresses = []net.IP{ip}
			template.DNSNames = []string{ip.String()}
		} else if _, _, ip = parseIPv6Address(mHost); ip != nil {
			template.IPAddresses = []net.IP{ip}
			template.DNSNames = []string{ip.String()}
		} else {
			template.DNSNames = []string{mHost}
		}
	}
	cer, err := x509.CreateCertificate(rand.Reader, &template, parent, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, &not, err
	}
	return pem.EncodeToMemory(&pem.Block{ // 证书
			Type:  "CERTIFICATE",
			Bytes: cer,
		}), pem.EncodeToMemory(&pem.Block{ // 私钥
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		}), &not, err
}

var tempNet map[*Sunny]map[string]byte
var tempNetLock sync.Mutex

func TempNetAdd(Sunny *Sunny, host string) {
	tempNetLock.Lock()
	if tempNet[Sunny] == nil {
		tempNet[Sunny] = make(map[string]byte)
	}
	if tempNet[Sunny][host] < 1 {
		tempNet[Sunny][host] = 1
	}
	tempNetLock.Unlock()
}

func init() {
	tempNet = make(map[*Sunny]map[string]byte)
	host := ""
	var obj *Sunny
	var rootCa *x509.Certificate //中间件CA证书
	var rootKey *rsa.PrivateKey  // 证书私钥
	var certificate tls.Certificate
	var certByte []byte
	var priByte []byte
	var not *time.Time
	var err error
	var rr *x509.Certificate
	go func() {
		for {
			tempNetLock.Lock()
			host = ""
			for k, v := range tempNet {
				for kk, vv := range v {
					if vv == 1 {
						host = kk
						obj = k
						break
					}
				}
			}
			if host == "" {
				tempNetLock.Unlock()
				time.Sleep(time.Second)
				continue
			}
			rootCa = obj.rootCa
			rootKey = obj.rootKey
			tempNetLock.Unlock()
			rr, err = GetIpAddressHost(obj.proxy, host, obj.outRouterIP)
			if rr == nil {
				goto gg
			}
			certByte, priByte, not, err = generatePem(rr, host, rootCa, rootKey)
			if err != nil {
				goto gg
			}
			certificate, err = tls.X509KeyPair(certByte, priByte)
			if err != nil {
				goto gg
			}
			{
				DNSNames := rr.DNSNames
				for _, v := range rr.IPAddresses {
					DNSNames = append(DNSNames, v.String())
				}
				whoisLock.Lock()
				whois[host] = &_cert{Cert: &certificate, Type: netCert, Expire: not, DNSNames: DNSNames}
				whoisLock.Unlock()
			}
		gg:
			tempNetLock.Lock()
			delete(tempNet[obj], host)
			tempNetLock.Unlock()

		}
	}()
}
