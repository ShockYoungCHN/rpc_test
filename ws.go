package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"

	"github.com/gobwas/ws"
)

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
	time.Sleep(2 * time.Second)
	defer conn.Close()
	/*	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				// handle error
				break
			}

			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				// handle error
				break
			}
		}
	}()*/
}

func main() {
	log.Printf("Starting server on port 8080")
	http.HandleFunc("/echo", echo)
	http.ListenAndServe(":8080", nil)
}

//183-105=78
//257-183=74
