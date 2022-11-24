# KubeSecondaryDNS
DNS for KubeVirt VirtualMachines secondary interfaces

### Mapping a container image to the code
In order to know which git commit hash was used to build a certain container image perform the following steps:
```bash
podman inspect --format '{{ index .Config.Labels "multi.GIT_SHA"}}' <IMAGE>
```
Note that it can also be inspected by skopeo without downloading the image.
