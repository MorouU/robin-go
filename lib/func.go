package lib

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)


// 判断是否为 windows
var isWindows = runtime.GOOS == "windows"

// 输入

func input(tag string, end byte) string{
	fmt.Print(tag)
	in := bufio.NewReader(os.Stdin)
	st, err := in.ReadString(end)
	if err != nil{
		return ""
	}
	return strings.TrimRight(st,string(end))
}

// 日志处理
func log(tag, contents string,err interface{}){

	// 日志标识符
	var x string

	switch tag {
		case "failed":
			x = "-"
		case "ok":
			x = "+"
		default:
			x ="*"
	}

	// 日志基础行
	var line = fmt.Sprintf("<%s>[%s] %s {%v}\n",tag,x,contents,err)

	// 回显
	if logFileConfig["display"] == "true"{
		fmt.Print(line)
	}

	// 记录到文件
	if logFileConfig["enable"] == "true"{
		src,err := os.OpenFile(logFileConfig["file"],os.O_APPEND|os.O_CREATE,600)
		// 如果无法记录到文件则忽略
		defer func() {
			_ = src.Close()
		}()
		if err == nil{
			// 写入日志文件
			_,_ = src.WriteString(logPrefix() + line)
			// 写入结尾
			if tag == "end"{
				_,_ = src.WriteString(logEnd())
			}
		}
	}

	// 判断日志大失败等级
	if tag == "failed"{
		os.Exit(1)
	}

}

// 路径处理
func format(filePath string) string{
	if isWindows{
		return strings.ReplaceAll(filePath,"\\","/")
	}
	return filePath
}

// 路径是否存在

func isExist(thisPath string) bool{
	_, err := os.Stat(thisPath)
	if err != nil{
		if os.IsExist(err){
			return true
		}
		if os.IsNotExist(err){
			return false
		}
		return false
	}
	return true
}

// 是否使用历史hash文件

func isUseHistory() bool{
	return strings.ToLower(input(history,'\n')) == "y"
}

// 是否回显

func isDisplay() bool{
	return strings.ToLower(input(display,'\n')) == "y"
}

// 遍历目录

func readDir(dirPath string) map[string]os.FileInfo {

	var files = make(map[string]os.FileInfo)
	var filePath string

	info, _ := ioutil.ReadDir(dirPath)
	for _ ,each := range info {
		// 获取文件路径
		filePath = format(path.Join(dirPath,each.Name()))
		// 记录文件信息
		files[filePath] = each
		// 若为目录则递归
		if each.IsDir() {
			for key, value := range readDir(filePath){
				files[key] = value
			}
		}
	}
	return files
}

// 加载输入配置

func loadConfig(config []string)map[string]string{

	var reConfig = make(map[string]string,len(config))

	// 载入默认值
	for key, value := range defaultConfig{
		reConfig[key] = value
	}

	// 检查目录是否为空
	if config[0] == "" {
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 需检查位置目录为空！使用默认值 -> " + defaultConfig["checkDir"],nil)
	}else{
		reConfig["checkDir"] = config[0]
	}
	realPath,err := filepath.Abs(reConfig["checkDir"])
	if err != nil{
		log("warning","-> 需检查位置目录无法转换为绝对路径！使用原先值 -> " + reConfig["checkDir"],nil)
	}else {
		reConfig["checkDir"] = realPath
	}
	if !isExist(reConfig["checkDir"]){
		log("failed","需检查目录不存 -> " + reConfig["checkDir"],"")
	}
	reConfig["checkDir"] = format(reConfig["checkDir"])

	// 检查日志目录是否为空
	if config[0] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 存储日志位置目录为空！使用默认值 -> " + defaultConfig["logDir"],nil)
	}else{
		reConfig["logDir"] = config[1]
	}
	realPath,err = filepath.Abs(reConfig["logDir"])
	if err != nil{
		log("warning","-> 存储日志位置目录无法转换为绝对路径！使用原先值 -> " + reConfig["logDir"],nil)
	}else {
		reConfig["logDir"] = realPath
	}
	reConfig["logDir"] = format(reConfig["logDir"])

	// 检查日志文件名称参数
	if config[2] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 日志名称为空！使用默认值 -> " + defaultConfig["logFileName"],nil)
	}else{
		reConfig["logFileName"] = config[2]
	}

	// 检查hash文件位置参数
	if config[3] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> hash文件位置为空！使用默认值 -> " + defaultConfig["hashFile"],nil)
	}else{
		reConfig["hashFile"] = config[3]
	}
	realPath,err = filepath.Abs(reConfig["hashFile"])
	if err != nil{
		log("warning","-> hash文件位置无法转换为绝对路径！使用原先值 -> " + reConfig["hashFile"],nil)
	}else {
		reConfig["hashFile"] = realPath
	}
	reConfig["hashFile"] = format(reConfig["hashFile"])

	// 检查备份目录参数
	if config[4] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 备份位置目录为空！使用默认值 -> " + defaultConfig["backDir"],nil)
	}else{
		reConfig["backDir"] = config[4]
	}
	realPath,err = filepath.Abs(reConfig["backDir"])
	if err != nil{
		log("warning","-> 备份目录无法转换为绝对路径！使用原先值 -> " + reConfig["backDir"],nil)
	}else {
		reConfig["backDir"] = realPath
	}
	reConfig["backDir"] = format(reConfig["backDir"])

	// 主动检查协程|被动检查协程|替换协程数
	if config[5] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 主动检查协程|被动检查协程|替换协程数为空！使用默认值 -> " + defaultConfig["coroutines"],nil)
	}else{
		// 检查格式
		c := strings.Split(config[5],"|")
		if len(c) != 3{
			log("failed","-> 主动检查协程|被动检查协程|替换协程数格式有误！",nil)
		}
		for _,num := range c{
			if _,err = strconv.Atoi(num); err != nil{
				log("failed","-> 主动检查协程|被动检查协程|替换协程数格式有误！",nil)
			}
		}
		reConfig["coroutines"] = config[5]
	}

	// 检查时间间隔
	if config[6] == ""{
		if defaultConfig["enable"] != "true"{
			log("failed","配置文件加载失败，无法使用默认值！",nil)
		}
		log("warning","-> 检查时间间隔为空！使用默认值 -> " + defaultConfig["timeSec"],nil)
	}else{
		reConfig["timeSec"] = config[6]
	}

	return reConfig

}

