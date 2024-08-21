package g2util

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/proxy"
)

// RandUseragent ...
func RandUseragent() string {
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.20 (KHTML, like Gecko) Chrome/19.0.1036.7 Safari/535.20",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71 Safari/537.1 LBBROWSER",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:2.0b13pre) Gecko/20110307 Firefox/4.0b13pre",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
		//"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.0.12) Gecko/20070731 Ubuntu/dapper-security Firefox/1.5.0.12",
		//"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Acoo Browser; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)",
		//"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		//"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0), Lynx/2.8.5rel.1 libwww-FM/2.14 SSL-MM/1.4.1 GNUTLS/1.2.9",
		//"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		//"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; QQBrowser/7.0.3698.400)",
		//"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; QQDownload 732; .NET4.0C; .NET4.0E)",
		//"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52",
		//"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.0.12) Gecko/20070731 Ubuntu/dapper-security Firefox/1.5.0.12",
		//"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; LBBROWSER)",
		//"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		//"Mozilla/5.0 (X11; U; Linux; en-US) AppleWebKit/527+ (KHTML, like Gecko, Safari/419.3) Arora/0.6",
		//"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; QQBrowser/7.0.3698.400)",
		//"Opera/9.25 (Windows NT 5.1; U; en), Lynx/2.8.5rel.1 libwww-FM/2.14 SSL-MM/1.4.1 GNUTLS/1.2.9",
	}

	return userAgents[MathRandInt(len(userAgents))]
}

// RestyAgent ...
type RestyAgent struct {
	Logger LevelLogger `inject:""`

	client *resty.Client
}

// New ...
func (a *RestyAgent) New() *RestyAgent { r := &RestyAgent{Logger: a.Logger}; return r.constructor() }

// Constructor ...
func (a *RestyAgent) Constructor() { a.constructor() }

// Client ...
func (a *RestyAgent) Client() *resty.Client { return a.client }

// SetDialTimeout ...
// 非线程安全
func (a *RestyAgent) SetDialTimeout(d time.Duration) *RestyAgent {
	if d < time.Second*1 {
		d = time.Second * 1
	}
	dialer := a.createDialer()
	dialer.Timeout = d
	transport := a.transport()
	transport.DialContext = dialer.DialContext
	a.client.SetTransport(transport)
	return a
}

// SetProxy ...
// 非线程安全
func (a *RestyAgent) SetProxy(proxyURL string, user *url.Userinfo) *RestyAgent {
	pURL, err := url.Parse(proxyURL)
	if err != nil {
		return a
	}
	pURL.User = user
	dialer := a.createDialer()
	proxyDialer, err := proxy.FromURL(pURL, dialer)
	if err != nil {
		return a
	}
	proxyContextDialer, ok := proxyDialer.(proxy.ContextDialer)
	if !ok {
		return a
	}
	transport := a.transport()
	transport.DialContext = proxyContextDialer.DialContext
	a.client.SetTransport(transport)
	return a
}

func (a *RestyAgent) transport() *http.Transport {
	/*if transport, ok := a.client.GetClient().Transport.(*http.Transport); ok {
		return transport, nil
	}
	return nil, errors.New("current transport is not an *http.Transport instance")*/
	return a.createTransport(nil).(*http.Transport)
}

func (a *RestyAgent) constructor() *RestyAgent {
	client := resty.New()
	client.SetLogger(a.Logger)
	client.SetAllowGetMethodPayload(true)

	//client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))

	client.SetHeaders(map[string]string{
		"Accept":        "*/*",
		"Pragma":        "no-cache",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
		//"User-Agent":    RandUseragent(),
	})
	client.OnBeforeRequest(func(client *resty.Client, request *resty.Request) error {
		request.SetHeader("User-Agent", RandUseragent())
		return nil
	})
	//client.SetContentLength(true)

	//retry
	client.SetRetryCount(2)
	client.SetRetryWaitTime(time.Millisecond * time.Duration(500))
	client.SetRetryMaxWaitTime(time.Millisecond * time.Duration(2000))

	//transport
	client.SetTransport(a.createTransport(nil))
	client.SetTimeout(time.Second * time.Duration(30))

	a.client = client
	return a
}

func (a *RestyAgent) createTransport(localAddr net.Addr) http.RoundTripper {
	dialer := a.createDialer()
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	tr := &http.Transport{
		Proxy:       http.ProxyFromEnvironment,
		DialContext: dialer.DialContext,
		//DialTLSContext:      nil,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: time.Second * time.Duration(10),
		//DisableKeepAlives:   false,
		//DisableCompression:  false,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(runtime.NumCPU()) + 1,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       time.Second * time.Duration(60*5),
		ResponseHeaderTimeout: time.Second * time.Duration(30),
		ExpectContinueTimeout: time.Second * time.Duration(20),
		ForceAttemptHTTP2:     true,
	}
	return tr
}

func (*RestyAgent) createDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:       time.Second * time.Duration(20),
		FallbackDelay: time.Millisecond * time.Duration(300),
		KeepAlive:     time.Second * time.Duration(15),
	}
}
