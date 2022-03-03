package controller

import (
	"Beescan/core/banner"
	"Beescan/core/config"
	"Beescan/core/db"
	"Beescan/core/json"
	"Beescan/core/scan/hostinfo"
	"Beescan/core/util"
	"Beescan/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/3
程序功能：控制器
*/
var (
	nowtime    string
	hostip     string
	hostinfos  *host.InfoStat
	parts      hostinfo.Parts
	cpuinfos   hostinfo.CpuInfo
	mempercent float64
	meminfos   hostinfo.MemInfo
	netinfos   []net.IOCountersStat
	netspeed   []hostinfo.SpeedInfo
	conn       *redis.Client
	nodesstate []db.NodeState
	tasksstate []db.TaskState
	nodenames  []string
	tasknames  []string
	es         *elastic.Client
)

type Page struct {
	SearchStr string
	NowPage   int
}

type Wapp struct {
	Uname      string `json:"uname"`
	Lname      string `json:"lname"`
	Confidence int    `json:"confidence"`
	Version    string `json:"version"`
	Icon       string `json:"icon"`
	Website    string `json:"website"`
	Cpe        string `json:"cpe"`
}

type Domain struct {
	Domainstr string
}

func init() {
	banner.Banner()
	fmt.Fprintln(color.Output, color.HiMagentaString("Initializing......"))
	config.Setup()
	var err error
	conn = db.RedisInit()
	es = db.ElasticSearchInit(config.GlobalConfig.DBConfig.Elasticsearch.Host, config.GlobalConfig.DBConfig.Elasticsearch.Port)
	hostip = hostinfo.GetLocalIP()
	hostinfos, err = hostinfo.GetHostInfo()
	if err != nil {
		log.Println(err)
	}
	parts, err = hostinfo.GetDiskInfo()
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[GetDiskInfo]:", err)
	}
	InfoInit()
	fmt.Fprintln(color.Output, color.HiMagentaString("Initialized!"))
}

