package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "listen" {
		log.Print("Listen mode")
		doServer()
		return
	}

	var wg sync.WaitGroup
	var perSecond uint32 = 0
	go measure(&perSecond)

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	lim := syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Current limit: %d", lim.Cur)

	var done uint32 = 0

	for i := uint64(0); i < lim.Cur-10; i++ {
		go connection(&wg, &perSecond, &done)
	}
	<-cancelChan
	atomic.AddUint32(&done, 1)
	log.Print("Caught signal, canceling...")
	wg.Wait()
}

func measure(count *uint32) {
	c := time.Tick(1 * time.Second)
	for {
		<-c
		log.Printf("%d", atomic.SwapUint32(count, 0))
	}
}

func connection(wg *sync.WaitGroup, count *uint32, done *uint32) {
	for atomic.LoadUint32(done) == 0 {
		wg.Add(1)

		tcp, err := net.Dial("tcp", "127.0.0.1:1223")
		if err != nil {
			log.Print(err)
			wg.Done()
			continue
		}
		_, err = tcp.Write([]byte("GET / HTTP/1.1\r"))
		if err != nil {
			log.Print(err)
			tcp.Close()
			wg.Done()
			continue
		}
		atomic.AddUint32(count, 1)
		tcp.Close()
		wg.Done()
	}
}
