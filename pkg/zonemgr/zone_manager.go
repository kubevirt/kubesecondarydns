package zonemgr

import (
	"errors"
	"fmt"
	"os"

	k8stypes "k8s.io/apimachinery/pkg/types"
	v1 "kubevirt.io/api/core/v1"

	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr/internal/zone-file"
	"github.com/kubevirt/kubesecondarydns/pkg/zonemgr/internal/zone-file-cache"
)

const (
	envVarDomain       = "DOMAIN"
	envVarNameServerIP = "NAME_SERVER_IP"
	zoneFileNamePrefix = "/zones/db."
	domainDefault      = "vm"
)

type ZoneManager struct {
	zoneFileCache *zone_file_cache.ZoneFileCache
	zoneFile      zone_file.ZoneFileInterface
}

func NewZoneManager() (*ZoneManager, error) {
	return NewZoneManagerWithParams(zone_file_cache.NewZoneFileCache, zone_file.NewZoneFile)
}

func NewZoneManagerWithParams(newZoneFileCache func(string, string, *int) *zone_file_cache.ZoneFileCache,
	newZoneFile func(string) zone_file.ZoneFileInterface) (*ZoneManager, error) {
	zoneMgr := &ZoneManager{}
	err := zoneMgr.prepare(newZoneFileCache, newZoneFile)
	return zoneMgr, err
}

func (zoneMgr *ZoneManager) prepare(newZoneFileCache func(string, string, *int) *zone_file_cache.ZoneFileCache,
	newZoneFile func(string) zone_file.ZoneFileInterface) error {
	domain := domainDefault
	nameServerIP := os.Getenv(envVarNameServerIP)
	if customDomain := os.Getenv(envVarDomain); customDomain != "" {
		domain = fmt.Sprintf("%s.%s", domain, customDomain)
	}
	zoneFileName := zoneFileNamePrefix + domain
	zoneMgr.zoneFile = newZoneFile(zoneFileName)

	soaSerial, err := zoneMgr.zoneFile.ReadSoaSerial()
	if err != nil {
		return err
	}
	zoneMgr.zoneFileCache = newZoneFileCache(nameServerIP, domain, soaSerial)
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
