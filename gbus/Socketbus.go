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

// Socketbus type
type Socketbus struct {
	log *logrus.Entry
	// remoteConnectionList map[string]*SocketConnection
	subscriberList SubscriberList
	serverListener net.Listener
	isServer       bool
	//	socketConnectionsLock sync.Mutex
	//	socketConnections     map[string]*SocketConnection
}

// Init will init an new socketbus
func (bus *Socketbus) Init() {

	bus.log = logrus.WithFields(
		logrus.Fields{
			"prefix": "SOCKETBUS",
		},
	)

	bus.subscriberList.Init()
	// bus.socketConnections = make(map[string]*SocketConnection)
}

// Serve [BLOCKING] will start the socket-server and run forever until an error occure
// if an connection occure and the handshake was okay, an goroutine will be started which handle incoming messages
func (bus *Socketbus) Serve(connectionString string) error {

	// remove socket if it already exists
	if err := os.RemoveAll(connectionString); err != nil {
		bus.log.Error(err)
		return err
	}

	// open the socket
	var err error
	bus.serverListener, err = net.Listen("unix", connectionString)
	if err != nil {
		bus.log.Error(err)
		return err
	}

	bus.log.Info(fmt.Sprintf("Create SOCKET on %s", connectionString))

	// wait for new clients
	for {

		// wait for a new client
		bus.log.Info("Wating for new client")
		conn, err := bus.serverListener.Accept()
		if err != nil {
			bus.log.Error(err)
			return err
		}

		bus.log.Info(fmt.Sprintf("Accept new client %v", conn.RemoteAddr()))

		// create a new session
		// the filter is empty, as server we accept every message
		var con SocketConnection
		con.Init(conn, Msg{
			NodeTarget:  mynodename.NodeName,
			GroupTarget: "",
		}, "IN-"+uuid.New().String())

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
		bus.log.Debug("Wait for OLEH-Message")
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
		bus.log.Debug("OLEH received")

		/*
			bus.socketConnectionsLock.Lock()
			con.logging.Debug(fmt.Sprintf("Register '%s'", con.ID))
			bus.socketConnections[con.ID] = con
			bus.socketConnectionsLock.Unlock()
		*/
		// we subscribe to the bus
		bus.subscriberList.Subscribe(con.ID, ehloMessage.NodeSource, ehloMessage.GroupSource, con.onPublish)

		// start goroutine which handle incoming messages
		go bus.eventLoopWaitForMessage(&con)
	}

}

// Connect [GOROUTINE] will connect to an server and return the remote Node-Name
func (bus *Socketbus) Connect(connectionString string, filter Msg) (string, error) {

	var conn net.Conn
	var err error

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
	con.Init(conn, filter, "OUT-"+uuid.New().String())

	// ################################# handshake #################################
	// we wait for HELO
	bus.log.Debug("Wait for HELO-Message")
	heloMessage, err := con.ReadMessage()
	if err != nil {
		con.Close()
		bus.log.Error(err)
		return "", err
	}
	if heloMessage.Command != "HELO" {
		con.Close()
		errNoHelo := errors.New("No OLEH was recieved")
		bus.log.Error(errNoHelo)
		return "", errNoHelo
	}
	bus.log.Debug("HELO received")

	// and informate the server about what we listen
	con.SendMessage(Msg{
		NodeSource:  filter.NodeSource,
		GroupSource: filter.GroupSource,
		NodeTarget:  heloMessage.NodeTarget,
		GroupTarget: "",
		Command:     "OLEH",
	})

	// we subscribe to the bus
	bus.subscriberList.Subscribe(con.ID, heloMessage.NodeSource, heloMessage.GroupSource, con.onPublish)

	// add it to the list
	/*
		bus.sessionListMutex.Lock()
		con.logging.Debug(fmt.Sprintf("Register '%s'", con.ID))
		bus.remoteConnectionList[con.ID] = &con
		bus.sessionListMutex.Unlock()
	*/

	go bus.eventLoopWaitForMessage(&con)

	return heloMessage.NodeSource, nil
}

func (bus *Socketbus) eventLoopWaitForMessage(con *SocketConnection) {

	// close and remove session if disconnect or error occure
	defer func() {
		con.logging.Debug(fmt.Sprintf("Close '%s'", con.ID))
		con.Close()

		bus.subscriberList.UnSubscribeID(con.ID)
	}()

	// message-loop
	for {
		message, err := con.ReadMessage()
		if err != nil {
			bus.log.Error(err)
			break
		}

		bus.subscriberList.PublishMsg(message)
	}

}

// Subscribe will subscribe to an message on the bus
func (bus *Socketbus) Subscribe(id string, listenForNodeName string, listenForGroupName string, onMessageFP OnMessageFct) error {
	return bus.subscriberList.Subscribe(id, listenForNodeName, listenForGroupName, onMessageFP)
}

// PublishPayload will place a message to the bus
func (bus *Socketbus) PublishPayload(nodeSource, nodeTarget, groupSource, groupTarget, command, payload string) error {
	return bus.subscriberList.PublishPayload(nodeSource, nodeTarget, groupSource, groupTarget, command, payload)
}

// PublishMsg will place a message to the bus
func (bus *Socketbus) PublishMsg(message Msg) error {
	return bus.subscriberList.PublishMsg(message)
}
