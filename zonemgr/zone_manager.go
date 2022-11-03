package zonemgr

import (
	"fmt"
	k8stypes "k8s.io/apimachinery/pkg/types"
	v1 "kubevirt.io/api/core/v1"
	"os"
)

//todo error handling

type SecIfaceData struct {
	interfaceName string
	interfaceIP   string
	namespaceName string
	vmName        string
}

type ZoneManager struct {
	zoneFileCopy *ZoneFile
	zoneFileDns  *DnsZoneFile
}

func NewZoneManager() *ZoneManager {
	zoneMgr := &ZoneManager{}
	zoneMgr.init()
	return zoneMgr
}

func (zoneMgr *ZoneManager) init() (err error) {

	fmt.Println("================ inside ZoneManager.init ===============")

	//todo fetch custom details
	clusterName := "cluster"
	zoneFileName := "/zones/db.secdns"

	userAdminEmail := ""
	userSubdomain := "domain.com"
	userNameserverName := ""
	userNameserverIP := ""

	zoneMgr.zoneFileCopy = NewZoneFile(clusterName, userAdminEmail,
		userNameserverName, userNameserverIP, userSubdomain)

	zoneMgr.zoneFileCopy.init()

	//fmt.Println("================ zoneMgr.zoneFileCopy.content ===============\n", zoneMgr.zoneFileCopy.content)

	zoneMgr.zoneFileDns = NewDnsZoneFile(zoneFileName)
	// todo return zoneFileDns.writeFile(zoneFileCopy.content) // is it necessary to create an empty zone file on load??
	return nil
}

func (zoneMgr *ZoneManager) UpdateZone(namespacedName k8stypes.NamespacedName, interfaces []v1.VirtualMachineInstanceNetworkInterface) (err error) {

	//fmt.Println("================ inside ZoneManager.UpdateZone ===============")
	//fmt.Println("================ namespacedName: ===============", namespacedName)
	//fmt.Println("================ interfaces: ===============", interfaces)

	isUpdated, err := zoneMgr.zoneFileCopy.updateVMIRecords(namespacedName.Name, namespacedName.Namespace, interfaces)
	if err != nil {
		return err
	}

	//fmt.Println("================ zoneMgr.zoneFileCopy.content: ===============\n", zoneMgr.zoneFileCopy.content)
	//fmt.Println("================ zoneMgr.zoneFileDns: ===============\n", zoneMgr.zoneFileDns)

	//override on disk, create if not exist
	if isUpdated {
		err = zoneMgr.zoneFileDns.writeFile(zoneMgr.zoneFileCopy.content)
		fmt.Println("********* err: ", err)

		/////test
		text, err := os.ReadFile("/zones/db.secdns")
		fmt.Println("********* err: ", err)
		fmt.Println("********* text: ", string(text))
	}
	return err
}
