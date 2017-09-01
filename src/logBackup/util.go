package logBackup

import (
	"path/filepath"
	"os"
)

func ChownR(path string, uid, gid int) (error) {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = Chown(name, uid, gid)
		} else {
			Debugf("walk %s failed %s", name, err)
		}
		return err
	})
}

func Chown(name string ,uid, gid int) (error) {
	err := os.Chown(name, uid, gid)
	if err != nil {
		Debugf("chown %s failed", name, err)
	}

	return err
}