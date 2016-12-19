# logBackupTool

## usage
```
Usage: ./logBackupTool [options]
Options:
  -c string
    	config (default "/etc/logBackup.conf")
  -file string
    	send file to server
  -mc
    	run as server mode (default false)
  -ms
    	run as client mode (default true)
  -path string
    	send file to server backup path
  -verbose
    	show run details (default true)
```

### logBackup.conf
```
[server]
bind-address=127.0.0.1:2010
datadir=/Users/xingqiba/workspace/go/logBackupTool/logs

[client]
server-address=127.0.0.1:2010
```

## server
```
./logBackupTool -c ../conf/logBackup.conf -ms
```

## client
```
./logBackupTool --file test.log --path data/your_path
```

### FAQ
更多疑问请+qq群 233415606 or [website http://www.hnphper.com](http://www.hnphper.com)
