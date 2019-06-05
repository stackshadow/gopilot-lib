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
	"github.com/sirupsen/logrus"
	"github.com/stackshadow/gopilot-lib/clog"
	"testing"
	"time"
)

var bus GBus

func TestInit(t *testing.T) {

	// we need our nodename
	clog.EnableDebug()
	clog.Init()

	bus.Init()

	isClosed := false

	bus.Subscribe("0", "local", "test1", func(message *Msg, group, command, payload string) {
		bus.PublishPayload("gotest", "local", "gotest", "test2", "", "")
	})

	bus.Subscribe("1", "local", "test2", func(message *Msg, group, command, payload string) {
		bus.PublishPayload("gotest", "local", "gotest", "end", "", "")
	})

	bus.Subscribe("3", "local", "end", func(message *Msg, group, command, payload string) {
		isClosed = true
	})

	bus.Subscribe("4", "local", "noreceive", func(message *Msg, group, command, payload string) {
		logrus.Panic("This should not happen")
		t.FailNow()
	})

	testCounter3 := 0
	bus.Subscribe("5", "", "", func(message *Msg, group, command, payload string) {
		testCounter3 = testCounter3 + 1
	})

	// now we send some messages
	bus.PublishPayload("gotest", "local", "gotest", "test1", "", "")

	for {
		if isClosed == true {
			break
		}
		time.Sleep(time.Second)
	}

}
