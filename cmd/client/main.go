package main

import (
	"crypto/tls"
	"fmt"
	"net/rpc"
	"bufio"
	"os"
	"strings"
	"time"
)

type RegisterRequest struct {
	Username	string
	Password	string
}

type RegisterResponse struct {
	Error			string
}

type Message struct {
	Token   string
	User    string
	Content string
}
type AuthRequest struct {
		Username string
		Password string
	}

type AuthResponse struct {
		Token string
		Error string
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


func main() {

	reader := bufio.NewReader(os.Stdin)

	

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "localhost:8080", tlsConfig)

	if err != nil {
		fmt.Println("Failed to connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := rpc.NewClient(conn)
	defer client.Close()
	
	fmt.Println("====RPC CHAT====")
	fmt.Println("Connected to server via TLS")
	fmt.Println("1. Login")
	fmt.Println("2. Register")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')

	if choice == "2" {
		var regResp RegisterResponse
		err = client.Call("ChatService.Register", RegisterRequest{
			Username: username, Password: password,
		}, &regResp)
		if err != nil {
			fmt.Println("Register RPC failed", err)
			os.Exit(1)
		}
		if regResp.Error != "" {
			fmt.Println("Registration failed:", regResp.Error)
			os.Exit(1)
		}
		fmt.Println("Registered! Now logging in...")
	}

	var authResp AuthResponse

	err = client.Call("ChatService.Authenticate", AuthRequest{
		Username: username, Password: password,
	}, &authResp)
	if err != nil {
		fmt.Println("Authenticate RPC failed:", err)
		os.Exit(1)
	}
	if authResp.Error != "" {
		fmt.Println("Login failed:", authResp.Error)
		os.Exit(1)
	}

	token := authResp.Token
	fmt.Printf("\nLogged in as %s. Type a message and hit Enter. Type /quit to exit. \n\n",username)
	go func(){
		sinceIndex := 0
		for{
			time.Sleep(1 * time.Second)
			var getResp GetMessagesResponse
			err := client.Call("ChatService.GetMessages", GetMessagesRequest{
				Token: token, SinceIndex: sinceIndex,
			}, &getResp)

			if err != nil{
				fmt.Printf("\n[poll error] %v\nYou > ", err)
				continue
			} 
			if getResp.Error != "" {
				fmt.Printf("\n[poll error] %s\nYou > ", err)
				continue
			}


			for _, m := range getResp.Messages {
				if m.From != username {
					fmt.Printf("\r[%s] %s\nYou>", m.From, m.Content)
				}
			}
			sinceIndex = getResp.NextIndex
		}
	}()

	for{
		fmt.Print("You> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "/quit" {
			fmt.Println("Goodbye!")
			break
		}
		var reply string
		err = client.Call("ChatService.SendMessage", Message{
			Token: token, User: username, Content: line,
		}, &reply)
		if err != nil {
			fmt.Println("Send failed:", err)
		}
	}

	}
