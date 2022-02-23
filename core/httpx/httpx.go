package httpx

import (
	"crypto/tls"
	"fmt"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	"github.com/projectdiscovery/rawhttp"
	"github.com/projectdiscovery/retryablehttp-go"
	"golang.org/x/net/http2"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

/*
创建人员：云深不知处
创建时间：2022/1/2
程序功能：
*/


// HTTPX represent an instance of the library client
type HTTPX struct {
	client        *retryablehttp.Client
	client2       *http.Client
	CustomHeaders map[string]string
	Dialer        *fastdialer.Dialer
	Options       *HTTPOptions
}


// New httpx instance
func NewHttpx(options *HTTPOptions) (*HTTPX, error) {
	httpx := &HTTPX{}
	dialer, err := fastdialer.NewDialer(fastdialer.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create resolver cache: %s", err)
	}
	httpx.Dialer = dialer
	httpx.Options = options

	var retryablehttpOptions = retryablehttp.DefaultOptionsSpraying
	retryablehttpOptions.Timeout = httpx.Options.Timeout
	retryablehttpOptions.RetryMax = httpx.Options.RetryMax

	var redirectFunc = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse // Tell the http client to not follow redirect
	}

	if httpx.Options.FollowRedirects {
		// Follow redirects
		redirectFunc = nil
	}

	transport := &http.Transport{
		DialContext:         httpx.Dialer.Dial,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}

	if httpx.Options.HTTPProxy != "" {
		proxyURL, parseErr := url.Parse(httpx.Options.HTTPProxy)
		if parseErr != nil {
			return nil, parseErr
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	httpx.client = retryablehttp.NewWithHTTPClient(&http.Client{
		Transport:     transport,
		Timeout:       httpx.Options.Timeout,
		CheckRedirect: redirectFunc,
	}, retryablehttpOptions)

	httpx.client2 = &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			AllowHTTP: true,
		},
		Timeout: httpx.Options.Timeout,
	}

	//httpx2.CustomHeaders = httpx2.Options.CustomHeaders
	return httpx, nil
}


// Do http request
func (h *HTTPX) Do(req *retryablehttp.Request) (*Response, error) {
	timeStart := time.Now()


	httpresp, err := h.getResponse(req)
	if err != nil {
		return nil, err
	}
	var resp Response
	resp.Headers = httpresp.Header.Clone()
	resp.HeaderStr = ""
	for h, v := range resp.Headers {
		resp.HeaderStr += fmt.Sprintf("%s: %S\n", h, strings.Join(v,""))
	}
	// httputil.DumpResponse does not handle websockets
	rawresp, err := DumpResponse(httpresp)
	if err != nil {
		return nil, err
	}

	resp.Raw = rawresp

	var respbody []byte
	// websockets don't have a readable body
	if httpresp.StatusCode != http.StatusSwitchingProtocols {
		var err error
		respbody, err = ioutil.ReadAll(httpresp.Body)
		if err != nil {
			return nil, err
		}
	}

	closeErr := httpresp.Body.Close()
	if closeErr != nil {
		return nil, closeErr
	}

	respbodystr := string(respbody)
	// Non UTF-8
	isgbk := false
	if contentTypes, ok := resp.Headers["Content-Type"]; ok {
		contentType := strings.Join(contentTypes, ";")

		// special cases
		if strings.Contains(contentType, "charset=GB2312") {

			bodyUtf8, err := Decodegbk([]byte(respbodystr))
			if err == nil {
				isgbk = true
				respbodystr = string(bodyUtf8)
			}
		}
	}
	if !isgbk {
		// special cases
		regx := regexp.MustCompile("(?i)<meta.*charset=['\"]?(gb2312|gbk)")
		if regx.MatchString(respbodystr) {
			titleUtf8, err := Decodegbk([]byte(respbodystr))
			if err == nil {
				respbodystr = string(titleUtf8)
			}
		}
	}

	resp.Data = respbody

	// fill metrics
	resp.StatusCode = httpresp.StatusCode
	resp.ContentLength = utf8.RuneCountInString(respbodystr)
	resp.DataStr = respbodystr
	resp.Title = ExtractTitle(respbodystr)


	if !h.Options.Unsafe  {
		// extracts TLS data if any
		resp.TLSData = h.TLSGrab(httpresp)
	}


	resp.Duration = time.Since(timeStart)

	return &resp, nil
}

// getResponse returns response from safe / unsafe request
func (h *HTTPX) getResponse(req *retryablehttp.Request) (*http.Response, error) {
	if h.Options.Unsafe {
		return h.doUnsafe(req)
	}

	return h.client.Do(req)
}

// doUnsafe does an unsafe http request
func (h *HTTPX) doUnsafe(req *retryablehttp.Request) (*http.Response, error) {
	method := req.Method
	headers := req.Header
	targetURL := req.URL.String()
	body := req.Body
	return rawhttp.DoRaw(method, targetURL, "", headers, body)
}

// NewRequest from url
func (h *HTTPX) NewRequest(method, targetURL string) (req *retryablehttp.Request, err error) {
	req, err = retryablehttp.NewRequest(method, targetURL, nil)
	if err != nil {
		return
	}

	// Skip if unsafe is used
	if !h.Options.Unsafe {
		// set default user agent
		req.Header.Set("User-Agent", h.Options.DefaultUserAgent)
		// set default encoding to accept utf8
		req.Header.Add("Accept-Charset", "utf-8")
	}
	return
}

// SetCustomHeaders on the provided request
func (h *HTTPX) SetCustomHeaders(r *retryablehttp.Request, headers map[string]string) {
	for name, value := range headers {
		r.Header.Set(name, value)
		// host header is particular
		if strings.EqualFold(name, "host") {
			r.Host = value
		}
	}
}