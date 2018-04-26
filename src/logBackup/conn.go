package logBackup

import (
	"net"
	"time"
)

type FConn struct {
	conn net.Conn
}

func (obj FConn) Read(p []byte) (int,error) {
	return obj.conn.Read(p)
}

func (obj FConn) Write(message []byte) bool {
	totalLen := len(message)
	writeLen := 0

	for {
		obj.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(30)))
		n, err := obj.conn.Write(message[writeLen:])
		if err != nil {
			Logger.Printf("server write client message %s failed %s", string(message), err)
			return false
		}

		writeLen += n
		if n >= totalLen {
			return true
		}
	}
}

func (obj FConn) ClientAddress() string {
	return obj.conn.RemoteAddr().String()
}

func (obj FConn) Close() error {
	return obj.conn.Close()
}

func (obj FConn) SetReadDeadline(t time.Time) error {
	return obj.conn.SetReadDeadline(t)
}