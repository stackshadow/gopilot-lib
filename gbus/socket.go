/*
Copyright (C) 2019 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, dersion 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

// This package provide the Socket-Server

package gbus

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.com/gopilot/lib/mynodename"
)

const recvBufferSize int = 2048

// SocketConnection represent an current socket-session ( socket connection )
type SocketConnection struct {
	log             *logrus.Entry
	id              string
	socket          net.Conn // our socket
	lastMessageID   int
	remoteNodeName  string
	remoteNodeGroup string
}

// SocketCallbacks provide different callbacks
// OnConnect - Will fire if you successful connect to an socket ( after handshake )
type SocketCallbacks struct {
	OnConnect           func(socket *SocketConnection)
	OnDisconnect        func(socket *SocketConnection)
	OnHandshakeFinished func(socket *SocketConnection)
	OnMessage           func(socket *SocketConnection, message Msg)
}

// SocketNew create a new Socket
func SocketNew() (socket *SocketConnection) {

	var newSocket SocketConnection

	UUID, err := uuid.NewRandom()
	if err != nil {
		newSocket.id = "0"
	} else {
		newSocket.id = UUID.String()
	}

	newSocket.log = logrus.WithFields(
		logrus.Fields{
			"prefix":   "SOCKET",
			"socketID": newSocket.id,
		},
	)

	newSocket.log.Debug("New Socket-connection created")
	return &newSocket
}

// ID return the connection-id
func (socket *SocketConnection) ID() string {
	return socket.id
}

// RemoteNodeName return the remote-node-name
//
// This is only useful on a client
func (socket *SocketConnection) RemoteNodeName() string {
	return socket.remoteNodeName
}

// RemoteNodeGroup return the remote-group-name
//
// This is only useful on a client
func (socket *SocketConnection) RemoteNodeGroup() string {
	return socket.remoteNodeGroup
}

// ReadMessage will call the onMessage if an message is recieved
// this function is synchron ( blocked if no message is aviable ! )
func (socket *SocketConnection) ReadMessage() (Msg, error) {

	socket.log.Debug("Wait for message")

	jsonString, _ := bufio.NewReader(socket.socket).ReadString('\n')
	socket.log.WithFields(logrus.Fields{
		"raw": jsonString,
	},
	).Debug("Raw message")

	newMessage, err := FromJSONString(jsonString)
	if err != nil {
		socket.log.Error(err)
		return Msg{}, err
	}

	// we tag the message with our connection id, so that we WONT send it out again
	newMessage.id = socket.lastMessageID
	newMessage.context = socket.ID

	// iterate id
	socket.lastMessageID = socket.lastMessageID + 1

	// debug
	socket.log.WithFields(logrus.Fields{
		"msgID":       newMessage.id,
		"source":      newMessage.NodeSource,
		"sourceGroup": newMessage.GroupSource,
		"target":      newMessage.NodeTarget,
		"targetGroup": newMessage.GroupTarget,
		"command":     newMessage.Command,
	},
	).Debug("Recieved Message")

	return newMessage, nil
}

// SendMessage will send a message over the socket connection
func (socket *SocketConnection) SendMessage(message Msg) {

	// we don't send messages that comes from us
	if message.context == socket.ID() {
		socket.log.WithFields(logrus.Fields{
			"msgID": message.id,
		},
		).Debug("We dont send to sender")

		return
	}

	// iterate id
	message.id = socket.lastMessageID
	socket.lastMessageID = socket.lastMessageID + 1

	// convert to string
	newMessageString, _ := message.ToJSONString()

	// debug
	socket.log.WithFields(logrus.Fields{
		"msgID":       message.id,
		"source":      message.NodeSource,
		"sourceGroup": message.GroupSource,
		"target":      message.NodeTarget,
		"targetGroup": message.GroupTarget,
		"command":     message.Command,
	},
	).Debug("Send Message")

	// send it
	fmt.Fprintf(socket.socket, "%s\n", newMessageString)
}

// Close will close the current socket connection
func (socket *SocketConnection) close() {

	// debug
	socket.log.Info("Close connection")
	socket.socket.Close()
}

// Serve [BLOCKING] will start the socket-server and run forever until an error occure
// if an connection occure and the handshake was okay, an goroutine will be started which handle incoming messages
func (socket *SocketConnection) Serve(filename string, cb SocketCallbacks) error {

	socket.log = socket.log.WithField("type", "server")

	// remove socket if it already exists
	if err := os.RemoveAll(filename); err != nil {
		socket.log.Error(err)
		return err
	}

	// open the socket
	var err error
	var serverListener net.Listener
	serverListener, err = net.Listen("unix", filename)
	if err != nil {
		socket.log.Error(err)
		return err
	}

	socket.log.Info(fmt.Sprintf("Create SOCKET on %s", filename))

	// wait for new clients
	for {

		// wait for a new client
		socket.log.Debug("Wating for new client")
		newSocketCon, err := serverListener.Accept()
		if err != nil {
			socket.log.Error(err)
			return err
		}

		// create a new session
		// the filter is empty, as server we accept every message
		newSocket := SocketNew()
		newSocket.socket = newSocketCon

		// ################################# handshake #################################
		// send a helo to the client
		newSocket.SendMessage(Msg{
			NodeSource:  mynodename.NodeName, // i'am the source
			GroupSource: "",                  // i hear on every group
			NodeTarget:  "",                  // i don't know you, so is just send it to all
			GroupTarget: "",                  // i don't know your group ( yet )
			Command:     "HELO",
		})

		// we wait for OLEH
		newSocket.log.Debug("Wait for OLEH-Message")
		ehloMessage, err := newSocket.ReadMessage()
		if err != nil {
			newSocket.log.Error(err)
			newSocket.close()
			continue
		}
		if ehloMessage.Command != "OLEH" {
			newSocket.log.Error("No OLEH was recieved")
			newSocket.close()
			continue
		}

		newSocket.remoteNodeName = ehloMessage.NodeSource
		newSocket.remoteNodeGroup = ehloMessage.GroupSource

		newSocket.log.Debug("")
		newSocket.log.Debug("############################ Handshake finished ############################")
		newSocket.log.Debug("")

		// ############################ handshake finished #############################

		// start goroutine which handle incoming messages
		go newSocket.eventLoopWaitForMessage(cb)

		// callback - finished with handshake
		if cb.OnHandshakeFinished != nil {
			cb.OnHandshakeFinished(newSocket)
		}
	}

}

// Connect [BLOCKING] Connect to an existing socket
func (socket *SocketConnection) Connect(filename, listenForNodeName, listenForGroupName string, cb SocketCallbacks) error {

	socket.log = socket.log.WithField("type", "client")

	var err error

	// this runs forever
	for {

		// wait for connections
		for {
			socket.socket, err = net.Dial("unix", filename)
			if err != nil {
				socket.log.Error(err)
				time.Sleep(10 * time.Second)
			}

			if socket.socket != nil {
				break
			}
		}

		// ################################# handshake #################################
		// we wait for HELO
		socket.log.Debug("Wait for HELO-Message")
		heloMessage, err := socket.ReadMessage()
		if err != nil {
			socket.close()
			socket.log.Error(err)
			return err
		}
		if heloMessage.Command != "HELO" {
			socket.close()
			errNoHelo := errors.New("No OLEH was recieved")
			socket.log.Error(errNoHelo)
			return errNoHelo
		}

		socket.remoteNodeName = heloMessage.NodeSource
		socket.remoteNodeGroup = heloMessage.GroupSource

		// and informate the server about what we listen
		socket.SendMessage(Msg{
			NodeSource:  listenForNodeName,
			GroupSource: listenForGroupName,
			NodeTarget:  heloMessage.NodeTarget,
			GroupTarget: "",
			Command:     "OLEH",
		})

		// callback - connected
		if cb.OnConnect != nil {
			cb.OnConnect(socket)
		}

		socket.log.Debug("")
		socket.log.Debug("############################ Handshake finished ############################")
		socket.log.Debug("")
		// ############################ handshake finished #############################

		// callback - finished with handshake
		if cb.OnHandshakeFinished != nil {
			cb.OnHandshakeFinished(socket)
		}

		socket.eventLoopWaitForMessage(cb)
	}
}

func (socket *SocketConnection) eventLoopWaitForMessage(cb SocketCallbacks) {

	// close and remove session if disconnect or error occure
	defer func() {
		socket.log.Debugf("Close '%s'", socket.ID())
		socket.close()
		cb.OnDisconnect(socket)
	}()

	// message-loop
	for {
		message, err := socket.ReadMessage()
		if err != nil {
			socket.log.Error(err)
			break
		}

		cb.OnMessage(socket, message)
	}

}
