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
	"encoding/json"
	"fmt"
)

// Msg represent a single message inside the bus
type Msg struct {
	id int

	// the source node when transfered over network
	NodeSource string `json:"s"`

	// target node where the message should send to
	// can be "" if you would like to send it to all
	NodeTarget string `json:"t"`

	// The source group, only be used by the internal socket
	GroupSource string `json:"o"`

	// The target group
	GroupTarget string `json:"g"`

	// The command
	Command string `json:"c"`

	// payload
	Payload string `json:"v"`
}

// ToJSONByteArray will convert an message to an byte-array for sending it out
func (curMessage *Msg) ToJSONByteArray() ([]byte, error) {

	b, err := json.Marshal(curMessage)
	if err != nil {
		fmt.Println("error:", err)
	}

	return b, nil
}

// ToJSONString will convert an message to an json string
func (curMessage *Msg) ToJSONString() (string, error) {

	b, err := json.Marshal(curMessage)
	if err != nil {
		fmt.Println("error:", err)
	}

	return string(b), nil
}

// FromJSONString will convert an json-string to an message
func FromJSONString(jsonString string) (Msg, error) {

	var newMessage Msg

	err := json.Unmarshal([]byte(jsonString), &newMessage)
	if err != nil {
		fmt.Println("error: ", err)
		return newMessage, err
	}

	return newMessage, nil
}

// Answer will answer to an message
func (curMessage *Msg) Answer(bus *MsgBus, command, payload string) {

	newMessage := Msg{
		NodeSource:  curMessage.NodeTarget,
		NodeTarget:  curMessage.NodeSource,
		GroupSource: curMessage.GroupTarget,
		GroupTarget: curMessage.GroupSource,
		Command:     command,
		Payload:     payload,
	}

	bus.PublishMsg(newMessage)

}
