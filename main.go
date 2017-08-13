package main

import (
	"os"
	"bytes"
	"encoding/json"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/websocket"
)

var port = flag.String("port", "8888", "Port for server")
var host = flag.String("host", "127.0.0.1:2735", "Docker host")

var contextPath = "/"

func main() {
	flag.Parse()

	if cp := os.Getenv("CONTEXT_PATH"); cp != "" {
		contextPath = strings.TrimRight(cp, "/")
	}

	http.Handle(contextPath + "/exec/", websocket.Handler(ExecContainer))
        http.Handle(contextPath + "/", http.StripPrefix(contextPath + "/", http.FileServer(http.Dir("./"))))
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		panic(err)
	}
}

func ExecContainer(ws *websocket.Conn) {
	wsParams := strings.Split(ws.Request().URL.Path[len(contextPath + "/exec/"):], ",")
	container := wsParams[0]	
	cmd, _ := base64.StdEncoding.DecodeString(wsParams[1])
		
	if container == "" {
		ws.Write([]byte("Container does not exist"))
		return
	}
	type stuff struct {
		Id string
	}
	var s stuff
	params := bytes.NewBufferString("{\"AttachStdin\":true,\"AttachStdout\":true,\"AttachStderr\":true,\"Tty\":true,\"Cmd\":[\"" + string(cmd) + "\"]}")
	resp, err := http.Post("http://" + *host + "/containers/" + container + "/exec", "application/json", params)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(data), &s)
	if err := hijack(*host, "POST", "/exec/"+s.Id+"/start", true, ws, ws, ws, nil, nil); err != nil {
		panic(err)
	}
	fmt.Println("Connection!")
	fmt.Println(ws)
	spew.Dump(ws)
}

func hijack(addr, method, path string, setRawTerminal bool, in io.ReadCloser, stdout, stderr io.Writer, started chan io.Closer, data interface{}) error {

	params := bytes.NewBufferString("{\"Detach\": false, \"Tty\": true}")
	req, err := http.NewRequest(method, path, params)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Docker-Client")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	req.Host = addr

	dial, err := net.Dial("tcp", addr)
	// When we set up a TCP connection for hijack, there could be long periods
	// of inactivity (a long running command with no output) that in certain
	// network setups may cause ECONNTIMEOUT, leaving the client in an unknown
	// state. Setting TCP KeepAlive on the socket connection will prohibit
	// ECONNTIMEOUT unless the socket connection truly is broken
	if tcpConn, ok := dial.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	if err != nil {
		return err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	defer clientconn.Close()

	// Server hijacks the connection, error 'connection closed' expected
	clientconn.Do(req)

	rwc, br := clientconn.Hijack()
	defer rwc.Close()

	if started != nil {
		started <- rwc
	}

	var receiveStdout chan error

	if stdout != nil || stderr != nil {
		go func() (err error) {
			if setRawTerminal && stdout != nil {
				_, err = io.Copy(stdout, br)
			}
			return err
		}()
	}

	go func() error {
		if in != nil {
			io.Copy(rwc, in)
		}

		if conn, ok := rwc.(interface {
			CloseWrite() error
		}); ok {
			if err := conn.CloseWrite(); err != nil {
			}
		}
		return nil
	}()

	if stdout != nil || stderr != nil {
		if err := <-receiveStdout; err != nil {
			return err
		}
	}
	spew.Dump(br)
	go func() {
		for {
			fmt.Println(br)
			spew.Dump(br)
		}
	}()

	return nil
}
