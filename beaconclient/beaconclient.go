package beaconclient

import (
	"errors"

	"github.com/godbus/dbus"
)

// Can be mocked for testing
var dbusCall = func(method string, params ...interface{}) ([]interface{}, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	obj := conn.Object("org.cacophony.beacon", "/org/cacophony/beacon")
	call := obj.Call(method, 0, params...)
	return call.Body, call.Err
}

var ErrorParsingOutput = errors.New("error with parsing dbus output")

func Ping() error {
	_, err := dbusCall("Ping")
	return err
}

func RecordingStarted() error {
	_, err := dbusCall("RecordingStarted")
	return err
}

func Classification() error {
	_, err := dbusCall("Classification")
	return err
}

func PowerOff(minutes uint16) error {
	_, err := dbusCall("PowerOff", minutes)
	return err
}
