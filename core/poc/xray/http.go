package xray

import (
	"Beescan/core/httpx"
	"fmt"
	"github.com/projectdiscovery/retryablehttp-go"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func ParseUrl(u *url.URL) *UrlType {
	nu := &UrlType{}
	nu.Scheme = u.Scheme
	nu.Domain = u.Hostname()
	nu.Host = u.Host
	nu.Port = u.Port()
	nu.Path = u.EscapedPath()
	nu.Query = u.RawQuery
	nu.Fragment = u.Fragment
	return nu
}

func ParseRequest(oReq *retryablehttp.Request) (*Request, error) {
	req := &Request{}
	req.Method = oReq.Method
	req.Url = ParseUrl(oReq.URL)
	header := make(map[string]string)
	for k := range oReq.Header {
		header[k] = oReq.Header.Get(k)
	}
	req.Headers = header
	req.ContentType = oReq.Header.Get("Content-Type")
	return req, nil
}

func ParseResponse(oResp *httpx.Response) *Response {
	var resp Response
	header := make(map[string]string)
	resp.Status = int32(oResp.StatusCode)
	//resp.Url =
	for k, v := range oResp.Headers {
		header[k] = strings.Join(v, "")
	}
	resp.Headers = header
	v, ok := header["Content-Type"]
	resp.ContentType = ""
	if ok {
		resp.ContentType = v
	}
	resp.Body = oResp.Data
	return &resp
}

func UrlTypeToString(u *UrlType) string {
	var buf strings.Builder
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}
	if u.Scheme != "" || u.Host != "" {
		if u.Host != "" || u.Path != "" {
			buf.WriteString("//")
		}
		if h := u.Host; h != "" {
			buf.WriteString(u.Host)
		}
	}
	path := u.Path
	if path != "" && path[0] != '/' && u.Host != "" {
		buf.WriteByte('/')
	}
	if buf.Len() == 0 {
		if i := strings.IndexByte(path, ':'); i > -1 && strings.IndexByte(path[:i], '/') == -1 {
			buf.WriteString("./")
		}
	}
	buf.WriteString(path)

	if u.Query != "" {
		buf.WriteByte('?')
		buf.WriteString(u.Query)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.Fragment)
	}
	return buf.String()
}
func replaceValue(v string, mm map[string]interface{}) string {
	regx, _ := regexp.Compile("{{([\\w|_]+)}}")
	result := regx.FindAllStringSubmatch(v, -1)
	if len(result) == 0 {
		return v
	} else {
		for _, sub := range result {
			if len(sub) == 2 {
				key := sub[1]
				value, ok := mm[key]

				if ok {
					var temp string
					switch value.(type) {
					case int:
						temp = strconv.Itoa(value.(int))
					case string:
						temp = value.(string)
					default:
						temp = fmt.Sprintf("%s", temp)
					}
					v = strings.ReplaceAll(v, sub[0], temp)
				}
			}
		}
		return v
	}
}
func doSearch(re string, body string) map[string]string {
	r, err := regexp.Compile(re)
	if err != nil {
		return nil
	}
	result := r.FindStringSubmatch(body)
	names := r.SubexpNames()
	if len(result) > 1 && len(names) > 1 {
		paramsMap := make(map[string]string)
		for i, name := range names {
			if i > 0 && i <= len(result) {
				paramsMap[name] = result[i]
			}
		}
		return paramsMap
	}
	return nil
}
