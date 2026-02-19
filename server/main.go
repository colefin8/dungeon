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

const MAX_NUM_CONNECTIONS = math.MaxUint16

var logFile *os.File

type Client struct {
	conn            *net.Conn
	isLoggedIn      bool
	username        string
	currentPos      world.Pos
	hasTornMeniscus bool
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
		shared.LOG_FILE,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		os.ModeAppend,
	)
	if err != nil {
		log("could not open log file: %v", err)
	}
	defer logFile.Close()

	// open socket
	err = os.Remove(shared.SOCKET_PATH)
	if err != nil {
		log("ERROR: could not remove existing socket file '%s': %v", shared.SOCKET_PATH, err)
	}
	listener, err := net.Listen("unix", shared.SOCKET_PATH)
	if err != nil {
		log("ERROR: could not listen on unix socket '%s' %v", shared.SOCKET_PATH, err)
	}
	defer listener.Close()
	defer os.Remove(shared.SOCKET_PATH)
	err = os.Chmod(shared.SOCKET_PATH, 0777)
	if err != nil {
		log("ERROR: Could not chmod socket file '%s': %v", shared.SOCKET_PATH, err)
		os.Exit(1)
	}
	log("\n\nListening on %s", shared.SOCKET_PATH)

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
