package logBackup

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "", log.Ldate | log.Lmicroseconds | log.Lshortfile)