// 加载历史hash文件

func loadHistory() *Control{

	// 获取用户输入的hash文件
	c := input("[*] 请输入历史hash路径(包括文件名,留空则使用默认值) -> ",'\n')

	// 若输入为空或不存在则使用默认值
	if c == "" || !isExist(c){
		log("warning","-> hash文件位置不存在或为空！使用默认值 -> " + defaultConfig["hashFile"],nil)
		c = defaultConfig["hashFile"]
	}

	// 若都不存在则回到开始菜单
	if !isExist(c){
		log("warning","hash文件位置有误，找不到对应文件 -> " + c,nil)
		return nil
	}

	// 读hash文件
	src,err := os.OpenFile(c,os.O_RDONLY,0)
	if err != nil{
		log("warning","无法读取hash文件 -> " + c,nil)
		return nil
	}
	defer func(){
		err = src.Close()
		if err != nil{
			log("warning","文件未能正常关闭 -> " + c,nil)
		}
	}()

	// 解析Gob
	var getGob = &hashFormat{}
	err = gob.NewDecoder(src).Decode(getGob)
	if err != nil{
		log("warning","无法读取/解析hash文件内容 -> " + c,nil)
		return nil
	}
	var control = &Control{Config: make(map[string]string),DirSet: make(map[string]*Dirs)}

	// 解析控制结构
	for configName, configValue := range getGob.BasicConfig{
		control.Config[configName] = configValue
	}
	// 解析目录结构
	for dirName, dirEach := range getGob.DirsConfig{
		dir := &Dirs{Data: make(map[string]string),FileSet: make(map[string]*Files)}
		for dirConfigName, dirConfigValue := range dirEach{
			dir.SetData(dirConfigName,dirConfigValue)
		}
		control.AddDirs(dirName,dir)
	}
	// 解析文件结构
	for fileName, fileEach := range getGob.FilesConfig{
		file := &Files{Data: make(map[string]string)}
		for fileConfigName, fileConfigValue := range fileEach{
			if fileConfigName != "fromDir" {
				file.SetData(fileConfigName, fileConfigValue)
			}
		}
		control.DirSet[fileEach["fromDir"]].addFiles(fileName,file)
	}


	return control


}

// 日志开头

func logPrefix()string{
	return fmt.Sprintf("[%s]>>>>> ",time.Now().Format("2006-01-02 15:04:05"))
}

// 日志结尾

func logEnd()string{
	return fmt.Sprintf("[%s]<<<<< \n",time.Now().Format("2006-01-02 15:04:05"))
}

// 生成日志文件

func logConfig(control *Control)bool{

	// 检查日志目录
	logDir := control.Config["logDir"]
	if !isExist(logDir){
		err := os.MkdirAll(logDir,500)
		if err != nil{
			log("warning","无法创建日志文件目录 -> " + logDir,nil)
			logFileConfig["enable"] = "false"
			return false
		}
	}

	// 创建日志文件
	logFile := format(filepath.Join(logDir,control.Config["logFileName"]))
	if isExist(logFile){
		log("warning","日志文件已存在 -> ",logFile)
	}
	src,err := os.OpenFile(logFile,os.O_APPEND|os.O_CREATE,0600)
	defer func() {
		err = src.Close()
		if err != nil{
			log("warning","此日志文件无法正常关闭 -> ",logFile)
		}
	}()
	if err != nil{
		log("warning","无法打开日志文件 -> ",logFile)
		logFileConfig["enable"] = "false"
		return false
	}

	// 写入开头
	_,err = src.WriteString(logPrefix() + "日志初始化...\n")
	if err != nil{
		log("warning","无法写入日志文件 -> ",logFile)
		logFileConfig["enable"] = "false"
	}

	// 配置完成
	logFileConfig["enable"] = "true"
	logFileConfig["file"] = logFile

	return true
}

