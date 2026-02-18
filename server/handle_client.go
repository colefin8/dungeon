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
		conn:            conn,
		username:        "",
		currentPos:      world.CenterPos,
		isLoggedIn:      false,
		hasTornMeniscus: false,
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
			switch strings.ToLower(client.username) {
			case "logan", "logo", "logomonkey", "lo":
				client.hasTornMeniscus = true
			}
			clientsMu.Unlock()
			log("%s has joined!", client.username)
			dataToSend := []byte{shared.ResponseTypeLogin}
			dataToSend = binary.LittleEndian.AppendUint16(dataToSend, uint16(len(client.username)))
			dataToSend = append(dataToSend, []byte(client.username)...)
			broadcast <- dataToSend
			updateLoggedInUsers <- true
		case shared.RequestTypeSay:
			txt := strings.TrimSpace(string(data[1:]))
			dataToSend := []byte{shared.ResponseTypeSay}
			dataToSend = binary.LittleEndian.AppendUint16(dataToSend, uint16(len(client.username)))
			dataToSend = append(dataToSend, []byte(client.username)...)
			dataToSend = binary.LittleEndian.AppendUint16(dataToSend, uint16(len(txt)))
			dataToSend = append(dataToSend, []byte(txt)...)
			broadcast <- dataToSend
		case shared.RequestTypeWho:
			writeDataToClient(conn, loggedInUsersBin)
		case shared.RequestTypeMovement:
			var movementType shared.Direction = data[1]
			// free space where they want to go?
			targetWorldPos := client.currentPos
			switch movementType {
			case shared.DirectionNorth:
				targetWorldPos.Y -= 1
			case shared.DirectionEast:
				targetWorldPos.X += 1
			case shared.DirectionSouth:
				targetWorldPos.Y += 1
			case shared.DirectionWest:
				targetWorldPos.X -= 1
			}
			if _, exists := world.Cells[targetWorldPos.Hash()]; exists && !client.hasTornMeniscus {
				client.currentPos = targetWorldPos
			} else {
				// send bonk message
				continue
			}
			fallthrough
		case shared.RequestTypeLook:
			sendLookResponse(conn, client)
		}
	}
}

func sendLookResponse(conn *net.Conn, client *Client) {
	dataToSend := []byte{shared.ResponseTypeLook}
	c := world.Cells[client.currentPos.Hash()]
	dataToSend = binary.LittleEndian.AppendUint16(dataToSend, uint16(len(c.Title)))
	dataToSend = append(dataToSend, []byte(c.Title)...)
	dataToSend = binary.LittleEndian.AppendUint16(dataToSend, uint16(len(c.Description)))
	dataToSend = append(dataToSend, []byte(c.Description)...)
	writeDataToClient(conn, dataToSend)
}

// Writes data to a client connection, encoded as the length of the data as a little-endian 16-bit number, followed by the data itself.
func writeDataToClient(conn *net.Conn, data []byte) (int, error) {
	lenData := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenData, uint16(len(data)))
	data = append(lenData, data...)
	return (*conn).Write(data)
}
