package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type Message struct {
	User    string
	Content string
}

func main() {
	client, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

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
