package main

import (
	"flag"
	"fmt"
	"github.com/gogf/greuse"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var mu sync.Mutex
var cond = sync.NewCond(&mu)
var start int64 = 0
var finish int64 = 0

var ioTimeout = flag.Duration("io_timeout", time.Millisecond*100, "i/o operations timeout")

type deadliner struct {
	net.Conn
	t time.Duration
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func echo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}
	defer ws.Close()
	time.Sleep(2 * time.Second)

	/*for {
			// Read message from browser
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("Failed to read message: %v", err)
				break
			}

			// Print the message to the console
			log.Printf("Received: %s\n", msg)

			// Write message back to browser
			if err = ws.WriteMessage(msgType, msg); err != nil {
				log.Printf("Failed to write message: %v", err)
				break
			}
	}*/
}

func echo_gobwas(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		// handle error
		return
	}
	port := conn.RemoteAddr().(*net.TCPAddr).Port
	log.Printf("port: %d", port)
	err = wsutil.WriteServerMessage(conn, 0x1, []byte(strconv.Itoa(port)))
	if err != nil {
		// handle error
		log.Printf("Failed to write message: %v", err)
		return
	}

	time.Sleep(2 * time.Second)
	conn.Close()
}

func epollEcho(conn net.Conn) {
	safeConn := deadliner{conn, *ioTimeout}
	_, _ = ws.Upgrade(safeConn)
	defer conn.Close()
	time.Sleep(20 * time.Second)
}

func MaxUpgrade(conn net.Conn) {
	cond.L.Lock()
	cond.Wait()
	cond.L.Unlock()

	safeConn := deadliner{conn, *ioTimeout}

	if start == 0 {
		start = time.Now().UnixNano()
	}
	_, _ = ws.Upgrade(safeConn)
	if time.Now().UnixNano() > finish {
		finish = time.Now().UnixNano()
	}

	defer conn.Close()
	time.Sleep(10 * time.Second)
}

func main() {
	/*	http.HandleFunc("/echo", echo_gobwas)
		http.ListenAndServe(":8080", nil)*/

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic err:", err)
		}
	}()

	var wg sync.WaitGroup

	n := 1 // default 1 acceptor
	n, _ = strconv.Atoi(os.Args[1])
	log.Printf("%d acceptor", n)

	var tempDelay time.Duration
	// create 4 goroutines to handle the connections
	for i := 0; i < n; i++ {
		wg.Add(1)

		port := "8080"
		l, err := greuse.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to create listener: %v", err)
		}
		log.Printf("Starting server on port 8080")

		go func() {
			for {
				rw, _ := l.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Temporary() {
						if tempDelay == 0 {
							tempDelay = 5 * time.Millisecond
						} else {
							tempDelay *= 2
						}
						if max := 1 * time.Second; tempDelay > max {
							tempDelay = max
						}
						time.Sleep(tempDelay)
						continue
					}
					log.Printf("Failed to accept connection: %v", err)
					break
				}
				tempDelay = 0
				//epoll, _ := CreateEpoll(nil)
				go MaxUpgrade(rw)
			}
		}()
	}
	// wait for all goroutines to be ready for upgrade
	time.Sleep(3 * time.Second)
	cond.Broadcast()

	wg.Wait()
	log.Printf("MaxUpgrade: %d", finish-start)
}
