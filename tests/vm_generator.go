package tests

import (
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/api/core/v1"
)

const virtIO = "virtio"

func getBaseVMISpec() *v1.VirtualMachineInstanceSpec {
	return &v1.VirtualMachineInstanceSpec{
		Domain: v1.DomainSpec{
			Resources: v1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					k8sv1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
	}
}

func addContainerDisk(spec *v1.VirtualMachineInstanceSpec, image string) *v1.VirtualMachineInstanceSpec {
	disk := &v1.Disk{
		Name: "containerdisk",
		DiskDevice: v1.DiskDevice{
			Disk: &v1.DiskTarget{
				Bus: virtIO,
			},
		},
	}
	spec.Domain.Devices.Disks = append(spec.Domain.Devices.Disks, *disk)
	volume := &v1.Volume{
		Name: "containerdisk",
		VolumeSource: v1.VolumeSource{
			ContainerDisk: &v1.ContainerDiskSource{
				Image: image,
			},
		},
	}
	spec.Volumes = append(spec.Volumes, *volume)
	return spec
}

func getBaseVMI(name string) *v1.VirtualMachineInstance {
	baseVMISpec := getBaseVMISpec()

	return &v1.VirtualMachineInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.GroupVersion.String(),
			Kind:       "VirtualMachineInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: *baseVMISpec,
	}
}

func addNoCloudDiskWitUserData(spec *v1.VirtualMachineInstanceSpec, data string) *v1.VirtualMachineInstanceSpec {
	spec.Domain.Devices.Disks = append(spec.Domain.Devices.Disks, v1.Disk{
		Name: "cloudinitdisk",
		DiskDevice: v1.DiskDevice{
			Disk: &v1.DiskTarget{
				Bus: virtIO,
			},
		},
	})

	spec.Volumes = append(spec.Volumes, v1.Volume{
		Name: "cloudinitdisk",
		VolumeSource: v1.VolumeSource{
			CloudInitNoCloud: &v1.CloudInitNoCloudSource{
				UserData: data,
			},
		},
	})
	return spec
}

func CreateVmiObject(name string, namespace string, interfaces []v1.Interface, networks []v1.Network) *v1.VirtualMachineInstance {
	vmi := getBaseVMI(randName(name))
	vmi.Namespace = namespace
	addContainerDisk(&vmi.Spec, "quay.io/kubevirt/alpine-container-disk-demo:v1.4.0")
	addNoCloudDiskWitUserData(&vmi.Spec, "#!/bin/sh\n\necho 'printed from cloud-init userdata'\n")
	vmi.Spec.Domain.Devices.Interfaces = interfaces
	vmi.Spec.Networks = networks
	return vmi
}
