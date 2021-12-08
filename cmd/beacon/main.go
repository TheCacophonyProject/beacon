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
	"sort"
	"sync"
	"time"

	goconfig "github.com/TheCacophonyProject/go-config"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/advertising"
	"github.com/snksoft/crc"
)

const (
	PingType           = 0x01
	RecordingType      = 0x02
	ClassificationType = 0x03
	PowerOffType       = 0x04
	version            = 0x01
	ManufactureID      = 0x1212
	AdapterID          = "hci0"
)

var deviceId uint16 = 0
var stopChannel = make(chan bool, 1)
var done sync.Mutex

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

func setDeviceID() {
	configRW, err := goconfig.New(goconfig.DefaultConfigDir)
	if err != nil {
		log.Printf("Cant read device config %v", err)
		return
	}

	var deviceConfig goconfig.Device
	if err := configRW.Unmarshal(goconfig.DeviceKey, &deviceConfig); err != nil {
		log.Printf("Cant read device config %v", err)
		return
	}
	deviceId = uint16(deviceConfig.ID)
	log.Printf("Using device id %v", deviceId)
}
func runMain() error {
	log.Println("Reading deviceId")
	setDeviceID()

	log.Println("Starting Beacon service")
	if err := startService(); err != nil {
		return err
	}
	return nil
}

func Ping() error {
	log.Println("Ping")
	expose(PingType, []byte{}, 30)
	return nil
}

func Stop(stopFunc func()) {
	done.Lock()
	for len(stopChannel) > 0 {
		<-stopChannel
	}
	select {
	case <-stopChannel:
	case <-time.After(10 * time.Second):
	}
	log.Println("Stopping")
	stopFunc()
	done.Unlock()
}
func Recording() error {
	log.Println("Recording")
	return expose(RecordingType, []byte{}, 10)
}

func Classification(classifications map[byte]byte) error {
	log.Println("Classification")
	return expose(ClassificationType, classificationToByteArray(classifications), 30)
}

func classificationToByteArray(classifications map[byte]byte) []byte {
	p := make(PairList, len(classifications))
	i := 0
	for k, v := range classifications {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	var maxPredictions int = 5
	if len(classifications) < 5 {
		maxPredictions = len(classifications)
	}
	data := make([]byte, maxPredictions*2+1, maxPredictions*2+1)
	data[0] = byte(maxPredictions)
	for i = 0; i < maxPredictions; i++ {
		data[(i*2)+1] = p[i].Key
		data[(i*2)+2] = p[i].Value

	}
	return data
}

func PowerOff(seconds uint16) error {
	log.Println("PowerOff")
	secondsByte := make([]byte, 2)
	binary.BigEndian.PutUint16(secondsByte, seconds)
	expose(PowerOffType, secondsByte, 30)
	return nil
}

func deviceIdInBytes() []byte {
	id := make([]byte, 2)
	binary.BigEndian.PutUint16(id, deviceId)
	return id
}

func expose(dataType byte, data []byte, timeout uint16) error {
	stopChannel <- true
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
	props.Timeout = timeout
	props.DiscoverableTimeout = timeout
	props.Duration = timeout

	props.Appearance = 0xFFFF // disables it
	//_, err := api.ExposeAdvertisement(AdapterID, props, uint32(timeout))
	f, err := api.ExposeAdvertisement(AdapterID, props, uint32(timeout))
	if err != nil {
		log.Printf("Error exposing %v", err)
		return err
	}
	go Stop(f)
	return err
}

type Pair struct {
	Key   byte
	Value byte
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
