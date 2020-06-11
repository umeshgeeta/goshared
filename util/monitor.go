// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

import (
	"errors"
	"time"
)

// The entity or service which we want to monitor. That entity should implement
// this interface. Implementation of this method should be 'lightweight' it is
// to be invoked at the higher frequency.
type Monitored interface {

	// Returns monitoring Data, some kind of system health check for example.

	GetData() Blob

	// Reference name, presumably unique so it is easier to locate in Logs
	Name() string
}

// Data returned by GetData method, Health Status Data
// Typically it should be Json string byte array.
type Blob struct {
	Data []byte
}

// Monitors the specified entity by invoking it's GetData at given frequency.
// For now it only Logs the Data.
type Monitor struct {
	continueRun bool
	frequency   int // in seconds
	monEntity   *Monitored
	MonDataChan chan Blob // exposed so anyone interested in Data can get handle
}

// Builds a new monitor for the given entity. GetData method on that entity will
// be invoked at the given frequency. Error is thrown when the entity is nil.
// There is no validation on the frequency number right now, it is Seconds.
func NewMonitor(freq int, chanBufSz int, entity Monitored) (*Monitor, error) {
	if entity == nil {
		err := errors.New("entity to be monitored is nil")
		return nil, err
	}
	m := &Monitor{}
	m.frequency = freq
	m.monEntity = &entity
	m.MonDataChan = make(chan Blob, chanBufSz)
	return m, nil
}

func NewBlob(ba []byte) *Blob {
	result := Blob{}
	result.Data = ba
	return &result
}

func (m *Monitor) Start() {
	m.continueRun = true
	go m.monitor()
}

func (m *Monitor) Stop() {
	m.continueRun = false
}

func (m *Monitor) monitor() {
	for m.continueRun {
		time.Sleep(time.Duration(m.frequency) * time.Second)
		blob := (*m.monEntity).GetData()
		m.MonDataChan <- blob
		Log((*m.monEntity).Name() + ":  " + string(blob.Data))
	}
	Log("Monitor for entity " + (*m.monEntity).Name() + " stopped.")
}