// LoginGet 登录页面
func LoginGet(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// LoginPost 登录页面
func LoginPost(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	//查询数据库，如果正确跳转，否则不跳转
	if username == config.GlobalConfig.UserPassConfig.UserName && password == config.GlobalConfig.UserPassConfig.PassWord {
		c.Request.URL.Path = "/info"

	}
}

// InfoGet 本机信息
func InfoGet(c *gin.Context) {
	var err error
	nowtime = time.Now().Format("2006-01-02 15:04:05")
	go func() {
		cpuinfos, err = hostinfo.GetCpuPercent()
		if err != nil {
			log.Println(err)
		}
		meminfos = hostinfo.GetMemInfo()
		netinfos, err = hostinfo.GetNetInfo()
		if err != nil {
			log.Println(err)
		}
		netspeed = hostinfo.GetNetSpeed()
	}()
	c.HTML(http.StatusOK, "info.html", gin.H{"hostip": hostip, "hostinfos": hostinfos, "parts": parts,
		"cpuinfos": cpuinfos, "mempercent": mempercent, "meminfos": meminfos,
		"netinfos": netinfos, "netspeed": netspeed, "nowtime": nowtime,
	})
}

// AssetsGet 资产展示
func AssetsGet(c *gin.Context) {

	var assetsnum, ipnum, portnum int
	var portsort []db.Port
	var countrysort []db.Country
	var serversort []db.Server
	var leftpages []Page
	var rightpages []Page

	searchstr := c.Query("search")
	page := c.Query("page")
	if searchstr != "" && page == "" {
		outputs := db.Query(es, searchstr, 10, 1)
		assetsnum, ipnum, portnum, portsort, countrysort, serversort = db.QuerySort(es, searchstr)
		if outputs != nil {
			for i := 2; i < 11; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}
		currentpage := 1
		leftpage := currentpage
		rightpage := currentpage + 1
		c.HTML(http.StatusOK, "assets.html", gin.H{"outputs": outputs, "currentpage": currentpage, "leftpage": leftpage, "rightpage": rightpage,
			"searchstr": searchstr, "leftpages": leftpages, "rightpages": rightpages,
			"assetsnum": assetsnum, "ipnum": ipnum, "portnum": portnum,
			"portsort": portsort, "countrysort": countrysort, "serversort": serversort,
		})
	} else if searchstr != "" && page != "" {
		condition := searchstr
		searchpage, _ := strconv.Atoi(page)
		outputs := db.Query(es, condition, 10, searchpage)
		assetsnum, ipnum, portnum, portsort, countrysort, serversort = db.QuerySort(es, searchstr)

		if searchpage > 0 && searchpage <= 5 {
			for i := 1; i < searchpage; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				leftpages = append(leftpages, a)
			}
			for i := searchpage + 1; i < 11; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}

		if searchpage > 5 {
			for i := searchpage - 4; i < searchpage; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				leftpages = append(leftpages, a)
			}
			for i := searchpage + 1; i < searchpage+6; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}
		leftpage := searchpage - 1
		rightpage := searchpage + 1
		c.HTML(http.StatusOK, "assets.html", gin.H{
			"outputs": outputs, "currentpage": searchpage, "leftpage": leftpage, "rightpage": rightpage,
			"searchstr": searchstr, "leftpages": leftpages, "rightpages": rightpages,
			"assetsnum": assetsnum, "ipnum": ipnum, "portnum": portnum,
			"portsort": portsort, "countrysort": countrysort, "serversort": serversort,
		})
	} else if searchstr == "" && page != "" {
		searchpage, _ := strconv.Atoi(page)

		outputs := db.Query(es, "", 10, searchpage)
		assetsnum, ipnum, portnum, portsort, countrysort, serversort = db.QuerySort(es, searchstr)
		if searchpage > 0 && searchpage <= 5 {
			for i := 1; i < searchpage; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				leftpages = append(leftpages, a)
			}
			for i := searchpage + 1; i < 11; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}

		if searchpage > 5 {
			for i := searchpage - 4; i < searchpage; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				leftpages = append(leftpages, a)
			}
			for i := searchpage + 1; i < searchpage+6; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}
		leftpage := searchpage - 1
		rightpage := searchpage + 1
		c.HTML(http.StatusOK, "assets.html", gin.H{"outputs": outputs, "currentpage": searchpage, "leftpage": leftpage, "rightpage": rightpage,
			"searchstr": searchstr, "leftpages": leftpages, "rightpages": rightpages,
			"assetsnum": assetsnum, "ipnum": ipnum, "portnum": portnum,
			"portsort": portsort, "countrysort": countrysort, "serversort": serversort,
		})
	} else if searchstr == "" && page == "" {
		outputs := db.Query(es, "", 10, 1)
		assetsnum, ipnum, portnum, portsort, countrysort, serversort = db.QuerySort(es, searchstr)
		if outputs != nil {
			for i := 2; i < 11; i++ {
				a := Page{NowPage: i, SearchStr: searchstr}
				rightpages = append(rightpages, a)
			}
		}
		currentpage := 1
		leftpage := 1
		rightpage := 1
		c.HTML(http.StatusOK, "assets.html", gin.H{"outputs": outputs, "currentpage": currentpage, "leftpage": leftpage, "rightpage": rightpage,
			"searchstr": searchstr, "leftpages": leftpages, "rightpages": rightpages,
			"assetsnum": assetsnum, "ipnum": ipnum, "portnum": portnum,
			"portsort": portsort, "countrysort": countrysort, "serversort": serversort,
		})
	}

}

// AssetsPost 资产展示
func AssetsPost(c *gin.Context) {
	c.HTML(http.StatusOK, "assets.html", gin.H{"outputs": ""})
}

// 资产导出
func AssetsExport(c *gin.Context) {
	searchstr := c.Query("search")
	file, err := utils.GetdataTocsv(db.QueryToExport(es, searchstr))
	if err != nil {
		log.Println(err)
	}
	defer func() {
		err := os.Remove("./" + file)
		if err != nil {
			fmt.Println(err)
		}
	}()
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file))
	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.File("./" + file)
}

