package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"

	"dungeon/shared"
)

const SOCKET = "/home/dungeon/.dungeon.sock"
const MAX_NUM_CONNECTIONS = math.MaxUint16

type Client struct {
	conn     *net.Conn
	username string
}

var broadcast = make(chan []byte)
var clients = make(map[*net.Conn]Client)
var clientsMu sync.Mutex

func main() {
	os.Remove(SOCKET)

	listener, err := net.Listen("unix", SOCKET)
	if err != nil {
		fmt.Printf("ERROR: could not listen on unix socket %s %e\n", SOCKET, err)
	}
	defer listener.Close()
	defer os.Remove(SOCKET)

	err = os.Chmod(SOCKET, 0777)
	if err != nil {
		fmt.Printf("Could not chmod socket file: %e\n", err)
		os.Exit(1)
	}

	fmt.Println("Listening on ", SOCKET)

	go startBroadcaster()

	for {
		conn, _ := listener.Accept()
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Someone connected!")

	clientsMu.Lock()
	client := Client{
		conn:     &conn,
		username: "",
	}
	clients[&conn] = client
	clientsMu.Unlock()

	// get input from user
	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if fmt.Sprintf("%v", err) == "EOF" {
				fmt.Printf("\x1b[34m%s\x1b[0m has left the dungeon...\n", client.username)
				username := client.username[:]
				removeClient(&conn)

				// send logout message to other clients
				data := []byte{shared.ResponseTypeLogout}
				data = binary.LittleEndian.AppendUint16(data, uint16(len(username)))
				data = append(data, []byte(client.username)...)
				broadcast <- data

				break
			} else {
				fmt.Printf("could not read from user %v", err)
				continue
			}
		}
		msgType := data[0]
		line := string(data[1:])
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch msgType {
		case shared.RequestTypeLogin:
			client.username = line
			fmt.Printf("%s has joined!\n", client.username)
			data := []byte{shared.ResponseTypeLogin}
			data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
			data = append(data, []byte(client.username)...)
			broadcast <- data
			// client.username = line
		case shared.RequestTypeSay:
			data := []byte{shared.ResponseTypeSay}
			data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
			data = append(data, []byte(client.username)...)
			data = binary.LittleEndian.AppendUint16(data, uint16(len(line)))
			data = append(data, []byte(line)...)
			broadcast <- data
		}
	}

	removeClient(&conn)
}

func startBroadcaster() {
	for msg := range broadcast {
		clientsMu.Lock()
		for conn := range clients {
			_, err := (*conn).Write(append(msg, '\n'))
			if err != nil {
				fmt.Printf("ERROR: could not broadcast to client: %v\n", err)
			}
		}
		clientsMu.Unlock()
	}
}

func removeClient(conn *net.Conn) {
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}
