package logBackup

import "os"

type Config struct {
	Addr       string
	BackupPath string
	ToChown bool
	Uid int
	Gid int
	Perm os.FileMode
}

var GloablConfig *Config;

func init() {
	GloablConfig = &Config{
		Addr:"0.0.0.0:2010",
		BackupPath:"/tmp",
		ToChown:false,
		Uid:0,
		Gid:0,
		Perm:0755,
	}
}
