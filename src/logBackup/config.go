package logBackup

type Config struct {
	Addr       string
	BackupPath string
	ToChown bool
	Uid int
	Gid int
}

var GloablConfig *Config;

func init() {
	GloablConfig = &Config{
		Addr:"0.0.0.0:2010",
		BackupPath:"/tmp",
		ToChown:false,
		Uid:0,
		Gid:0,
	}
}
