package zonemgr

import (
	"os"
)

const zoneFilePerm = 0644

type DnsZoneFile struct {
	zoneFileFullName string
}

func NewDnsZoneFile(fileName string) *DnsZoneFile {
	return &DnsZoneFile{
		zoneFileFullName: fileName,
	}
}

func (dnsZoneFile *DnsZoneFile) writeFile(content string) (err error) {
	return os.WriteFile(dnsZoneFile.zoneFileFullName, []byte(content), zoneFilePerm) //todo check permissions correct
}