// 加载文件

func loadFiles(control *Control) bool{

	// 随机salt
	getSalt := func(length int)string{
		var randBytes = make([]byte,length)
		for i:=0;i<length;i++{
			randBytes[i] = uint8(Seed.Intn(127))
		}
		return string(randBytes)
	}

	// 列目录/文件
	realPath := control.Config["checkDir"]
	err := filepath.Walk(realPath,func(path string, info os.FileInfo, err error) error{
		// 格式化
		path = format(path)
		// 记录文件权限
		fileMode := strconv.Itoa(int(info.Mode().Perm()))
		// 记录文件最后修改时间
		fileMTime := info.ModTime().Format("2006-01-02 15:04:05")
		// 若为 Linux 记录文件所有者/所有组
		var fileUid = ""
		var fileGid = ""
		if !isWindows{
		// 若编译为windows把这两行注释了
			fileUid = strconv.Itoa(int(info.Sys().(*syscall.Stat_t).Uid))
			fileGid = strconv.Itoa(int(info.Sys().(*syscall.Stat_t).Gid))
		}
		// 如果为目录
		if info.IsDir(){
			// 如果目录未记录
			if _,ok := control.DirSet[path]; !ok{

				dir := &Dirs{Data: map[string]string{
					"mtime":fileMTime,
					"mode":fileMode,
					"uid":fileUid,
					"gid":fileGid,
				},
				FileSet: make(map[string]*Files),
				}
				control.AddDirs(path,dir)
			}
			return nil
		}
		// 如果为文件
		// 如果文件的所属目录未被记录
		dirPath := format(filepath.Dir(path))
		if _,ok := control.DirSet[dirPath]; !ok{
			dir := &Dirs{Data: map[string]string{
				"mtime":fileMTime,
				"mode":fileMode,
				"uid":fileUid,
				"gid":fileGid,
			},
			FileSet: make(map[string]*Files),
			}
			control.AddDirs(dirPath,dir)
		}
		// 如果文件未记录
		if _,ok := control.DirSet[dirPath].FileSet[path]; !ok{
			file := &Files{Data: map[string]string{
				"mtime":fileMTime,
				"mode":fileMode,
				"uid":fileUid,
				"gid":fileGid,
				"salt":getSalt(16),
			}}
			control.DirSet[dirPath].addFiles(path,file)
		}
		return nil
	})


	if err != nil{
		log("failed","列出需检查目录失败 -> ",err)
	}

	return true

}

// 备份文件
func backupFiles(control *Control)bool{

	// 备份目录绝对路径
	backDir := control.Config["backDir"]

	// 如果备份目录不存在则尝试新建
	if !isExist(control.Config["backDir"]){
		err := os.MkdirAll(backDir,0500)
		if err != nil{
			log("failed","尝试创建备份目录失败 -> " + backDir,err)
		}
	}

	// 复制备份
	copyFiles := func(backDir, fileName string) string{

		// 备份文件绝对路径
		backFile := format(path.Join(backDir,filepath.Base(fileName)))
		isBackup := false
		// 如果已存在备份文件
		if isExist(backFile){
			log("warning","已存在备份 -> " + backFile,nil)
			isBackup = true
		}
		// 如果备份文件所在备份目录不存在则新建
		if !isExist(backDir){
			err := os.MkdirAll(backDir,0500)
			if err != nil{
				log("failed","尝试创建备份目录失败 -> " + backDir,err)
			}
		}

		// 读取文件内容
		var src *os.File
		src, err := os.OpenFile(fileName, os.O_RDONLY, 0)
		if err != nil{
			log("warning","(备份)无法读取文件 -> " + fileName,nil)
			if isBackup{
				return backFile
			}
			return ""
		}
		defer func(){
			err = src.Close()
			if err !=  nil{
				log("warning","(备份)无法正常关闭文件 -> " + fileName,nil)
			}
		}()


		//将内容写入文件
		var dst *os.File
		dst, err = os.OpenFile(backFile,os.O_WRONLY|os.O_CREATE,0400 )
		if err != nil{
			log("warning","(备份)无法写入文件 -> " + backFile,nil)
			if isBackup{
				return backFile
			}
			return ""
		}
		defer func(){
			err = dst.Close()
			if err !=  nil{
				log("warning","(备份)无法正常关闭文件 -> " + backFile,nil)
			}
		}()


		// 复制
		_,err = io.Copy(dst,src)
		if err != nil{
			log("warning","(备份)无法复制文件 -> " + fileName + ">" + backFile,nil)
			if isBackup{
				return backFile
			}
			return ""
		}

		// 重置权限
		err = os.Chmod(backFile, 0400)
		if err != nil{
			log("warning","(备份)无法设置备份文件权限 -> " + backFile,nil)
		}

		return backFile
	}

	// 获取需检查目录的绝对路径
	realPath := control.Config["checkDir"]

	// 遍历目录集
	for dirPath, dirEach := range control.DirSet{

		// 相对于备份目录的绝对目录路径
		backRealDir := format(filepath.Join(backDir,strings.ReplaceAll(dirPath,realPath,"")))

		// 设置相对于备份目录的绝对目录路径
		dirEach.SetData("backPath",backRealDir)

		// 遍历目录集中的目录中的文件集
		for fileName, fileEach := range dirEach.FileSet{

			// 设置相对于备份目录的绝对文件路径
			fileEach.SetData("backPath",copyFiles(backRealDir,fileName))

		}
	}
	return true
}

