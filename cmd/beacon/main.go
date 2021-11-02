// beacon - helper for sending BLE beacons to other devices
// Copyright (C) 2021, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"runtime"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/advertising"
	"github.com/snksoft/crc"
)

const (
	PingType             = 0x01
	RecordingStartedType = 0x02
	ClassificationType   = 0x03
	PowerOffType         = 0x04
	deviceId             = 0xFFFF
	version              = 0x01
	ManufactureID        = 0x1212
	AdapterID            = "hci0"
)

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
	runtime.Goexit()
	log.Println("Exiting")
}

type Data struct {
	Animal     [2]byte
	Confidence [2]byte
}

func runMain() error {
	log.Println("Starting Beacon service")
	if err := startService(); err != nil {
		return err
	}
	return nil
}

func Ping() error {
	return expose(PingType, []byte{}, 30)
}

func RecordingStarted() error {
	return expose(RecordingStartedType, []byte{}, 30)
}

func Classification(classifications map[byte]byte) error {
	return expose(ClassificationType, classificationToByteArray(classifications), 30)
}

func classificationToByteArray(classifications map[byte]byte) []byte {
	data := []byte{}
	i := 0
	for {
		if i >= 5 || len(classifications) == 0 {
			break
		}
		var maxKey byte = 0x00
		var maxCon byte = 0x00
		for key, con := range classifications {
			if con >= maxCon {
				maxKey = key
				maxCon = con
			}
		}
		data = append(data, maxKey, maxCon)
		i++
		delete(classifications, maxKey)
	}
	data = append([]byte{byte(i)}, data...)
	return data
}

func PowerOff(seconds uint16) error {
	secondsByte := make([]byte, 2)
	binary.BigEndian.PutUint16(secondsByte, seconds)
	return expose(PowerOffType, secondsByte, 30)
}

func deviceIdInBytes() []byte {
	id := make([]byte, 2)
	binary.BigEndian.PutUint16(id, deviceId)
	return id
}

func expose(dataType byte, data []byte, timeout uint16) error {
	d := []byte{version, deviceIdInBytes()[0], deviceIdInBytes()[1], dataType}
	data = append(d, data...)
	log.Println(data)
	//Calculate CRC
	crcTable := crc.NewTable(crc.CRC32)
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, uint32(crcTable.CalculateCRC(data)))
	data = append(data, crc...)
	log.Println(data)
	log.Println(hex.EncodeToString(data))

	props := new(advertising.LEAdvertisement1Properties)
	props.AddManifacturerData(ManufactureID, data)
	props.Type = advertising.AdvertisementTypeBroadcast

	props.Appearance = 0xFFFF // disables it
	_, err := api.ExposeAdvertisement(AdapterID, props, uint32(timeout))
	return err
}
