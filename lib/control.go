package lib

var E = `控制集 > 目录集 > 文件集

控制集：主目录，其他参数

主目录：
- 目录最后修改时间
- 备份目录路径
- 文件集

文件集：
- 文件原始hash
- 文件最后修改时间
- 备份文件路径`

type Control struct {
	DirSet map[string]*Dirs
	Config map[string]string
}

type Dirs struct {
	Data    map[string]string
	FileSet map[string]*Files
}

type Files struct {
	Data map[string]string
}

func (c *Control) AddDirs(dirName string,dir *Dirs) {
	c.DirSet[dirName] = dir
}
func (c *Control) SetConfig(key string, value string) {
	c.Config[key] = value
}

func (d *Dirs) addFiles(fileName string,file *Files){
	d.FileSet[fileName] = file
}
func (d *Dirs) SetData(key, value string) {
	d.Data[key] = value
}

func (f *Files) SetData(key, value string) {
	f.Data[key] = value
}