// 计算hash
func hashFiles(control *Control)bool{

	// 遍历需检查目录的所有文件

	sha256File := func(fileName,salt string)string{

		src,err := os.OpenFile(fileName,os.O_RDONLY,0)
		if err != nil{
			log("warning","无法计算文件hash -> " + fileName,nil)
			return ""
		}
		defer func() {
			err = src.Close()
			if err != nil{
				log("warning","无法正常关闭文件 -> " + fileName,nil)
			}
		}()
		toHash := sha256.New()

		_, err = io.Copy(toHash,src)
		if err != nil{
			log("warning","无法计算文件hash -> " + fileName,nil)
			return ""
		}

		toHash.Write([]byte(salt))

		return fmt.Sprintf("%x",toHash.Sum(nil))

	}

	for _, dirEach := range control.DirSet{
		for fileName, fileEach := range dirEach.FileSet{

			fileEach.SetData("hash",sha256File(fileName,fileEach.Data["salt"]))

		}
	}
	return true

}

// 输出hash文件
func outputHashFile(control * Control)bool{

	var reGob = hashFormat{
		BasicConfig: make(map[string]string),
		DirsConfig: make(map[string]map[string]string),
		FilesConfig: make(map[string]map[string]string),
	}

	// 控制结构
	for configName, configValue := range control.Config{
		reGob.BasicConfig[configName] = configValue
	}

	// 目录结构
	for dirName, dirEach := range control.DirSet {

		reGob.DirsConfig[dirName] = make(map[string]string)
		for dirConfigName, dirConfigValue := range dirEach.Data {
			reGob.DirsConfig[dirName][dirConfigName] = dirConfigValue
		}

		// 文件结构
		for fileName, fileEach := range dirEach.FileSet{
			reGob.FilesConfig[fileName] = make(map[string]string)
			for fileConfigName, fileConfigValue := range fileEach.Data{
				reGob.FilesConfig[fileName][fileConfigName] = fileConfigValue
			}
			reGob.FilesConfig[fileName]["fromDir"] = dirName
		}

	}



	// 输出的hash文件
	var hashFile = control.Config["hashFile"]

	if isExist(hashFile){
		log("warning","hash文件已存在 -> " + hashFile ,nil)
	}

	// 写入hash文件
	dst,err := os.OpenFile(hashFile, os.O_RDWR|os.O_CREATE,0600)
	if err != nil{
		log("warning","无法写入hash文件 -> " + hashFile ,nil)
	}
	defer func(){
		err = dst.Close()
		if err != nil{
			log("warning","无法正常关闭文件 -> " + hashFile ,nil)
		}

	}()

	err = gob.NewEncoder(dst).Encode(reGob)

	if err != nil{
		log("warning","包装/写入hash文件数据失败" ,nil)
		return false
	}

	return true
}

