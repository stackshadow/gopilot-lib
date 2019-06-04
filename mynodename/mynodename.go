package mynodename

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
)

// NodeName contains the current nodeName ( normally localhost if not set by command line paramater)
var NodeName string

// ParseCmdLine parse your command line parameter to internal variables
func ParseCmdLine() {

	// nodename
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	flag.StringVar(&NodeName, "nodeName", hostname, "Set the name of this node ( normaly the hostname is used )")
}

// Init the nodename
func Init() {
	log.WithFields(log.Fields{
		"nodename": NodeName,
	}).Info()
}
