package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"time"
)

// LangFile is a language file data struct
type LangFile struct {
	inv1            byte
	invFileSize     []byte
	inv2            byte
	checksum        []byte
	execSectionSize uint32
	Name            string
	fileSize        uint32
	InternalName    string
	VersionNumber   string
	DateCreated     time.Time
	Salutation      string
	FileName        string
	Messages        []string
	unknown         []byte
}

func main() {
	dat, err := ioutil.ReadFile("English.g3l")
	if err != nil {
		log.Fatal(err)
	}

	file, err := readFile(dat)
	if err != nil {
		log.Fatal(err)
	}

	json, err := json.MarshalIndent(file, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	//log.Printf("%s\n", string(json))
	//log.Printf("%#v\n", file)
	err = ioutil.WriteFile("out.json", json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readFile(data []byte) (LangFile, error) {
	var file LangFile

	// Header
	index := 0

	index += 14 // Const: b3 86 c8 ca ca df df df d3 ff fe ff fe ff
	file.inv1 = data[index]
	index++
	index++ // Const: fe
	file.invFileSize = data[index : index+4]
	index += 4
	file.inv2 = data[index]
	index++
	index += 11 // Const: 00 ff ff 00 00 00 00 00 00 00 00
	file.checksum = data[index : index+4]
	index += 4
	index += 2  // Const: 04 01
	index += 16 // Padding
	file.execSectionSize = binary.BigEndian.Uint32(data[index : index+4])
	index += 4
	index += 6 // Padding
	file.Name = string(bytes.Trim(data[index:index+28], "\x00"))
	index += 28
	file.fileSize = binary.BigEndian.Uint32(data[index : index+4])
	index += 4
	file.InternalName = string(bytes.Trim(data[index:0x130], "\x00"))
	// String end / padding start is undefined so just set to 0x130
	index = 0x130
	file.VersionNumber = string(data[index : index+10])
	index += 10
	index += 2 // Padding
	var err error
	file.DateCreated, err = time.Parse("2006.0102.1504", string(data[index:index+14]))
	if err != nil {
		return file, err
	}
	index += 14
	// Warp!
	index = 0x0e9c
	// TODO: check names are equal?
	// _ = string(bytes.Trim(data[index:index+16], "\x00"))
	index += 16
	file.Salutation = string(bytes.Trim(data[index:index+16], "\x00"))
	index += 16
	file.FileName = string(bytes.Trim(data[index:index+16], "\x00"))
	index += 16

	// Executable section
	index = 0x1000

	index += 10 // Const: 4c 59 37 35 35 00 00 00 02 01
	// +1 as it is zero-based?
	messageCount := binary.BigEndian.Uint32(data[index:index+4]) + 1
	index += 4
	index += 2 // Padding
	// Make message arrays
	messageOffsets := make([]int, messageCount)
	file.Messages = make([]string, messageCount)
	for i := range messageOffsets {
		messageOffset := binary.BigEndian.Uint32(data[index : index+4])
		index += 4
		messageOffsets[i] = int(messageOffset)
	}
	// After getting to first message contents, read actual message data
	for i, v := range messageOffsets {
		// offset == 4294967295 means it doesn't exist
		if v != 4294967295 && v < (len(data)-index) {
			// Search index + offset to find 0x00
			numBytes := bytes.IndexByte(data[index+v:], 0)
			if numBytes != -1 {
				file.Messages[i] = string(data[index+v : index+v+numBytes])
			} else {
				return file, errors.New("Invalid string, could not find null byte")
			}
		}
	}

	file.unknown = data[index : index+100]
	return file, nil
}
