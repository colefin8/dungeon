package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"dungeon/server/world"
	"dungeon/shared"
)

const SOCKET = "/home/dungeon/.dungeon.sock"
const LOG_FILE = "/home/dungeon/.dungeon.log"
const MAX_NUM_CONNECTIONS = math.MaxUint16

var logFile *os.File

type Client struct {
	conn       *net.Conn
	username   string
	currentPos world.Pos
	isLoggedIn bool
}

var broadcast = make(chan []byte)
var clients = make(map[*net.Conn]*Client)
var clientsMu sync.Mutex

// 1. open log file
// 2. open socket
// 3. create world (cells)
// (world is now ready for players)
// 4. listen for player connections
func main() {
	// open log file
	var err error
	logFile, err = os.OpenFile(
		LOG_FILE,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		os.ModeAppend,
	)
	if err != nil {
		log("could not open log file: %v", err)
	}
	defer logFile.Close()

	// open socket
	err = os.Remove(SOCKET)
	if err != nil {
		log("ERROR: could not remove existing socket file '%s': %v", SOCKET, err)
	}
	listener, err := net.Listen("unix", SOCKET)
	if err != nil {
		log("ERROR: could not listen on unix socket '%s' %v", SOCKET, err)
	}
	defer listener.Close()
	defer os.Remove(SOCKET)
	err = os.Chmod(SOCKET, 0777)
	if err != nil {
		log("ERROR: Could not chmod socket file '%s': %v", SOCKET, err)
		os.Exit(1)
	}
	log("\n\nListening on %s", SOCKET)

	world.CreateWorld()

	// start listening for player connections
	go startUpdatingLoggedInUsers()
	go startBroadcaster()
	for {
		conn, _ := listener.Accept()
		go HandleClient(&conn)
	}
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
	(*conn).Close()
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
