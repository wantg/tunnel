package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type endpoint struct {
	Host string
	Port int
}

// Tunnel Tunnel
type Tunnel struct {
	Title   string
	Enabled bool
	Gate    struct {
		endpoint     `yaml:",inline"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		IdentityFile string `yaml:"identityFile"`
	}
	Source endpoint
	Mirror endpoint
}

func (endpoint *endpoint) String() string {
	return endpoint.Host + ":" + strconv.Itoa(endpoint.Port)
}

func getPublicKey(file string) ssh.AuthMethod {
	p := file
	if strings.HasPrefix(file, "~") {
		usr, _ := user.Current()
		userHomeDir := usr.HomeDir
		p = userHomeDir + file[1:]
	}
	p, _ = filepath.Abs(p)
	buffer, err := ioutil.ReadFile(p)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (tunnel *Tunnel) start() error {
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
		ctx0 := context.WithValue(context.Background(), 1, time.Now().UnixNano())
		go tunnel.forward(ctx0, mirrorConn)
	}
}

func (tunnel *Tunnel) forward(ctx0 context.Context, mirrorConn net.Conn) {
	fmt.Println()
	// fmt.Println("parent goroutine", ctx0.Value(key(1)).(string))
	authMethod := ssh.Password(tunnel.Gate.Password)
	if len(tunnel.Gate.IdentityFile) > 0 {
		authMethod = getPublicKey(tunnel.Gate.IdentityFile)
	}
	sshConfig := &ssh.ClientConfig{
		User:            tunnel.Gate.Username,
		Auth:            []ssh.AuthMethod{authMethod},
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
		// ctx1 := context.WithValue(context.Background(), key(1), time.Now().UnixNano())
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

func loadConfig() ([]byte, error) {
	configPtr := flag.String("c", "", "config file path")
	flag.Parse()
	configPath := strings.TrimSpace(*configPtr)
	if len(configPath) == 0 {
		return nil, fmt.Errorf(`send config to me by "-c config.yml"`)
	}
	configPath, _ = filepath.Abs(configPath)
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func main() {
	bts, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	var tunnels = make([]Tunnel, 0)
	yaml.Unmarshal(bts, &tunnels)

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
		go func(tunnel Tunnel) {
			fmt.Println(tunnel.start())
		}(tunnel)
	}
	for {
	}
}
