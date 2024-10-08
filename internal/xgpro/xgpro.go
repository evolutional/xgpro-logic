package xgpro

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/komkom/toml"
)

const (
	lgcFileFlag     = 0xABABABEE
	lgcMaxItemCount = 512

	lgcFileVoltageLevel5v0 = 0
	lgcFileVoltageLevel3v3 = 1
	lgcFileVoltageLevel2v5 = 2
	lgcFileVoltageLevel1v8 = 3

	lgcVectorInputLow   = 0
	lgcVectorInputHigh  = 1
	lgcVectorOutputLow  = 2
	lgcVectorOutputHigh = 3
	lgcVectorInputPulse = 4
	lgcVectorHighZ      = 5
	lgcVectorIgnore     = 6
	lgcVectorGround     = 7
	lgcVectorVCC        = 8
)

type lgcFileHeader struct {
	AllCrc32  uint32
	UIFlag    uint32
	ItemCount uint32
	Res       uint32
	ItemStart [lgcMaxItemCount]uint32
}

type lgcFileItem struct {
	VectorCount  uint32
	ItemName     [32]byte
	VoltageLevel byte
	PinCount     byte
	Res0         byte
	Res1         byte
	UIRes        uint32
}

type lgcLogicVectors struct {
	Vectors [24]byte
}

type lgcFileEntry struct {
	item    lgcFileItem
	vectors []lgcLogicVectors
}

type lgcFile struct {
	header  lgcFileHeader
	entries []lgcFileEntry
}

type tomlFile struct {
	ICs []tomlIC
}

type tomlIC struct {
	Name    string
	Pins    uint32
	Vcc     float64
	Vectors []string
}

func ParseJsonFile(fileName string) (*lgcFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	f := tomlFile{}
	err = decoder.Decode(&f)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode json file: %s", err)
	}

	return parseJsonFile(&f)
}

func ParseTomlFile(fileName string) (*lgcFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(toml.New(file))

	f := tomlFile{}
	err = decoder.Decode(&f)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode toml file: %s", err)
	}
	return parseJsonFile(&f)
}

func DescribeToml(lgc *lgcFile, file *os.File) error {
	writer := bufio.NewWriter(file)
	return writeToml(writer, lgc)
}

func DescribeJson(lgc *lgcFile, file *os.File) error {
	writer := bufio.NewWriter(file)
	return writeJson(writer, lgc)
}

func DescribeXml(lgc *lgcFile, file *os.File) error {
	writer := bufio.NewWriter(file)
	return writeXml(writer, lgc)
}

func WriteToml(fileName string, lgc *lgcFile) error {

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Failed to create file: %s", err)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	return writeToml(writer, lgc)
}

func WriteLgc(fileName string, lgc *lgcFile) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Failed to open file: %s", err)
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, &lgc.header)
	if err != nil {
		return fmt.Errorf("Failed to write file header: %s", err)
	}

	for itemID := 0; itemID < len(lgc.entries); itemID++ {
		entry := &lgc.entries[itemID]
		offset, _ := file.Seek(0, io.SeekCurrent)
		lgc.header.ItemStart[itemID] = uint32(offset)
		err = binary.Write(file, binary.LittleEndian, entry.item)
		if err != nil {
			return fmt.Errorf("Failed to write file item (%d): %s", itemID, err)
		}

		for vectorID := 0; vectorID < len(entry.vectors); vectorID++ {
			vector := &entry.vectors[vectorID]
			err = binary.Write(file, binary.LittleEndian, vector)
			if err != nil {
				return fmt.Errorf("Failed to write file item (%d) vector (%d): %s", itemID, vectorID, err)
			}
		}
	}

	// Patch up the offset table
	file.Seek(0, io.SeekStart)
	err = binary.Write(file, binary.LittleEndian, &lgc.header)
	if err != nil {
		return fmt.Errorf("Failed to write file header: %s", err)
	}

	return nil
}

func ParseLGCFile(fileName string) (*lgcFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %s", err)
	}
	defer file.Close()
	fileHeader := lgcFileHeader{}

	err = binary.Read(file, binary.LittleEndian, &fileHeader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file header: %s", err)
	}

	lgc := lgcFile{}
	lgc.header = fileHeader
	lgc.entries = make([]lgcFileEntry, fileHeader.ItemCount)

	for itemID := 0; itemID < int(fileHeader.ItemCount); itemID++ {
		entry, vectors, err := readFileItem(file, &fileHeader, itemID)
		if err != nil {
			return nil, err
		}

		lgc.entries[itemID].item = *entry
		lgc.entries[itemID].vectors = vectors
	}

	return &lgc, nil
}

