package runner

import (
	"Beescan/pkg/httpx"
	"Beescan/pkg/scan/fringerprint"
	"fmt"
	"github.com/projectdiscovery/hmap/store/hybrid"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：
*/


type Runner struct {
	Ip			string
	Port 		string
	ht          *httpx.HTTPX
	hm          *hybrid.HybridMap
	fofa        *fringerprint.FofaPrints
}

// 创建runner实例
func NewRunner()(*Runner,error){
	runner := &Runner{
	}
	if hm, err := hybrid.New(hybrid.DefaultDiskOptions); err != nil {
		log.Fatalf("Could not create temporary input file: %s\n", err)
	} else {
		runner.hm = hm
	}


	// http
	HttpOptions := &httpx.HTTPOptions{
		Timeout:          3 * time.Second,
		RetryMax:         3,
		FollowRedirects:  true,
		Unsafe:           false,
		DefaultUserAgent: httpx.GetRadnomUserAgent(),
	}
	ht, err := httpx.NewHttpx(HttpOptions)
	if err != nil {
		return nil, err
	}
	runner.ht = ht

	// fofa
	FofaPrints, err := fringerprint.InitFofa()
	if err != nil {
		return	nil, err
	}
	runner.fofa = &FofaPrints
	return runner, nil

}


// http请求
func do(r *Runner, fullUrl string) (*httpx.Response, error) {
	req, err := r.ht.NewRequest("GET", fullUrl)

	if err != nil {
		return &httpx.Response{}, err
	}
	resp, err2 := r.ht.Do(req)
	return resp, err2
}


func runRequest(r *Runner, domain string, output chan Result) {

	retried := false
	protocol := httpx.HTTPS
retry:
	fullUrl := fmt.Sprintf("%s://%s", protocol, domain)
	timeStart := time.Now()

	resp, err := do(r,fullUrl)
	if err != nil {
		if !retried {
			if protocol == httpx.HTTPS {
				protocol = httpx.HTTP
			} else {
				protocol = httpx.HTTPS
			}
			retried = true
			goto retry
		}
		return
	}
	builder := &strings.Builder{}
	builder.WriteString(fullUrl)



	title := resp.Title

	p, err := url.Parse(fullUrl)
	var ip string
	var ipArray []string
	if err != nil {
		ip = ""
	} else {
		hostname := p.Hostname()
		ip = r.ht.Dialer.GetDialedIP(hostname)
		// ip为空，看看p.host是不是ip
		if ip == "" {
			address := net.ParseIP(hostname)
			if address != nil {
				ip = address.String()
			}
		}
	}
	dnsData, err := r.ht.Dialer.GetDNSData(p.Host)
	if dnsData != nil && err == nil {
		ipArray = append(ipArray, dnsData.CNAME...)
		ipArray = append(ipArray, dnsData.A...)
		ipArray = append(ipArray, dnsData.AAAA...)
	}
	var cdn string
	// 指纹处理
	fofaResults, err := r.fofa.Matcher(resp)
	if err != nil {
		fmt.Println(err)
	}
	if fofaResults != nil {
		s := strings.Join(fofaResults, ",")
		builder.WriteString(fmt.Sprintf(" [%s] ", s))
	}
	result := fingerResult{
		URL:           fullUrl,
		IP:            ip,
		Title:         title,
		TLSData:       resp.TLSData,
		ContentLength: resp.ContentLength,
		StatusCode:    resp.StatusCode,
		ResponseTime:  time.Since(timeStart).String(),
		str:           builder.String(),
		CDN:           cdn,
	}
	output <- &result
}

func Close(r *Runner) {
	r.ht.Dialer.Close()
	_ = r.hm.Close()
}
