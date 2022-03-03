package nuclei

import (
	"errors"
	"fmt"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog"
	"github.com/projectdiscovery/nuclei/v2/pkg/output"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/protocolinit"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"github.com/projectdiscovery/nuclei/v2/pkg/types"
	"go.uber.org/ratelimit"
)

type NucleiPoC struct {
	option protocols.ExecuterOptions
}

func New(limiter ratelimit.Limiter) (*NucleiPoC, error) {
	fakeWriter := fakeWrite{}
	progress := &fakeProgress{}
	o := types.Options{
		RateLimit:               5,
		BulkSize:                25,
		TemplateThreads:         25,
		HeadlessBulkSize:        10,
		HeadlessTemplateThreads: 10,
		Timeout:                 3,
		Retries:                 1,
		MaxHostError:            30,
	}
	r := NucleiPoC{}
	err := protocolinit.Init(&o)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not initialize protocols: %s", err))
	}
	catalog2 := catalog.New("")
	var executerOpts = protocols.ExecuterOptions{
		Output:      &fakeWriter,
		Options:     &o,
		Progress:    progress,
		Catalog:     catalog2,
		RateLimiter: limiter,
	}
	r.option = executerOpts
	return &r, nil
}
func (n *NucleiPoC) ParsePocFile(filePath string) (*templates.Template, error) {
	var err error
	template := &templates.Template{}
	template, err = templates.Parse(filePath, nil, n.option)
	if err != nil {
		return nil, err
	}
	if template != nil {
		return template, nil
	}

	return nil, nil

}

func ExecuteNucleiPoc(input string, poc *templates.Template) ([]string, error) {
	var ret []string
	var results bool
	e := poc.Executer
	name := fmt.Sprint(poc.ID)
	err := e.ExecuteWithResults(input, func(result *output.InternalWrappedEvent) {
		for _, r := range result.Results {
			results = true
			if r.ExtractorName != "" {
				ret = append(ret, name+":"+r.ExtractorName)
			} else if r.MatcherName != "" {
				ret = append(ret, name+":"+r.MatcherName)
			}
		}
	})
	if err != nil || !results {
		return nil, nil
	}
	if len(ret) == 0 {
		ret = append(ret, name)
	}
	return ret, err
}
