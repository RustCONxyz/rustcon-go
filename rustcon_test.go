package rustcon

import (
	"testing"
)

func TestConfig(t *testing.T) {
	connOptions := RconConnectionOptions{
		IP:       "127.0.0.1",
		Port:     28016,
		Password: "password",
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
