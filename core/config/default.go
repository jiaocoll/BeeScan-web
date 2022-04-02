package config

/*
创建人员：云深不知处
创建时间：2022/1/16
程序功能：默认配置
*/

var defaultYamlByte = []byte(`
# 节点配置
NodeConfig:
  # 节点名称配置
  NodeNames:
    BeeScan_node_1,
    BeeScan_node_2,
    BeeScan_node_3

# 数据库配置
DBConfig:
  # Redis配置
  redis:
    host: "127.0.0.1"
    password: ""
    port: "6379"
    database: ""
  # Elasticsearch配置
  Elasticsearch:
    host: "127.0.0.1"
    port: "9200"
    username: ""
    password: ""
    index: "beescan"

# web端账号密码配置
UserPassConfig:
  UserName: Ameng
  PassWord: Ameng

# nuclei路径
NucleiPath: /Users/ameng/Tools/nuclei/nuclei

`)
