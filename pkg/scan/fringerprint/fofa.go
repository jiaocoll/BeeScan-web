package fringerprint

import (
	"Beescan/pkg/httpx"
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：指纹识别
*/


type Fofa struct {
	RuleId         string `json:"rule_id"`
	Level          string `json:"level"`
	SoftHard       string `json:"softhard"`
	Product        string `json:"product"`
	Company        string `json:"company"`
	Category       string `json:"category"`
	ParentCategory string `json:"parent_category"`
	Condition      string `json:"Condition"`
}
type FofaPrints []Fofa


var FofaJson []byte

func InitFofa() (FofaPrints, error) {
	//datas, err := ioutil.ReadFile(filename)
	//if err != nil {
	//	return nil, err
	//}
	var fofas FofaPrints
	err := json.Unmarshal(FofaJson, &fofas)
	if err != nil {
		return nil, err
	}
	return fofas, err
}

func (f *Fofa) Matcher(response *httpx.Response) (bool, error) {
	expString := f.Condition
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(expString, HelperFunctions(response))
	if err != nil {
		return false, err
	}
	paramters := make(map[string]interface{})
	paramters["title"] = response.Title
	paramters["server"] = response.GetHeader("server")
	paramters["protocol"] = "http"

	result, err := expression.Evaluate(paramters)
	if err != nil {
		return false, err
	}
	t := result.(bool)
	return t, err
}
func (f *FofaPrints) Matcher(response *httpx.Response) ([]string, error) {
	ret := make([]string, 0)
	for _, item := range *f {
		v, err := item.Matcher(response)
		if err != nil {
			return nil, err
		}
		if v {
			n := item.Product
			ret = append(ret, n)
		}
	}
	return ret, nil
}

// HelperFunctions contains the dsl functions
func HelperFunctions(resp *httpx.Response) (functions map[string]govaluate.ExpressionFunction) {
	functions = make(map[string]govaluate.ExpressionFunction)

	functions["title_contains"] = func(args ...interface{}) (interface{}, error) {
		pattern := strings.ToLower(toString(args[0]))
		title := strings.ToLower(resp.Title)
		return strings.Index(title, pattern) != -1, nil
	}

	functions["body_contains"] = func(args ...interface{}) (interface{}, error) {
		pattern := strings.ToLower(toString(args[0]))
		data := strings.ToLower(resp.DataStr)
		return strings.Index(data, pattern) != -1, nil
	}

	functions["protocol_contains"] = func(args ...interface{}) (interface{}, error) {
		return false, nil
	}

	functions["banner_contains"] = func(args ...interface{}) (interface{}, error) {
		return false, nil
	}

	functions["header_contains"] = func(args ...interface{}) (interface{}, error) {
		pattern := strings.ToLower(toString(args[0]))
		data := strings.ToLower(resp.HeaderStr)
		return strings.Index(data, pattern) != -1, nil
	}

	functions["server_contains"] = func(args ...interface{}) (interface{}, error) {
		pattern := strings.ToLower(toString(args[0]))
		server := resp.GetHeader("server")
		return strings.Index(server, pattern) != -1, nil
	}

	functions["cert_contains"] = func(args ...interface{}) (interface{}, error) {
		return false, nil
	}

	functions["port_contains"] = func(args ...interface{}) (interface{}, error) {
		return false, nil
	}

	return functions
}

func toString(v interface{}) string {
	return fmt.Sprint(v)
}

func toInt(v interface{}) int {
	return int(v.(float64))
}