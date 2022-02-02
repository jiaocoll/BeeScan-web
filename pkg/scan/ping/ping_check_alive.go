package ping

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

/*
创建人员：云深不知处
创建时间：2022/1/1
程序功能：ping扫描
*/

func PingCheckAlive(host string)bool{
	sysType := runtime.GOOS
	if sysType == "windows"{
		cmd := exec.Command("ping", "-n", "2", host)
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Run()
		if strings.Contains(output.String(), "TTL=") && strings.Contains(output.String(), host)  {
			return true
		}
	}else if (sysType == "linux" || sysType=="darwin") {
		cmd := exec.Command("ping", "-c", "2", host)
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Run()
		if strings.Contains(output.String(), "ttl=") && strings.Contains(output.String(), host) {
			return true
		}
	}
	return false
}
