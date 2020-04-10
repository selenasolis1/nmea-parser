package main

import (
	"fmt"
	"time"
)

// WindData PGN 130306
type WindData struct {
	SID           byte
	WindSpeed     float64
	WindDirection float64
	WindReference float64
	ReservedBits  float64
}

type fieldData struct {
	bitSize          int
	requestParameter string
	commandParameter string
	maxRange         float64
	minRange         float64
	resolution       float64
}

// 16 2 147 19 2 2 253 1 255 36 108 176 2 0 8 64 15 0 102 108 248 255 255 248 16 3

// Message nmea2000 data
type Message struct {
	PacketLength byte
	Priority     byte
	PGN          uint32
	Destination  byte
	Source       byte
	TimeStamp    time.Time
	DataLength   byte
	Data         NMEAData
	CRC          byte
}

// NMEAData holds the 8 bytes of data that will need to be parsed dependent on PGN
type NMEAData struct {
	pgn  uint32
	data []byte
}

// PGNData is an interface that allows for the data to be parsed appropriately
type PGNData interface {
	parseData([]byte)
}

var pgnMap = map[uint32]PGNData{
	130306: &WindData{},
	// 127250: "VesselHeading",
}

var dataFieldMap = map[string]fieldData{
	"Wind Speed": fieldData{
		bitSize:          16,
		requestParameter: "optional",
		commandParameter: "optional",
		maxRange:         655.32,
		minRange:         0,
		resolution:       .01,
	},
	"Wind Direction": fieldData{
		bitSize:          16,
		requestParameter: "optional",
		commandParameter: "optional",
		maxRange:         6.2831853,
		minRange:         0,
		resolution:       .0001,
	},
	"Wind Reference": fieldData{
		bitSize:          3,
		requestParameter: "optional",
		commandParameter: "optional",
		maxRange:         6.2831853,
		minRange:         0,
		resolution:       1,
	},
}

func main() {
	bs := []byte{16, 2, 147, 19, 2, 2, 253, 1, 255, 36, 108, 176, 2, 0, 8, 64, 15, 0, 102, 108, 248, 255, 255, 248, 16, 3}
	msg := getRequest(bs)
	fmt.Println("msg: ", msg)
	msg.Data.Parse()
}

func getRequest(frame []byte) Message {
	pgn := uint32(frame[5]) | uint32(frame[6])<<8 | uint32(frame[7])<<16
	d := NMEAData{
		pgn:  pgn,
		data: frame[15 : len(frame)-3],
	}
	fmt.Println(pgn)
	msg := Message{
		PacketLength: frame[3],
		Priority:     frame[4],
		PGN:          pgn,
		Destination:  frame[8],
		Source:       frame[9],
		TimeStamp:    time.Now(),
		DataLength:   frame[14],
		Data:         d,
	}
	return msg
}

// Parse will use the PGN to create the correct data type and calculate the appropriate
// readable values for each data field.
func (nd *NMEAData) Parse() {
	t, err := createDataType(nd.pgn)
	if err != nil {
		fmt.Println(err)
	}
	t.parseData(nd.data)
	fmt.Println("t ", t)
}

func createDataType(pgn uint32) (PGNData, error) {
	v, ok := pgnMap[pgn]
	if !ok {
		err := fmt.Errorf("Unrecognizable pgn")
		return nil, err
	}
	fmt.Println("v ", v)
	return v, nil
}

func (wd *WindData) parseData(data []byte) {
	wd.SID = data[0]
	wd.WindSpeed = getValue("Wind Speed", data[1:3])
	wd.WindDirection = getValue("Wind Direction", data[3:5])
	wd.WindReference = getValue("Wind Reference", data[5:6])
	fmt.Println("wind data: ", wd)
}

func getValue(name string, values []byte) float64 {
	v, ok := dataFieldMap[name]
	if !ok {
		err := fmt.Errorf("Unrecognizable field name")
		fmt.Println(err)
	}
	len := len(values)

	switch len {
	case 1:
		if size := v.bitSize; size < 8 {
			fmt.Println("less than 8 bits")
			var intVal byte
			switch size {
			case 1:
				intVal = values[0] & 1
			case 2:
				intVal = values[0] & 3
			case 3:
				intVal = values[0] & 7
			case 4:
				intVal = values[0] & 15
			case 5:
				intVal = values[0] & 31
			case 6:
				intVal = values[0] & 63
			case 7:
				intVal = values[0] & 127
			}
			fmt.Println("intVal ", intVal)
			value := float64(intVal) * v.resolution
			fmt.Printf("name: %s, value: %v\n", name, value)
			return value
		}
		intVal := values[0]
		value := float64(intVal) * v.resolution
		fmt.Printf("name: %s, value: %v\n", name, value)
		return value
	case 2:
		intVal := uint16(values[0]) | uint16(values[1])<<8
		value := float64(intVal) * v.resolution
		fmt.Printf("name: %s, value: %v\n", name, value)
		return value
	case 3:
		intVal := uint32(values[0]) | uint32(values[1])<<8 | uint32(values[2])<<16
		value := float64(intVal) * v.resolution
		fmt.Printf("name: %s, value: %v\n", name, value)
		return value
	default:
		return 0
	}
}
