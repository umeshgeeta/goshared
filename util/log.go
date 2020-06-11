// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

// There are three methods to initialize logs and set log settings:
//
// 1) Caller can pass LoggingCfg structure as defined here to initialize logs.
//
// 2) Alternatively, assuming caller has used 'cfg' program in the 'util' package,
// that 'uber' configuration file can be passed as long as it has a JSON member
// of name "LogSettings". The configuration file can be passed:
// - either with the absolute path
// - or the file name only in which case GP_CFG_HOME environment variable value
// is used as the directory name where the give configuration file is located.
//
// 3) Finally, caller can explicitly call InitializeLog method with arguments.
// The file path can be absolute. It is not absolute, it is appended to the
// current working directory.
//
// Until LoggingCfg is set, GlobalLogSettings will be nil. All log calls will be
// handled by the built in logging facility of Golang. Once the logging is
// initialized by one of the above mentioned 3 methods; log calls will be
// directing the output to the specified rotating log file.
package util

import (
	"encoding/json"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"path/filepath"
)

type LoggingCfg struct {
	LogFileName  string
	MaxSizeInMb  int
	Backups      int
	AgeInDays    int
	Compress     bool
	LogOnConsole bool
	DebugLog     bool
}

// Name of the Json element in any Json Configuration file which contains
// LoggingCfg structure value.
const LoggingCfgJsonElementName = "LogSettings"

// Pointer to a structure which holds log settings in effect.
var GlobalLogSettings *LoggingCfg

// Initialize logging to given inputs:
// fn	:	Log files with fill path
// ms	:	Maximum allowed log file size  in Megabytes
// bk	:	How many backups to be retained.
// age	:	Past logs of how many days to be retained.
// compress:	Whether logs are compressed or not.
func InitializeLog(fn string, ms int, bk int, age int, compress bool) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   fn,
		MaxSize:    ms, // megabytes
		MaxBackups: bk,
		MaxAge:     age,      //days
		Compress:   compress, // disabled by default
	})
	logFilePath, _ := filepath.Abs(fn)
	log.Printf("logFilePath: %s\n", logFilePath)
	GlobalLogSettings = &LoggingCfg{}
	GlobalLogSettings.LogFileName = fn
	GlobalLogSettings.MaxSizeInMb = ms
	GlobalLogSettings.Backups = bk
	GlobalLogSettings.AgeInDays = age
	GlobalLogSettings.Compress = compress
	if GlobalLogSettings.LogOnConsole {
		// typically it will be false when GlobalLogSettings is created
		fmt.Printf("logFilePath: %s\n", logFilePath)
	}
}

func SetLoggingCfg(ls *LoggingCfg) {
	if ls != nil {
		InitializeLog(ls.LogFileName, ls.MaxSizeInMb, ls.Backups, ls.AgeInDays, ls.Compress)
		GlobalLogSettings.DebugLog = ls.DebugLog
		GlobalLogSettings.LogOnConsole = ls.LogOnConsole
	} else {
		log.Fatal("Logging configuration is nil")
	}
}

// It assumes argument configuration file in JSON format and it contains an
// element named LogSettings. Contents of that member are used to build the
// log setting configuration.
func SetLogSettings(cfgFileName string) {
	if len(cfgFileName) > 0 {
		ls, err := ExtractCfgJsonEleFromFile(cfgFileName, LoggingCfgJsonElementName)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error extracting LogSettings from the given config file (%s): %v\n", cfgFileName, err))
		} else {
			ls, err := FormLoggingCfg(ls)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error forming LoggingCfg from the given config file (%s): %v\n", cfgFileName, err))
			} else {
				SetLoggingCfg(ls)
				// Log the setting values which will be used
				Log(fmt.Sprintf("Log Settings: %v\n", ls))
			}
		}
	} else {
		msg := "configuration file is nil"
		fmt.Println(msg)
		log.Fatal(msg)
	}
}

// Make logging configuration struct from the given input string.
func FormLoggingCfg(ls []byte) (*LoggingCfg, error) {
	var lc LoggingCfg
	err := json.Unmarshal([]byte(ls), &lc)
	return &lc, err
}

// Enable or disable the console logging
func SetConsoleLog(val bool) {
	GlobalLogSettings.LogOnConsole = val
}

// Enable or disable debug logging.
func SetDebugLog(val bool) {
	GlobalLogSettings.DebugLog = val
}

// Log the given message.
func Log(msg string) {
	log.Println(msg)
	if GlobalLogSettings.LogOnConsole {
		fmt.Println(msg)
	}
}

// Log debug messages. Invocation of this call will result in adding the message
// to the log provided SetDebugLog(true) is called.
func LogDebug(msg string) {
	if GlobalLogSettings.DebugLog {
		Log(msg)
	}
}

// Indicates whether logging has been configured or not.
func IsLoggingConfigured() bool {
	if GlobalLogSettings != nil {
		return true
	} else {
		return false
	}
}
