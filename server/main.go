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
	"time"

	"dungeon/server/cell"
	"dungeon/shared"
)

const SOCKET = "/home/dungeon/.dungeon.sock"
const LOG_FILE = "/home/dungeon/.dungeon.log"
const MAX_NUM_CONNECTIONS = math.MaxUint16

var logFile *os.File

type Client struct {
	conn       *net.Conn
	username   string
	currentPos WorldPos
	isLoggedIn bool
}

type WorldPos struct {
	X uint16
	Y uint16
	Z uint16
}

func (p WorldPos) Hash() uint64 {
	return (uint64(p.Z) << 32) | (uint64(p.Y) << 16) | uint64(p.X)
}

var cells = make(map[uint64]cell.Cell)

var broadcast = make(chan []byte)
var clients = make(map[*net.Conn]*Client)
var clientsMu sync.Mutex

func main() {
	var err error

	go startUpdatingLoggedInUsers()

	logFile, err = os.OpenFile(LOG_FILE, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModeAppend)
	if err != nil {
		log("could not open log file: %v", err)
	}
	defer logFile.Close()

	cells[WorldPos{X: 0, Y: 0, Z: 0}.Hash()] = cell.Cell{
		Title:       "The Chapel",
		Description: "You wonder as to why there is a chapel deep beneath the earth, where the stone corridors of the dungeon wind like the roots of a great oak tree, yet here exists one, hewn from the living rock. It is not so vast as the cathedrals of great cities, yet neither is it small; its vaulted ceiling rises high enough that a tall banner might hang untroubled, and its nave would seat a modest gathering without crowding. The air within is cool and still, carrying the faint scent of old incense long settled into the stone. Its wooden pews that line the two sides of the room have seen better days, but have also seen days of use recently.",
	}

	os.Remove(SOCKET)
	listener, err := net.Listen("unix", SOCKET)
	if err != nil {
		log("ERROR: could not listen on unix socket %s %v", SOCKET, err)
	}
	defer listener.Close()
	defer os.Remove(SOCKET)

	err = os.Chmod(SOCKET, 0777)
	if err != nil {
		log("ERROR: Could not chmod socket file: %v", err)
		os.Exit(1)
	}

	// extra newline to separate running sessions
	log("\nListening on %s", SOCKET)

	go startBroadcaster()

	for {
		conn, _ := listener.Accept()
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	log("Someone connected!")

	clientsMu.Lock()
	client := &Client{
		conn:       &conn,
		username:   "",
		currentPos: WorldPos{0, 0, 0},
		isLoggedIn: false,
	}
	clients[&conn] = client
	clientsMu.Unlock()

	// get input from user
	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if fmt.Sprintf("%v", err) == "EOF" { // Go error checking is so dumb
				log("%s has left the dungeon", client.username)

				if !client.isLoggedIn {
					break
				}

				// send logout message to other clients
				data := []byte{shared.ResponseTypeLogout}
				data = binary.LittleEndian.AppendUint16(data, uint16(len(client.username)))
				data = append(data, []byte(client.username)...)
				removeClient(&conn)
				broadcast <- data

				updateLoggedInUsers <- true
				break
			} else {
				log("ERROR: could not read from user %s: %v", client.username, err)
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
			writeDataToClient(&conn, loggedInUsersBin)
		case shared.RequestTypeLook:
			data = []byte{shared.ResponseTypeLook}
			c := cells[client.currentPos.Hash()]
			data = binary.LittleEndian.AppendUint16(data, uint16(len(c.Title)))
			data = append(data, []byte(c.Title)...)
			data = binary.LittleEndian.AppendUint16(data, uint16(len(c.Description)))
			data = append(data, []byte(c.Description)...)
			writeDataToClient(&conn, data)
		}
	}

	removeClient(&conn)
}

var updateLoggedInUsers = make(chan bool, 1)
var loggedInUsersBin = []byte{shared.ResponseTypeLoggedInUsers, 0x00, 0x00}

func startUpdatingLoggedInUsers() {
	for range updateLoggedInUsers {
		loggedInUsersBin = []byte{shared.ResponseTypeLoggedInUsers}
		var numUsers uint16 = 0
		usernamesBin := []byte{}
		clientsMu.Lock()
		for _, client := range clients {
			if !client.isLoggedIn {
				continue
			}
			numUsers++
			usernamesBin = binary.LittleEndian.AppendUint16(usernamesBin, uint16(len(client.username)))
			usernamesBin = append(usernamesBin, []byte(client.username)...)
		}
		clientsMu.Unlock()
		loggedInUsersBin = binary.LittleEndian.AppendUint16(loggedInUsersBin, numUsers)
		loggedInUsersBin = append(loggedInUsersBin, usernamesBin...)

		// show live list of logged in users to all clients on the login page
		for conn, client := range clients {
			if client.isLoggedIn {
				continue
			}
			fmt.Println("found a not logged in user!!!!")
			_, err := writeDataToClient(conn, loggedInUsersBin)
			if err != nil {
				log("ERROR: could not write to client %s: %v", client.username, err)
			}
		}
	}
}

func startBroadcaster() {
	for msg := range broadcast {
		clientsMu.Lock()
		for conn, client := range clients {
			if !(*client).isLoggedIn {
				continue
			}
			_, err := writeDataToClient(conn, msg)
			if err != nil {
				log("ERROR: could not broadcast to client: %v", err)
			}
		}
		clientsMu.Unlock()
	}
}

// Writes data to a client connection, encoded as the length of the data as a little-endian 16-bit number, followed by the data itself.
func writeDataToClient(conn *net.Conn, data []byte) (int, error) {
	lenData := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenData, uint16(len(data)))
	data = append(lenData, data...)
	return (*conn).Write(data)
}

func removeClient(conn *net.Conn) {
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

func log(msg string, args ...any) {
	if logFile == nil {
		fmt.Printf(msg+"\n", args...)
	} else {
		nowFmt := time.Now().Local().Format(time.UnixDate)
		fmt.Fprint(logFile, nowFmt+" - ")
		fmt.Fprintf(logFile, msg+"\n", args...)
	}
}
