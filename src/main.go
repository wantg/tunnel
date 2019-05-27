package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type key int

type endpoint struct {
	Host string
	Port int
}

type tunnel struct {
	Title   string
	Enabled bool
	Gate    struct {
		endpoint
		Username string
		Password string
	}
	Source endpoint
	Mirror endpoint
}

func appPath(subPath *string) *string {
	rootPath, _ := os.Executable()
	s := filepath.Join(filepath.Dir(rootPath), *subPath)
	return &s
}

func randString() string {
	rand.Seed(time.Now().UnixNano())
	s := strconv.Itoa(rand.Int())
	return s
}

func (endpoint *endpoint) String() string {
	return endpoint.Host + ":" + strconv.Itoa(endpoint.Port)
}

func (tunnel *tunnel) start() error {
	listener, err := net.Listen("tcp", tunnel.Mirror.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		mirrorConn, err := listener.Accept()
		if err != nil {
			return err
		}
		ctx0 := context.WithValue(context.Background(), key(1), randString())
		go tunnel.forward(ctx0, mirrorConn)
	}
}

func (tunnel *tunnel) forward(ctx0 context.Context, mirrorConn net.Conn) {
	fmt.Println()
	// fmt.Println("parent goroutine", ctx0.Value(key(1)).(string))
	sshConfig := &ssh.ClientConfig{
		User:            tunnel.Gate.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(tunnel.Gate.Password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}
	sshServerConn, err := ssh.Dial("tcp", tunnel.Gate.String(), sshConfig)
	fmt.Println("Open Gate", sshServerConn.RemoteAddr())
	if err != nil {
		fmt.Printf("Server dial error: %s\n", err)
		return
	}
	sourceConn, err := sshServerConn.Dial("tcp", tunnel.Source.String())
	fmt.Println("Connected", tunnel.Source.String())
	if err != nil {
		fmt.Printf("Remote dial error: %s\n", err)
		return
	}
	copyConn := func(writer, reader net.Conn) {
		// ctx1 := context.WithValue(context.Background(), key(1), randString())
		// fmt.Println(" child goroutine", ctx1.Value(key(1)).(string))
		_, err := io.Copy(writer, reader)
		writer.Close()
		reader.Close()
		sshServerConn.Close()
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}
	go copyConn(mirrorConn, sourceConn)
	go copyConn(sourceConn, mirrorConn)
}

func main() {
	var tunnels = make([]tunnel, 0)
	configPath := "./config.json"
	configBytes, _ := ioutil.ReadFile(*(appPath(&configPath)))
	json.Unmarshal(configBytes, &tunnels)

	var titleLength, GateTitleLength, SourceTitleLength float64 = 0, 0, 0
	for _, tunnel := range tunnels {
		if !tunnel.Enabled {
			continue
		}
		titleLength = math.Max(titleLength, float64(len(tunnel.Title)))
		GateTitleLength = math.Max(GateTitleLength, float64(len(tunnel.Gate.String())))
		SourceTitleLength = math.Max(SourceTitleLength, float64(len(tunnel.Source.String())))
	}
	logFormat := fmt.Sprintf("%%-%ds %%-%ds %%-%ds => %%s\n", int(titleLength), int(GateTitleLength), int(SourceTitleLength))
	// fmt.Println(logFormat)
	for _, tunnel := range tunnels {
		if !tunnel.Enabled {
			continue
		}
		fmt.Printf(logFormat, tunnel.Title, tunnel.Gate.String(), tunnel.Source.String(), tunnel.Mirror.String())
		go tunnel.start()
	}
	time.Sleep(time.Hour * 24 * 7)
}