func ConvertFile(inputFileName string, inputFormat string, outputFileName string) error {
	var lgc *lgcFile
	var err error

	switch inputFormat {
	case "toml":
		lgc, err = ParseTomlFile(inputFileName)
	case "json":
		lgc, err = ParseJsonFile(inputFileName)
	}

	if err != nil {
		return err
	}
	return WriteLgc(outputFileName, lgc)
}

func DumpLGCFile(lgc *lgcFile, file *os.File) error {
	fileHeader := &lgc.header

	fmt.Fprintf(file, "File contains %d entries\n", fileHeader.ItemCount)
	for itemID := 0; itemID < int(fileHeader.ItemCount); itemID++ {

		entry := lgc.entries[itemID]
		vectors := lgc.entries[itemID].vectors

		fmt.Fprintf(file, "Entry #%d\n", itemID)
		fmt.Fprintf(file, "\tName:\t%s\n", string(entry.item.ItemName[:]))
		fmt.Fprintf(file, "\tPins:\t%d\n", entry.item.PinCount)

		vcc := "INVALID"

		switch entry.item.VoltageLevel {
		case lgcFileVoltageLevel5v0:
			vcc = "5.0V"
		case lgcFileVoltageLevel3v3:
			vcc = "3.3V"
		case lgcFileVoltageLevel2v5:
			vcc = "2.5V"
		case lgcFileVoltageLevel1v8:
			vcc = "1.8V"
		}

		fmt.Fprintf(file, "\tVCC:\t%s\n", vcc)
		fmt.Fprintf(file, "\tVectors: %d\n", entry.item.VectorCount)

		for vectorID := 0; vectorID < int(entry.item.VectorCount); vectorID++ {
			vector := vectors[vectorID]
			fmt.Fprintf(file, "\t\t#%03d: ", vectorID)

			for vecByte := 0; vecByte < int(entry.item.PinCount/2); vecByte++ {
				pinLow := mapVector(vector.Vectors[vecByte] >> 4)
				pinHigh := mapVector(vector.Vectors[vecByte] & 0x0F)
				fmt.Fprintf(file, "%s %s ", pinHigh, pinLow)
			}
			fmt.Fprintln(file)
		}
	}

	return nil
}

func readFileItem(file *os.File, fileHeader *lgcFileHeader, itemID int) (*lgcFileItem, []lgcLogicVectors, error) {

	crcTable := crc32.IEEETable
	hashValue := uint32(0)

	item := lgcFileItem{}
	file.Seek(int64(fileHeader.ItemStart[itemID]), 0)

	err := binary.Read(file, binary.LittleEndian, &item)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read record %d: %s", itemID, err)
	}

	binBuf := bytes.Buffer{}
	binary.Write(&binBuf, binary.LittleEndian, item)
	hashValue = crc32.Update(hashValue, crcTable, binBuf.Bytes())

	vectors := make([]lgcLogicVectors, item.VectorCount)

	for vectorID := 0; vectorID < int(item.VectorCount); vectorID++ {
		vector := lgcLogicVectors{}
		err = binary.Read(file, binary.LittleEndian, &vector)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to read record %d vector %d: %s", itemID, vectorID, err)
		}
		vectors[vectorID] = vector
		binBuf.Reset()
		binary.Write(&binBuf, binary.LittleEndian, vector)
		hashValue = crc32.Update(hashValue, crcTable, binBuf.Bytes())
	}

	return &item, vectors, nil
}

func mapVector(vector byte) string {
	switch vector {
	case lgcVectorInputLow:
		return "0"
	case lgcVectorInputHigh:
		return "1"
	case lgcVectorOutputLow:
		return "L"
	case lgcVectorOutputHigh:
		return "H"
	case lgcVectorHighZ:
		return "Z"
	case lgcVectorInputPulse:
		return "C"
	case lgcVectorIgnore:
		return "X"
	case lgcVectorVCC:
		return "V"
	case lgcVectorGround:
		return "G"
	}
	return " "
}

