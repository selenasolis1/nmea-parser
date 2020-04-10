package main

import (
	"fmt"
	"time"
)

// WindData PGN 130306
type WindData struct {
	SID           byte
	WindSpeed     float64
	WindDirection uint16
	WindReference uint16
	ReservedBits  uint16
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
	windSpeed := uint16(data[1]) | uint16(data[2])<<8
	wd.WindSpeed = float64(windSpeed) * .01
	wd.WindDirection = uint16(data[3]) | uint16(data[4])<<8
	fmt.Println("wind data: ", wd)
}
