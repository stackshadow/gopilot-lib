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
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

const recvBufferSize int = 2048

// SocketConnection represent an current socket-session ( socket connection )
type SocketConnection struct {
	logging *logrus.Entry
	ID      string
	socket  net.Conn // our socket
	lastID  int
}

// Init will init an socket-connection
func (con *SocketConnection) Init(socket net.Conn, id string) {

	con.ID = id
	con.socket = socket
	con.logging = logrus.WithFields(
		logrus.Fields{
			"prefix": "SCON",
			"conID":  con.ID,
		},
	)

	con.logging.Debug("New Socket-connection created")
}

// ReadMessage will call the onMessage if an message is recieved
// this function is synchron ( blocked if no message is aviable ! )
func (con *SocketConnection) ReadMessage() (Msg, error) {

	con.logging.Debug("Wait for message")

	jsonString, _ := bufio.NewReader(con.socket).ReadString('\n')
	con.logging.WithFields(logrus.Fields{
		"raw": jsonString,
	},
	).Debug("Raw message")

	newMessage, err := FromJSONString(jsonString)
	if err != nil {
		con.logging.Error(err)
		return Msg{}, err
	}

	// we tag the message with our connection id, so that we WONT send it out again
	newMessage.id = con.lastID
	newMessage.context = con.ID

	// iterate id
	con.lastID = con.lastID + 1

	// debug
	con.logging.WithFields(logrus.Fields{
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
func (con *SocketConnection) SendMessage(message Msg) {

	// we don't send messages that comes from us
	if message.context == con.ID {
		con.logging.WithFields(logrus.Fields{
			"msgID": message.id,
		},
		).Debug("We dont send to sender")

		return
	}

	// iterate id
	message.id = con.lastID
	con.lastID = con.lastID + 1

	// convert to string
	newMessageString, _ := message.ToJSONString()

	// debug
	con.logging.WithFields(logrus.Fields{
		"msgID":       message.id,
		"source":      message.NodeSource,
		"sourceGroup": message.GroupSource,
		"target":      message.NodeTarget,
		"targetGroup": message.GroupTarget,
		"command":     message.Command,
	},
	).Debug("Send Message")

	// send it
	fmt.Fprintf(con.socket, "%s\n", newMessageString)
}

// Close will close the current socket connection
func (con *SocketConnection) Close() {

	// debug
	con.logging.Info("Close connection")
	con.socket.Close()
}

func (con *SocketConnection) onPublish(message *Msg, group, command, payload string) {
	con.SendMessage(*message)
}
