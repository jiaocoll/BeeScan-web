package httpx

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

/*
创建人员：云深不知处
创建时间：2022/1/2
程序功能：
*/

const (
	// HTTP defines the plain http scheme
	HTTP = "http"
	// HTTPS defines the secure http scheme
	HTTPS = "https"
	// HTTPorHTTPS defines the both http and https scheme
	HTTPorHTTPS = "http|https"
)
// DumpResponse to string
func DumpResponse(resp *http.Response) (string, error) {
	// httputil.DumpResponse does not work with websockets
	if resp.StatusCode == http.StatusContinue {
		raw := resp.Status + "\n"
		for h, v := range resp.Header {
			raw += fmt.Sprintf("%s: %s\n", h, v)
		}
		return raw, nil
	}

	raw, err := httputil.DumpResponse(resp, true)
	return string(raw), err
}