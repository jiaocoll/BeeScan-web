package options

import "flag"

/*
创建人员：云深不知处
创建时间：2022/1/3
程序功能：
*/


type Options struct {
	Target              multiStringFlag
	Targets             string
	Output              string
	ProxyURL            string
	TimeOut             int
	JSON                bool
	Verbose             bool
	OutputStatusCode    bool
	OutputWithNoColor   bool
	OutputContentLength bool
	OutputTitle         bool
	OutputIP            bool
	OutputFingerPrint   bool
	RateLimit           int
	OutputCDN           bool
}
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return ""
}

func (m *multiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}
func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.Targets, "l", "", "目标地址的列表")
	flag.StringVar(&options.Output, "o", "", "输出的文件")
	flag.IntVar(&options.TimeOut, "timeout", 30, "超时时间(s)")
	flag.StringVar(&options.ProxyURL, "proxy-url", "", "URL of the proxy server")
	flag.BoolVar(&options.Verbose, "verbose", false, "输出更多调试信息")
	flag.BoolVar(&options.OutputStatusCode, "status-code", false, "Extracts status code")
	flag.BoolVar(&options.OutputContentLength, "content-length", false, "Extracts content length")
	flag.BoolVar(&options.OutputTitle, "title", false, "Extracts title")
	flag.BoolVar(&options.OutputIP, "ip", false, "Extracts ip")
	flag.IntVar(&options.RateLimit, "limit", 200, "限制每秒的并发数量")
	flag.BoolVar(&options.OutputCDN, "cdn", false, "检测目标是否含有CDN")
	flag.BoolVar(&options.OutputFingerPrint, "fringerprint", false, "输出指纹识别结果")
	flag.Parse()

	return options
}
