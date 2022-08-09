package banner

import (
	"fmt"
	"github.com/fatih/color"
)

/*
创建人员：云深不知处
创建时间：2022/1/19
程序功能：Banner
*/

const (
	Version = "0.1.3"
)

func Banner() {
	banner := " ____            ____\n" +
		"| __ )  ___  ___/ ___|  ___ __ _ _ __\n" +
		"|  _ \\ / _ \\/ _ \\___ \\ / __/ _` | '_ \\\n" +
		"| |_) |  __/  __/___) | (_| (_| | | | |\n" +
		"|____/ \\___|\\___|____/ \\___\\__,_|_| |_| version:" + Version + "\n\n"
	fmt.Fprintf(color.Output, color.HiCyanString(banner))
}
