package rustcon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type RconConnection struct {
	IP                      string
	Port                    int
	Password                string
	ws                      *websocket.Conn
	OnConnected             func()
	OnMessage               func(*GenericMessage)
	OnChatMessage           func(*ChatMessage)
	OnDisconnected          func()
}

type RconConnectionOptions struct {
	IP                      string
	Port                    int
	Password                string
	OnConnected             func()
	OnMessage               func(*GenericMessage)
	OnChatMessage           func(*ChatMessage)
	OnDisconnected          func()
}

type GenericMessage struct {
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

func NewRconConnection(options RconConnectionOptions) *RconConnection {
	return &RconConnection{
		IP:                      options.IP,
		Port:                    options.Port,
		Password:                options.Password,
		OnConnected:             options.OnConnected,
		OnMessage:               options.OnMessage,
		OnChatMessage:           options.OnChatMessage,
		OnDisconnected:          options.OnDisconnected,
	}
}

func (r *RconConnection) Connect() error {
	if net.ParseIP(r.IP) == nil {
		return errors.New("invalid IP address")
	}

	if r.Port < 1 || r.Port > 65535 {
		return errors.New("invalid port number")
	}

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", r.IP, r.Port), Path: fmt.Sprintf("/%s", r.Password)}

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	r.ws = ws

	if r.OnConnected != nil {
		r.OnConnected()
	}

	go r.readPump()

	return nil
}

func (r *RconConnection) SendCommand(command string) error {
	if r.ws == nil {
		return errors.New("not connected to server")
	}

	msg := &Command{
		Identifier: 0,
		Message:    command,
		Name:       "RCON",
	}

	err := r.ws.WriteJSON(msg)
	if err != nil {
		return err
	}

	return nil
}

func (r *RconConnection) Disconnect() error {
	if r.ws == nil {
		return errors.New("not connected to server")
	}

	err := r.ws.Close()
	if err != nil {
		return err
	}

	if r.OnDisconnected != nil {
		r.OnDisconnected()
	}

	return nil
}

func (r *RconConnection) readPump() {
	defer r.Disconnect()

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
	genericMessage := GenericMessage{}
	err := json.Unmarshal(data, &genericMessage)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return
	}

	if genericMessage.Type == "Chat" && r.OnChatMessage != nil {
		chatMessage := ChatMessage{}
		err := json.Unmarshal([]byte(genericMessage.Message), &chatMessage)
		if err != nil {
			fmt.Println("Error unmarshalling chat message:", err)
			return
		}

		r.OnChatMessage(&chatMessage)
		return
	}

	if r.OnMessage != nil {
		r.OnMessage(&genericMessage)
	}
}
