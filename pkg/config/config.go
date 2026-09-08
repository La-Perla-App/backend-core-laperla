package config

import (
	"fmt"
)

// FileName file name that contain a json object.
// By default filename: "config.yaml"
var FileName string = "config.yaml"

// FilePath file config path.
// By default file config path: "./"
var FilePath string = "./"

var configurator Configurator

// RegisterConfigurator a new implement of Configurator interface
func RegisterConfigurator(conf Configurator) {
	configurator = conf
}

// ValidateConfiguratorRegister validate if the Configurator interfaces has been implement
func ValidateConfiguratorRegister() error {
	if configurator != nil {
		return nil
	}
	return fmt.Errorf("the implementation of the 'config.Configurator' interface has not been registered")
}

func init() {
	// By default register json configurator implement.
	RegisterConfigurator(JSONConfigurator{})
	var err error
	file := fmt.Sprintf("%v%v", FilePath, FileName)
	// Load config file
	err = LoadConfigFile(file)
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %s", err))
	}
}
