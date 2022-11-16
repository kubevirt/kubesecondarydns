package zonemgr

import (
	"fmt"
	v1 "kubevirt.io/api/core/v1"
	"strconv"
)

const (
	refresh = "3600"    // 1 hour (seconds) - how long a nameserver should wait prior to checking for a Serial Number increase within the primary zone file
	retry   = "3600"    // 1 hour (seconds) - how long a nameserver should wait prior to retrying to update a zone after a failed attempt.
	expire  = "1209600" // 2 weeks (seconds) - how long a nameserver should wait prior to considering data from a secondary zone invalid and stop answering queries for that zone
	ttl     = "3600"    // 1 hour (seconds) - the duration that the record may be cached by any resolver
)

type ZoneFile struct {
	soaSerial          int
	clusterName        string
	userAdminEmail     string
	userNameserverName string
	userNameserverIP   string
	userSubdomain      string

	domainName string
	headerPref string
	headerSuf  string

	header   string
	aRecords string
	content  string

	vmiRecordsMap map[string]map[string]ARecord // key=vmi_name+namespace, value=map[key=iface_name, value=ip]
	//ifacesMap map[string]string
}

func NewZoneFile(clusterName string, userAdminEmail string, userNameserverName string,
	userNameserverIP string, userSubdomain string) *ZoneFile {
	return &ZoneFile{
		clusterName:        clusterName,
		userAdminEmail:     userAdminEmail,
		userNameserverName: userNameserverName,
		userNameserverIP:   userNameserverIP,
		userSubdomain:      userSubdomain,
	}
}

func (zoneFile *ZoneFile) init() {
	zoneFile.setDefaultValues()
	//before SOA serial
	zoneFile.generateHeaderPrefix()
	//after SOA serial
	zoneFile.generateHeaderSuffix()
	zoneFile.soaSerial = 0 // TODO make default RANDOM: time.Now().UnixNano()
	zoneFile.header = zoneFile.generateHeader()
	zoneFile.content = zoneFile.header
	zoneFile.vmiRecordsMap = make(map[string]map[string]ARecord)
}

func (zoneFile *ZoneFile) setDefaultValues() {
	zoneFile.domainName = fmt.Sprintf("%s.secondary", zoneFile.clusterName)

	if zoneFile.userSubdomain != "" {
		zoneFile.domainName = fmt.Sprintf("%s.%s", zoneFile.domainName, zoneFile.userSubdomain)
	}

	if zoneFile.userNameserverName == "" {
		zoneFile.userNameserverName = fmt.Sprintf("ns1.%s", zoneFile.domainName)
	}
	/*
		if zoneFile.userNameserverIP == "" {
			//TODO default???
		}
	*/
	if zoneFile.userAdminEmail == "" {
		zoneFile.userAdminEmail = fmt.Sprintf("admin.%s", zoneFile.domainName)
	}
}

/*
$ORIGIN mycluster.secondary.mydomain.com
$TTL 1750
@       IN      SOA ns1.mycluster.secondary.mydomain.com admin@secdns.com (
*/
func (zoneFile *ZoneFile) generateHeaderPrefix() {
	zoneFile.headerPref = fmt.Sprintf("\n$ORIGIN %s \n$TTL %s \n@ IN SOA %s %s (", zoneFile.domainName, ttl,
		zoneFile.userNameserverName, zoneFile.userAdminEmail)

	//fmt.Println("======= prefix:\n", zoneFile.headerPref)

}

/*
									7200       ; refresh (2 hours)
	                                3600       ; retry (1 hour)
	                                1209600    ; expire (2 weeks)
	                                3600       ; minimum (1 hour)
	                                )
	        IN      NS  ns1.mycluster.secondary.mydomain.com
	        IN      A   1.2.3.4
*/
func (zoneFile *ZoneFile) generateHeaderSuffix() {
	zoneFile.headerSuf = fmt.Sprintf(" %s %s %s %s)\n", refresh, retry, expire, ttl)

	if zoneFile.userNameserverName != "" {
		zoneFile.headerSuf += fmt.Sprintf("IN NS %s\n", zoneFile.userNameserverName)
	}

	if zoneFile.userNameserverIP != "" {
		zoneFile.headerSuf += fmt.Sprintf("IN A %s\n", zoneFile.userNameserverIP)
	}
}

func (zoneFile *ZoneFile) generateHeader() (header string) {
	return zoneFile.headerPref + strconv.Itoa(zoneFile.soaSerial) + zoneFile.headerSuf
}

//////////// maps impl

func (zoneFile *ZoneFile) updateVMIRecords(vmiName string, vmiNamespace string, interfaces []v1.VirtualMachineInstanceNetworkInterface) (isUpdated bool, err error) {
	key := fmt.Sprintf("%s_%s", vmiName, vmiNamespace)
	isUpdated = false
	//delete vmi records from the list
	if interfaces == nil {
		delete(zoneFile.vmiRecordsMap, key)
		isUpdated = true
	} else {
		ifacesMap := zoneFile.vmiRecordsMap[key]
		//if not exist or changed - override
		if ifacesMap == nil || isIfaceListChanged(ifacesMap, interfaces) {
			zoneFile.vmiRecordsMap[key] = buildInterfacesMap(vmiName, vmiNamespace, interfaces)
			isUpdated = true
		}
	}

	fmt.Println("********** zoneFile.vmiRecordsMap: ", zoneFile.vmiRecordsMap)

	if isUpdated {
		zoneFile.updateContent()
	}
	return isUpdated, nil
}

type ARecord struct {
	value string
	ip    string
}

func isIfaceListChanged(ifacesMap map[string]ARecord, interfaces []v1.VirtualMachineInstanceNetworkInterface) (isChanged bool) {
	if len(ifacesMap) != len(interfaces) {
		return true
	}

	for _, newIface := range interfaces {
		if ifacesMap[newIface.Name].ip != newIface.IP {
			return true
		}
	}
	return false
}

func buildInterfacesMap(vmiName string, vmiNamespace string, interfaces []v1.VirtualMachineInstanceNetworkInterface) (ifaceMap map[string]ARecord) {
	ifaceMap = make(map[string]ARecord)
	for _, iface := range interfaces {
		ifaceMap[iface.Name] = ARecord{value: generateARecord(vmiName, vmiNamespace, iface.Name, iface.IP), ip: iface.IP}
	}
	return ifaceMap
}

func generateARecord(vmiName string, vmiNamespace string, ifaceName string, ifaceIP string) (aRecord string) {
	//iface1.ns1.vm1   IN      A           5.6.7.8
	fqdn := fmt.Sprintf("%s.%s.%s", ifaceName, vmiNamespace, vmiName)
	return fmt.Sprintf("%s IN A %s\n", fqdn, ifaceIP)
}

func (zoneFile ZoneFile) generateARecords() (aRecords string) {
	for _, ifaceMap := range zoneFile.vmiRecordsMap {
		for _, aRecord := range ifaceMap {
			aRecords += aRecord.value
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
