package httpx

import "time"

/*
创建人员：云深不知处
创建时间：2022/1/2
程序功能：
*/

type HTTPOptions struct {
	Timeout          time.Duration
	RetryMax         int
	FollowRedirects  bool
	HTTPProxy        string
	Unsafe           bool
	DefaultUserAgent string
}