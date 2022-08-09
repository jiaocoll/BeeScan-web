package db

import (
	"Beescan/core/config"
	"Beescan/core/httpx"
	"Beescan/core/scan"
	"Beescan/core/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/14
程序功能：elasticsearch模块
*/

type Output struct {
	Ip         string       `json:"ip"`
	TargetId   string       `json:"target_id"`
	Port       string       `json:"port"`
	Protocol   string       `json:"protocol"`
	Domain     string       `json:"domain"`
	Webbanner  FingerResult `json:"webbanner"`
	Servers    Result       `json:"servers"`
	CityId     int64        `json:"cityId"`
	Country    string       `json:"country"`
	Region     string       `json:"region"`
	Province   string       `json:"province"`
	City       string       `json:"city"`
	ISP        string       `json:"isp"`
	Servername string       `json:"servername"`
	Wappalyzer *Wapp        `json:"wappalyzer"`
	Banner     string       `json:"banner"`
	Target     string       `json:"target"`
	LastTime   string       `json:"lastTime"`
	TlsDatas   Tls          `json:"tlsDatas"`
}

type FingerResult struct {
	Title         string              `json:"title"`
	ContentLength int                 `json:"content-length"`
	TLSData       *httpx.TLSData      `json:"tls,omitempty"`
	StatusCode    int                 `json:"status-code"`
	ResponseTime  string              `json:"response-time"`
	CDN           string              `json:"cdn"`
	Fingers       []Fofa              `json:"fingers"`
	Str           string              `json:"str"`
	Header        string              `json:"header"`
	FirstLine     string              `json:"firstLine"`
	Headers       map[string][]string `json:"headers"`
	DataStr       string              `json:"dataStr"`
}

type Tls struct {
	DNSNames         string `json:"dns_names,omitempty"`
	Emails           string `json:"emails,omitempty"`
	CommonName       string `json:"common_name,omitempty"`
	Organization     string `json:"organization,omitempty"`
	IssuerCommonName string `json:"issuer_common_name,omitempty"`
	IssuerOrg        string `json:"issuer_organization,omitempty"`
}

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

// 待探测的目标端口
type Target struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

func (t *Target) GetAddress() string {
	return t.IP + ":" + strconv.Itoa(t.Port)
}

// 输出的结果数据
type Result struct {
	Target  `json:"target"`
	Service `json:"service"`

	Timestamp int32  `json:"timestamp"`
	Error     string `json:"error"`
}

// 获取的端口服务信息
type Service struct {
	Target

	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	Banner      string `json:"banner"`
	BannerBytes []byte `json:"banner_bytes"`

	//IsSSL	    bool `json:"is_ssl"`

	Extras  `json:"extras"`
	Details `json:"details"`
}

// 对应 NMap versioninfo 信息
type Extras struct {
	VendorProduct   string `json:"vendor_product,omitempty"`
	Version         string `json:"version,omitempty"`
	Info            string `json:"info,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	OperatingSystem string `json:"operating_system,omitempty"`
	DeviceType      string `json:"device_type,omitempty"`
	CPE             string `json:"cpe,omitempty"`
}

// 详细的结果数据（包含具体的 Probe 和匹配规则信息）
type Details struct {
	ProbeName     string `json:"probe_name"`
	ProbeData     string `json:"probe_data"`
	MatchMatched  string `json:"match_matched"`
	IsSoftMatched bool   `json:"soft_matched"`
}

type Wapp struct {
	URLs         []ScrapedURL `json:"urls,omitempty"`
	Technologies []Technology `json:"technologies,omitempty"`
}

type ScrapedURL struct {
	URL    string `json:"url,omitempty"`
	Status int    `json:"status,omitempty"`
}

type Technology struct {
	Slug       string             `json:"slug"`
	Name       string             `json:"name"`
	Confidence int                `json:"confidence"`
	Version    string             `json:"version"`
	Icon       string             `json:"icon"`
	Website    string             `json:"website"`
	CPE        string             `json:"cpe"`
	Categories []ExtendedCategory `json:"categories"`
}

type ExtendedCategory struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Priority int    `json:"-"`
}

type Port struct {
	Port string
	Num  int
}

type Country struct {
	Country string
	Num     int
}

type Server struct {
	Server string
	Num    int
}

type NodeLog struct {
	Log      string `json:"log"`
	LastTime string `json:"lastTime"`
}

func (x Result) IsStructureEmpty() bool {
	return reflect.DeepEqual(x, Result{})
}

func ElasticSearchInit(ip string, port string) *elastic.Client {
	host := "http://" + ip + ":" + port
	client, err := elastic.NewClient(elastic.SetURL(host), elastic.SetBasicAuth(config.GlobalConfig.DBConfig.Elasticsearch.Username, config.GlobalConfig.DBConfig.Elasticsearch.Password), elastic.SetSniff(false))
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "[ElasticSearchInit]:", err)
		os.Exit(1)
	}
	return client
}

// QueryByID 通过id搜索
func QueryByID(client *elastic.Client, id string) (Output, []string) {
	var res *elastic.SearchResult
	var res2 *elastic.SearchResult
	var domains []string
	var out Output
	var out2 Output
	var outs []Output
	var err error
	ip := strings.Split(id, "-")
	matchquery := elastic.NewMatchPhraseQuery("target_id", id)
	matchquery2 := elastic.NewMatchPhraseQuery("ip", ip[0])
	size, err := client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(matchquery2).Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(matchquery).Do(context.Background())
	res2, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(size)).From(0).Query(matchquery2).Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if res != nil {
		if res.Hits != nil {
			if res.Hits.Hits != nil {
				for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
					err = json.Unmarshal(item.Source, &out)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	if res2 != nil {
		if res2.Hits != nil {
			if res2.Hits.Hits != nil {
				for _, item := range res2.Hits.Hits { //从搜索结果中取数据的方法
					err = json.Unmarshal(item.Source, &out2)
					if err != nil {
						log.Println(err)
					}
					outs = append(outs, out2)
				}
			}
		}
	}

	for _, v := range outs {
		if v.Domain != "" {
			domains = append(domains, v.Domain)
		}
	}

	return out, domains
}

// Query 资产搜索
func Query(client *elastic.Client, condition string, size, page int) []Output {
	var res *elastic.SearchResult
	var outs []Output
	var err error
	if size > 0 && page > 0 {
		if condition == "" {
			// 取所有
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(size).From((page - 1) * size).Do(context.Background())
		} else if !strings.Contains(condition, "&&") && !strings.Contains(condition, "||") && condition != "" { //单个查询条件

			tmp := strings.Split(condition, "=\"")
			key := tmp[0]
			key = strings.Replace(key, " ", "", -1)
			tmp2 := strings.Split(tmp[1], "\"")
			value := strings.Replace(tmp2[0], "\"", "", -1)

			if key == "body" || key == "header" || key == "title" || key == "domain" {
				// key中包含value
				matchquery := elastic.NewWildcardQuery(key, value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(matchquery).Size(size).From((page - 1) * size).Do(context.Background())
			} else if key == "app" {
				query := elastic.NewQueryStringQuery(value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(query).Size(size).From((page - 1) * size).Do(context.Background())
			} else {
				// 单个条件字段相等
				keyquery := elastic.NewMatchQuery(key, value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(keyquery).Size(size).From((page - 1) * size).Do(context.Background())
			}

		} else if strings.Contains(condition, "&&") && !strings.Contains(condition, "||") { //and逻辑
			tmpcondition := strings.Split(condition, " && ")
			bollQ := elastic.NewBoolQuery()
			for _, v := range tmpcondition {
				tmpv1 := strings.Split(v, "=\"")
				key := tmpv1[0]
				key = strings.Replace(key, " ", "", -1)
				tmpv2 := strings.Split(tmpv1[1], "\"")
				value := strings.Replace(tmpv2[0], "\"", "", -1)

				if key == "body" || key == "header" || key == "title" || key == "domain" {
					// key中包含value
					bollQ.Must(elastic.NewWildcardQuery(key, value))
				} else if key == "app" {
					bollQ.Must(elastic.NewQueryStringQuery(value))
				} else {
					// 单个条件字段相等
					bollQ.Must(elastic.NewMatchQuery(key, value))
				}

			}
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Size(size).From((page - 1) * size).Do(context.Background())
		} else if strings.Contains(condition, "||") && !strings.Contains(condition, "&&") { //or逻辑
			tmpcondition := strings.Split(condition, " || ")
			bollQ := elastic.NewBoolQuery()
			for _, v := range tmpcondition {
				tmpv1 := strings.Split(v, "=\"")
				key := tmpv1[0]
				key = strings.Replace(key, " ", "", -1)
				tmpv2 := strings.Split(tmpv1[1], "\"")
				value := strings.Replace(tmpv2[0], "\"", "", -1)

				if key == "body" || key == "header" || key == "title" || key == "domain" {
					// key中包含value
					bollQ.Should(elastic.NewWildcardQuery(key, value))
				} else if key == "app" {
					bollQ.Should(elastic.NewQueryStringQuery(value))
				} else {
					// 单个条件字段相等
					bollQ.Should(elastic.NewMatchQuery(key, value))
				}

			}
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Size(size).From((page - 1) * size).Do(context.Background())
		}
		if res != nil {
			if res.Hits != nil {
				if res.Hits.Hits != nil {
					for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
						out := Output{}
						err = json.Unmarshal(item.Source, &out)
						if err != nil {
							log.Println(err)
						}
						if out.Target != "" {
							outs = append(outs, out)
						}
					}
				}
			}
		}
	} else {
		log.Println("Page error")
	}

	return outs
}

// QueryVul 漏洞搜索
func QueryVul(client *elastic.Client, condition string, size, page int) []scan.NucleiOutput {
	var res *elastic.SearchResult
	var outs []scan.NucleiOutput
	var err error
	if size > 0 && page > 0 {
		if condition == "" {
			// 取所有
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Size(size).From((page - 1) * size).Do(context.Background())
		} else if !strings.Contains(condition, "&&") && !strings.Contains(condition, "||") && condition != "" { //单个查询条件

			tmp := strings.Split(condition, "=\"")
			key := tmp[0]
			key = strings.Replace(key, " ", "", -1)
			tmp2 := strings.Split(tmp[1], "\"")
			value := strings.Replace(tmp2[0], "\"", "", -1)

			if key == "TaskName" || key == "Target" || key == "PocName" || key == "PocAuthor" {
				// key中包含value
				matchquery := elastic.NewWildcardQuery(key, value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(matchquery).Size(size).From((page - 1) * size).Do(context.Background())
			} else if key == "vul" {
				query := elastic.NewQueryStringQuery(value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(query).Size(size).From((page - 1) * size).Do(context.Background())
			} else {
				// 单个条件字段相等
				keyquery := elastic.NewMatchQuery(key, value)
				res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(keyquery).Size(size).From((page - 1) * size).Do(context.Background())
			}

		} else if strings.Contains(condition, "&&") && !strings.Contains(condition, "||") { //and逻辑
			tmpcondition := strings.Split(condition, " && ")
			bollQ := elastic.NewBoolQuery()
			for _, v := range tmpcondition {
				tmpv1 := strings.Split(v, "=\"")
				key := tmpv1[0]
				key = strings.Replace(key, " ", "", -1)
				tmpv2 := strings.Split(tmpv1[1], "\"")
				value := strings.Replace(tmpv2[0], "\"", "", -1)

				if key == "TaskName" || key == "Target" || key == "PocName" || key == "PocAuthor" {
					// key中包含value
					bollQ.Must(elastic.NewWildcardQuery(key, value))
				} else if key == "vul" {
					bollQ.Must(elastic.NewQueryStringQuery(value))
				} else {
					// 单个条件字段相等
					bollQ.Must(elastic.NewMatchQuery(key, value))
				}

			}
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(bollQ).Size(size).From((page - 1) * size).Do(context.Background())
		} else if strings.Contains(condition, "||") && !strings.Contains(condition, "&&") { //or逻辑
			tmpcondition := strings.Split(condition, " || ")
			bollQ := elastic.NewBoolQuery()
			for _, v := range tmpcondition {
				tmpv1 := strings.Split(v, "=\"")
				key := tmpv1[0]
				key = strings.Replace(key, " ", "", -1)
				tmpv2 := strings.Split(tmpv1[1], "\"")
				value := strings.Replace(tmpv2[0], "\"", "", -1)

				if key == "TaskName" || key == "Target" || key == "PocName" || key == "PocAuthor" {
					// key中包含value
					bollQ.Should(elastic.NewWildcardQuery(key, value))
				} else if key == "vul" {
					bollQ.Should(elastic.NewQueryStringQuery(value))
				} else {
					// 单个条件字段相等
					bollQ.Should(elastic.NewMatchQuery(key, value))
				}

			}
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(bollQ).Size(size).From((page - 1) * size).Do(context.Background())
		}
		if res != nil {
			if res.Hits != nil {
				if res.Hits.Hits != nil {
					for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
						out := scan.NucleiOutput{}
						err = json.Unmarshal(item.Source, &out)
						if err != nil {
							log.Println(err)
						}
						if out.Template != "" {
							outs = append(outs, out)
						}
					}
				}
			}
		}
	} else {
		log.Println("Page error")
	}

	return outs
}

// QueryVulByID 通过id搜索
func QueryVulByID(client *elastic.Client, id string) (scan.NucleiOutput, []string) {
	var res *elastic.SearchResult
	var res2 *elastic.SearchResult
	var vuls []string
	var out scan.NucleiOutput
	var out2 scan.NucleiOutput
	var outs []scan.NucleiOutput
	var err error
	host := strings.Split(id, "-")
	matchquery := elastic.NewMatchPhraseQuery("id", id)
	matchquery2 := elastic.NewMatchPhraseQuery("target", host[1])
	size, err := client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(matchquery2).Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Query(matchquery).Do(context.Background())
	res2, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Size(int(size)).From(0).Query(matchquery2).Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if res != nil {
		if res.Hits != nil {
			if res.Hits.Hits != nil {
				for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
					err = json.Unmarshal(item.Source, &out)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	if res2 != nil {
		if res2.Hits != nil {
			if res2.Hits.Hits != nil {
				for _, item := range res2.Hits.Hits { //从搜索结果中取数据的方法
					err = json.Unmarshal(item.Source, &out2)
					if err != nil {
						log.Println(err)
					}
					outs = append(outs, out2)
				}
			}
		}
	}

	for _, v := range outs {
		if v.Info.Name != "" {
			vuls = append(vuls, v.Info.Name)
		}
	}

	return out, vuls
}

// QuerySort 搜索所有统计排名
func QuerySort(client *elastic.Client, condition string) (int, int, int, []Port, []Country, []Server) {
	var res *elastic.SearchResult
	var count int64
	var outs []Output
	var err error

	if condition == "" {
		// 取所有
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Do(context.Background())
	} else if !strings.Contains(condition, "&&") && !strings.Contains(condition, "||") && condition != "" { //单个查询条件

		tmp := strings.Split(condition, "=\"")
		key := tmp[0]
		key = strings.Replace(key, " ", "", -1)
		tmp2 := strings.Split(tmp[1], "\"")
		value := strings.Replace(tmp2[0], "\"", "", -1)

		if key == "body" || key == "header" || key == "title" || key == "domain" {
			// key中包含value
			matchquery := elastic.NewWildcardQuery(key, value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(matchquery).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(matchquery).Do(context.Background())
		} else if key == "app" {
			quert := elastic.NewQueryStringQuery(value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(quert).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(quert).Do(context.Background())
		} else {
			// 单个条件字段相等
			keyquery := elastic.NewMatchQuery(key, value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(keyquery).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(keyquery).Do(context.Background())
		}

	} else if strings.Contains(condition, "&&") && !strings.Contains(condition, "||") { //and逻辑
		tmpcondition := strings.Split(condition, " && ")
		bollQ := elastic.NewBoolQuery()
		for _, v := range tmpcondition {
			tmpv1 := strings.Split(v, "=\"")
			key := tmpv1[0]
			key = strings.Replace(key, " ", "", -1)
			tmpv2 := strings.Split(tmpv1[1], "\"")
			value := strings.Replace(tmpv2[0], "\"", "", -1)

			if key == "body" || key == "header" || key == "title" || key == "domain" {
				// key中包含value
				bollQ.Must(elastic.NewWildcardQuery(key, value))
			} else if key == "app" {
				bollQ.Must(elastic.NewQueryStringQuery(value))
			} else {
				// 单个条件字段相等
				bollQ.Must(elastic.NewMatchQuery(key, value))
			}

		}
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(bollQ).Do(context.Background())
	} else if strings.Contains(condition, "||") && !strings.Contains(condition, "&&") { //or逻辑
		tmpcondition := strings.Split(condition, " || ")
		bollQ := elastic.NewBoolQuery()
		for _, v := range tmpcondition {
			tmpv1 := strings.Split(v, "=\"")
			key := tmpv1[0]
			key = strings.Replace(key, " ", "", -1)
			tmpv2 := strings.Split(tmpv1[1], "\"")
			value := strings.Replace(tmpv2[0], "\"", "", -1)

			if key == "body" || key == "header" || key == "title" || key == "domain" {
				// key中包含value
				bollQ.Should(elastic.NewWildcardQuery(key, value))
			} else if key == "app" {
				bollQ.Should(elastic.NewQueryStringQuery(value))
			} else {
				// 单个条件字段相等
				bollQ.Should(elastic.NewMatchQuery(key, value))
			}

		}
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(bollQ).Do(context.Background())
	}

	if res != nil {
		if res.Hits != nil {
			if res.Hits.Hits != nil {
				for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
					out := Output{}
					err = json.Unmarshal(item.Source, &out)
					if err != nil {
						log.Println(err)
					}
					if out.Target != "" {
						outs = append(outs, out)
					}
				}
			}
		}
	}

	var assetsnum int
	var ipnum int
	var portnum int
	var portsort []Port
	var countrysort []Country
	var serversort []Server
	var ips []string

	// 统计、排序
	if len(outs) != 0 {
		for _, v := range outs {
			ips = append(ips, v.Ip)
			portnum++
			if v.Webbanner.Header != "" {
				assetsnum++
			}
			tmpport := Port{Port: v.Port, Num: 1}
			tmpcountry := Country{Country: v.Country, Num: 1}
			tmpserver := Server{Server: v.Servername, Num: 1}

			if v.Port != "" {
				// 端口集合
				if len(portsort) == 0 {
					portsort = append(portsort, tmpport)
				} else {
					for i := 0; i < len(portsort); i++ {
						if tmpport.Port == portsort[i].Port {
							portsort[i].Num++
							tmpport.Num--
						} else if i == len(portsort)-1 && tmpport.Num == 1 {
							portsort = append(portsort, tmpport)
							tmpport.Num--
							break
						}
					}
				}
			}
			if v.Country != "" {
				// 国家集合
				if len(countrysort) == 0 {
					countrysort = append(countrysort, tmpcountry)
				} else {
					for i := 0; i < len(countrysort); i++ {
						if tmpcountry.Country == countrysort[i].Country {
							countrysort[i].Num++
							tmpcountry.Num--
						} else if i == len(countrysort)-1 && tmpcountry.Num == 1 {
							countrysort = append(countrysort, tmpcountry)
							tmpcountry.Num--
							break
						}
					}
				}
			}
			if v.Servername != "" {
				// 服务集合
				if len(serversort) == 0 {
					serversort = append(serversort, tmpserver)
				} else {
					for i := 0; i < len(serversort); i++ {
						if tmpserver.Server == serversort[i].Server {
							serversort[i].Num++
							tmpserver.Num--
						} else if i == len(serversort)-1 && tmpserver.Num == 1 {
							serversort = append(serversort, tmpserver)
							tmpserver.Num--
							break
						}
					}
				}
			}

		}
	}

	// 端口排序（从大到小）
	for i := 0; i < len(portsort)-1; i++ {
		for j := i + 1; j < len(portsort); j++ {
			if portsort[i].Num < portsort[j].Num {
				tmp := portsort[i]
				portsort[i] = portsort[j]
				portsort[j] = tmp
			}
		}
	}
	// 国家排序（从大到小）
	for i := 0; i < len(countrysort)-1; i++ {
		for j := i + 1; j < len(countrysort); j++ {
			if countrysort[i].Num < countrysort[j].Num {
				tmp := countrysort[i]
				countrysort[i] = countrysort[j]
				countrysort[j] = tmp
			}
		}
	}
	// 服务排序（从大到小）
	for i := 0; i < len(serversort)-1; i++ {
		for j := i + 1; j < len(serversort); j++ {
			if serversort[i].Num < serversort[j].Num {
				tmp := serversort[i]
				serversort[i] = serversort[j]
				serversort[j] = tmp
			}
		}
	}

	ips = util.Removesamesip(ips)
	ipnum = len(ips)

	return assetsnum, ipnum, portnum, portsort, countrysort, serversort
}

func QueryToExport(client *elastic.Client, condition string) []Output {
	var res *elastic.SearchResult
	var count int64
	var outs []Output
	var err error

	if condition == "" {
		// 取所有
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Do(context.Background())
	} else if !strings.Contains(condition, "&&") && !strings.Contains(condition, "||") && condition != "" { //单个查询条件

		tmp := strings.Split(condition, "=\"")
		key := tmp[0]
		key = strings.Replace(key, " ", "", -1)
		tmp2 := strings.Split(tmp[1], "\"")
		value := strings.Replace(tmp2[0], "\"", "", -1)

		if key == "body" || key == "header" || key == "title" || key == "domain" {
			// key中包含value
			matchquery := elastic.NewWildcardQuery(key, value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(matchquery).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(matchquery).Do(context.Background())
		} else if key == "app" {
			quert := elastic.NewQueryStringQuery(value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(quert).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(quert).Do(context.Background())
		} else {
			// 单个条件字段相等
			keyquery := elastic.NewMatchQuery(key, value)
			count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(keyquery).Do(context.Background())
			res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(keyquery).Do(context.Background())
		}

	} else if strings.Contains(condition, "&&") && !strings.Contains(condition, "||") { //and逻辑
		tmpcondition := strings.Split(condition, " && ")
		bollQ := elastic.NewBoolQuery()
		for _, v := range tmpcondition {
			tmpv1 := strings.Split(v, "=\"")
			key := tmpv1[0]
			key = strings.Replace(key, " ", "", -1)
			tmpv2 := strings.Split(tmpv1[1], "\"")
			value := strings.Replace(tmpv2[0], "\"", "", -1)

			if key == "body" || key == "header" || key == "title" || key == "domain" {
				// key中包含value
				bollQ.Must(elastic.NewWildcardQuery(key, value))
			} else if key == "app" {
				bollQ.Must(elastic.NewQueryStringQuery(value))
			} else {
				// 单个条件字段相等
				bollQ.Must(elastic.NewMatchQuery(key, value))
			}

		}
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(bollQ).Do(context.Background())
	} else if strings.Contains(condition, "||") && !strings.Contains(condition, "&&") { //or逻辑
		tmpcondition := strings.Split(condition, " || ")
		bollQ := elastic.NewBoolQuery()
		for _, v := range tmpcondition {
			tmpv1 := strings.Split(v, "=\"")
			key := tmpv1[0]
			key = strings.Replace(key, " ", "", -1)
			tmpv2 := strings.Split(tmpv1[1], "\"")
			value := strings.Replace(tmpv2[0], "\"", "", -1)

			if key == "body" || key == "header" || key == "title" || key == "domain" {
				// key中包含value
				bollQ.Should(elastic.NewWildcardQuery(key, value))
			} else if key == "app" {
				bollQ.Should(elastic.NewQueryStringQuery(value))
			} else {
				// 单个条件字段相等
				bollQ.Should(elastic.NewMatchQuery(key, value))
			}

		}
		count, err = client.Count().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index).Query(bollQ).Do(context.Background())
		res, err = client.Search(config.GlobalConfig.DBConfig.Elasticsearch.Index).Size(int(count)).From(0).Query(bollQ).Do(context.Background())
	}

	if res != nil {
		if res.Hits != nil {
			if res.Hits.Hits != nil {
				for _, item := range res.Hits.Hits { //从搜索结果中取数据的方法
					out := Output{}
					err = json.Unmarshal(item.Source, &out)
					if err != nil {
						log.Println(err)
					}
					outs = append(outs, out)
				}
			}
		}
	}

	return outs
}

// QueryLogByID 节点日志查询
func QueryLogByID(client *elastic.Client, nodename string) NodeLog {
	var res *elastic.GetResult
	var err error
	var TheNodeLog NodeLog
	id := nodename + "_log"
	res, err = client.Get().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_log").Id(id).Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if res != nil {
		if res.Found {
			if res.Source != nil {
				err = json.Unmarshal(res.Source, &TheNodeLog)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	return TheNodeLog
}

// EsAdd 添加结果到es数据库
func EsAdd(client *elastic.Client, res scan.NucleiOutput) {

	// 文档件存在则更新，否则插入
	_, err := client.Update().Index(config.GlobalConfig.DBConfig.Elasticsearch.Index + "_vul").Id(res.ID).Doc(res).Upsert(res).Refresh("true").Do(context.Background())
	if err != nil {
		log.Println("[DBEsUpInsert]:", err)
	}

}
