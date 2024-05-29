package rustcon

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gorilla/websocket"
)

const rconPassword = "password"

var upgrader = websocket.Upgrader{}

func createMockRustServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/"+rconPassword {
			upgrader.Upgrade(w, r, nil)
		}
	}))
}

func TestConfig(t *testing.T) {
	connOptions := RconConnectionOptions{
		IP:       "127.0.0.1",
		Port:     28016,
		Password: rconPassword,
	}

	conn, err := NewRconConnection(connOptions)
	if err != nil {
		t.Error(err)
	}

	if conn.IP != connOptions.IP {
		t.Errorf("expected IP to be %s, got %s", connOptions.IP, conn.IP)
	}
	if conn.Port != connOptions.Port {
		t.Errorf("expected port to be %d, got %d", connOptions.Port, conn.Port)
	}
	if conn.Password != connOptions.Password {
		t.Errorf("expected password to be %s, got %s", connOptions.Password, conn.Password)
	}
}

func TestConnection(t *testing.T) {
	mockServer := createMockRustServer()
	defer mockServer.Close()

	parsedURL, _ := url.Parse(mockServer.URL)

	serverIP := parsedURL.Hostname()
	serverPort, _ := strconv.Atoi(parsedURL.Port())

	conn := RconConnection{
		IP:       serverIP,
		Port:     serverPort,
		Password: rconPassword,
	}

	if err := conn.Connect(); err != nil {
		t.Error(err)
	}
	defer conn.Disconnect()
}
