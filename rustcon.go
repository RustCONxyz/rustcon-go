package rustcon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/gorilla/websocket"
)

type RconConnection struct {
	IP             string
	Port           int
	Password       string
	ws             *websocket.Conn
	OnConnected    func()
	OnMessage      func(*Message)
	OnChatMessage  func(*ChatMessage)
	OnDisconnected func()
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
	Message := Message{}
	err := json.Unmarshal(data, &Message)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return
	}

	if Message.Type == "Chat" && r.OnChatMessage != nil {
		chatMessage := ChatMessage{}
		err := json.Unmarshal([]byte(Message.Message), &chatMessage)
		if err != nil {
			fmt.Println("Error unmarshalling chat message:", err)
			return
		}

		r.OnChatMessage(&chatMessage)
		return
	}

	if r.OnMessage != nil {
		r.OnMessage(&Message)
	}
}
