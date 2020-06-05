package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	setForwarding("127.0.0.1:3389", "10.0.0.1:3389")
}

func setForwarding(listen, target string) {
	fmt.Printf("%s > %s\n", listen, target)
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handleRequest(conn, target)
	}
}

func handleRequest(conn net.Conn, addr string) {
	fmt.Println("new client")
	proxy, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("proxy connected")
	go copyIO(conn, proxy)
	go copyIO(proxy, conn)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
