/*
Copyright (C) 2019 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

package gbus

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stackshadow/gopilot-lib/mynodename"
	"net"
	"os"
	"time"
)

// Socket type
type Socket struct {
	log            *logrus.Entry
	serverListener net.Listener
	isServer       bool
}

// SocketCallbacks provide different callbacks
// OnConnect - Will fire if you successful connect to an socket
type SocketCallbacks struct {
	OnConnect func( /*remoteNodeName*/ string)
}

// Init will init an new socketbus
func (socket *Socket) Init() {

	socket.log = logrus.WithFields(
		logrus.Fields{
			"prefix": "SOCKETBUS",
		},
	)

}

// Serve [BLOCKING] will start the socket-server and run forever until an error occure
// if an connection occure and the handshake was okay, an goroutine will be started which handle incoming messages
func (socket *Socket) Serve(bus *GBus, connectionString string) error {

	// remove socket if it already exists
	if err := os.RemoveAll(connectionString); err != nil {
		bus.log.Error(err)
		return err
	}

	// open the socket
	var err error
	socket.serverListener, err = net.Listen("unix", connectionString)
	if err != nil {
		bus.log.Error(err)
		return err
	}

	socket.log.Info(fmt.Sprintf("Create SOCKET on %s", connectionString))

	// wait for new clients
	for {

		// wait for a new client
		socket.log.Info("Wating for new client")
		conn, err := socket.serverListener.Accept()
		if err != nil {
			bus.log.Error(err)
			return err
		}

		connectionID := uuid.New().String()

		// create a new session
		// the filter is empty, as server we accept every message
		var con SocketConnection
		con.Init(conn, connectionID)

		// ################################# handshake #################################
		// send a helo to the client
		con.SendMessage(Msg{
			NodeSource:  mynodename.NodeName, // i'am the source
			GroupSource: "",                  // i hear on every group
			NodeTarget:  "",                  // i don't know you, so is just send it to all
			GroupTarget: "",                  // i don't know your group ( yet )
			Command:     "HELO",
		})

		// we wait for OLEH
		socket.log.Debug("Wait for OLEH-Message")
		ehloMessage, err := con.ReadMessage()
		if err != nil {
			bus.log.Error(err)
			conn.Close()
			continue
		}
		if ehloMessage.Command != "OLEH" {
			bus.log.Error("No OLEH was recieved")
			conn.Close()
			continue
		}

		/*
			bus.socketConnectionsLock.Lock()
			con.logging.Debug(fmt.Sprintf("Register '%s'", con.ID))
			bus.socketConnections[con.ID] = con
			bus.socketConnectionsLock.Unlock()
		*/
		// we subscribe to the bus
		bus.Subscribe(con.ID, ehloMessage.NodeSource, ehloMessage.GroupSource, con.onPublish)

		socket.log.Debug("")
		socket.log.Debug("############################ Handshake finished ############################")
		socket.log.Debug("")
		// ############################ handshake finished #############################

		// start goroutine which handle incoming messages
		go socket.eventLoopWaitForMessage(bus, &con)
	}

}

// Connect [BLOCKING] will connect to an server and return the remote Node-Name
func (socket *Socket) Connect(bus *GBus, connectionString, listenForNodeName, listenForGroupName string, callbacks SocketCallbacks) error {

	var conn net.Conn
	var err error

	// this runs forever
	for {

		for {
			conn, err = net.Dial("unix", connectionString)
			if err != nil {
				bus.log.Error(err)
				time.Sleep(10 * time.Second)
			}

			if conn != nil {
				break
			}
		}

		// create a new socket connection
		var con SocketConnection
		con.Init(conn, uuid.New().String())

		// ################################# handshake #################################
		// we wait for HELO
		socket.log.Debug("Wait for HELO-Message")
		heloMessage, err := con.ReadMessage()
		if err != nil {
			con.Close()
			bus.log.Error(err)
			return err
		}
		if heloMessage.Command != "HELO" {
			con.Close()
			errNoHelo := errors.New("No OLEH was recieved")
			bus.log.Error(errNoHelo)
			return errNoHelo
		}

		// and informate the server about what we listen
		con.SendMessage(Msg{
			NodeSource:  listenForNodeName,
			GroupSource: listenForGroupName,
			NodeTarget:  heloMessage.NodeTarget,
			GroupTarget: "",
			Command:     "OLEH",
		})

		// callback
		if callbacks.OnConnect != nil {
			callbacks.OnConnect(heloMessage.NodeSource)
		}

		// we subscribe to the bus
		bus.Subscribe(con.ID, heloMessage.NodeSource, heloMessage.GroupSource, con.onPublish)

		socket.log.Debug("")
		socket.log.Debug("############################ Handshake finished ############################")
		socket.log.Debug("")
		// ############################ handshake finished #############################

		socket.eventLoopWaitForMessage(bus, &con)
	}
}

func (socket *Socket) eventLoopWaitForMessage(bus *GBus, con *SocketConnection) {

	// close and remove session if disconnect or error occure
	defer func() {
		con.logging.Debug(fmt.Sprintf("Close '%s'", con.ID))
		con.Close()

		bus.UnSubscribeID(con.ID)
	}()

	// message-loop
	for {
		message, err := con.ReadMessage()
		if err != nil {
			bus.log.Error(err)
			break
		}

		bus.PublishMsg(message)
	}

}
