package config

import (
	"encoding/json"
	"gopilot/clog"

	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var logging clog.Logger

// ConfigPath the path where all config files will stored
var ConfigPath string

var jsonConfigNew map[string]interface{}

// ParseCmdLine parse your command line parameter to internal variables
func ParseCmdLine() {
	// config path
	flag.StringVar(&ConfigPath, "configPath", ".", "The base path")
}

// Init create an new empty config
func Init() {
	logging = clog.New("CORE")

	jsonConfigNew = make(map[string]interface{})
}

// Read will read the config from an file calles core.json
func Read() {

	// current path
	ex, err := os.Executable()
	if err == nil {
		exPath := filepath.Dir(ex)
		logging.Debug("CONFIG Our current path is '" + exPath + "'")
	}

	// Open our jsonFile
	jsonFile, err := os.Open(ConfigPath + "/core.json")
	if err != nil {
		// okay non existant, create a new one
		Save()
		return
	}
	defer jsonFile.Close()

	logging.Debug("CONFIG Successfully Opened '" + ConfigPath + "/core.json'")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &jsonConfigNew)

}

// GetJSONObject Return an json object
func GetJSONObject(name string) (map[string]interface{}, error) {

	// then we try to get the object
	if jsonObject, ok := jsonConfigNew[name].(map[string]interface{}); ok {
		return jsonObject, nil
	}

	return nil, fmt.Errorf("No Object found with name '%s'", name)
}

// SetJSONObject set an json object
func SetJSONObject(name string, jsonNode map[string]interface{}) error {

	// save it
	jsonConfigNew[name] = jsonNode

	return nil
}

// Save save you config to the core.json
func Save() {
	byteValue, _ := json.MarshalIndent(jsonConfigNew, "", "    ")
	err := ioutil.WriteFile(ConfigPath+"/core.json", byteValue, 0644)
	if err != nil {
		logging.Error("CONFIG" + err.Error())
		os.Exit(-1)
	}
}
