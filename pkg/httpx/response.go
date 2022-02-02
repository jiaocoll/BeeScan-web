package httpx

import (
	"github.com/projectdiscovery/rawhttp/client"
	"strings"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/2
程序功能：
*/

// Response contains the response to a server
type Response struct {
	StatusCode    int
	Headers       map[string][]string
	HeaderStr     string
	Data          []byte
	DataStr       string
	ContentLength int
	Raw           string
	TLSData       *TLSData
	Duration      time.Duration
	Title         string
	FirstLine     string
}

// GetHeader value
func (r *Response) GetHeader(name string) string {
	v, ok := r.Headers[name]
	if ok {
		return strings.Join(v, " ")
	}
	return ""
}

// GetHeaderPart with offset
func (r *Response) GetHeaderPart(name, sep string) string {
	v, ok := r.Headers[name]
	if ok && len(v) > 0 {
		tokens := strings.Split(strings.Join(v, " "), sep)
		return tokens[0]
	}
	return ""
}

func (r *Response) DumpResponse() string {
	return r.FirstLine + client.NewLine + r.HeaderStr + client.NewLine + r.DataStr
}
