package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type key int

type Endpoint struct {
	Host string
	Port int
}

type SSHtunnel struct {
	Server *Endpoint
	Config *ssh.ClientConfig
	Remote *Endpoint
	Local  *Endpoint
}

type TunnelConfig struct {
	Gate struct {
		Endpoint
		Username string
		Password string
	}
	Source Endpoint
	Mirror Endpoint
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

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

func (tunnel *SSHtunnel) start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			return err
		}
		ctx0 := context.WithValue(context.Background(), key(1), randString())
		go tunnel.forward(ctx0, localConn)
	}
}

func (tunnel *SSHtunnel) forward(ctx0 context.Context, localConn net.Conn) {
	fmt.Println()
	fmt.Println("parent goroutine", ctx0.Value(key(1)).(string))
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		fmt.Printf("Server dial error: %s\n", err)
		return
	}

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		fmt.Printf("Remote dial error: %s\n", err)
		return
	}

	copyConn := func(writer, reader net.Conn) {
		ctx1 := context.WithValue(context.Background(), key(1), randString())
		fmt.Println(" child goroutine", ctx1.Value(key(1)).(string))
		_, err := io.Copy(writer, reader)
		writer.Close()
		reader.Close()
		serverConn.Close()
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func main() {
	var tunnelConfig = make(map[string]TunnelConfig)
	configPath := "./config.json"
	configBytes, _ := ioutil.ReadFile(*(appPath(&configPath)))
	json.Unmarshal(configBytes, &tunnelConfig)

	for flag, config := range tunnelConfig {
		tunnel := &SSHtunnel{
			&Endpoint{config.Gate.Host, config.Gate.Port},
			&ssh.ClientConfig{
				User:            config.Gate.Username,
				Auth:            []ssh.AuthMethod{ssh.Password(config.Gate.Password)},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
			},
			&Endpoint{config.Source.Host, config.Source.Port},
			&Endpoint{config.Mirror.Host, config.Mirror.Port},
		}
		go func(flag string, tunnel *SSHtunnel) {
			fmt.Printf("%s %s %s:%-5d => %s:%-5d\n", flag, tunnel.Server.Host, tunnel.Remote.Host, tunnel.Remote.Port, tunnel.Local.Host, tunnel.Local.Port)
			fmt.Println(tunnel.start())
		}(flag, tunnel)
	}
	time.Sleep(time.Hour * 24 * 7)
}
