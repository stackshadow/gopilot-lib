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
	"fmt"
	"gopilot/clog"
	"gopilot/nodeName"
	"testing"
	"time"
)

var bus Socketbus
var client Socketbus

func TestInit(t *testing.T) {

	// we need our nodename
	clog.Init()
	clog.EnableDebug()
	mynodename.ParseCmdLine()
	mynodename.Init()

	bus.Init()
	bus.Subscribe("UUID1", "", "servergroup", func(message *Msg, group, command, payload string) {
		fmt.Printf("%s/%s", group, command)
	})
	go bus.Serve("/tmp/gopilottest.go")
	time.Sleep(time.Second * 2)

	// client
	client.Init()
	go client.Connect("/tmp/gopilottest.go", Msg{
		NodeTarget:  "client",
		GroupTarget: "clientgroup",
	})
	time.Sleep(time.Second * 2)

	client.PublishPayload("client", mynodename.NodeName, "clientgroup", "servergroup", "test", "")

	for {
		time.Sleep(time.Second * 10)
	}

}