// 实时检查
func checkCore(control *Control){

	// 设置差异信道
	var evilData = make(chan map[string]string)
	defer close(evilData)

	// 获取判断间隔
	timeSec,err := strconv.ParseFloat(control.Config["timeSec"],64)
	if err != nil{
		log("warning","无法获取判断间隔秒数 -> 使用默认值 0.5 " ,nil)
		 timeSec = .5
	}
	timeDuration := time.Duration(int64(float64(time.Second) * timeSec))

	// 获取检查目录绝对路径
	realPath := control.Config["checkDir"]

	// 检查hash
	checkHashSha256 := func(fileName, salt, correctHash string) bool{

		// 若正确hash为空则跳过检查
		if correctHash == ""{
			return true
		}

		// 获取文件读流
		var src *os.File
		src,err = os.OpenFile(fileName,os.O_RDONLY,0)
		defer func() {
			err = src.Close()
			if err != nil {
				log("warning", "该文件在计算hash时无法正常关闭 -> "+fileName, nil)
			}
		}()
		if err  != nil{
			log("warning","无法读取此文件内容 -> " + fileName ,nil)
		}

		// 计算hash
		toHash := sha256.New()

		_, err = io.Copy(toHash,src)
		if err != nil{
			log("warning","无法计算此文件hash -> " + fileName,nil)
			return false
		}

		toHash.Write([]byte(salt))


		return fmt.Sprintf("%x",toHash.Sum(nil)) == correctHash

	}

	// 获取当前时区的时间戳
	getTimeNow := func(timeStr string)time.Time{
		// 获取当前时区时间，若无法获取到则使用默认当前时间戳
		var t time.Time
		t,err = time.ParseInLocation("2006-01-02 15:04:05",timeStr,time.Local)
		if err != nil{
			return time.Now()
		}
		return t
	}

	// 文件替换
	fileReplace := func(fromFileName, toFileName string,fileTime time.Time, uid, gid int, mode os.FileMode) bool {

		// 若备份文件为空则跳过
		if fromFileName == "" {
			return true
		}

		// 读取
		var src *os.File
		src, err = os.OpenFile(fromFileName, os.O_RDONLY, 0)
		defer func() {
			err = src.Close()
			if err != nil {
				log("warning", "备份文件无法正常关闭 -> "+fromFileName+" > "+toFileName, nil)
			}
		}()
		if err != nil {
			log("warning", "无法读取备份文件 -> "+fromFileName+" > "+toFileName, nil)
			return false
		}
		// 删除
		if isExist(toFileName) {
			err = os.Remove(toFileName)
			if err != nil {
				log("warning", "无法替换备份文件 -> "+fromFileName+" > "+toFileName, nil)
			}
		}
		// 写入
		var dst *os.File
		dst,err = os.OpenFile(toFileName,os.O_WRONLY|os.O_CREATE,0777)
		defer func() {
			err = dst.Close()
			if err != nil{
				log("warning","写入文件无法正常关闭 -> " + fromFileName + " > " + toFileName,nil)
			}
		}()
		if err != nil{
			log("warning","无法重新写入文件 -> " + fromFileName + " > " + toFileName,nil)
			return false
		}

		// 替换内容
		_,err = io.Copy(dst,src)
		if err != nil{
			log("warning","无法从备份文件替换文件 -> " + fromFileName + " > " + toFileName,nil)
			return false
		}

		// 重置最后修改时间
		err = os.Chtimes(toFileName,fileTime,fileTime)
		if err != nil{
			log("warning","无法重置修改时间 -> " + fromFileName + " > " + toFileName,nil)
		}

		// 若为 Linux 重置所有者
		if !isWindows{
			err = os.Chown(toFileName,uid,gid)
			if err != nil{
				log("warning","无法重置所有者 -> " + fromFileName + " > " + toFileName,nil)
			}
		}

		// 重置权限
		err = os.Chmod(toFileName,mode)
		if err != nil{
			log("warning","无法重置权限 -> " + fromFileName + " > " + toFileName,nil)
		}

		return true

	}

	// 目录替换
	dirReplace := func(dirName string, dirTime time.Time, dirUid, dirGid int, dirMode os.FileMode)bool{

		// 获取目录权限
		err = os.MkdirAll(dirName, dirMode)
		if err != nil{
			log("warning", "无法从备份创建目录 -> "+dirName, nil)
			return false
		}

		// 若能成功创建目录
		// 若为 Linux 则更改目录所有者/目录所有组
		if !isWindows{
			err = os.Chown(dirName,dirUid,dirGid)
			if err != nil{
				log("warning","无法重置此目录所有者 -> " + dirName ,nil)
			}
		}

		// 重置最后修改时间
		err = os.Chtimes(dirName,dirTime,dirTime)
		if err != nil{
			log("warning","无法重置修改时间 -> " + dirName,nil)
		}

		return true

	}

	// 遍历所有文件检查(被动)
	check := func() {
		// 文件记录
		var fileRecord = make(map[string]os.FileInfo)
		// 主循环
		for{
			// 遍历文件
			fileRecord = readDir(realPath)

			for filePath, info := range fileRecord{
				// 忽略条件竞争错误
				if info == nil{
					continue
				}
				fileMode := strconv.Itoa(int(info.Mode().Perm()))
				fileMTime := info.ModTime().Format("2006-01-02 15:04:05")
				// 如果为目录
				if info.IsDir(){
					// 如果目录不存在于记录中
					if _,ok := control.DirSet[filePath]; !ok{
						log("check","此目录不存在于记录中 -> "+filePath,nil)
						// 尝试递归删除整个目录
						err = os.RemoveAll(filePath)
						continue
					}

					// 检查目录
					var dirSelf = control.DirSet[filePath]
					// 若在 Linux 中，检查所有者/所有组
					if !isWindows{

						// 若所有者记录是无效的
						if dirSelf.Data["uid"] == "" || dirSelf.Data["gid"] == ""{

							log("warning","此目录所有者/所有组记录无效 -> " + filePath ,nil)

						}else {
							// 若所有者记录是有效的
							var dirUid,dirGid = 0,0

							// 获取UID
							dirUid,err = strconv.Atoi(dirSelf.Data["uid"])
							if err != nil{
								log("warning","此目录无法获取UID -> " + filePath ,nil)
								dirUid = 0
							}
							// 获取GID记录
							dirGid,err = strconv.Atoi(dirSelf.Data["gid"])
							if err != nil{
								log("warning","此目录无法获取GID -> " + filePath ,nil)
								dirGid = 0
							}

							// 若不相同则进行替换，若相同则不进行操作
							// 若编译为windows把这两行注释了
							var thisDirUid = int(info.Sys().(*syscall.Stat_t).Uid)
							var thisDirGid = int(info.Sys().(*syscall.Stat_t).Gid)
							// ==============================
							if dirUid != thisDirUid || dirGid != thisDirGid {
								log("check","此目录所有者/所有组不匹配 -> "+filePath,nil)
								// 判断转换UID
								if dirUid != thisDirUid {
									thisDirUid = dirUid
								}
								// 判断转换GID
								if dirGid != thisDirGid {
									thisDirGid = dirGid
								}
							}
								err = os.Chown(filePath,thisDirUid,thisDirGid)
								if err != nil{
									log("warning","无法重置此目录所有者 -> " + filePath ,nil)
								}
							}
						}

					// 若检查权限不通过则修改权限
					if fileMode != dirSelf.Data["mode"]{
						log("check","此目录权限不匹配 -> "+filePath,nil)
						// 获取目录权限记录
						var mode = 0
						mode,err = strconv.Atoi(dirSelf.Data["mode"])
						if err != nil{
							log("warning","此目录权限记录无效 -> " + filePath ,nil)
						}
						err = os.Chmod(filePath,os.FileMode(mode))
						if err != nil{
							log("warning","无法重置此目录权限 -> " + filePath ,nil)
						}

					}

					// 若检查最后修改时间不通过则更改时间
					if fileMTime != dirSelf.Data["mtime"]{
						log("check","此目录最后修改时间不匹配 -> "+filePath,nil)
						err = os.Chtimes(filePath,getTimeNow(dirSelf.Data["mtime"]),getTimeNow(dirSelf.Data["mtime"]))
						if err != nil{
							log("warning","无法重置此目录时间戳 -> " + filePath ,nil)
						}
					}

					continue

				}

				// 如果为文件
				// 如果文件目录不存在于记录中
				dirPath := format(filepath.Dir(filePath))
				if _,ok := control.DirSet[dirPath]; !ok{
					log("check","此目录不存在于记录中 -> "+dirPath,nil)
					// 尝试递归删除整个目录
					err = os.RemoveAll(filePath)
					continue
				}
				// 如果文件不存在于记录中
				if _,ok := control.DirSet[dirPath].FileSet[filePath]; !ok{
					log("check","此文件不存在于记录中 -> "+filePath,nil)
					// 尝试删除文件
					err = os.Remove(filePath)
					continue
				}

				// 如果文件存在记录中，则检查所有者/所有组|检查权限|检查hash|检查最后修改时间
				var fileSelf = control.DirSet[dirPath].FileSet[filePath]
				var dirSelf = control.DirSet[dirPath]
				// 若在 Linux 中，检查所有者/所有组
				if !isWindows{

					// 若所有者记录是无效的
					if fileSelf.Data["uid"] == "" || fileSelf.Data["gid"] == ""{

						log("warning","此文件所有者/所有组记录无效 -> " + filePath ,nil)

					}else {
						// 若所有者记录是有效的
						var fileUid,fileGid = 0,0

						// 获取UID
						fileUid,err = strconv.Atoi(fileSelf.Data["uid"])
						if err != nil{
							log("warning","此文件无法获取UID -> " + filePath ,nil)
							fileUid = 0
						}
						// 获取GID记录
						fileGid,err = strconv.Atoi(fileSelf.Data["gid"])
						if err != nil{
							log("warning","此文件无法获取GID -> " + filePath ,nil)
							fileGid = 0
						}

						// 若不相同则进行替换，若相同则不进行操作
						// 若编译为windows把这两行注释了
						var thisFileUid =  int(info.Sys().(*syscall.Stat_t).Uid)
						var thisFileGid = int(info.Sys().(*syscall.Stat_t).Gid)
						// ==============================
						if fileUid != thisFileUid || fileGid != thisFileGid{

							log("check","此文件所有者/所有组不匹配 -> "+filePath,nil)

							// 判断转换UID
							if fileUid != thisFileUid{
								thisFileUid = fileUid
							}
							// 判断转换GID
							if fileGid != thisFileGid{
								thisFileGid = fileGid
							}
							err = os.Chown(filePath,thisFileUid,thisFileGid)
							if err != nil{
								log("warning","无法重置此文件所有者 -> " + filePath ,nil)
							}
						}
					}
				}

				// 若检查权限不通过则更改权限
				if fileMode != fileSelf.Data["mode"]{

					log("check","此文件权限不匹配 -> "+filePath,nil)

					// 获取文件权限记录
					var mode = 0
					mode,err = strconv.Atoi(fileSelf.Data["mode"])
					if err != nil{
						log("warning","此文件权限记录无效 -> " + filePath ,nil)
					}
					err = os.Chmod(filePath,os.FileMode(mode))
					if err != nil{
						log("warning","无法重置此文件权限 -> " + filePath ,nil)
					}

				}

				// 若检查hash不通过则替换文件
				if !checkHashSha256(filePath,fileSelf.Data["salt"],fileSelf.Data["hash"]) {

					log("check","此文件hash不匹配 -> "+filePath,nil)

					evilData <- map[string]string{
						"fileName":filePath,
						"backPath":fileSelf.Data["backPath"],
						"mtime":fileSelf.Data["mtime"],
						"mode":fileSelf.Data["mode"],
						"uid":fileSelf.Data["uid"],
						"gid":fileSelf.Data["gid"],
						"dirName":dirPath,
						"dirMode": dirSelf.Data["mode"],
						"dirMTime":dirSelf.Data["mtime"],
						"dirUid":dirSelf.Data["uid"],
						"dirGid":dirSelf.Data["gid"],
					}
				}


				// 若检查最后修改时间不通过则更改时间
				if fileMTime != fileSelf.Data["mtime"]{

					log("check","此文件最后修改时间不匹配 -> "+filePath,nil)

					err = os.Chtimes(filePath,getTimeNow(fileSelf.Data["mtime"]),getTimeNow(fileSelf.Data["mtime"]))
					if err != nil{
						log("warning","无法重置此文件时间戳 -> " + filePath ,nil)
					}
				}

				continue
			}

			time.Sleep(timeDuration)
		}

		WG.Done()
	}

	// 遍历备份数据检查(主动)
	shot := func() {

		// 对文件
		var mode int
		var fileTime time.Time
		var uid,gid = 0, 0
		var backFileName string

		// 对目录
		var dirMode int
		var dirTime time.Time
		var dirUid, dirGid = 0, 0


		// 主循环
		for{

			for dirName, dirEach := range control.DirSet{

			checkDirLoop:
				// 若目录不存在
				if !isExist(dirName){
					log("shot","尝试恢复此目录 -> "+dirName,nil)
					// 获取目录权限
					dirMode,err  = strconv.Atoi(dirEach.Data["mode"])
					if err != nil {
						log("warning", "无法获取此目录备份记录的权限 -> "+dirName, nil)
						dirMode = 0777
					}
					// 若为 Linux 则更改目录所有者/目录所有组
					if !isWindows{
						// 若所有者记录是无效的
						if dirEach.Data["uid"] == "" || dirEach.Data["gid"] == ""{
							log("warning","此目录所有者/所有组记录无效 -> " + dirName ,nil)
						}else {
							// 获取UID
							dirUid,err = strconv.Atoi(dirEach.Data["uid"])
							if err != nil{
								log("warning","此目录无法获取UID -> " + dirName ,nil)
								dirUid = 0
							}
							// 获取GID记录
							dirGid,err = strconv.Atoi(dirEach.Data["gid"])
							if err != nil{
								log("warning","此目录无法获取GID -> " + dirName ,nil)
								dirGid = 0
							}
						}
					}
					// 重置最后修改时间
					dirTime = getTimeNow(dirEach.Data["mtime"])

					// 若目录不存在则替换

					// 替换目录
					dirReplace(dirName,dirTime,dirUid,dirGid,os.FileMode(dirMode))
				}

				for fileName, fileEach := range dirEach.FileSet{
					// 若文件不存在
					if !isExist(fileName){
						log("shot","尝试恢复此文件 -> "+fileName,nil)
						// 避免条件竞争
						if !isExist(dirName){
							goto checkDirLoop
						}

						backFileName = fileEach.Data["backPath"]

						// 若为 Linux 获取所有者
						if !isWindows{
							uid,err = strconv.Atoi(fileEach.Data["uid"])
							if err != nil{
								log("warning","无法获取此文件备份记录的UID -> " + fileName ,nil)
							}
							gid,err = strconv.Atoi(fileEach.Data["gid"])
							if err != nil{
								log("warning","无法获取此文件备份记录的GID -> " + fileName ,nil)
							}
						}

						// 获取权限
						mode,err = strconv.Atoi(fileEach.Data["mode"])
						if err != nil{
							log("warning","无法获取此文件备份记录的权限 -> " + fileName ,nil)
							mode = 0777
						}

						// 获取最后修改时间
						fileTime = getTimeNow(fileEach.Data["mtime"])

						// 替换文件
						fileReplace(backFileName,fileName,fileTime,uid,gid,os.FileMode(mode))

					}
				}
			}
			time.Sleep(timeDuration)
		}

		WG.Done()

	}


	// 替换
	replace := func(){

		// 若能从信道取到值则进行替换
		var get = make(map[string]string)

		// 基础数据
		// 对文件
		var fileName string
		var mode int
		var fileTime time.Time
		var uid,gid = 0, 0
		var backFileName string

		// 对目录
		var dirName string
		var dirMode int
		var dirTime time.Time
		var dirUid, dirGid = 0, 0

		for{
			// 获取数据
			get = <- evilData

			fileName = get["fileName"]
			backFileName = get["backPath"]

			log("replace","尝试恢复此文件 -> "+fileName,nil)

			// 若为 Linux 获取所有者
			if !isWindows{
				uid,err = strconv.Atoi(get["uid"])
				if err != nil{
					log("warning","无法获取此文件备份记录的UID -> " + fileName ,nil)
				}
				gid,err = strconv.Atoi(get["gid"])
				if err != nil{
					log("warning","无法获取此文件备份记录的GID -> " + fileName ,nil)
				}
			}

			// 获取权限
			mode,err = strconv.Atoi(get["mode"])
			if err != nil{
				log("warning","无法获取此文件备份记录的权限 -> " + fileName ,nil)
				mode = 0777
			}

			// 获取最后修改时间
			fileTime = getTimeNow(get["mtime"])

			// 检查对应目录
			dirName = get["dirName"]

			// 若目录不存在则递归创建对应目录
			if !isExist(dirName){
				log("replace","尝试恢复此目录 -> "+dirName,nil)
				// 获取目录权限
				dirMode,err  = strconv.Atoi(get["dirMode"])
				if err != nil {
					log("warning", "无法获取此目录备份记录的权限 -> "+dirName, nil)
					dirMode = 0777
				}

				// 若为 Linux 则更改目录所有者/目录所有组
				if !isWindows{
					// 若所有者记录是无效的
					if get["dirUid"] == "" || get["dirGid"] == ""{
						log("warning","此目录所有者/所有组记录无效 -> " + dirName ,nil)
					}else {
						// 获取UID
						dirUid,err = strconv.Atoi(get["dirUid"])
						if err != nil{
							log("warning","此目录无法获取UID -> " + dirName ,nil)
							dirUid = 0
						}
						// 获取GID记录
						dirGid,err = strconv.Atoi(get["dirGid"])
						if err != nil{
							log("warning","此目录无法获取GID -> " + dirName ,nil)
							dirGid = 0
						}
					}
				}

				// 重置最后修改时间
				dirTime = getTimeNow(get["dirMTime"])

				// 替换目录
				dirReplace(dirName,dirTime,dirUid,dirGid,os.FileMode(dirMode))

			}
			// 替换文件
			fileReplace(backFileName,fileName,fileTime,uid,gid,os.FileMode(mode))

		}

		WG.Done()


	}

	// 获取检查协程|替换协程数
	var cs = strings.Split(control.Config["coroutines"],"|")
	var p,c,r = 3,2,2
	if len(cs) != 3{
		log("warning","无法获取主动检查协程|被动检查协程|替换协程数数 -> 使用默认值 1|1|1 " ,nil)
	}else{
		p,err = strconv.Atoi(cs[0])
		if err != nil{
			log("warning","无法获取动检查协程数 -> 使用默认值 1 " ,nil)
		}
		c,err = strconv.Atoi(cs[1])
		if err != nil{
			log("warning","无法获取被动检查协程数 -> 使用默认值 1 " ,nil)
		}
		r,err = strconv.Atoi(cs[2])
		if err != nil{
			log("warning","无法获取替换协程数 -> 使用默认值 1 " ,nil)
		}
	}

	// 设置主动检查协程
	WG.Add(p)
	for ;p>0;p--{
		go shot()

	}

	// 设置被动检查协程
	WG.Add(c)
	for ;c>0;c--{
		go check()
	}

	// 设置替换协程
	WG.Add(r)
	for ;r>0;r--{
		go replace()
	}

	// 阻塞
	WG.Wait()

}
