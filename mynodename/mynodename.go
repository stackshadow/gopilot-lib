package mynodename

import (
	"flag"
	"github.com/stackshadow/gopilot-lib/clog"
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
	logging := clog.New("NODENAME")

	logging.Info("HOST MyNode: " + NodeName)
}
