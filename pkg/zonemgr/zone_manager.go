package zonemgr

import (
	"errors"
	"fmt"
	"os"

	k8stypes "k8s.io/apimachinery/pkg/types"
	v1 "kubevirt.io/api/core/v1"
)

const (
	envVarDomain       = "DOMAIN"
	envVarNameServerIP = "NAME_SERVER_IP"
	zoneFileNamePrefix = "/zones/db."
	domainDefault      = "vm"
)

type SecIfaceData struct {
	interfaceName string
	interfaceIP   string
	namespaceName string
	vmName        string
}

type ZoneManager struct {
	zoneFileCache *ZoneFileCache
	zoneFile      *ZoneFile
}

func NewZoneManager() (*ZoneManager, error) {
	zoneMgr := &ZoneManager{}
	err := zoneMgr.prepare()
	return zoneMgr, err
}

func (zoneMgr *ZoneManager) prepare() error {
	domain := domainDefault
	nameServerIP := os.Getenv(envVarNameServerIP)
	if customDomain := os.Getenv(envVarDomain); customDomain != "" {
		domain = fmt.Sprintf("%s.%s", domain, customDomain)
	}
	zoneFileName := zoneFileNamePrefix + domain
	zoneMgr.zoneFile = NewZoneFile(zoneFileName)

	soaSerial, err := zoneMgr.zoneFile.ReadSoaSerial()
	if err != nil {
		return err
	}
	zoneMgr.zoneFileCache = NewZoneFileCache(nameServerIP, domain, soaSerial)
	return nil
}

func (zoneMgr *ZoneManager) UpdateZone(namespacedName k8stypes.NamespacedName, interfaces []v1.VirtualMachineInstanceNetworkInterface) error {
	if namespacedName.Name == "" {
		return errors.New("VM name in empty")
	}
	if namespacedName.Namespace == "" {
		return errors.New("VM namespace is empty")
	}

	if isUpdated := zoneMgr.zoneFileCache.UpdateVMIRecords(namespacedName, interfaces); isUpdated {
		return zoneMgr.zoneFile.WriteFile(zoneMgr.zoneFileCache.Content)
	}

	return nil
}
