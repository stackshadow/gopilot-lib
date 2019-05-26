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
	"github.com/stackshadow/gopilot-lib/clog"
	"net"
)

// SocketFileName is the default socket-filename
const SocketFileName string = "/tmp/gopilot.sock"
const recvBufferSize int = 2048

// SocketConnection represent an current socket-session ( socket connection )
type SocketConnection struct {
	logging clog.Logger
	ID      string
	socket  net.Conn // our socket
	lastID  int
}

// Init will init an socket-connection
func (s *SocketConnection) Init(socket net.Conn, filter Msg, id string) {

	s.logging = clog.New(id)
	s.ID = id
	s.socket = socket

	s.logging.Debug(fmt.Sprintf("Created new session with filter %+v", filter))
}

// ReadMessage will call the onMessage if an message is recieved
// this function is synchron ( blocked if no message is aviable ! )
func (s *SocketConnection) ReadMessage() (Msg, error) {

	s.logging.Debug(fmt.Sprintf("Wait for message [%s]", s.socket.RemoteAddr().Network()))

	jsonString, _ := bufio.NewReader(s.socket).ReadString('\n')
	s.logging.Debug(jsonString)

	newMessage, err := FromJSONString(jsonString)
	if err != nil {
		s.logging.Err(err)
		return Msg{}, err
	}

	// we tag the message with our connection id, so that we WONT send it out again
	newMessage.id = s.lastID
	newMessage.context = s.ID

	// iterate id
	s.lastID = s.lastID + 1

	// debug
	s.logging.Debug(fmt.Sprintf(
		"[CON %s][MSG %d] Recieved Message from %s/%s to %s/%s/%s",
		s.ID, newMessage.id, newMessage.NodeSource, newMessage.GroupSource, newMessage.NodeTarget, newMessage.GroupTarget, newMessage.Command,
	))

	return newMessage, nil
}

// SendMessage will send a message over the socket connection
func (s *SocketConnection) SendMessage(message Msg) {

	// we don't send messages that comes from us
	if message.context == s.ID {
		s.logging.Debug(fmt.Sprintf("[MSG %d] We dont send to sender", message.id))
		return
	}

	// iterate id
	message.id = s.lastID
	s.lastID = s.lastID + 1

	// convert to string
	newMessageString, _ := message.ToJSONString()

	// debug
	s.logging.Debug(fmt.Sprintf(
		"[CON %s][MSG %d] Send Message from %s/%s to %s/%s/%s",
		s.ID, message.id, message.NodeSource, message.GroupSource, message.NodeTarget, message.GroupTarget, message.Command,
	))

	// send it
	fmt.Fprintf(s.socket, "%s\n", newMessageString)
}

// Close will close the current socket connection
func (s *SocketConnection) Close() {

	// debug
	s.logging.Info(fmt.Sprintf(
		"[CON %s] Close connection",
		s.ID,
	))

	s.socket.Close()
}

func (s *SocketConnection) onPublish(message *Msg, group, command, payload string) {
	s.SendMessage(*message)
}
