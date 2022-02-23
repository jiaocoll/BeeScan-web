package runner

import (
	"Beescan/pkg/httpx"
	"encoding/json"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：扫描结果
*/

type Result interface {
	STR() string
	JSON() string
}

type fingerResult struct {
	URL           string         `json:"url"`
	IP            string         `json:"ip"`
	Title         string         `json:"title"`
	TLSData       *httpx.TLSData `json:"tls,omitempty"`
	ContentLength int            `json:"content-length"`
	StatusCode    int            `json:"status-code"`
	ResponseTime  string         `json:"response-time"`
	CDN           string         `json:"cdn"`
	Fingers       []string
	str           string
}

func (r *fingerResult) JSON() string {
	if js, err := json.Marshal(r); err == nil {
		return string(js)
	}

	return ""
}
func (r *fingerResult) STR() string {
	return r.str
}