func unmapVector(vector byte) byte {
	switch vector {
	case '0':
		return lgcVectorInputLow
	case '1':
		return lgcVectorInputHigh
	case 'L':
		return lgcVectorOutputLow
	case 'H':
		return lgcVectorOutputHigh
	case 'Z':
		return lgcVectorHighZ
	case 'C':
		return lgcVectorInputPulse
	case 'X':
		return lgcVectorIgnore
	case 'V':
		return lgcVectorVCC
	case 'G':
		return lgcVectorGround
	}
	panic("invalid vector")
}

func mapVoltageLevel(vcc float64) byte {
	switch vcc {
	case 5.0:
		return lgcFileVoltageLevel5v0
	case 3.3:
		return lgcFileVoltageLevel3v3
	case 2.5:
		return lgcFileVoltageLevel2v5
	case 1.8:
		return lgcFileVoltageLevel1v8
	}
	return 255
}

func unmapVoltageLevel(vcc byte) float64 {
	switch vcc {
	case lgcFileVoltageLevel5v0:
		return 5.0
	case lgcFileVoltageLevel3v3:
		return 3.3
	case lgcFileVoltageLevel2v5:
		return 2.5
	case lgcFileVoltageLevel1v8:
		return 1.8
	}
	return 255
}

func parseVectorString(vectorStr string) (*lgcLogicVectors, error) {
	result := lgcLogicVectors{}

	vByte := byte(0)
	vOffset := 0
	for i := 0; i < len(vectorStr); i++ {
		nib := unmapVector(vectorStr[i])
		if i%2 == 0 {
			vByte = nib
		} else {
			vByte |= (nib << 4)
			result.Vectors[vOffset] = vByte
			vOffset = vOffset + 1
		}
	}
	return &result, nil
}

func cleanItemName(input [32]byte) string {
	itemName := string(input[:])
	itemName = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, itemName)
	itemName = strings.Trim(itemName, " ")
	return itemName
}

func writeToml(writer *bufio.Writer, lgc *lgcFile) error {
	for icID := 0; icID < int(lgc.header.ItemCount); icID++ {
		entry := &lgc.entries[icID]

		itemName := cleanItemName(entry.item.ItemName)
		writer.WriteString("[[ics]]\n")
		fmt.Fprintf(writer, "name = \"%s\"\n", itemName)
		fmt.Fprintf(writer, "pins = %d\n", entry.item.PinCount)
		fmt.Fprintf(writer, "vcc = %0.1f\n", unmapVoltageLevel(entry.item.VoltageLevel))
		fmt.Fprintf(writer, "vectors = [\n")

		for vectorID := 0; vectorID < int(entry.item.VectorCount); vectorID++ {

			vectorStr := ""
			vector := entry.vectors[vectorID]

			for vecByte := 0; vecByte < int(entry.item.PinCount/2); vecByte++ {
				pinLow := mapVector(vector.Vectors[vecByte] >> 4)
				pinHigh := mapVector(vector.Vectors[vecByte] & 0x0F)
				vectorStr = fmt.Sprintf("%s%s%s", vectorStr, pinHigh, pinLow)
			}

			fmt.Fprintf(writer, "\t\"%s\",\n", vectorStr)
		}
		fmt.Fprintf(writer, "]\n")
	}
	writer.Flush()
	return nil
}

func writeJson(writer *bufio.Writer, lgc *lgcFile) error {
	writer.WriteString("{ \"ics\": [")
	for icID := 0; icID < int(lgc.header.ItemCount); icID++ {
		entry := &lgc.entries[icID]

		itemName := cleanItemName(entry.item.ItemName)

		writer.WriteString("{")
		fmt.Fprintf(writer, " \"name\": \"%s\",", itemName)
		fmt.Fprintf(writer, " \"pins\": %d,", entry.item.PinCount)
		fmt.Fprintf(writer, " \"vcc\": %0.1f,", unmapVoltageLevel(entry.item.VoltageLevel))
		fmt.Fprintf(writer, " \"vectors\": [")

		for vectorID := 0; vectorID < int(entry.item.VectorCount); vectorID++ {

			vectorStr := ""
			vector := entry.vectors[vectorID]

			for vecByte := 0; vecByte < int(entry.item.PinCount/2); vecByte++ {
				pinLow := mapVector(vector.Vectors[vecByte] >> 4)
				pinHigh := mapVector(vector.Vectors[vecByte] & 0x0F)
				vectorStr = fmt.Sprintf("%s%s%s", vectorStr, pinHigh, pinLow)
			}
			sep := ", "
			if vectorID == int(entry.item.VectorCount)-1 {
				sep = ""
			}
			fmt.Fprintf(writer, "\"%s\"%s", vectorStr, sep)
		}
		fmt.Fprintf(writer, "] }")

		if icID < int(lgc.header.ItemCount)-1 {
			fmt.Fprintf(writer, ", ")
		}

	}

	writer.WriteString("]}\n")
	writer.Flush()
	return nil
}

