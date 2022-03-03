package runner

import "encoding/json"

/*
创建人员：云深不知处
创建时间：2022/3/2
程序功能：
*/

type Result interface {
	JSON() string
}

type pocResult struct {
	URL            string   `json:"url"`
	PocName        string   `json:"poc_name"`
	PocLink        []string `json:"poc_link"`
	PocAuthor      string   `json:"poc_author"`
	PocDescription string   `json:"poc_description"`
}

func (r *pocResult) JSON() string {
	if js, err := json.Marshal(r); err == nil {
		return string(js)
	}
	return ""
}
