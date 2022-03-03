package job

import "Beescan/core/runner"

/*
创建人员：云深不知处
创建时间：2022/1/3
程序功能：任务
*/

type Job struct {
	Targets []*runner.Runner
	State   string
}
