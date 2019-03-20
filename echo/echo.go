package main

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var (
	receivedBytes int64
	receivedMsgs  int64
)

func serve(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 32*1024)
	for {
		rn, err := conn.Read(buf)
		if err != nil {
			break
		}

		wn, err := conn.Write(buf[:rn])
		if wn != rn || err != nil {
			break
		}

		atomic.AddInt64(&receivedBytes, int64(rn))
		atomic.AddInt64(&receivedMsgs, 1)
	}
}

func printThroughputTillDie() {
	duration := time.Second * 3

	ticker := time.NewTicker(duration)
	for {
		select {

		case <-ticker.C:
			bytes := atomic.SwapInt64(&receivedBytes, 0)
			msgs := atomic.SwapInt64(&receivedMsgs, 0)

			fmt.Printf("%.3f MiB/s %.3f Ki Msgs/s %.3f bytes per msg\n",
				float64(bytes)/duration.Seconds()/1024/1024,
				float64(msgs)/duration.Seconds()/1024,
				float64(bytes)/float64(msgs))
		}
	}
}

func main() {
	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :20001\n", arg[0], arg[0])
		return
	}

	listener, err := net.Listen("tcp", arg[1])

	if err != nil {
		panic(err)
	}

	defer listener.Close()

	go printThroughputTillDie()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go serve(conn)
	}
}
