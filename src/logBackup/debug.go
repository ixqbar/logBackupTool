package logBackup

import (
	"os"
	"runtime"
	"fmt"
)

func Debugf(format string, a ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "<unknown>"
			line = -1
		}

		fmt.Fprintf(os.Stderr, fmt.Sprintf("[%d] %s:%d %s\n", os.Getpid(), file, line, format), a...)
	}
}