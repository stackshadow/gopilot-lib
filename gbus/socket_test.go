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
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestSocketHandshake(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	var finished sync.Mutex
	finished.Lock()

	server := SocketNew()
	go server.Serve("/tmp/inttest.sock", SocketCallbacks{
		OnHandshakeFinished: func(socket *SocketConnection) {
			if socket.RemoteNodeName() != "testnode" {
				t.Fail()
			}
			if socket.RemoteNodeGroup() != "test" {
				t.Fail()
			}
			finished.Unlock()
		},
	})

	client := SocketNew()
	go client.Connect("/tmp/inttest.sock", "testnode", "test", SocketCallbacks{})

	finished.Lock()
}

func TestMessage(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	var finished sync.Mutex
	finished.Lock()

	server := SocketNew()
	go server.Serve("/tmp/inttest.sock", SocketCallbacks{
		OnMessage: func(socket *SocketConnection, message Msg) {

			if message.Command != "ping" {
				t.Fail()
			}
			finished.Unlock()
			return
		},
	})

	time.Sleep(time.Second * 2)

	client := SocketNew()
	go client.Connect("/tmp/inttest.sock", "testnode", "test", SocketCallbacks{
		OnHandshakeFinished: func(socket *SocketConnection) {
			socket.SendMessage(Msg{
				NodeSource:  "testnode",
				GroupSource: "test",
				NodeTarget:  "server",
				GroupTarget: "serverGroup",
				Command:     "ping",
			})
		},
	})

	finished.Lock()

}
