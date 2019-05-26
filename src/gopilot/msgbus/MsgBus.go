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

package msgbus

import (
	"fmt"
	"github.com/google/uuid"
	"gopilot/clog"
	"strconv"
	"sync"
)

// MsgBus represents a single bus
type MsgBus struct {
	log            clog.Logger
	list           chan Msg
	lastID         int
	listeners      []MsgListener
	listenersMutex sync.Mutex
}

type MsgListener struct {
	id string

	// nodeTarget can be "", then it will be ignored
	listenForNodeName  string
	listenForGroupName string // can be ""
	command            string // can be ""
	onMessage          onMessageFct
}

// callbacks
type onMessageFct func(*Msg, string /* group */, string /*command*/, string /*payload*/) // For example: onMessage(message *msgbus.Msg, group, command, payload string)

// New will create a new Messagebus
func New() *MsgBus {

	newBus := MsgBus{
		log:    clog.New("MSGBUS"),
		list:   make(chan Msg, 10),
		lastID: 0,
	}

	// start the worker
	for w := 1; w <= 1; w++ {
		newBus.log.Debug("WORKER " + strconv.Itoa(w) + " Start")
		go newBus.worker2(w, newBus.list)
	}

	return &newBus
}

// ListenForGroup will register the onMessageFP-Function and call it when a message
// with the group is published
func (bus *MsgBus) ListenForGroup(listenForNodeName string, listenForGroupName string, onMessageFP onMessageFct) MsgListener {

	// create new plugin and append it
	newListener := MsgListener{
		id:                 uuid.New().String(),
		listenForNodeName:  listenForNodeName,
		listenForGroupName: listenForGroupName,
		command:            "",
		onMessage:          onMessageFP,
	}

	bus.listenersMutex.Lock()
	bus.listeners = append(bus.listeners, newListener)
	bus.listenersMutex.Unlock()

	if newListener.listenForGroupName != "" {
		bus.log.Debug(fmt.Sprintf("[PLUGIN %s] Listen for %s/%s", newListener.id, newListener.listenForNodeName, newListener.listenForGroupName))
	} else {
		bus.log.Debug(fmt.Sprintf("[PLUGIN %s] Listen for %s/%s", newListener.id, "all groups", newListener.listenForNodeName))
	}

	return newListener
}

// Deregister will remove all callback listeners of an plugin
func (bus *MsgBus) Deregister(listener MsgListener) {

	bus.listenersMutex.Lock()
	var NewMessageListeners []MsgListener

	for _, curListener := range bus.listeners {
		//curListener := messageListeners[listenerIndex]

		if curListener.id != listener.id {
			NewMessageListeners = append(NewMessageListeners, curListener)
		}

	}

	bus.listeners = NewMessageListeners
	bus.listenersMutex.Unlock()

}

// ListenersCount return the amount if listeners
func (bus *MsgBus) ListenersCount() int {
	return len(bus.listeners)
}

// Publish will publish an message to the bus
func (bus *MsgBus) Publish(pluginName string, nodeSource, nodeTarget, groupSource, groupTarget, command, payload string) {

	newMessage := Msg{
		id:          bus.lastID,
		NodeSource:  nodeSource,
		NodeTarget:  nodeTarget,
		GroupSource: groupSource,
		GroupTarget: groupTarget,
		Command:     command,
		Payload:     payload,
	}

	bus.PublishMsg(newMessage)
}

// PublishMsg will publish an message to the bus
func (bus *MsgBus) PublishMsg(newMessage Msg) {

	bus.listenersMutex.Lock()
	newMessage.id = bus.lastID

	bus.log.Debug(
		fmt.Sprintf(
			"[MSG %d] FROM %s/%s TO %s/%s/%s",
			newMessage.id,
			newMessage.NodeSource, newMessage.GroupSource,
			newMessage.NodeTarget, newMessage.GroupTarget, newMessage.Command,
		),
	)
	bus.lastID++

	bus.list <- newMessage
	bus.listenersMutex.Unlock()
}

func (bus *MsgBus) worker2(no int, messages <-chan Msg) {
	workerName := fmt.Sprintf("WORKER %d", no)
	bus.log.Debug(workerName + " Run")

	for curMessage := range messages {

		bus.log.Debug(fmt.Sprintf("%s [MSG %d] Process %s/%s/%s", workerName, curMessage.id, curMessage.NodeTarget, curMessage.GroupTarget, curMessage.Command))

		//logging.Debug(workerName, fmt.Sprintf("pluginList contains %d Plugins", len(pluginList)))

		// bus.listenersMutex.Lock()
		for _, curListener := range bus.listeners {

			// if target is set it must match
			if curListener.listenForNodeName != "" && curMessage.NodeTarget != "" && curListener.listenForNodeName != curMessage.NodeTarget {
				continue
			}

			// if group is set, it must match
			if curListener.listenForGroupName != "" && curMessage.GroupTarget != "" && curListener.listenForGroupName != curMessage.GroupTarget {
				continue
			}

			bus.log.Debug(
				fmt.Sprintf(
					"%s [MSG %d] CALL", workerName,
					curMessage.id,
				),
			)
			curListener.onMessage(&curMessage, curMessage.GroupTarget, curMessage.Command, curMessage.Payload)
		}
		// bus.listenersMutex.Unlock()

	}

}
