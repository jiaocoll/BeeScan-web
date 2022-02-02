package httpx

import (
	"html"
	"regexp"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/2
程序功能：
*/

// ExtractTitle from a response
func ExtractTitle(resp string) (title string) {
	var re = regexp.MustCompile(`(?im)<\s*title.*>(.*?)<\s*/\s*title>`)
	for _, match := range re.FindAllString(resp, -1) {
		title = html.UnescapeString(trimTitleTags(match))
		break
	}
	return
}

func trimTitleTags(title string) string {
	// trim <title>*</title>
	titleBegin := strings.Index(title, ">")
	titleEnd := strings.Index(title, "</")
	return title[titleBegin+1 : titleEnd]
}