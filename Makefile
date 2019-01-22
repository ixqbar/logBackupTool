SHELL:=/bin/bash
TARGET=BackupServer

all: win linux mac

win: 
	GOOS=windows GOARCH=amd64 go build -o ./bin/${TARGET}.exe ./src
	GOOS=windows GOARCH=386 go build -o ./bin/${TARGET}-x32.exe ./src
	
linux: 
	GOOS=linux GOARCH=amd64 go build -o ./bin/${TARGET}_${@} ./src
	GOOS=linux GOARCH=386 go build -o ./bin/${TARGET}_${@}_x32 ./src

mac: 
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${TARGET}_${@} ./src
	
clean:
	rm -rf ./bin/${TARGET}*
