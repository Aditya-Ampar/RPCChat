package main

import (
	"fmt"
	"github.com/Aditya-Ampar/RPCChat/internal/database"
	"log"
	"net"
	"net/rpc"
)

type ChatService struct{}

type Message struct {
	User    string
	Content string
}

func (c *ChatService) SendMessage(msg Message, reply *string) error {
	log.Printf("[%s] %s", msg.User, msg.Content)

	*reply = "Message received"

	return nil
}

func main() {

	db, err := database.Open("chat.db")
	if err != nil {
		log.Fatal("Failed to Open Database", err)
	}

	if err := database.InitializeSchema(db); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	defer db.Close()

	chat := new(ChatService)

	err = rpc.Register(chat)

	if err != nil {
		log.Fatal("RPC registration failed:", err)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	defer listener.Close()

	fmt.Println("RPC Chat Server")
	fmt.Println("Listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}
