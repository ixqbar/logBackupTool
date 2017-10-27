package logBackup

import (
	"net"
	"fmt"
	"os"
	"strings"
	"gopkg.in/cheggaaa/pb.v1"
	"crypto/md5"
	"io"
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb/errors"
)
// name 文件新名称
func Transfer(server string, file string, path string, name string) error {
	fi, err := os.Stat(file)
	if err != nil || fi.IsDir() {
		fmt.Printf("Sorry,not found transfer file %s\n", file)
		os.Exit(1)
	}

	f, err := os.Open(file)
	if err != nil {
		Logger.Printf("Open file %s failed %v", file, err)
		return err
	}
	defer f.Close();

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		Logger.Printf("Get file %s md5sum failed %v", file, err)
		return err
	}

	m := hex.EncodeToString(h.Sum(nil))
	Logger.Printf("Get file %s md5sum %s", server, m)

	//reset to begin
	f.Seek(0, io.SeekStart);

	conn, err := net.Dial("tcp", server)
	if err != nil {
		Logger.Printf("Connect target server %s failed %v", server, err)
		return err
	}

	defer conn.Close()

	header := ""
	if len(name) == 0 {
		header = fmt.Sprintf("%s@%d@%s@%s\r\n", fi.Name(), fi.Size(), m, path)
	} else {
		header = fmt.Sprintf("%s@%d@%s@%s\r\n", name, fi.Size(), m, path)
	}

	headerByte := []byte(header)
	headerLen := len(header)
	n := 0
	for {
		nr, err := conn.Write(headerByte[n:])
		if err != nil {
			Logger.Printf("Transfer target server %s with header failed %v", server, err)
			return err
		}
		n += nr
		if n >= headerLen {
			break
		}
	}

	bar := pb.New64(fi.Size())
	bar.Start()

	buf := make([]byte, 1024)
	for {
		nr, err := f.Read(buf)
		if nr > 0 {
			nw, err := conn.Write(buf[0:nr])
			if err != nil {
				break
			}

			bar.Add(nw)
		}
		if err != nil {
			break
		}
	}

	bar.Finish()

	response := make([]byte, 1024)
	l, err := conn.Read(response)
	if err != nil {
		Logger.Printf("Transfer target server %s failed %v", server, err)
		return err
	}

	responseStr := strings.Trim(string(response[:l]), "\r\n")
	if l < 2 || responseStr != "OK" {
		Logger.Printf("Transfer target server %s resposne \n",server, responseStr)
		return errors.New("Transfer target server response error")
	}

	return nil
}