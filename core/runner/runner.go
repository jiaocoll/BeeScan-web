package runner

import (
	"Beescan/core/httpx"
	"Beescan/core/poc/nuclei"
	"Beescan/core/poc/xray"
	"github.com/projectdiscovery/fastdialer/fastdialer"
	"github.com/projectdiscovery/hmap/store/hybrid"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"go.uber.org/ratelimit"
	"log"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：
*/

type Runner struct {
	hp         *httpx.HTTPX
	hm         *hybrid.HybridMap
	result     []Result
	nuclei     *nuclei.NucleiPoC
	xrayPocs   []*xray.Poc
	nucleiPoCs []*templates.Template
	pocChanel  chan string
	taskname   string
}

func NewRunner(TaskName string, targets []string) *Runner {
	runner := &Runner{taskname: TaskName}

	dialer, err := fastdialer.NewDialer(fastdialer.DefaultOptions)
	if err != nil {
		log.Println(err)
		return nil
	}
	httpOptions := &httpx.HTTPOptions{
		Timeout:          3 * time.Second,
		RetryMax:         3,
		FollowRedirects:  true,
		HTTPProxy:        "",
		Unsafe:           false,
		DefaultUserAgent: httpx.GetRadnomUserAgent(),
		Dialer:           dialer,
	}
	hp, err := httpx.NewHttpx(httpOptions)
	if err != nil {
		log.Println(err)
		return nil
	}
	runner.hp = hp

	nu, err := nuclei.New(ratelimit.New(5))
	if err != nil {
		log.Println(err)
	}
	runner.nuclei = nu
	runner.pocChanel = make(chan string, 2000)
	for i := 0; i < len(targets); i++ {
		runner.pocChanel <- targets[i]
	}
	return runner
}

func (r *Runner) RunPoc() chan PocResult {
	output := make(chan PocResult)
	for target := range r.pocChanel {
		r.RunPocs(target, output)
	}
	return output
}
