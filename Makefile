TARGET=BackupServer

all: mac linux win

linux: 
	GOOS=linux GOARCH=amd64 go build -o ./bin/${TARGET}_${@} ./src

mac: 
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${TARGET}_${@} ./src

win:
	GOOS=windows GOARCH=amd64 go build -o ./bin/${TARGET}.exe ./src
	
clean:
	rm -rf ./bin/${TARGET}_*
