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
	"github.com/stackshadow/gopilot-lib/msgbus"
)

// OnMessageFct represents the callback-function when a new message arrived and match the filter
type OnMessageFct func(*Msg, string /* group */, string /*command*/, string /*payload*/) // For example: onMessage(message *msgbus.Msg, group, command, payload string)

type gbus interface {
	// This init create a new bus
	Init()

	// This start the bus-server
	Serve(connectionString string) error

	// This connect to an server
	// filter is sendet to the server, every non emtpy field must be matched
	// to recieve an message
	// for example if you provide the following filter:
	// Msg {
	//   GroupTarget: "test"
	// }
	// will only deliver messages to us, when an published message
	// contains the group-target "test"
	Connect(connectionString string, filter msgbus.Msg) error

	// this publish ( send a message )
	PublishPayload(nodeSource, nodeTarget, groupSource, groupTarget, command, payload string) error

	// this call the callback-function
	// if error is not null, we can not listening !
	onMessage(listenForNodeName string, listenForGroupName string, onMessageFP OnMessageFct) error

	// this will start the eventloop ( blocking function ! )
	eventLoop() error
}
