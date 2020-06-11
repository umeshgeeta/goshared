// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/umeshgeeta/goshared/util"
	"log"
)

type ExecutionService struct {
	taskDispatcher *Dispatcher
	Monitor        *util.Monitor // exposed for testing purposes
}

// Configuration for the entire execution service which comprises of
// configuration for Dispatcher, Executor Pool and for each Executor.
type ExecServiceCfg struct {
	Dispatcher DispatcherCfg `json:"DispatcherSettings"`
	ExexPool   ExecPoolCfg   `json:"ExecPoolSettings"`
	Executor   ExecCfg       `json:"ExecutorSettings"`
	Monitoring MonitoringCfg `json:"MonitoringSettings"`
}

type MonitoringCfg struct {
	MonitoringFrequency int `json:"MonitoringFrequency"`
	MonDataChanBufSz    int `json:"ChannelBufferSize"`
}

// Name of the Json element in any Json Configuration file which contains
// ExecServiceCfg structure value. Note that we do not support only part
// settings, the constant refers to a Json segment which will contain
// values for all 3 config structures.
const ExecServiceCfgJsonElementName = "ExecServiceSettings"

// Name of a configuration file which contains default values; in the same
// folder where you would find execution-service.go. If user allows to use
// the default configuration, then in absence of user provided configuration
// values in this file will be used.
const DefaultCfgFileName = "default-cfg.json"

var StaticBox *packr.Box

var GlobalExecServiceCfg *ExecServiceCfg

// Initialize the box so that static files are available for consumption.
// Default configuration file is one important static content file.
func init() {
	StaticBox = packr.New("Static Files", "./static")
}

// Caller can pass the configuration file name which will contain all parameters
// needed to start the execution service. The file will be searched in the
// directory as pointed by the environmental variable GO_CFG_HOME. If the
// environmental variable is not set or file is not found; caller can indicate
// whether default configuration file is to be used or not. If configuration is
// found, method returns with a Fatal Log call.
func NewExecutionService(cfgFileName string, useDefault bool) *ExecutionService {
	cfg, err := util.ExtractCfgJsonEleFromFile(cfgFileName, ExecServiceCfgJsonElementName)
	if err != nil {
		if useDefault {
			cfgJsonBa, err := StaticBox.Find(DefaultCfgFileName)
			if err != nil {
				msg := fmt.Sprintf("Invalid config file name %s and error %v while sourcing default config file. "+
					"Pass second argument true to use default config file or fix te config file issues.\n", cfgFileName, err)
				fmt.Print(msg)
				log.Fatal(msg)
			} else {
				cfg, err = util.ExtractCfgJsonEleFromBytes(cfgJsonBa, ExecServiceCfgJsonElementName)
				if err != nil {
					msg := fmt.Sprintf("Error reading configuration: %v\n", err)
					fmt.Print(msg)
					log.Fatal(msg)
				}
			}
		} else {
			msg := fmt.Sprintf("Invalid config file name %s and default config file not allowed. "+
				"Pass second argument true to use default config file or fix te config file issues.\n", cfgFileName)
			fmt.Print(msg)
			log.Fatal(msg)
		}
	}
	err = json.Unmarshal(cfg, &GlobalExecServiceCfg)
	if err != nil {
		msg := fmt.Sprintf("Error parsing configuration: %v\n", err)
		fmt.Print(msg)
		log.Fatal(msg)
	}
	// let us see if logging configuration is set or not; else we try default
	if !util.IsLoggingConfigured() {
		logCfgJsonBa, err := StaticBox.Find(DefaultCfgFileName)
		if err == nil {
			// form the logging configuration and set it
			logCfgBa, err := util.ExtractCfgJsonEleFromBytes(logCfgJsonBa, util.LoggingCfgJsonElementName)
			logCfg, err := util.FormLoggingCfg(logCfgBa)
			if err == nil {
				util.SetLoggingCfg(logCfg)
			}
		}
		// else got an error getting default logging configuration,
		// we will fall back to default go lang builtin logging
	}
	// else if it configured, nothing to worry

	es := new(ExecutionService)
	es.taskDispatcher = NewDispatcher(GlobalExecServiceCfg.Dispatcher,
		NewExecutorPool(GlobalExecServiceCfg.ExexPool,
			GlobalExecServiceCfg.Executor))
	util.Log(fmt.Sprintf("Started ExecutorService %v", es))

	// start monitoring service
	es.Monitor, _ = util.NewMonitor(GlobalExecServiceCfg.Monitoring.MonitoringFrequency,
		GlobalExecServiceCfg.Monitoring.MonDataChanBufSz,
		*es)
	return es
}

func (es *ExecutionService) Start() {
	es.taskDispatcher.Start()
	es.Monitor.Start()
}

func (es *ExecutionService) Submit(tsk Task) (error, *Response) {
	return es.taskDispatcher.Submit(tsk)
}

func (es *ExecutionService) Stop() {
	es.taskDispatcher.Stop()
	es.Monitor.Stop()
}

func (es ExecutionService) GetData() util.Blob {
	return *(util.NewBlob(es.taskDispatcher.JobStats.byteArray()))
}

func (es ExecutionService) Name() string {
	return "ExecutionService"
}