func writeXml(writer *bufio.Writer, lgc *lgcFile) error {
	writer.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	writer.WriteString("<infoic>\n")
	writer.WriteString("  <database device=\"TL866II\">\n")
	writer.WriteString("    <manufacturer name=\"Logic Ic\">\n")
	for icID := 0; icID < int(lgc.header.ItemCount); icID++ {
		entry := &lgc.entries[icID]

		itemName := cleanItemName(entry.item.ItemName)

		writer.WriteString("      <ic ")
		fmt.Fprintf(writer, "name=\"%s\" ", itemName)
		fmt.Fprintf(writer, "pins=\"%d\" ", entry.item.PinCount)
		fmt.Fprintf(writer, "voltage=\"%0.1fV\" type=\"5\">\n", unmapVoltageLevel(entry.item.VoltageLevel))

		for vectorID := 0; vectorID < int(entry.item.VectorCount); vectorID++ {

			vectorStr := fmt.Sprintf("        <vector id=\"%02d\">", vectorID)
			vector := entry.vectors[vectorID]

			for vecByte := 0; vecByte < int(entry.item.PinCount/2); vecByte++ {
				pinLow := mapVector(vector.Vectors[vecByte] >> 4)
				pinHigh := mapVector(vector.Vectors[vecByte] & 0x0F)
				vectorStr = fmt.Sprintf("%s %s %s", vectorStr, pinHigh, pinLow)
			}
			fmt.Fprintf(writer, "%s </vector>\n", vectorStr)
		}
		fmt.Fprintf(writer, "      </ic>\n")

	}
	writer.WriteString("    </manufacturer>\n")
	writer.WriteString("  </database>\n")
	writer.WriteString("</infoic>\n")
	writer.Flush()
	return nil
}

func parseJsonFile(f *tomlFile) (*lgcFile, error) {

	lgc := lgcFile{
		header: lgcFileHeader{
			AllCrc32:  0,
			ItemCount: uint32(len(f.ICs)),
			Res:       0,
			UIFlag:    lgcFileFlag,
		},
		entries: make([]lgcFileEntry, len(f.ICs)),
	}

	crcTable := crc32.IEEETable
	hashValue := uint32(0)

	for icID := 0; icID < len(f.ICs); icID++ {
		s := f.ICs[icID]

		paddedName := []byte(fmt.Sprintf("%-32s", s.Name))
		paddedName[len(s.Name)] = 0

		item := lgcFileItem{
			PinCount:     byte(s.Pins),
			VoltageLevel: mapVoltageLevel(s.Vcc),
			VectorCount:  uint32(len(s.Vectors)),
		}

		copy(item.ItemName[:], paddedName[:32])

		binBuf := bytes.Buffer{}
		binary.Write(&binBuf, binary.LittleEndian, item)
		hashValue = crc32.Update(hashValue, crcTable, binBuf.Bytes())

		entry := &lgc.entries[icID]
		entry.item = item
		entry.vectors = make([]lgcLogicVectors, len(s.Vectors))

		for vectorID := 0; vectorID < len(s.Vectors); vectorID++ {
			v, err := parseVectorString(s.Vectors[vectorID])
			if err != nil {
				return nil, fmt.Errorf("Failed to process vector (%d): %s", vectorID, err)
			}

			entry.vectors[vectorID] = *v

			binBuf.Reset()
			binary.Write(&binBuf, binary.LittleEndian, *v)
			hashValue = crc32.Update(hashValue, crcTable, binBuf.Bytes())
		}
	}

	lgc.header.AllCrc32 = hashValue

	return &lgc, nil
}
