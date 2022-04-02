package hostinfo

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"log"
	systemNet "net"
	"strconv"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/15
程序功能：获取本机信息
*/

// 分区
type Part struct {
	Path        string  `json:"path"`
	FsType      string  `json:"fstype"`
	Total       float64 `json:"total"`
	Free        float64 `json:"free"`
	Used        float64 `json:"used"`
	UsedPercent int     `json:"usedPercent"`
}

// 分区集合
type Parts []Part

// CPU
type CpuSingle struct {
	Num     string `json:"num"`
	Percent int    `json:"percent"`
}

type CpuInfo struct {
	CpuAvg float64     `json:"cpuAvg"`
	CpuAll []CpuSingle `json:"cpuAll"`
}

// 内存使用情况
type MemInfo struct {
	Percent float64     `json:"Percent"`
	All     interface{} `json:"All"`
	Used    int         `json:"Used"`
	Fare    int         `json:"Fare"`
}

// 上下行带宽
type SpeedInfo struct {
	Name      string `json:"Name"`
	Send      uint64 `json:"Send"`
	Recv      uint64 `json:"Recv"`
	Upspeed   string `json:"Upspeed"`
	Downspeed string `json:"Downspeed"`
}

const GB = 1024 * 1024 * 1024

func decimal(v string) float64 {
	value, _ := strconv.ParseFloat(v, 64)
	return value
}

// 本地主机IP
func GetLocalIP() (ip string) {
	addresses, err := systemNet.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addresses {
		ipAddr, ok := addr.(*systemNet.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String()
	}
	return ""
}

// 本地主机信息
func GetHostInfo() (result *host.InfoStat, err error) {
	result, err = host.Info()
	return result, err
}

// 本地主机磁盘信息
func GetDiskInfo() (result Parts, err error) {
	parts, err := disk.Partitions(true)
	if err != nil {
		return result, err
	}
	for _, part := range parts {
		diskInfo, err := disk.Usage(part.Mountpoint)
		if err == nil {
			result = append(result, Part{
				Path:        diskInfo.Path,
				FsType:      diskInfo.Fstype,
				Total:       decimal(fmt.Sprintf("%.2f", float64(diskInfo.Total/GB))),
				Free:        decimal(fmt.Sprintf("%.2f", float64(diskInfo.Free/GB))),
				Used:        decimal(fmt.Sprintf("%.2f", float64(diskInfo.Used/GB))),
				UsedPercent: int(diskInfo.UsedPercent),
			})
		} else {
			return result, err
		}
	}
	return result, err
}

// 本地主机CPU使用率
func GetCpuPercent() (result CpuInfo, err error) {
	infos, err := cpu.Percent(1*time.Second, true)
	if err != nil {
		return result, err
	}
	var total float64 = 0
	for index, value := range infos {
		result.CpuAll = append(result.CpuAll, CpuSingle{
			Num:     fmt.Sprintf("#%d", index+1),
			Percent: int(value),
		})
		total += value
	}
	result.CpuAvg = decimal(fmt.Sprintf("%.1f", total/float64(len(infos))))
	return result, err
}

// 本地主机内存信息
func GetMemInfo() MemInfo {
	info, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
	}
	result := MemInfo{
		Percent: decimal(fmt.Sprintf("%.1f", info.UsedPercent)),
		All:     int(info.Total / GB),
		Used:    int(info.Used / GB),
		Fare:    int(info.Total/GB) - int(info.Used/GB),
	}
	return result
}

// 获取本地主机网卡信息
func GetNetInfo() (result []net.IOCountersStat, err error) {
	var infos []net.IOCountersStat
	info, err := net.IOCounters(true)
	if err != nil {
		return result, err
	}
	for _, v := range info {
		if v.BytesSent != 0 || v.BytesRecv != 0 {
			infos = append(infos, v)
		}
	}
	return infos, err
}

// 计算上下行带宽
func GetNetSpeed() (sppedinfos2 []SpeedInfo) {
	var sppedinfos1 []SpeedInfo
	info, err := net.IOCounters(true)
	if err != nil {
		log.Println(err)
	}
	for _, item := range info {
		if item.BytesSent != 0 && item.BytesRecv != 0 {
			speedinfo := SpeedInfo{
				Name: item.Name,
				Send: item.BytesSent,
				Recv: item.BytesRecv,
			}
			sppedinfos1 = append(sppedinfos1, speedinfo)
		}
	}

	time.Sleep(1 * time.Second)

	info, err = net.IOCounters(true)
	if err != nil {
		log.Println(err)
	}
	for _, item := range info {
		if item.BytesSent != 0 && item.BytesRecv != 0 {
			for _, v := range sppedinfos1 {
				if v.Name == item.Name {
					speedinfo := SpeedInfo{
						Name: item.Name,
						Send: item.BytesSent - v.Send,
						Recv: item.BytesRecv - v.Recv,
					}
					if speedinfo.Send >= 1024 {
						speedinfo.Upspeed = strconv.Itoa(int(speedinfo.Send)/1024) + "KB/S"
					} else if speedinfo.Send < 1024 {
						speedinfo.Upspeed = strconv.Itoa(int(speedinfo.Send)) + "B/S"
					}
					if speedinfo.Recv >= 1024 {
						speedinfo.Downspeed = strconv.Itoa(int(speedinfo.Recv)/1024) + "KB/S"
					} else if speedinfo.Recv < 1024 {
						speedinfo.Downspeed = strconv.Itoa(int(speedinfo.Recv)) + "B/S"
					}
					sppedinfos2 = append(sppedinfos2, speedinfo)
				}
			}
		}
	}
	return sppedinfos2
}
