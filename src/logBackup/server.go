package logBackup

import (
	"net"
	"fmt"
	"time"
	"errors"
	"bufio"
	"os"
	"path/filepath"
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

func (srv *Server) handleConn(conn net.Conn) error {
	var clientAddr = conn.RemoteAddr().String()
	Debugf("client %s connected", clientAddr)

	defer func() {
		Debugf("client %s disconnected", clientAddr)
		conn.Close()
	}()

	//file\r\n
	//data

	r := bufio.NewReader(conn)
	content, err := r.ReadString('\n')
	if err != nil {
		Debugf("handleConn occur error %v", err)
		return err
	}

	fname := ""
	fsize := 0
	fpath := ""

	if _, err := fmt.Sscanf(content, "%s %d %s\r\n", &fname, &fsize, &fpath); err != nil {
		Debugf("handleConn occur error %v", err)
		return err
	}

	Debugf("start recevie %v file %s", clientAddr, fname)

	fileName := ""
	if len(fpath) > 0 {
		fileName = fmt.Sprintf("%s%s%s", srv.backupPath, fpath, fname)
	} else {
		fileName = fmt.Sprintf("%s%s", srv.backupPath, fname)
	}

	parentDir := filepath.Dir(fileName)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			Debugf("creaet folder %s failed %v", parentDir, err)
			return err
		}
	} else if _, err := os.Stat(fileName); err == nil {
		Debugf("target file %s exists", fileName)
		os.Remove(fileName)
	}

	f, err := os.OpenFile(fileName, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644)
	if err != nil {
		Debugf("handleConn occur error %v", err)
		return err
	}

	defer f.Close()

	data := make([]byte, 1024)

	m := 0
	for {
		n, err := r.Read(data)
		if err != nil {
			break;
		}

		f.Write(data[:n])
		m +=n
		if m >= fsize {
			break
		}
	}

	conn.Write([]byte("OK\r\n"))

	return nil
}