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

	"github.com/sirupsen/logrus"
)

// subscriber represents an subscription to a message ( filter ) on the bus
// all fields in the filter must match the message which arrived on the bus
// fields with "" means "ignore the value"
// this is an INTERNAL struct, so will not be used by userland
type subscriber struct {
	id          string
	filter      Msg
	displayName string
	onMessage   OnMessageFct
}

// SubscriberList represents all subscribers in the list
type SubscriberList struct {
	Subscriber map[string]SubscriberListEntry
}

type SubscriberListEntry struct {
	NodeTarget  string `json:"nodeTarget"`
	GroupTarget string `json:"groupTarget"`
}

// callbacks
type OnMessageFct func(*Msg, string /* group */, string /*command*/, string /*payload*/) // For example: onMessage(message *msgbus.Msg, group, command, payload string)

// SubscriberList represent the list for all subcribers
type GBus struct {
	log             *logrus.Entry
	subscribersLock sync.Mutex
	subscribers     []subscriber

	// messages
	lastMsgNo int
	messages  chan *Msg
}

// Init [NONBLOCKING] the message-bus, you need to call Run() to start it
func (bus *GBus) Init() {

	bus.log = logrus.WithFields(
		logrus.Fields{
			"prefix": "SLIST",
		},
	)

	// message
	bus.lastMsgNo = 0
	bus.messages = make(chan *Msg, 5)

}

// Run [NONBLOCKING] will start the bus
func (bus *GBus) Run() {
	go bus.onPublishListWorker()
}

func (bus *GBus) onPublishListWorker() {
	// blocking until message arive
	for message := range bus.messages {

		// lock the list
		bus.subscribersLock.Lock()

		bus.log.WithFields(logrus.Fields{
			"msgID":               message.id,
			"message.NodeTarget":  message.NodeTarget,
			"message.GroupTarget": message.GroupTarget,
			"message.Command":     message.Command,
		}).Debug("Handle message")

		// send it to all subscribers
		for _, subscriber := range bus.subscribers {
			if message.NodeTarget != "" && subscriber.filter.NodeTarget != "" && message.NodeTarget != subscriber.filter.NodeTarget {
				continue
			}
			if message.GroupTarget != "" && subscriber.filter.GroupTarget != "" && message.GroupTarget != subscriber.filter.GroupTarget {
				continue
			}

			bus.log.WithFields(logrus.Fields{
				"subID":                  subscriber.id,
				"subscriber.NodeTarget":  subscriber.filter.NodeTarget,
				"subscriber.GroupTarget": subscriber.filter.GroupTarget,
			}).Debug("Message match, call onMessage()")

			subscriber.onMessage(message, message.GroupTarget, message.Command, message.Payload)
		}

		// finished
		bus.log.WithFields(logrus.Fields{"msgID": message.id}).Debug("Handle message finished")

		// unlock the list
		bus.subscribersLock.Unlock()
	}
}

// Subscribe will register an callback function
// this function is called wenn a new message arrive and the listenForNodeName and listenForGroupName matches the target node/group in the message
func (bus *GBus) Subscribe(id string, listenForNodeName string, listenForGroupName string, onMessageFP OnMessageFct) error {

	var newSubscriber subscriber
	newSubscriber.id = id
	newSubscriber.filter.NodeTarget = listenForNodeName
	newSubscriber.filter.GroupTarget = listenForGroupName
	newSubscriber.onMessage = onMessageFP

	// append it to the list
	bus.subscribersLock.Lock()

	bus.log.WithFields(logrus.Fields{
		"subID":                 newSubscriber.id,
		"subscriberNodeTarget":  newSubscriber.filter.NodeTarget,
		"subscriberGroupTarget": newSubscriber.filter.GroupTarget,
	}).Debug("Subscribe")

	bus.subscribers = append(bus.subscribers, newSubscriber)
	bus.subscribersLock.Unlock()

	return nil
}

// UnSubscribeID will remove an listener function from the subscriber list
func (bus *GBus) UnSubscribeID(id string) error {
	return bus.unSubscribeFull(id, "", "")
}

// UnSubscribe will remove an listener function from the subscriber list
func (bus *GBus) UnSubscribe(listenForNodeName string, listenForGroupName string) error {
	return bus.unSubscribeFull("", listenForNodeName, listenForGroupName)
}

func (bus *GBus) unSubscribeFull(id string, listenForNodeName string, listenForGroupName string) error {

	var newList []subscriber

	bus.subscribersLock.Lock()
	for _, subscriber := range bus.subscribers {

		if id != "" && subscriber.id != id {
			newList = append(newList, subscriber)
			continue
		}

		if listenForNodeName != "" && subscriber.filter.NodeTarget != listenForNodeName {
			newList = append(newList, subscriber)
			continue
		}

		if listenForGroupName != "" && subscriber.filter.GroupTarget != listenForGroupName {
			newList = append(newList, subscriber)
			continue
		}

		bus.log.WithFields(logrus.Fields{
			"subID": subscriber.id,
		}).Debug("UnSubscribe")
	}
	bus.subscribers = newList
	bus.subscribersLock.Unlock()

	return nil
}

// PublishPayload [BLOCKING] will place a new message to the bus
// for socket-connections it will write directly to the socket itselfe
func (bus *GBus) PublishPayload(nodeSource, nodeTarget, groupSource, groupTarget, command, payload string) error {

	bus.PublishMsg(Msg{
		NodeSource:  nodeSource,
		NodeTarget:  nodeTarget,
		GroupSource: groupSource,
		GroupTarget: groupTarget,
		Command:     command,
		Payload:     payload,
	})

	return nil
}

// PublishMsg [BLOCKING] will place a new message to the bus
// for socket-connections it will write directly to the socket itselfe
func (bus *GBus) PublishMsg(message Msg) error {

	// set message id
	message.id = bus.lastMsgNo
	bus.lastMsgNo = bus.lastMsgNo + 1

	bus.messages <- &message
	return nil
}

// SubscriberListGet return a list with all subscribers
func (bus *GBus) SubscriberListGet() SubscriberList {

	var newSubscriberList SubscriberList
	newSubscriberList.Subscriber = make(map[string]SubscriberListEntry)

	// bus.subscribersLock.Lock()
	for _, subscriber := range bus.subscribers {

		newSubscriberList.Subscriber[subscriber.id] = SubscriberListEntry{
			NodeTarget:  subscriber.filter.NodeTarget,
			GroupTarget: subscriber.filter.GroupTarget,
		}

	}
	// bus.subscribersLock.Unlock()

	return newSubscriberList
}
