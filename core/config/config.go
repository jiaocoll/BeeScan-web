package config

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/16
程序功能：配置模块
*/

var GlobalConfig *Config

type Redis struct {
	Host     string
	Password string
	Port     string
	User     string
	Database string
}

type Elasticsearch struct {
	Host     string
	Password string
	Port     string
	Username string
	Index    string
}

type DBConfig struct {
	Redis         Redis
	Elasticsearch Elasticsearch
}

type NodeConfig struct {
	NodeNames []string
}

type DicConfig struct {
	Dic_user string
	Dic_pwd  string
}

type WorkerConfig struct {
	WorkerNumber int
	Thread       int
}

type UserPassConfig struct {
	UserName string
	PassWord string
}

type NucleiConfig struct {
	Enable     bool
	NucleiPath string
}

type XrayConfig struct {
	Enable   bool
	XrayPath string
}

type Config struct {
	NodeConfig     NodeConfig
	DicConfig      DicConfig
	WorkerConfig   WorkerConfig
	DBConfig       DBConfig
	UserPassConfig UserPassConfig
	NucleiConfig   NucleiConfig
	XrayConfig     XrayConfig
}

// 加载配置
func Setup() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", err)
	}
	// 配置文件
	configFile := path.Join(dir, "config.yaml")

	if !Exists(configFile) {
		WriteYamlConfig(configFile)
	}
	ReadYamlConfig(configFile)
}

func (cfg *Config) Level() zapcore.Level {
	return zapcore.DebugLevel
}

func (cfg *Config) LogPath() string {
	return ""
}

func (cfg *Config) InfoOutput() string {
	return ""
}

func (cfg *Config) ErrorOutput() string {
	return ""
}

func (cfg *Config) DebugOutput() string {
	return ""
}

func ReadYamlConfig(configFile string) {
	// 加载config
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[config_Setup]:fail to read 'config.yaml':", err)
		os.Exit(1)
	}
	err = viper.Unmarshal(&GlobalConfig)
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[config_Setup]:fail to parse 'config.yaml', check format:", err)
		os.Exit(1)
	}

}

func WriteYamlConfig(configFile string) {
	// 生成默认config
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(defaultYamlByte))
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[config_Setup]:fail to read default config bytes:", err)
		os.Exit(1)
	}
	// 写文件
	//err = viper.SafeWriteConfigAs(configFile)
	//if err != nil {
	//	log.Fatalf("[config_Setup]:fail to write 'config.yaml': %v", err)
	//}
	f, err := os.Create("config.yaml")
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", "[config_Setup]:fail to write config yaml", err)
		os.Exit(1)
	}
	_, err = f.Write(defaultYamlByte)
	if err != nil {
		fmt.Fprintln(color.Output, color.HiRedString("[ERRO]"), "["+time.Now().Format("2006-01-02 15:04:05")+"]", err)
		os.Exit(1)
	}
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
