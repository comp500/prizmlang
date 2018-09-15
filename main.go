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
	Name            string // Limited to 16 bytes
	fileSize        uint32
	InternalName    string
	VersionNumber   string // Must be XX.XX.XXXX
	DateCreated     time.Time
	Salutation      string // Limited to 16 bytes
	FileName        string // Limited to 16 bytes
	Messages        map[int]string
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

	err = ioutil.WriteFile("out.json", json, 0644)
	if err != nil {
		log.Fatal(err)
	}

	data2, err := writeFile(file)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("out.bin", data2, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readFile(data []byte) (LangFile, error) {
	var file LangFile

	// Header
	index := 0

	index += 14             // Const: b3 86 c8 ca ca df df df d3 ff fe ff fe ff
	file.inv1 = data[index] // TODO: check this is correct?
	index++
	index++                                  // Const: fe
	file.invFileSize = data[index : index+4] // TODO: check this is correct?
	index += 4
	file.inv2 = data[index] // TODO: check this is correct?
	index++
	index += 11 // Const: 00 ff ff 00 00 00 00 00 00 00 00
	// check this is correct?
	// calculator does not check this, could break stuff
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
	// file.Name = string(bytes.Trim(data[index:index+16], "\x00"))
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
	file.Messages = make(map[int]string, messageCount)
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
	// TODO: check checksums are equal?
	index = int(file.execSectionSize) + 0x1000
	// file.checksum = data[index : index+4]
	index += 4

	return file, nil
}

// padBuf pads a buffer with n of 0x00
func padBuf(b *bytes.Buffer, n int) (int, error) {
	return b.Write(make([]byte, n))
}

// writePadString pads/truncates s to n, and then writes to b
func writePadString(b *bytes.Buffer, s string, n int) (int, error) {
	slice := make([]byte, n)
	copy(slice, s)
	return b.Write(slice)
}

func getMaxIndex(messages map[int]string) int {
	var maxNumber int
	for maxNumber = range messages {
		break
	}
	for n := range messages {
		if n > maxNumber {
			maxNumber = n
		}
	}
	return maxNumber
}

func writeFile(file LangFile) ([]byte, error) {
	var b bytes.Buffer

	// Set up variables
	const fixedLength = 87138 // TODO: make this variable, calculate at end?
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, fixedLength)
	invLengthBytes := make([]byte, 4)
	for i := range invLengthBytes {
		// Calculate inverses
		invLengthBytes[i] = lengthBytes[i] ^ 0xff
	}
	// Low byte of inverse - 0x41
	inv1 := invLengthBytes[3] - 0x41
	// Low byte of inverse - 0xB8
	inv2 := invLengthBytes[3] - 0xB8
	checksum := []byte("\x00\x00\x00\x00")        // Null checksum
	execSectionSize := (fixedLength - 4) - 0x1000 // Executable section length
	execSectionSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(execSectionSizeBytes, uint32(execSectionSize))
	timeString := file.DateCreated.Format("2006.0102.1504")
	messageCount := getMaxIndex(file.Messages) + 1
	messageCountBytes := make([]byte, 4)
	// -1 as it is zero-based?
	binary.BigEndian.PutUint32(messageCountBytes, uint32(messageCount-1))
	messageOffsetBytes := make([]byte, messageCount*4)
	messageOffsetBlankBytes := []byte{0xff, 0xff, 0xff, 0xff}
	currentOffset := 0
	for i := 0; i < messageCount; i++ { // Go in index order
		if v, ok := file.Messages[i]; ok { // If exists, write
			// Put uint32s every 4 bytes into messageOffsetBytes
			binary.BigEndian.PutUint32(messageOffsetBytes[i*4:(i*4)+4], uint32(currentOffset))
			currentOffset += len(v) + 1 // +1 for null byte
		} else {
			copy(messageOffsetBytes[i*4:(i*4)+4], messageOffsetBlankBytes)
		}
	}

	// Allocate at least 0x1000, as header is 0x1000
	//b.Grow(0x1000)
	b.Grow(fixedLength)
	// Const: b3 86 c8 ca ca df df df d3 ff fe ff fe ff
	b.WriteString("\xb3\x86\xc8\xca\xca\xdf\xdf\xdf\xd3\xff\xfe\xff\xfe\xff")
	b.WriteByte(inv1)
	b.WriteByte(0xfe) // Const: 0xfe
	b.Write(invLengthBytes)
	b.WriteByte(inv2)
	// Const: 00 ff ff 00 00 00 00 00 00 00 00
	b.WriteString("\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")
	b.Write(checksum)
	b.WriteString("\x04\x01") // Const: 04 01
	padBuf(&b, 16)
	b.Write(execSectionSizeBytes)
	padBuf(&b, 6)
	writePadString(&b, file.Name, 28)
	b.Write(lengthBytes)
	// Internal name starts at 0x60, goes to 0x130
	writePadString(&b, file.InternalName, (0x130 - 0x60))
	writePadString(&b, file.VersionNumber, 10)
	padBuf(&b, 2)
	b.WriteString(timeString)
	// Move to 0x0e9c
	padBuf(&b, 0x0e9c-0x14a)
	writePadString(&b, file.Name, 16)
	writePadString(&b, file.Salutation, 16)
	writePadString(&b, file.FileName, 16)

	// Move to 0x1000, executable section
	padBuf(&b, 0x1000-0xecc)
	// Const: 4c 59 37 35 35 00 00 00 02 01
	b.WriteString("\x4c\x59\x37\x35\x35\x00\x00\x00\x02\x01")
	b.Write(messageCountBytes)
	padBuf(&b, 2)
	b.Write(messageOffsetBytes)
	for i := 0; i < messageCount; i++ { // Go in index order
		if v, ok := file.Messages[i]; ok { // If exists, write
			b.WriteString(v)
			b.WriteByte(0x00) // null-terminated
		}
	}

	// Pad out to fixedLength
	existingLen := b.Len()
	output := b.Bytes()[:fixedLength]
	for i := existingLen; i < fixedLength; i++ {
		log.Print(i)
		output[i] = 0xff // pad with 0xff
	}
	copy(output[fixedLength-4:fixedLength], checksum) // Write checksum to end

	return output, nil
}
