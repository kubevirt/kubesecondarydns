package zonemgr

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

const zoneFilePerm = 0644

var soaSerialReg = regexp.MustCompile("SOA .*\\(([0-9]+) ")

type ZoneFile struct {
	zoneFileFullName string
}

func NewZoneFile(fileName string) *ZoneFile {
	return &ZoneFile{
		zoneFileFullName: fileName,
	}
}

func (zoneFile *ZoneFile) WriteFile(content string) (err error) {
	return os.WriteFile(zoneFile.zoneFileFullName, []byte(content), zoneFilePerm)
}

func (zoneFile *ZoneFile) readFile() ([]byte, error) {
	return os.ReadFile(zoneFile.zoneFileFullName)
}

func (zoneFile *ZoneFile) isFileExist() (bool, error) {
	var err error
	isExist := false
	if _, err = os.Stat(zoneFile.zoneFileFullName); err == nil {
		isExist = true
	} else if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return isExist, err
}

func (zoneFile *ZoneFile) ReadSoaSerial() (*int, error) {
	if isFileExist, err := zoneFile.isFileExist(); !isFileExist || err != nil {
		return nil, err
	}
	if content, err := zoneFile.readFile(); content == nil || err != nil {
		return nil, err
	} else {
		return fetchSoaSerial(string(content))
	}
}

func fetchSoaSerial(content string) (*int, error) {
	if result := soaSerialReg.FindStringSubmatch(content); result != nil && len(result) > 0 {
		soaSerial := result[1]
		if soaSerialInt, err := strconv.Atoi(soaSerial); err == nil {
			return &soaSerialInt, nil
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New(fmt.Sprintf("failed to fetch SOA serial value from the zone file content: %s", content))
	}
}
