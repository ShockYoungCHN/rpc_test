package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

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
	time.Sleep(2 * time.Second)
}

func main() {
	log.Printf("Starting server on port 8080")
	http.HandleFunc("/echo", echo_gobwas)
	http.ListenAndServe(":8080", nil)

	l, _ := net.Listen("tcp", ":8080/echo")

	var tempDelay time.Duration
	// create 4 goroutines to handle the connections
	for i := 0; i < 4; i++ {
		go func() {
			for {
				rw, err := l.Accept()
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
					break
				}
				tempDelay = 0
				//epoll, _ := CreateEpoll(nil)
				go epollEcho(rw)
			}
		}()
	}

}
