package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/rpc"
)

type Message struct {
	User    string
	Content string
}

func main() {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "localhost:8080", tlsConfig)

	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	client := rpc.NewClient(conn)
	defer client.Close()

	message := Message{
		User:    "Aditya",
		Content: "Hello from the RPC client!",
	}

	var reply string

	err = client.Call(
		"ChatService.SendMessage",
		message,
		&reply,
	)

	if err != nil {
		log.Fatal("RPC call failed:", err)
	}

	fmt.Println("Server:", reply)
}