// SingleAssetsDetail 单个资产详细展示
func SingleAssetsDetail(c *gin.Context) {
	var wapps []Wapp
	var domains []Domain
	id := c.Query("detail")
	output, domainstr := db.QueryByID(es, id)
	for _, v := range domainstr {
		tmpdomain := Domain{
			Domainstr: v,
		}
		domains = append(domains, tmpdomain)
	}
	tls := db.Tls{}
	if output.Webbanner.ContentLength != 0 {
		for k, v := range output.Webbanner.TLSData.DNSNames {
			if k != len(output.Webbanner.TLSData.DNSNames)-1 {
				tls.DNSNames += v + " / "
			} else {
				tls.DNSNames += v
			}
		}
		for k, v := range output.Webbanner.TLSData.Emails {
			if k != len(output.Webbanner.TLSData.Emails)-1 {
				tls.Emails += v + " / "
			} else {
				tls.Emails += v
			}
		}
		for k, v := range output.Webbanner.TLSData.CommonName {
			if k != len(output.Webbanner.TLSData.CommonName)-1 {
				tls.CommonName += v + " / "
			} else {
				tls.CommonName += v
			}
		}
		for k, v := range output.Webbanner.TLSData.Organization {
			if k != len(output.Webbanner.TLSData.Organization)-1 {
				tls.Organization += v + " / "
			} else {
				tls.Organization += v
			}
		}
		for k, v := range output.Webbanner.TLSData.IssuerCommonName {
			if k != len(output.Webbanner.TLSData.IssuerCommonName)-1 {
				tls.IssuerCommonName += v + " / "
			} else {
				tls.IssuerCommonName += v
			}
		}
		for k, v := range output.Webbanner.TLSData.IssuerOrg {
			if k != len(output.Webbanner.TLSData.IssuerOrg)-1 {
				tls.IssuerOrg += v + " / "
			} else {
				tls.IssuerOrg += v
			}
		}
		output.TlsDatas = tls

		if output.Wappalyzer != nil {
			for _, v := range output.Wappalyzer.Technologies {
				Uname := ""
				for k, vv := range v.Categories {
					if k != len(v.Categories)-1 {
						Uname += vv.Name + " / "
					} else {
						Uname += vv.Name
					}
				}
				wap := Wapp{
					Uname:      Uname,
					Lname:      v.Name,
					Confidence: v.Confidence,
					Version:    v.Version,
					Icon:       v.Icon,
					Cpe:        v.CPE,
					Website:    v.Website,
				}
				wapps = append(wapps, wap)
			}
		}

	}

	c.HTML(http.StatusOK, "assetsdetail.html", gin.H{"output": output, "wapps": wapps, "domains": domains})
}

// ScanGet 资产探测
func ScanGet(c *gin.Context) {

	nodesstate = db.GetNodeStates(conn, config.GlobalConfig.NodeConfig.NodeNames)
	tasksstate = db.GetTaskStates(conn, tasknames)
	var tasks int
	var running int
	var finished int

	for _, v := range nodesstate {
		tmptasks, _ := strconv.Atoi(v.Tasks)
		tmprunning, _ := strconv.Atoi(v.Running)
		tmpfinished, _ := strconv.Atoi(v.Finished)
		tasks += tmptasks
		running += tmprunning
		finished += tmpfinished
	}
	c.HTML(200, "scan.html", gin.H{
		"NodeStates": nodesstate,
		"TaskStates": tasksstate,
		"Tasks":      tasks,
		"Running":    running,
		"Finished":   finished,
	})
}

// ScanPost 资产探测
func ScanPost(c *gin.Context) {

	//获取用户输入的目标
	TaskName := c.PostForm("task_name")
	TargetsHost := c.PostForm("targets_host")
	NodeName := c.PostForm("node_name")
	Port := c.PostForm("targets_ports")
	if NodeName != "" {
		if TargetsHost != "" && Port != "" {
			TargetsHost = strings.Replace(TargetsHost, "\r\n", ",", -1)
			TargetsHosts := strings.Split(TargetsHost, ",")

			Ports := strings.Split(Port, ",")

			NodeQueue := NodeName + "_queue"
			TargetsHosts = util.Removesamesip(TargetsHosts)
			if TaskName == "" {
				TaskName = "BeeScan_task_1"
				if !util.In(TaskName, tasknames) {
					tasknames = append(tasknames, TaskName)
				}
			} else {
				if !util.In(TaskName, tasknames) {
					tasknames = append(tasknames, TaskName)
				}
			}

			var targets string
			targets += TaskName + ","
			// 每一个ip和端口构成一个扫描目标,组成目标集合
			for _, p := range Ports {
				for _, i := range TargetsHosts {
					targets += fmt.Sprintf("%s:%s,", i, p)
				}
			}
			tmpnum := strings.Split(targets, ",")

			// 将目标装换成bytes数据
			jsjob, err := json.MarshalBinary(targets)
			if err != nil {
				log.Println(err)
			}

			// 将目标送进redis消息队列中
			err1 := db.AddJob(conn, jsjob, NodeQueue)
			if err1 != nil {
				log.Println("[ADDJob]:", err1)
			}
			db.TaskRegister(conn, TaskName, strconv.Itoa(len(tmpnum)-2))
		}
	}

	nodesstate = db.GetNodeStates(conn, config.GlobalConfig.NodeConfig.NodeNames)
	tasksstate = db.GetTaskStates(conn, tasknames)

	var tasks int
	var running int
	var finished int

	for _, v := range nodesstate {
		tmptasks, _ := strconv.Atoi(v.Tasks)
		tmprunning, _ := strconv.Atoi(v.Running)
		tmpfinished, _ := strconv.Atoi(v.Finished)
		tasks += tmptasks
		running += tmprunning
		finished += tmpfinished
	}

	c.HTML(200, "scan.html", gin.H{
		"NodeStates": nodesstate,
		"TaskStates": tasksstate,
		"Tasks":      tasks,
		"Running":    running,
		"Finished":   finished,
	})

}

