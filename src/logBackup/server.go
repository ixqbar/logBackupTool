package logBackup

import (
	"net"
	"fmt"
	"time"
	"errors"
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"strconv"
	"io"
	"path"
)

type Server struct {
	addr       string
	backupPath string
	socket     *net.TCPListener
	cm         *ConnManager
}

func NewServer(config *Config) (*Server, error) {
	Debugf("server run %s %s", config.Addr, config.BackupPath)
	return &Server{
		addr       : config.Addr,
		backupPath : config.BackupPath,
		socket     : nil,
		cm         : NewConnManager(),
	}, nil
}

func (srv *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", srv.addr)
	if err != nil {
		return fmt.Errorf("fail to resolve addr: %v", err)
	}

	sock, err := net.ListenTCP("tcp", addr)
	if err != nil {
		Debugf("server run %s failed %s", srv.addr, err)
		return fmt.Errorf("fail to listen tcp: %v", err)
	}

	srv.socket = sock

	return srv.acceptConn()
}

func (srv *Server) Stop() error {
	srv.socket.SetDeadline(time.Now())

	tt := time.NewTimer(time.Second * time.Duration(30))
	wait := make(chan struct{})
	go func() {
		srv.cm.Wait()
		wait <- struct{}{}
	}()

	select {
	case <-tt.C:
		return errors.New("stop server wait timeout")
	case <-wait:
		return nil
	}
}

func (srv *Server)acceptConn() error {
	defer func() {
		srv.socket.Close()
	}()

	for {
		conn, err := srv.socket.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				break;
			}
			continue
		}

		go func() {
			srv.cm.Add(1)
			srv.handleConn(conn)
			srv.cm.Done()
		}()
	}

	return nil
}

func (srv *Server) handleConn(conn net.Conn) {
	var clientAddr = conn.RemoteAddr().String()
	Debugf("%s connected", clientAddr)

	defer func() {
		Debugf("%s disconnected", clientAddr)
		conn.Close()
	}()

	//file size path\r\n
	//data
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(30)))

		r := bufio.NewReader(conn)
		content, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				Debugf("%s parse transfer header failed %v", clientAddr, err)
			}
			break
		}

		summaryInfo := strings.Split(strings.Trim(content, "\r\n"), "@")
		summaryLen := len(summaryInfo)
		if summaryLen < 3 {
			if summaryInfo[0] == "PING" {
				conn.Write([]byte("PONG\r\n"))
				continue
			} else {
				Debugf("%s parse transfer header failed %s", clientAddr, content)
				conn.Write([]byte("parse transfer header failed\r\n"))
				break
			}
		}

		fpath := summaryInfo[summaryLen-1]
		fsize := 0

		if len(summaryInfo[summaryLen-2]) > 0 {
			num, err := strconv.Atoi(summaryInfo[summaryLen-2])
			if err != nil {
				Debugf("%s parse transfer header failed %s %s", clientAddr, content, err)
				conn.Write([]byte("parse transfer header failed error size\r\n"))
				break
			}
			fsize = num
		}

		if fsize == 0 {
			Debugf("%s parse transfer header failed %s error size", clientAddr, content)
			conn.Write([]byte("parse transfer header failed error size\r\n"))
			break
		}

		fname := strings.Trim(strings.Join(summaryInfo[:summaryLen-2], "@"), "@")

		fileName := ""
		if len(fpath) > 0 {
			fileName = path.Join(srv.backupPath, fpath, fname)
		} else {
			fileName = path.Join(srv.backupPath, fname)
		}

		Debugf("%s backup file %s size %d", clientAddr, fileName, fsize)

		parentDir := filepath.Dir(fileName)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				Debugf("%s creaet folder %s failed %v", clientAddr, parentDir, err)
				break
			}
		} else if _, err := os.Stat(fileName); err == nil {
			Debugf("%s target file %s exists", clientAddr, fileName)
			os.Remove(fileName)
		}

		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			Debugf("%s open file %s failed %v", clientAddr, fileName, err)
			break
		}

		data := make([]byte, 1024)

		m := 0
		for {
			conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(180)))
			n, err := r.Read(data)
			if err != nil {
				break;
			}

			f.Write(data[:n])
			m += n
			if m >= fsize {
				break
			}
		}

		f.Close()

		conn.Write([]byte("OK\r\n"))
	}

}