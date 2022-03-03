package xray

import (
	"Beescan/core/httpx"
	"errors"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Poc struct {
	Name   string             `yaml:"name"`
	App    []string           `yaml:"app"`
	Set    map[string]string  `yaml:"set"`
	Rules  []Rules            `yaml:"rules"`
	Groups map[string][]Rules `yaml:"groups"`
	Detail Detail             `yaml:"detail"`
	setCel map[string]cel.Program
}

type Rules struct {
	Method          string            `yaml:"method"`
	Path            string            `yaml:"path"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	Search          string            `yaml:"search"`
	FollowRedirects bool              `yaml:"follow_redirects"`
	Expression      string            `yaml:"expression"`
	cel             cel.Program
}

type Detail struct {
	Author      string   `yaml:"author"`
	Links       []string `yaml:"links"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
}

func ParsePocFile(file string) (*Poc, error) {
	template := &Poc{}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(template)
	if err != nil {
		return nil, err
	}
	f2, _ := os.Open(file)
	content, err := ioutil.ReadAll(f2)
	defer f2.Close()
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(content), "newReverse()") {
		return nil, errors.New(file + " 不支持 newReverse()函数")
	}
	if template.Name == "" {
		return nil, errors.New("name field is null")
	}
	return template, nil
}

// Check 对表达式进行预编译和静态类型检查,yaml 中的 set 和 expression 部分
func (p *Poc) Check() error {
	c := NewEnvOption()
	c.UpdateCompileOptions(p.Set)
	env, err := cel.NewEnv(cel.Lib(c))
	if err != nil {
		return errors.New(fmt.Sprintf("environment creation error: %v", err))
	}
	// 编译Expression
	for i, rule := range p.Rules {
		ast, iss := env.Compile(rule.Expression)
		if iss.Err() != nil {
			return errors.New(fmt.Sprintf("Expression %s compile error:%s", p.Name, iss.String()))
		}
		prg, err := env.Program(ast)
		if err != nil {
			return errors.New(fmt.Sprintf("Program creation error: %v", err))
		}
		p.Rules[i].cel = prg
	}
	// 编译Groups
	for key, groups := range p.Groups {
		for i2, rule := range groups {
			ast, iss := env.Compile(rule.Expression)
			if iss.Err() != nil {
				return errors.New(fmt.Sprintf("Expression %s compile error:%s", p.Name, iss.String()))
			}
			prg, err := env.Program(ast)
			if err != nil {
				return errors.New(fmt.Sprintf("Program creation error: %v", err))
			}
			p.Groups[key][i2].cel = prg
		}
	}
	// 编译set
	p.setCel = make(map[string]cel.Program)
	for k, v := range p.Set {
		ast, iss := env.Compile(v)
		if iss.Err() != nil {
			fmt.Printf("Set %s compile error:%s", k, iss.String())
			continue
		}
		prg, err := env.Program(ast)
		if err != nil {
			fmt.Println("Program creation error:", err)
			continue
		}
		p.setCel[k] = prg
	}
	return nil
}
func (p *Poc) ToMsg() string {
	message := fmt.Sprintf("Loading Xray PoC %s (%s)",
		aurora.Bold(p.Name).String(),
		aurora.BrightYellow("@"+p.Detail.Author).String())
	return message
}
func (p *Poc) Execute(url string, hp *httpx.HTTPX) (bool, error) {
	variableMap := make(map[string]interface{})
	req, err := hp.NewRequest("GET", url)
	if err != nil {
		return false, err
	}
	req2, err2 := ParseRequest(req)
	if err2 != nil {
		return false, err
	}
	variableMap["request"] = req2

	// 处理Set
	// 处理非payload字段
	for k, prg := range p.setCel {
		if k != "payload" {
			out, _, err := prg.Eval(variableMap)
			if err != nil {
				return false, errors.New(fmt.Sprintf("Evaluation %s error: %v", k, err))
			}
			switch value := out.Value().(type) {
			case *UrlType:
				variableMap[k] = UrlTypeToString(value)
			case int64:
				variableMap[k] = int(value)
			default:
				variableMap[k] = fmt.Sprintf("%v", out)
			}
		}
	}
	// 处理payload字段
	prg, ok := p.setCel["payload"]
	if ok {
		out, _, err := prg.Eval(variableMap)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Evaluation payload error: %v", err))
		} else {
			variableMap["payload"] = fmt.Sprintf("%v", out)
		}
	}
	if len(p.Rules) > 0 {
		return p.handleRule(url, variableMap, p.Rules)
	}
	if len(p.Groups) > 0 {
		for _, v := range p.Groups {
			vbool, err := p.handleRule(url, variableMap, v)
			if err != nil {
				return vbool, err
			}
			if vbool {
				return true, nil
			}
		}
	}
	return false, nil
}
func (p *Poc) handleRule(url string, variableMap map[string]interface{}, rules []Rules) (bool, error) {
	success := true
	for _, rule := range rules {
		fullUrl := url + replaceValue(rule.Path, variableMap)
		// init httpx
		httpOptions := &httpx.HTTPOptions{
			Timeout:          3 * time.Second,
			RetryMax:         1,
			FollowRedirects:  rule.FollowRedirects,
			HTTPProxy:        "",
			Unsafe:           false,
			DefaultUserAgent: httpx.GetRadnomUserAgent(),
		}
		hp, err := httpx.NewHttpx(httpOptions)
		if err != nil {
			return false, err
		}
		req, err := hp.NewRequest(rule.Method, fullUrl)
		if err != nil {
			return false, err
		}
		newHeader := make(map[string]string)
		if rule.Headers != nil {
			for k, v := range rule.Headers {
				newHeader[k] = replaceValue(v, variableMap)
			}
			hp.SetCustomHeaders(req, newHeader)
		}
		if rule.Body != "" {
			body := rule.Body
			body = replaceValue(body, variableMap)
			req.ContentLength = int64(len(body))
			req.Body = ioutil.NopCloser(strings.NewReader(body))
		}
		resp, err := hp.Do(req)
		if err != nil {
			return false, err
		}
		variableMap["response"] = ParseResponse(resp)

		// 判断响应页面是否匹配search规则
		if rule.Search != "" {
			result := doSearch(strings.TrimSpace(rule.Search), resp.DataStr)
			if result != nil && len(result) > 0 { // 正则匹配成功
				for k, v := range result {
					variableMap[k] = v
				}
				//return false, nil
			} else {
				return false, nil
			}
		}
		out, _, err := rule.cel.Eval(variableMap)
		if err != nil {
			return false, errors.New(fmt.Sprintf("requests eval error:%v domain:%s rule:%s", err, url, rule.Expression))
		}
		if fmt.Sprintf("%v", out) == "false" { //如果false不继续执行后续rule
			success = false // 如果最后一步执行失败，就算前面成功了最终依旧是失败
		}
		if !success {
			break
		}
	}
	return success, nil
}
