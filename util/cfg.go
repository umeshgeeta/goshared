// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

// Package util contains common functionality which is required by a typical
// server side any GO application. Namely it contains functionality to read
// JSON configuration file (from ilyakaznacheev) and rotating logging
// functionality (from lumberjack). Essentially these are few simple useful
// wrappers convenient for any GO application.
//
// If user passes a configuration file with absolute path, it is read as is.
// Otherwise environment variable GO_CFG_HOME is checked to see if contains
// the directory name in which the passed configuration file is checked.
// If found, it is read and configuration structure is returned.
package util

import (
	"encoding/json"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ConfigHome struct {
	Dir string `env:"GO_CFG_HOME" env-description:"Directory where we can find configuration file"`
}

var cfgHome ConfigHome

// Read configurations
func ReadCfg(cfg interface{}, fileName string) error {
	err := cleanenv.ReadEnv(&cfgHome)
	if err != nil {
		msg := fmt.Sprintf("Error reading environment variable GO_CFG_HOME: %v", err)
		fmt.Println(msg)
		log.Fatal(msg)
	}
	if len(cfgHome.Dir) == 0 {
		msg := fmt.Sprintf("Environment variable GO_CFG_HOME is not defined")
		fmt.Println(msg)
		log.Fatal(msg)
	}
	cfgFileName := cfgHome.Dir + "/" + fileName
	fmt.Printf("Config file in use %s\n", cfgFileName)
	return cleanenv.ReadConfig(cfgFileName, cfg)
}

// Returns directory which holds config files.
func GetCfgHomeDir() (string, error) {
	err := cleanenv.ReadEnv(&cfgHome)
	return cfgHome.Dir, err
}

// Returns the byte array corresponding to json segment of specified config
// element in the passed filename. If file name is with the absolute path,
// reading is attempted for that location, else the given file name is looked
// up in to the directory as pointed by GO_CFG_HOME. Second argument is the
// name of the configuration structure member caller is seeking.
func ExtractCfgJsonEleFromFile(fileName string, cfgJsonElementName string) ([]byte, error) {
	cfgFileName := makeCfgFilePath(fileName)
	// Open our jsonFile
	jsonFile, err := os.Open(cfgFileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		return []byte{}, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return []byte{}, err
	}
	return ExtractCfgJsonEleFromBytes(byteValue, cfgJsonElementName)
}

// Returns the byte array corresponding to json segment of the specified config
// element in the passed byte array. Second argument is the name of the
// configuration structure member caller is seeking.
func ExtractCfgJsonEleFromBytes(byteValue []byte, cfgJsonElementName string) ([]byte, error) {
	var result map[string]interface{}
	json.Unmarshal(byteValue, &result)

	// get value of requested cfg element
	rr := result[cfgJsonElementName]
	fmt.Printf("LogSettings: %v\n", rr)
	buf, err := json.Marshal(rr)
	fmt.Println(string(buf))
	return buf, err
}

func makeCfgFilePath(fn string) string {
	if isAbsFilepath(fn) {
		return fn
	} else {
		err := cleanenv.ReadEnv(&cfgHome)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error reading environment variable GO_CFG_HOME: %v", err))
		}
		if len(cfgHome.Dir) == 0 {
			fmt.Println(fmt.Sprintf("Environment variable GO_CFG_HOME is not defined"))
		}
		return cfgHome.Dir + "/" + fn
	}
}

func isAbsFilepath(fn string) bool {
	result := false
	filePath, err := filepath.Abs(fn)
	if err == nil {
		if filePath == fn {
			result = true
		}
	}
	return result
}
