package zonemgr

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/net"

	v1 "kubevirt.io/api/core/v1"
)

const (
	refresh = "3600"    // 1 hour (seconds) - how long a nameserver should wait prior to checking for a Serial Number increase within the primary zone file
	retry   = "3600"    // 1 hour (seconds) - how long a nameserver should wait prior to retrying to update a zone after a failed attempt.
	expire  = "1209600" // 2 weeks (seconds) - how long a nameserver should wait prior to considering data from a secondary zone invalid and stop answering queries for that zone
	ttl     = "3600"    // 1 hour (seconds) - the duration that the record may be cached by any resolver

	domainDefault     = "vm"
	nameServerDefault = "ns"
	adminEmailDefault = "email"
)

type ZoneFile struct {
	soaSerial      int
	adminEmail     string
	nameServerName string
	nameServerIP   string
	domain         string

	headerPref string
	headerSuf  string

	header   string
	aRecords string
	content  string

	vmiRecordsMap map[string][]string
}

func NewZoneFile(nameServerIP string, domain string) *ZoneFile {
	return &ZoneFile{
		nameServerIP: nameServerIP,
		domain:       domain,
	}
}

func (zoneFile *ZoneFile) init() {
	zoneFile.initCustomFields()
	zoneFile.generateHeaderPrefix()
	zoneFile.generateHeaderSuffix()
	zoneFile.soaSerial = 0
	zoneFile.header = zoneFile.generateHeader()
	zoneFile.content = zoneFile.header
	zoneFile.vmiRecordsMap = make(map[string][]string)
}

func (zoneFile *ZoneFile) initCustomFields() {
	if zoneFile.domain == "" {
		zoneFile.domain = domainDefault
	} else {
		zoneFile.domain = fmt.Sprintf("%s.%s", domainDefault, zoneFile.domain)
	}
	zoneFile.nameServerName = fmt.Sprintf("%s.%s", nameServerDefault, zoneFile.domain)
	zoneFile.adminEmail = fmt.Sprintf("%s.%s", adminEmailDefault, zoneFile.domain)
}

func (zoneFile *ZoneFile) generateHeaderPrefix() {
	zoneFile.headerPref = fmt.Sprintf("$ORIGIN %s. \n$TTL %s \n@ IN SOA %s. %s. (", zoneFile.domain, ttl,
		zoneFile.nameServerName, zoneFile.adminEmail)
}

func (zoneFile *ZoneFile) generateHeaderSuffix() {
	zoneFile.headerSuf = fmt.Sprintf(" %s %s %s %s)\n", refresh, retry, expire, ttl)

	if zoneFile.nameServerIP != "" {
		zoneFile.headerSuf += fmt.Sprintf("IN NS %s.\n", zoneFile.nameServerName)
		zoneFile.headerSuf += fmt.Sprintf("IN A %s\n", zoneFile.nameServerIP)
	}
}

func (zoneFile *ZoneFile) generateHeader() string {
	return zoneFile.headerPref + strconv.Itoa(zoneFile.soaSerial) + zoneFile.headerSuf
}

func (zoneFile *ZoneFile) updateVMIRecords(namespacedName k8stypes.NamespacedName, interfaces []v1.VirtualMachineInstanceNetworkInterface) bool {
	key := fmt.Sprintf("%s_%s", namespacedName.Name, namespacedName.Namespace)
	isUpdated := false

	if interfaces == nil {
		if zoneFile.vmiRecordsMap[key] != nil {
			delete(zoneFile.vmiRecordsMap, key)
			isUpdated = true
		}
	} else {
		newRecords := buildARecordsArr(namespacedName.Name, namespacedName.Namespace, interfaces)
		isUpdated = !reflect.DeepEqual(newRecords, zoneFile.vmiRecordsMap[key])
		if isUpdated {
			zoneFile.vmiRecordsMap[key] = newRecords
		}
	}

	if isUpdated {
		zoneFile.updateContent()
	}
	return isUpdated
}

func buildARecordsArr(name string, namespace string, interfaces []v1.VirtualMachineInstanceNetworkInterface) []string {
	var recordsArr []string
	for _, iface := range interfaces {
		if iface.Name != "" {
			IPs := iface.IPs
			for _, IP := range IPs {
				if net.IsIPv4String(IP) {
					recordsArr = append(recordsArr, generateARecord(name, namespace, iface.Name, IP))
					break
				}
			}
		}
	}
	sort.Strings(recordsArr)
	return recordsArr
}

func generateARecord(name string, namespace string, ifaceName string, ifaceIP string) string {
	fqdn := fmt.Sprintf("%s.%s.%s", ifaceName, name, namespace)
	return fmt.Sprintf("%s IN A %s\n", fqdn, ifaceIP)
}

func (zoneFile ZoneFile) generateARecords() string {
	aRecords := ""
	for _, recordsArr := range zoneFile.vmiRecordsMap {
		for _, aRecord := range recordsArr {
			aRecords += aRecord
		}
	}
	return aRecords
}

func (zoneFile *ZoneFile) updateContent() {
	zoneFile.soaSerial++
	zoneFile.header = zoneFile.generateHeader()
	zoneFile.aRecords = zoneFile.generateARecords()

	zoneFile.content = zoneFile.header + zoneFile.aRecords
}
