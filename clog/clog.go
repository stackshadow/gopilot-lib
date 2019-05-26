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

package clog

import (
	"flag"
	"fmt"
	"runtime"
)

var isDebug bool

type Logger struct {
	prefix string
}

func getCaller() string {
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return details.Name()
	}

	return "unknown"
}

func ParseCmdLine() {
	flag.BoolVar(&isDebug, "v", false, "Enable debug")
}

func Init() {

}

func New(prefix string) Logger {
	newLog := Logger{
		prefix: prefix,
	}
	return newLog
}

func EnableDebug() {
	isDebug = true
}

func (logging *Logger) Debug(message string) {
	if isDebug == true {
		fmt.Printf("[DEBUG] [%s] [%s] %s\n", logging.prefix, getCaller(), message)
	}
}

func (logging *Logger) Info(message string) {
	fmt.Printf("[INFO] [%s] [%s] %s\n", logging.prefix, getCaller(), message)
}

func (logging *Logger) Error(message string) {
	fmt.Printf("[ERROR] [%s] [%s] %s\n", logging.prefix, getCaller(), message)
}
func (logging *Logger) Err(err error) error {
	fmt.Printf("[ERROR] [%s] [%s] %s\n", logging.prefix, getCaller(), err.Error())
	return err
}
