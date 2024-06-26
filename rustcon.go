package rustcon

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type RconConnection struct {
	IP              string
	Port            int
	Password        string
	ws              *websocket.Conn
	pendingMessages map[int]chan *Message
	OnConnected     func()
	OnMessage       func(*Message)
	OnChatMessage   func(*ChatMessage)
	OnDisconnected  func()
}

type RconConnectionOptions struct {
	IP             string
	Port           int
	Password       string
	OnConnected    func()
	OnMessage      func(*Message)
	OnChatMessage  func(*ChatMessage)
	OnDisconnected func()
}

type Message struct {
	Message    string `json:"Message"`
	Identifier int    `json:"Identifier"`
	Type       string `json:"Type"`
	Stacktrace string `json:"Stacktrace"`
}

type ChatMessage struct {
	Channel  int64  `json:"Channel"`
	Message  string `json:"Message"`
	UserId   string `json:"UserId"`
	Username string `json:"Username"`
	Color    string `json:"Color"`
	Time     int64  `json:"Time"`
}

type Command struct {
	Identifier int    `json:"Identifier"`
	Message    string `json:"Message"`
	Name       string `json:"Name"`
}

func NewRconConnection(options RconConnectionOptions) (*RconConnection, error) {
	if options.IP == "" || net.ParseIP(options.IP) == nil {
		return nil, errors.New("invalid IP address")
	}
	if options.Port < 1 || options.Port > 65535 {
		return nil, errors.New("invalid port number")
	}
	if options.Password == "" {
		return nil, errors.New("password cannot be empty")
	}

	conn := &RconConnection{
		IP:             options.IP,
		Port:           options.Port,
		Password:       options.Password,
		OnConnected:    options.OnConnected,
		OnMessage:      options.OnMessage,
		OnChatMessage:  options.OnChatMessage,
		OnDisconnected: options.OnDisconnected,
	}

	return conn, nil
}

func (r *RconConnection) Connect() error {
	if r.IP == "" || net.ParseIP(r.IP) == nil {
		return errors.New("invalid IP address")
	}
	if r.Port < 1 || r.Port > 65535 {
		return errors.New("invalid port number")
	}
	if r.Password == "" {
		return errors.New("password cannot be empty")
	}

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", r.IP, r.Port), Path: fmt.Sprintf("/%s", r.Password)}

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	r.ws = ws
	r.pendingMessages = make(map[int]chan *Message)

	if r.OnConnected != nil {
		r.OnConnected()
	}

	go r.readPump()

	return nil
}

func (r *RconConnection) SendCommand(command string) (*Message, error) {
	if r.ws == nil {
		return nil, errors.New("not connected to server")
	}

	identifier := r.generateIdentifier()

	msg := &Command{
		Identifier: identifier,
		Message:    command,
		Name:       "RCON",
	}

	r.pendingMessages[identifier] = make(chan *Message)

	err := r.ws.WriteJSON(msg)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-r.pendingMessages[identifier]:
		delete(r.pendingMessages, identifier)
		return resp, nil
	case <-time.After(10 * time.Second):
		delete(r.pendingMessages, identifier)
		return nil, errors.New("timeout waiting for response")
	}
}

func (r *RconConnection) Disconnect() error {
	if r.ws == nil {
		return errors.New("not connected to server")
	}

	for identifier, respChan := range r.pendingMessages {
		close(respChan)
		delete(r.pendingMessages, identifier)
	}

	err := r.ws.Close()
	if err != nil {
		return err
	}

	r.ws = nil

	if r.OnDisconnected != nil {
		r.OnDisconnected()
	}

	return nil
}

func (r *RconConnection) readPump() {
	for {
		_, message, err := r.ws.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		go r.handleMessage(message)
	}
}

func (r *RconConnection) handleMessage(data []byte) {
	message := Message{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return
	}

	if respChan, ok := r.pendingMessages[message.Identifier]; ok {
		respChan <- &message
	}

	if message.Type == "Chat" && r.OnChatMessage != nil {
		chatMessage := ChatMessage{}
		err := json.Unmarshal([]byte(message.Message), &chatMessage)
		if err != nil {
			fmt.Println("Error unmarshalling chat message:", err)
			return
		}

		r.OnChatMessage(&chatMessage)
		return
	}

	if r.OnMessage != nil {
		r.OnMessage(&message)
	}
}

func (r *RconConnection) generateIdentifier() int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := rng.Intn(1000) + 1

	for r.pendingMessages[randomNumber] != nil {
		randomNumber = rng.Intn(1000) + 1
	}

	return randomNumber
}
