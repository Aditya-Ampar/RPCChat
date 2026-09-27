package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/rpc"
	"regexp"
	"strings"

	"github.com/Aditya-Ampar/RPCChat/internal/auth"
	"github.com/Aditya-Ampar/RPCChat/internal/database"
)

const (
	MaxContentBytes   = 64 * 1024 //1MiB
	MaxUsernameLength = 32
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

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

func validateMessage(msg Message) error {
	if len(msg.User) == 0 || len(msg.User) > MaxUsernameLength {
		return fmt.Errorf("invalid Username length: must be 1-%d characters and underscores allowed")
	}
	if !usernamePattern.MatchString(msg.User) {
		return fmt.Errorf("invalid username: only alphanumeric characters and underscores allowed")
	}

	if len(msg.Content) == 0 {
		return fmt.Errorf("message content cannot be empty")
	}

	if len(msg.Content) > MaxContentBytes {
		return fmt.Errorf("message content too large: %d bytes exceesds limit of %d bytes", len(msg.Content), MaxContentBytes)
	}

	if strings.ContainsRune(msg.Content, '\x00') {
		return fmt.Errorf("message content contains null bytes")
	}
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

	if err := validateMessage(msg); err != nil {
		return fmt.Errorf("validation failed: %v", err)
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
