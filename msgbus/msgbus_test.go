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

import "testing"
import "fmt"
import "time"
import "github.com/stackshadow/gopilot/lib/clog"
import "os"

var testBus *MsgBus

func TestInit(t *testing.T) {

	testBus = New()
	clog.EnableDebug()

	t.Run("Test register/deregister", RegisterDeregister)
}

func RegisterDeregister(t *testing.T) {

	// this test, register an listener, which should ONLY be called once,
	// because all listeners will be deleted after
	listenera := testBus.ListenForGroup("", "groupa", onlyOnSingleMessage)
	listenerb := testBus.ListenForGroup("", "groupb", onlyOnSingleMessage)
	listenerc := testBus.ListenForGroup("", "groupc", onlyOnSingleMessage)
	listenerd := testBus.ListenForGroup("", "groupd", onlyOnSingleMessage)

	if testBus.ListenersCount() != 4 {
		t.Error("There should be 4 listeners...")
		t.FailNow()
		return
	}

	// remove listener
	testBus.Deregister(listenera)
	testBus.Deregister(listenerc)
	testBus.Deregister(listenerd)

	if testBus.ListenersCount() != 1 {
		t.Error("There should be 1 listeners...")
		t.FailNow()
		return
	}

	// send a message ( this now should be fired only once )
	testBus.Publish("DUMMY", "sourceNode", "targetNode", "group", "group", "command", "payload")

	// remove the last one
	testBus.Deregister(listenerb)
	if testBus.ListenersCount() != 0 {
		t.Error("There should be 0 listeners...")
		t.FailNow()
		return
	}

	time.Sleep(time.Second * 1)
}

var onlyOnSingleMessageCount int

func onlyOnSingleMessage(message *Msg, group, command, payload string) {
	if onlyOnSingleMessageCount > 0 {
		os.Exit(-1)
	}
	onlyOnSingleMessageCount++
}

var testOnMessageCounter = 0

func testOnMessage(message *Msg, group, command, payload string) {
	fmt.Println("GROUP: ", group, " CMD: ", command, " PAYLOAD: ", payload)
	testOnMessageCounter++
}

func onMultipleMessage(message *Msg, group, command, payload string) {
	if group != "group" && group != "tst" {
		os.Exit(-1)
	}
}
func onSingleMessage(message *Msg, group, command, payload string) {
	if group != "tst" {
		os.Exit(-1)
	}
}
func onNeverMessage(message *Msg, group, command, payload string) {
	fmt.Println("GROUP: ", group, " CMD: ", command, " PAYLOAD: ", payload)
}
