package main

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

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
	GateHost     string
	GatePort     int
	GateUsername string
	GatePassword string
	DataHost     string
	DataPort     int
	LocalHost    string
	LocalPort    int
}

var tunnelConfigMap = map[string]TunnelConfig{
	"PostgreSQL": {"10.0.0.11", 22, "root", "password", "127.0.0.1", 5432, "127.0.0.1", 5432},
	"Redis     ": {"10.0.0.11", 22, "root", "password", "127.0.0.1", 6379, "127.0.0.1", 6379},
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
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
	for flag, config := range tunnelConfigMap {
		tunnel := &SSHtunnel{
			&Endpoint{config.GateHost, config.GatePort},
			&ssh.ClientConfig{
				User:            config.GateUsername,
				Auth:            []ssh.AuthMethod{ssh.Password(config.GatePassword)},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
			},
			&Endpoint{config.DataHost, config.DataPort},
			&Endpoint{config.LocalHost, config.LocalPort},
		}
		go func(flag string, tunnel *SSHtunnel) {
			fmt.Printf("%s %s %s:%-5d => %s:%-5d\n", flag, tunnel.Server.Host, tunnel.Remote.Host, tunnel.Remote.Port, tunnel.Local.Host, tunnel.Local.Port)
			fmt.Println(tunnel.Start())
		}(flag, tunnel)
	}
	for {
	}
}
