package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func doServer() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		ln, err := net.Listen("tcp", ":1223")
		if err != nil {
			log.Fatal(err)
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleConnection(conn)

		}
	}()
	<-cancelChan
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Hello"))
}
