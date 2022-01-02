package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

// 横幅

var Banner = `
===================================
			robin v1.00a
===================================
`

// 随机种子

var Seed = rand.New(rand.NewSource(int64(float64(time.Now().Unix())*math.Pi)))

// 同步

var WG =sync.WaitGroup{}

// 默认值

var defaultConfig = map[string]string{}

// 选项

var options = []string{

	"[*] 请输入需检查的目录(留空则使用默认值) -> ",
	"[*] 请输入日志存储的目录(留空则使用默认值) -> ",
	"[*] 请输入日志文件名称(留空则使用默认值) -> ",
	"[*] 请输入hash文件存储位置(留空则使用默认值) -> ",
	"[*] 请输入备份目录(留空则使用默认值) -> ",
	"[*] 请输入被动检查协程|主动检查协程|替换协程数(留空则使用默认值) -> ",
	"[*] 请输入检查时间间隔(留空则使用默认值) -> ",

}

var history = "[*] 是否加载以前的hash文件(y/N) -> "

var display = "[*] 请输入是否回显(留空则使用默认值|y/N) -> "

// hash文件format

type hashFormat struct {

	BasicConfig map[string]string `json:"basic_config"`
	DirsConfig map[string]map[string]string `json:"dirs_config"`
	FilesConfig map[string]map[string]string `json:"files_config"`

}

// config文件format

type configFormat struct {

	CheckDir string	`json:"check_dir"`
	LogDir string	`json:"log_dir"`
	BackDir string	`json:"back_dir"`
	LogFileName string	`json:"log_file_name"`
	HashFile string	`json:"hash_file"`
	Coroutines string	`json:"coroutines"`
	TimeSec string	`json:"time_sec"`

}

// 日志文件配置

var logFileConfig = map[string]string{}

// 准备参数

func prepare() *Control{

	var optionData = make([]string,len(options))
	var optionDict map[string]string

	// 控制结构
	var control *Control

	// 显示横幅
	fmt.Print(Banner)

	// 读取配置文件
	src,err := os.OpenFile("_config.json",os.O_RDONLY,0)
	defer func() {
		err = src.Close()
		if err != nil{
			fmt.Println("[*] 配置文件无法正常关闭 -> _config.json")
		}
	}()
	if err != nil{
		fmt.Println("[*] 无法加载配置文件 -> _config.json")
	}
	configData := &configFormat{}
	contents,err := io.ReadAll(src)
	if err != nil{
		fmt.Println("[*] 无法读取配置文件内容 -> _config.json")
	}
	err = json.Unmarshal(contents,configData)
	// 映射至字典
	if err != nil{
		fmt.Println("[*] 无法解析配置文件内容 -> _config.json")

		defaultConfig = map[string]string{
			"enable":"false",
		}

	}else{

		defaultConfig = map[string]string{

			"enable":"true",
			"checkDir":configData.CheckDir,
			"logDir":configData.LogDir,
			"backDir":configData.BackDir,
			"logFileName":configData.LogFileName,
			"hashFile":configData.HashFile,
			"coroutines":configData.Coroutines,
			"timeSec":configData.TimeSec,

		}

		fmt.Println("[+] 配置文件加载成功！")

	}

	// 判断是否回显
	if isDisplay(){
		logFileConfig["display"] = "true"
	}

	// 判断是否使用历史hash文件
	if !isUseHistory() {

		// 加入用户输入数据
		for key,each := range options{
			optionData[key] = input(each, '\n')
		}
		optionDict = loadConfig(optionData)
		control = &Control{Config: optionDict,DirSet: make(map[string]*Dirs)}
	}else{

		// 载入hash文件
		control = loadHistory()
		// 若失败则返回错误
		if control == nil {
			log("warning","加载hash文件失败````",nil)
			return prepare()
		}


	}

	return control

}

// 生成日志

func record(control *Control){

	log("run","开始生成日志````",nil)
	if logConfig(control){
		log("ok","完成生成日志````",nil)
	}else{
		log("warning","未完成生成文件````",nil)
	}

}

// 遍历文件

func load(control *Control){
	log("run","开始遍历文件````",nil)
	if loadFiles(control) {
		log("ok","完成遍历文件````",nil)
	}else{
		log("warning","未完成遍历文件````",nil)
	}
}

// 备份文件

func backup(control *Control){
	log("run","开始备份文件````",nil)
	if backupFiles(control){
		log("ok","完成备份文件````",nil)
	}else{
		log("warning","未完成备份文件````",nil)
	}
}

// 计算hash

func hash(control *Control){
	log("run","开始计算文件hash````",nil)
	if hashFiles(control) {
		log("ok","完成文件hash计算````",nil)
	}else{
		log("warning","未完成文件hash计算````",nil)
	}
}

// 输出hash文件

func out(control *Control){
	log("run","开始输出hash文件````",nil)
	if outputHashFile(control){
		log("ok","完成输出hash文件````",nil)
	}else{
		log("warning","未完成输出hash文件````",nil)
	}
}

// 运行

func Run(){

	var control = prepare()
	// 生成日志
	record(control)
	// 遍历文件
	load(control)
	// 备份文件
	backup(control)
	// 计算文件hash
	hash(control)
	// 输出hash文件
	out(control)
	// 开始
	log("run","监控开始````",nil)
	checkCore(control)
	// 结束
	log("end","监控结束````",nil)

}