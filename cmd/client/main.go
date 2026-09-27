package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/rpc"
)

type Message struct {
	Token   string
	User    string
	Content string
}

func main() {

	type AuthRequest struct {
		Username string
		Password string
	}

	type AuthResponse struct {
		Token string
		Error string
	}

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

	fmt.Println("Connected to server via TLS")

	fmt.Println("\n[1] Authenticating...")
	authReq := AuthRequest{
		Username: "ampar",
		Password: "password123",
	}

	var authResp AuthResponse

	err = client.Call("ChatService.Authenticate", authReq, &authResp)
	if err != nil {
		log.Fatalf("Authentication failed: %s, authResp.Error")
	}

	token := authResp.Token
	fmt.Printf("Authenticated! Token: %s\n", token[:16]+"...")

	fmt.Println("n[2] Sending message...")
	message := Message{
		Token:   token,
		User:    "ampar",
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
