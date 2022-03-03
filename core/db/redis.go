package db

import (
	"Beescan/core/config"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/14
程序功能：redis模块
*/

type NodeState struct {
	Name        string
	Tasks       string
	Running     string
	Finished    string
	ScanPercent string
	LastTime    string
	RunTime     string
	State       string
}

type TaskState struct {
	Name        string
	TargetNum   string
	Tasks       string
	Running     string
	Finished    string
	LastTime    string
	ScanPercent string
}

var state map[string]interface{}

// RedisInit 初始化连接
func RedisInit() *redis.Client {
	addr := config.GlobalConfig.DBConfig.Redis.Host + ":" + config.GlobalConfig.DBConfig.Redis.Port
	db, _ := strconv.Atoi(config.GlobalConfig.DBConfig.Redis.Database)
	conn := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.GlobalConfig.DBConfig.Redis.Password, // no password set
		DB:       db,                                          // use default DB
	})
	Pong, err := conn.Ping().Result()
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[RedisInit]:", err)
		os.Exit(1)
	} else if Pong == "PONG" {
		return conn
	}
	return conn
}

// AddJob 添加任务到消息队列
func AddJob(c *redis.Client, data []byte, NodeName string) error {
	_, err := c.Do("lpush", NodeName, data).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetNodeState 获取节点运行情况
func GetNodeState(c *redis.Client, NodeName string) NodeState {
	var nodestate NodeState
	tasks := c.HGet(NodeName, "tasks").Val()
	running := c.HGet(NodeName, "running").Val()
	finished := c.HGet(NodeName, "finished").Val()
	runtime := c.HGet(NodeName, "runtime").Val()
	nowstate := c.HGet(NodeName, "state").Val()
	lasttime := c.HGet(NodeName, "lasttime").Val()

	nodestate.Name = NodeName
	nodestate.Tasks = tasks
	nodestate.Running = running
	nodestate.Finished = finished
	nodestate.LastTime = lasttime
	nodestate.State = nowstate
	nodestate.RunTime = runtime + "小时"
	tmpfinished, _ := strconv.Atoi(finished)
	tmptasks, _ := strconv.Atoi(tasks)
	if tmpfinished == 0 {
		nodestate.ScanPercent = "0"
	} else {
		nodestate.ScanPercent = fmt.Sprintf("%.2f", (float64(tmpfinished)/float64(tmptasks))*100)
	}
	loc, _ := time.LoadLocation("Local")
	dt, _ := time.ParseInLocation("2006-01-02 15:04:05", nodestate.LastTime, loc)
	nowtime := time.Now()

	if nowtime.Sub(dt).Minutes() > 5 {
		nodestate.State = "Invalid"
	}

	return nodestate
}

// GetNodeStates 得到节点队列
func GetNodeStates(c *redis.Client, nodenames []string) []NodeState {
	var nodestates []NodeState

	for _, v := range nodenames {
		if v != "" {
			nodestates = append(nodestates, GetNodeState(c, v))
		}
	}
	return nodestates

}

// GetTaskState 获取任务运行情况
func GetTaskState(c *redis.Client, TaskName string) TaskState {
	var taskstate TaskState
	tasks := c.HGet(TaskName, "tasks").Val()
	running := c.HGet(TaskName, "running").Val()
	finished := c.HGet(TaskName, "finished").Val()
	lasttime := c.HGet(TaskName, "lasttime").Val()
	targetnum := c.HGet(TaskName, "targetnum").Val()

	taskstate.Name = TaskName
	taskstate.Tasks = tasks
	taskstate.Running = running
	taskstate.Finished = finished
	taskstate.LastTime = lasttime
	taskstate.TargetNum = targetnum

	tmpfinished, _ := strconv.Atoi(finished)
	tmptasks, _ := strconv.Atoi(tasks)

	if tmpfinished == 0 {
		taskstate.ScanPercent = "0"
	} else {
		taskstate.ScanPercent = fmt.Sprintf("%.2f", (float64(tmpfinished)/float64(tmptasks))*100)
	}

	return taskstate
}

// GetTaskStates 得到任务队列
func GetTaskStates(c *redis.Client, tasknames []string) []TaskState {
	var taskstates []TaskState

	for _, v := range tasknames {
		if v != "" {
			taskstates = append(taskstates, GetTaskState(c, v))
		}
	}
	return taskstates
}

// DelTask 删除任务
func DelTask(c *redis.Client, tasknames string) {
	c.Del(tasknames)
}

// 任务注册
func TaskRegister(c *redis.Client, taskname string, tasknum string) {
	state = make(map[string]interface{})
	state["tasks"] = tasknum
	state["running"] = 0
	state["finished"] = 0
	state["targetnum"] = tasknum
	state["lasttime"] = time.Now().Format("2006-01-02 15:04:05")
	c.HMSet(taskname, state)
}
