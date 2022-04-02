package scan

import (
	"Beescan/core/config"
	"Beescan/core/util"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/3/29
程序功能：漏洞扫描
*/

type NucleiOutput struct {
	ID            string      `json:"id"`
	TaskName      string      `json:"task-name"`
	Template      string      `json:"template"`
	TemplateURL   string      `json:"template-url"`
	TemplateID    string      `json:"template-id"`
	Info          Info        `json:"info"`
	MatcherName   string      `json:"matcher-name"`
	Type          string      `json:"type"`
	Host          string      `json:"host"`
	MatchedAt     string      `json:"matched-at"`
	IP            string      `json:"ip"`
	Timestamp     time.Time   `json:"timestamp"`
	CurlCommand   string      `json:"curl-command"`
	MatcherStatus bool        `json:"matcher-status"`
	MatchedLine   interface{} `json:"matched-line"`
}
type Info struct {
	Name        string      `json:"name"`
	Author      []string    `json:"author"`
	Tags        []string    `json:"tags"`
	Description string      `json:"description"`
	Reference   interface{} `json:"reference"`
	Severity    string      `json:"severity"`
}

func VulnScan(targets []string) {
	// Print Go Version
	var args []string
	dirpath := util.GetCurrentDirectory()
	filepath := dirpath + "/vulns.txt"
	args = append(args, "-u")
	args = append(args, targets...)
	args = append(args, "-severity")
	args = append(args, "critical,medium,high")
	args = append(args, "-t")
	//args = append(args, "/Users/ameng/nuclei-templates/dns/txt-fingerprint.yaml")
	args = append(args, "-json")
	args = append(args, "-o")
	args = append(args, filepath)
	cmd := exec.Command(config.GlobalConfig.NucleiPath, args...)
	err := cmd.Run()
	if err != nil {
		fmt.Print("VulnScan", err)
	}

}

func UnMarshal(line string) NucleiOutput {
	var nucleioutput NucleiOutput
	if line != "" {
		err := json.Unmarshal([]byte(line), &nucleioutput)
		if err != nil {
			log.Println("UnMarshal", err)
		}
	}
	return nucleioutput
}
