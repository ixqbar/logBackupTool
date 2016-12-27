package logBackup

import (
	"net"
	"fmt"
	"os"
	"strings"
	"gopkg.in/cheggaaa/pb.v1"
)

func Transerf(server string, file string, path string, name string) error {
	fi, err := os.Stat(file)
	if err != nil || fi.IsDir() {
		fmt.Printf("Sorry,not found transfer file %s\n", file)
		os.Exit(1)
	}

	f, err := os.Open(file)
	if err != nil {
		Debugf("Open file %s failed %v", file, err)
		return err
	}
	defer f.Close();

	conn, err := net.Dial("tcp", server)
	if err != nil {
		Debugf("Connect target server %s failed %v", server, err)
		return err
	}

	defer conn.Close()

	if len(name) == 0 {
		conn.Write([]byte(fmt.Sprintf("%s@%d@%s\r\n", fi.Name(), fi.Size(), path)))
	} else {
		conn.Write([]byte(fmt.Sprintf("%s@%d@%s\r\n", name, fi.Size(), path)))
	}

	bar := pb.New64(fi.Size())
	bar.Start()

	buf := make([]byte, 1024)
	for {
		nr, er := f.Read(buf)
		if nr > 0 {
			nw, ew := conn.Write(buf[0:nr])
			if ew != nil {
				break
			}

			bar.Add(nw)
		}
		if er != nil {
			break
		}
	}

	bar.Finish()

	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err == nil {
		fmt.Printf("Server Response `%s`\n", strings.Trim(string(buff[:n]), "\r\n"))
	}

	return nil
}