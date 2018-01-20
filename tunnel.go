package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var tunnelConfigMap = map[string]([]interface{}){
	"PostgreSQL": []interface{}{"10.0.0.11", 22, "root", "password", "127.0.0.1", 5432, "127.0.0.1", 5432},
	"Redis     ": []interface{}{"10.0.0.11", 22, "root", "password", "127.0.0.1", 6379, "127.0.0.1", 6379},
}

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHtunnel struct {
	Server *Endpoint
	Config *ssh.ClientConfig
	Remote *Endpoint
	Local  *Endpoint
}

func (tunnel *SSHtunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHtunnel) forward(localConn net.Conn) {
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
		_, err := io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func main() {
	for flag, c := range tunnelConfigMap {
		tunnel := &SSHtunnel{
			&Endpoint{c[0].(string), c[1].(int)},
			&ssh.ClientConfig{
				User:            c[2].(string),
				Auth:            []ssh.AuthMethod{ssh.Password(c[3].(string))},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
			},
			&Endpoint{c[4].(string), c[5].(int)},
			&Endpoint{c[6].(string), c[7].(int)},
		}
		go func(flag string, tunnel *SSHtunnel) {
			fmt.Printf("%s %s %s:%-5d => %s:%-5d\n", flag, tunnel.Server.Host, tunnel.Remote.Host, tunnel.Remote.Port, tunnel.Local.Host, tunnel.Local.Port)
			fmt.Println(tunnel.Start())
		}(flag, tunnel)
	}
	for {
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}
