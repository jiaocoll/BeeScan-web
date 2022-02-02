package util

import (
	"sort"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/3
程序功能：
*/

// TrimProtocol removes the HTTP scheme from an URI
func TrimProtocol(targetURL string) string {
	URL := strings.TrimSpace(targetURL)
	if strings.HasPrefix(strings.ToLower(URL), "http://") || strings.HasPrefix(strings.ToLower(URL), "https://") {
		URL = URL[strings.Index(URL, "//")+2:]
	}
	URL = strings.TrimRight(URL, "/")
	return URL
}

// Removesamesip 去重函数
func Removesamesip(ips []string) (result []string) {
	result = make([]string, 0)
	tempMap := make(map[string]bool, len(ips))
	for _, e := range ips {
		if tempMap[e] == false {
			tempMap[e] = true
			result = append(result, e)
		}
	}
	return result
}

func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}
