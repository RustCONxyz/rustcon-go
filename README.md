# RustCON Go

RustCON Go is a Go library for interacting with Rust servers via WebSockets (Web RCON)

## Links

- [Website](https://rustcon.xyz)
- [Documentation](https://rustcon.xyz/developers)
- [GitHub](https://github.rustcon.xyz/)
- [Support](https://support.rustcon.xyz/)

## Installation

```bash
go get github.com/RustCONxyz/rustcon-go
```

## Example Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/RustCONxyz/rustcon-go"
)

func main() {
    connection := &rustcon.RconConnection{
        IP:       "127.0.0.1",
        Port:     28016,
        Password: "password",
        OnConnected: func() {
            fmt.Println("Connected to RCON")
        },
        OnMessage: func(message *rustcon.Message) {
            fmt.Println(message.Message)
        },
        OnChatMessage: func(chatMessage *rustcon.ChatMessage) {
            fmt.Println(chatMessage.Message)
        },
        OnDisconnected: func() {
            fmt.Println("Disconnected from RCON")
        },
    }

    if err := connection.Connect(); err != nil {
        log.Fatal(err)
    }

    serverInfo, err := connection.SendCommand("serverinfo")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(serverInfo.Message)
}
```

## License

[MIT](/LICENSE)
