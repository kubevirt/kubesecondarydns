package zonemgr

import (
	"os"
)

const zoneFilePerm = 0644

type ZoneFile struct {
	zoneFileFullName string
}

func NewZoneFile(fileName string) *ZoneFile {
	return &ZoneFile{
		zoneFileFullName: fileName,
	}
}

func (zoneFile *ZoneFile) writeFile(content string) (err error) {
	return os.WriteFile(zoneFile.zoneFileFullName, []byte(content), zoneFilePerm)
}
