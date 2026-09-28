package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/rpc"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Aditya-Ampar/RPCChat/internal/auth"
	"github.com/Aditya-Ampar/RPCChat/internal/database"
	"github.com/Aditya-Ampar/RPCChat/internal/ratelimit"
	"golang.org/x/crypto/bcrypt"
)

const (
	MaxContentBytes   = 64 * 1024 //1MiB
	MaxUsernameLength = 32
	MinPasswordLength = 8

	RateLimitPerSecond = 20.0
	RateLimitBurst     = 30.0
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type RegisterRequest struct {
	Username string
	Password string
}

type RegisterResponse struct {
	Error string
}

type AuthRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	Token string
	Error string
}

type ChatService struct {
	auth    *auth.Authenticator
	limiter *ratelimit.Limiter
	db      *sql.DB
	room    *ChatRoom
}

type Message struct {
	Token   string
	User    string
	Content string
}

type StoredMessage struct {
	From      string
	Content   string
	Timestamp time.Time
}

type GetMessagesRequest struct {
	Token      string
	SinceIndex int
}

type GetMessagesResponse struct {
	Messages  []StoredMessage
	NextIndex int
	Error     string
}

type ChatRoom struct {
	mu      sync.Mutex
	history []StoredMessage
}

func (r *ChatRoom) Append(from, content string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history = append(r.history, StoredMessage{From: from, Content: content, Timestamp: time.Now()})
	return len(r.history)
}

func (r *ChatRoom) Since(index int) ([]StoredMessage, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if index < 0 || index > len(r.history) {
		index = 0
	}
	return r.history[index:], len(r.history)
}

func (c *ChatService) Register(req RegisterRequest, resp *RegisterResponse) error {
	if len(req.Username) == 0 || len(req.Username) > MaxUsernameLength || !usernamePattern.MatchString(req.Username) {
		resp.Error = "invalid username"
		return nil
	}

	if len(req.Password) < MinPasswordLength {
		resp.Error = fmt.Sprintf("password must be atleast %d characters", MinPasswordLength)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		resp.Error = "internal error hashing password"
		return nil
	}

	if err := database.CreateUser(c.db, req.Username, string(hash)); err != nil {
		resp.Error = err.Error()
		return nil
	}

	log.Printf("[REGISTER] New user %q registered", req.Username)
	return nil
}

func (c *ChatService) Authenticate(req AuthRequest, resp *AuthResponse) error {
	hash, err := database.GetUserPasswordHash(c.db, req.Username)
	if err != nil {
		resp.Error = "invalid username or password"
		return nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		resp.Error = "invalid username or password"
		return nil
	}

	token, err := c.auth.IssueToken(req.Username)
	if err != nil {
		resp.Error = err.Error()
		return nil
	}

	resp.Token = token.Value
	log.Printf("[AUTH] User %q authenticated", req.Username)
	return nil
}

func validateMessage(user, content string) error {
	if len(user) == 0 || len(user) > MaxUsernameLength {
		return fmt.Errorf("invalid Username length: must be 1-%d characters and underscores allowed",MaxUsernameLength)
	}
	if !usernamePattern.MatchString(user) {
		return fmt.Errorf("invalid username: only alphanumeric characters and underscores allowed")
	}

	if len(content) == 0 {
		return fmt.Errorf("message content cannot be empty")
	}

	if len(content) > MaxContentBytes {
		return fmt.Errorf("message content too large: %d bytes exceesds limit of %d bytes", len(content), MaxContentBytes)
	}

	if strings.ContainsRune(content, '\x00') {
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

	if !c.limiter.Allow(username) {
		return fmt.Errorf("rate limit exceeded: max %.0f calls/sec, please slow down", RateLimitPerSecond)
	}

	if err := validateMessage(msg.User, msg.Content); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	
	c.room.Append(username, msg.Content)
	log.Printf("[%s] %s", msg.User, msg.Content)
	*reply = "Message received"
	return nil
}

func (c *ChatService) GetMessages(req GetMessagesRequest, resp *GetMessagesResponse) error {
	username, err := c.auth.Validate(req.Token)
	if err != nil {
		resp.Error = "Authentication required"
		return nil
	}

	_ = username

	msgs, nextIndex := c.room.Since(req.SinceIndex)
	resp.Messages = msgs
	resp.NextIndex = nextIndex
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
	limiter := ratelimit.NewLimiter(RateLimitBurst, RateLimitPerSecond)
	chat := &ChatService{
		auth:    authenticator,
		limiter: limiter,
		db:      db,
		room:    &ChatRoom{},
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
