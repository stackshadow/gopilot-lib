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

import ()
import "sync"

// subscriber represents an subscription to a message ( filter ) on the bus
// all fields in the filter must match the message which arrived on the bus
// fields with "" means "ignore the value"
// this is an INTERNAL struct, so will not be used by userland
type subscriber struct {
	id        string
	filter    Msg
	onMessage OnMessageFct
}

// SubscriberList represent the list for all subcribers
type SubscriberList struct {
	lock     sync.Mutex
	list     []subscriber
	messages chan *Msg
}

// Init the subscriber list
func (list *SubscriberList) Init() {
	list.messages = make(chan *Msg, 5)

	// start the worker
	go list.onPublishListWorker()
}

func (list *SubscriberList) onPublishListWorker() {
	// blocking until message arive
	for message := range list.messages {

		// lock the list
		list.lock.Lock()

		// send it to all subscribers
		for _, subscriber := range list.list {
			if subscriber.filter.NodeTarget != "" && message.NodeTarget != subscriber.filter.NodeTarget {
				continue
			}
			if subscriber.filter.GroupTarget != "" && message.GroupTarget != subscriber.filter.GroupTarget {
				continue
			}

			subscriber.onMessage(message, message.GroupTarget, message.Command, message.Payload)
		}

		// unlock the list
		list.lock.Unlock()
	}
}

// Subscribe will register an callback function
// this function is called wenn a new message arrive and the listenForNodeName and listenForGroupName matches the target node/group in the message
func (list *SubscriberList) Subscribe(id string, listenForNodeName string, listenForGroupName string, onMessageFP OnMessageFct) error {

	var newSubscriber subscriber
	newSubscriber.filter.NodeTarget = listenForNodeName
	newSubscriber.filter.GroupTarget = listenForGroupName
	newSubscriber.onMessage = onMessageFP

	// append it to the list
	list.lock.Lock()
	list.list = append(list.list, newSubscriber)
	list.lock.Unlock()

	return nil
}

// UnSubscribeID will remove an listener function from the subscriber list
func (list *SubscriberList) UnSubscribeID(id string) error {
	return list.unSubscribeFull(id, "", "")
}

// UnSubscribe will remove an listener function from the subscriber list
func (list *SubscriberList) UnSubscribe(listenForNodeName string, listenForGroupName string) error {
	return list.unSubscribeFull("", listenForNodeName, listenForGroupName)
}

func (list *SubscriberList) unSubscribeFull(id string, listenForNodeName string, listenForGroupName string) error {

	var newList []subscriber

	list.lock.Lock()
	for _, subscriber := range list.list {

		if id != "" && subscriber.id != id {
			newList = append(newList, subscriber)
		}

		if listenForNodeName != "" && subscriber.filter.NodeTarget != listenForNodeName {
			newList = append(newList, subscriber)
		}

		if listenForGroupName != "" && subscriber.filter.GroupTarget != listenForGroupName {
			newList = append(newList, subscriber)
		}

	}
	list.list = newList
	list.lock.Unlock()

	return nil
}

// PublishPayload [BLOCKING] will place a new message to the bus
// for socket-connections it will write directly to the socket itselfe
func (list *SubscriberList) PublishPayload(nodeSource, nodeTarget, groupSource, groupTarget, command, payload string) error {

	list.PublishMsg(Msg{
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
func (list *SubscriberList) PublishMsg(message Msg) error {
	list.messages <- &message
	return nil
}
