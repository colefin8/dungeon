package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"dungeon/server/world"
	"dungeon/shared"
)

func HandleClient(conn *net.Conn) {
	defer removeClient(conn)

	log("someone connected!")

	// add user to client list
	clientsMu.Lock()
	client := &Client{
		conn:       conn,
		username:   "",
		currentPos: world.CenterPos,
		isLoggedIn: false,
	}
	clients[conn] = client
	clientsMu.Unlock()

	// get input from user
	reader := bufio.NewReader(*conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if fmt.Sprintf("%v", err) == "EOF" { // Go error checking is so dumb
				if !client.isLoggedIn {
					log("a user exited from the login screen")
					break
				} else {
					log("%s has left the dungeon", client.username)
				}

				// send logout message to other clients
				data := []byte{shared.ResponseTypeLogout}
				data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
				data = append(data, []byte(client.username)...)
				removeClient(conn)
				broadcast <- data

				updateLoggedInUsers <- true
				break
			} else {
				log("ERROR: could not read from user '%s': %v", client.username, err)
				continue
			}
		}
		msgType := data[0]
		switch msgType {
		case shared.RequestTypeLogin:
			clientsMu.Lock()
			client.username = strings.TrimSpace(string(data[1:]))
			client.isLoggedIn = true
			clientsMu.Unlock()
			log("%s has joined!", client.username)
			data := []byte{shared.ResponseTypeLogin}
			data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
			data = append(data, []byte(client.username)...)
			broadcast <- data
			updateLoggedInUsers <- true
		case shared.RequestTypeSay:
			txt := strings.TrimSpace(string(data[1:]))
			data := []byte{shared.ResponseTypeSay}
			data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
			data = append(data, []byte(client.username)...)
			data = binary.LittleEndian.AppendUint16(data, uint16(len(txt)))
			data = append(data, []byte(txt)...)
			broadcast <- data
		case shared.RequestTypeWho:
			writeDataToClient(conn, loggedInUsersBin)
		case shared.RequestTypeLook:
			data = []byte{shared.ResponseTypeLook}
			c := world.Cells[client.currentPos.Hash()]
			data = binary.LittleEndian.AppendUint16(data, uint16(len(c.Title)))
			data = append(data, []byte(c.Title)...)
			data = binary.LittleEndian.AppendUint16(data, uint16(len(c.Description)))
			data = append(data, []byte(c.Description)...)
			writeDataToClient(conn, data)
		}
	}
}
