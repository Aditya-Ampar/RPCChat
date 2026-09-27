package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Aditya-Ampar/RPCChat/internal/auth"
	"github.com/Aditya-Ampar/RPCChat/internal/database"
	"log"
	"net/rpc"
)

type AuthRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	Token string
	Error string
}

type ChatService struct {
	auth *auth.Authenticator
}

type Message struct {
	Token   string
	User    string
	Content string
}

func (c *ChatService) Authenticate(req AuthRequest, resp *AuthResponse) error {
	token, err := c.auth.Authenticate(req.Username, req.Password)

	if err != nil {
		resp.Error = err.Error()
		return nil
	}

	resp.Token = token.Value
	log.Printf("[AUTH] User %q authenticated, token issued", req.Username)
	return nil
}

func (c *ChatService) SendMessage(msg Message, reply *string) error {
	username, err := c.auth.Validate(msg.Token)
	if err != nil {
		return fmt.Errorf("Authentication required: %v", err)
	}

	if msg.User != username {
		return fmt.Errorf("user mismatch: token is for %q, but message claims %q", username, msg.User)
	}

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

	authenticator := auth.NewAuthenticator("dev-secret-key")

	chat := &ChatService{
		auth: authenticator,
	}

	err = rpc.Register(chat)

	if err != nil {
		log.Fatal("RPC registration failed:", err)
	}

	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")

	if err != nil {
		log.Fatalf("Failde to load TLS certification %v\n"+"Run ./generate-cert.sh first to create certs/", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", ":8080", tlsConfig)
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