// TaskDelete 删除任务
func TaskDelete(c *gin.Context) {
	name := c.Query("delete")
	if name != "" {
		db.DelTask(conn, name)
	}
	nodesstate = db.GetNodeStates(conn, config.GlobalConfig.NodeConfig.NodeNames)
	tasksstate = db.GetTaskStates(conn, tasknames)
	var tasks int
	var running int
	var finished int

	for _, v := range nodesstate {
		tmptasks, _ := strconv.Atoi(v.Tasks)
		tmprunning, _ := strconv.Atoi(v.Running)
		tmpfinished, _ := strconv.Atoi(v.Finished)
		tasks += tmptasks
		running += tmprunning
		finished += tmpfinished
	}
	c.HTML(200, "scan.html", gin.H{
		"NodeStates": nodesstate,
		"TaskStates": tasksstate,
		"Tasks":      tasks,
		"Running":    running,
		"Finished":   finished,
	})
}

// VulGet 漏洞检测
func VulGet(c *gin.Context) {
	c.HTML(http.StatusOK, "vul.html", nil)
}

// VulPost 漏洞检测
func VulPost(c *gin.Context) {
	c.HTML(http.StatusOK, "vul.html", nil)
}

// PocGet POC管理
func PocGet(c *gin.Context) {
	c.HTML(http.StatusOK, "poc.html", nil)
}

// PocAdd 添加POC
func PocAdd(c *gin.Context) {
	file, err := c.FormFile("pocfile")
	if file != nil {
		if err != nil {
			c.HTML(http.StatusOK, "poc.html", gin.H{
				"uploadfilemsg": err,
			})
		}
		files := strings.Split(file.Filename, ".")
		if files[1] != "yaml" {
			c.HTML(http.StatusOK, "poc.html", gin.H{
				"uploadfilemsg": "上传失败文件格式不正确！",
			})
		} else {
			c.HTML(http.StatusOK, "poc.html", gin.H{"uploadfilemsg": "上传成功！"})
		}
	}
}

// PocDelete 删除POC
func PocDelete(c *gin.Context) {
	pocid := c.PostForm("delete_poc")
	if pocid != "" {
		c.HTML(http.StatusOK, "poc.html", gin.H{"deletepocmsg": "已删除" + pocid + "!"})
	} else {
		c.Request.URL.Path = "/poc"
		c.HTML(http.StatusOK, "poc.html", gin.H{"deletepocmsg": pocid + "删除失败!"})
	}
}

// PocSearch 搜索POC
func PocSearch(c *gin.Context) {
	searchkey := c.PostForm("search_key")
	if searchkey != "" {
		c.HTML(http.StatusOK, "poc.html", nil)
	} else {
		c.HTML(http.StatusOK, "poc.html", nil)
	}
}

// LogsGet 日志管理
func LogsGet(c *gin.Context) {
	var logs []byte
	var err error
	if nodesstate == nil {
		nodesstate = db.GetNodeStates(conn, config.GlobalConfig.NodeConfig.NodeNames)
	}
	if config.Exists("BeeScanLog-Web.log") {
		logs, err = ioutil.ReadFile("BeeScanLog-Web.log")
		if err != nil {
			log.Println(err)
		}
	}

	c.HTML(http.StatusOK, "logs.html", gin.H{"NodeStates": nodesstate, "Logs": string(logs)})
}

// NodeLog 日志管理
func NodeLog(c *gin.Context) {
	nodename := c.Query("nodelog")
	TheNodeLog := db.QueryLogByID(es, nodename)
	c.HTML(http.StatusOK, "logs.html", gin.H{"NodeLog": TheNodeLog, "NodeName": nodename})
}

// InfoInit 本机信息初始化
func InfoInit() {
	var err error
	cpuinfos, err = hostinfo.GetCpuPercent()
	if err != nil {
		log.Println(err)
	}
	meminfos = hostinfo.GetMemInfo()
	netinfos, err = hostinfo.GetNetInfo()
	if err != nil {
		log.Println(err)
	}
	netspeed = hostinfo.GetNetSpeed()

}
