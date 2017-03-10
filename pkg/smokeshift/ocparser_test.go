package smokeshift

import "testing"

func TestNodeCount(t *testing.T) {

	ko := OCOutput{
		Success:     true,
		CombinedOut: SampleNodeRespones,
		RawOut:      []byte(SampleNodeRespones),
	}
	if nodesCount := ko.NodeCount(); nodesCount != 3 {
		t.Errorf("Wrong number of nodes, expeted 3, got %d", nodesCount)
	}
}

const SampleNodeRespones = `
{
    "kind": "List",
    "apiVersion": "v1",
    "metadata": {},
    "items": [
        {
            "kind": "Node",
            "apiVersion": "v1",
            "metadata": {
                "name": "node1",
                "selfLink": "/api/v1/nodes/node1",
                "uid": "12c68944-964d-11e6-82b1-525400c3c0db",
                "resourceVersion": "161",
                "creationTimestamp": "2016-10-19T22:40:45Z",
                "labels": {
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/os": "linux",
                    "kubernetes.io/hostname": "node1"
                },
                "annotations": {
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                }
            },
            "spec": {
                "podCIDR": "172.16.0.0/24",
                "externalID": "node1",
                "unschedulable": true
            },
            "status": {
                "capacity": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "allocatable": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "type": "OutOfDisk",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:17Z",
                        "lastTransitionTime": "2016-10-19T22:40:45Z",
                        "reason": "KubeletHasSufficientDisk",
                        "message": "kubelet has sufficient disk space available"
                    },
                    {
                        "type": "MemoryPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:17Z",
                        "lastTransitionTime": "2016-10-19T22:40:45Z",
                        "reason": "KubeletHasSufficientMemory",
                        "message": "kubelet has sufficient memory available"
                    },
                    {
                        "type": "DiskPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:17Z",
                        "lastTransitionTime": "2016-10-19T22:40:45Z",
                        "reason": "KubeletHasNoDiskPressure",
                        "message": "kubelet has no disk pressure"
                    },
                    {
                        "type": "Ready",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:17Z",
                        "lastTransitionTime": "2016-10-19T22:40:45Z",
                        "reason": "KubeletNotReady",
                        "message": "ConfigureCBR0 requested, but PodCIDR not set. Will not configure CBR0 right now"
                    }
                ],
                "addresses": [
                    {
                        "type": "LegacyHostIP",
                        "address": "192.168.205.11"
                    },
                    {
                        "type": "InternalIP",
                        "address": "192.168.205.11"
                    }
                ],
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10250
                    }
                },
                "nodeInfo": {
                    "machineID": "d8468efd20354b6b9633a4c0dd4b5c06",
                    "systemUUID": "D8468EFD-2035-4B6B-9633-A4C0DD4B5C06",
                    "bootID": "b6170f6b-9675-4ddc-9e2c-d00f888cfdf2",
                    "kernelVersion": "3.10.0-327.22.2.el7.x86_64",
                    "osImage": "CentOS Linux 7 (Core)",
                    "containerRuntimeVersion": "docker://1.11.2",
                    "kubeletVersion": "v1.4.3",
                    "kubeProxyVersion": "v1.4.3",
                    "operatingSystem": "linux",
                    "architecture": "amd64"
                },
                "images": [
                    {
                        "names": [
                            "calico/node:v0.22.0"
                        ],
                        "sizeBytes": 91482858
                    }
                ]
            }
        },
        {
            "kind": "Node",
            "apiVersion": "v1",
            "metadata": {
                "name": "node2",
                "selfLink": "/api/v1/nodes/node2",
                "uid": "151e0cca-964d-11e6-82b1-525400c3c0db",
                "resourceVersion": "142",
                "creationTimestamp": "2016-10-19T22:40:49Z",
                "labels": {
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/os": "linux",
                    "kubernetes.io/hostname": "node2"
                },
                "annotations": {
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                }
            },
            "spec": {
                "podCIDR": "172.16.3.0/24",
                "externalID": "node2"
            },
            "status": {
                "capacity": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "allocatable": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "type": "OutOfDisk",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:49Z",
                        "reason": "KubeletHasSufficientDisk",
                        "message": "kubelet has sufficient disk space available"
                    },
                    {
                        "type": "MemoryPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:49Z",
                        "reason": "KubeletHasSufficientMemory",
                        "message": "kubelet has sufficient memory available"
                    },
                    {
                        "type": "DiskPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:49Z",
                        "reason": "KubeletHasNoDiskPressure",
                        "message": "kubelet has no disk pressure"
                    },
                    {
                        "type": "Ready",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:49Z",
                        "reason": "KubeletNotReady",
                        "message": "ConfigureCBR0 requested, but PodCIDR not set. Will not configure CBR0 right now"
                    }
                ],
                "addresses": [
                    {
                        "type": "LegacyHostIP",
                        "address": "192.168.205.12"
                    },
                    {
                        "type": "InternalIP",
                        "address": "192.168.205.12"
                    }
                ],
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10250
                    }
                },
                "nodeInfo": {
                    "machineID": "aec706480f614ded888bcdd3dadb3818",
                    "systemUUID": "AEC70648-0F61-4DED-888B-CDD3DADB3818",
                    "bootID": "a3ae9d83-6b12-48ed-972b-3bbf3a96c1c0",
                    "kernelVersion": "3.10.0-327.22.2.el7.x86_64",
                    "osImage": "CentOS Linux 7 (Core)",
                    "containerRuntimeVersion": "docker://1.11.2",
                    "kubeletVersion": "v1.4.3",
                    "kubeProxyVersion": "v1.4.3",
                    "operatingSystem": "linux",
                    "architecture": "amd64"
                },
                "images": [
                    {
                        "names": [
                            "calico/node:v0.22.0"
                        ],
                        "sizeBytes": 91482858
                    }
                ]
            }
        },
        {
            "kind": "Node",
            "apiVersion": "v1",
            "metadata": {
                "name": "node3",
                "selfLink": "/api/v1/nodes/node3",
                "uid": "14e27293-964d-11e6-82b1-525400c3c0db",
                "resourceVersion": "140",
                "creationTimestamp": "2016-10-19T22:40:48Z",
                "labels": {
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/os": "linux",
                    "kubernetes.io/hostname": "node3"
                },
                "annotations": {
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                }
            },
            "spec": {
                "podCIDR": "172.16.1.0/24",
                "externalID": "node3"
            },
            "status": {
                "capacity": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "allocatable": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "type": "OutOfDisk",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasSufficientDisk",
                        "message": "kubelet has sufficient disk space available"
                    },
                    {
                        "type": "MemoryPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasSufficientMemory",
                        "message": "kubelet has sufficient memory available"
                    },
                    {
                        "type": "DiskPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasNoDiskPressure",
                        "message": "kubelet has no disk pressure"
                    },
                    {
                        "type": "Ready",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletNotReady",
                        "message": "ConfigureCBR0 requested, but PodCIDR not set. Will not configure CBR0 right now"
                    }
                ],
                "addresses": [
                    {
                        "type": "LegacyHostIP",
                        "address": "192.168.205.13"
                    },
                    {
                        "type": "InternalIP",
                        "address": "192.168.205.13"
                    }
                ],
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10250
                    }
                },
                "nodeInfo": {
                    "machineID": "0e6d0d4831514913ae97550d71753057",
                    "systemUUID": "0E6D0D48-3151-4913-AE97-550D71753057",
                    "bootID": "b9e07e2f-4b19-47d8-b370-8c51a09df1c9",
                    "kernelVersion": "3.10.0-327.22.2.el7.x86_64",
                    "osImage": "CentOS Linux 7 (Core)",
                    "containerRuntimeVersion": "docker://1.11.2",
                    "kubeletVersion": "v1.4.3",
                    "kubeProxyVersion": "v1.4.3",
                    "operatingSystem": "linux",
                    "architecture": "amd64"
                },
                "images": [
                    {
                        "names": [
                            "calico/node:v0.22.0"
                        ],
                        "sizeBytes": 91482858
                    }
                ]
            }
        },
        {
            "kind": "Node",
            "apiVersion": "v1",
            "metadata": {
                "name": "node4",
                "selfLink": "/api/v1/nodes/node4",
                "uid": "14ed6fea-964d-11e6-82b1-525400c3c0db",
                "resourceVersion": "141",
                "creationTimestamp": "2016-10-19T22:40:48Z",
                "labels": {
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/os": "linux",
                    "kubernetes.io/hostname": "node4"
                },
                "annotations": {
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                }
            },
            "spec": {
                "podCIDR": "172.16.2.0/24",
                "externalID": "node4"
            },
            "status": {
                "capacity": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "allocatable": {
                    "alpha.kubernetes.io/nvidia-gpu": "0",
                    "cpu": "1",
                    "memory": "1016860Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "type": "OutOfDisk",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasSufficientDisk",
                        "message": "kubelet has sufficient disk space available"
                    },
                    {
                        "type": "MemoryPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasSufficientMemory",
                        "message": "kubelet has sufficient memory available"
                    },
                    {
                        "type": "DiskPressure",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletHasNoDiskPressure",
                        "message": "kubelet has no disk pressure"
                    },
                    {
                        "type": "Ready",
                        "status": "False",
                        "lastHeartbeatTime": "2016-10-19T22:41:09Z",
                        "lastTransitionTime": "2016-10-19T22:40:48Z",
                        "reason": "KubeletNotReady",
                        "message": "ConfigureCBR0 requested, but PodCIDR not set. Will not configure CBR0 right now"
                    }
                ],
                "addresses": [
                    {
                        "type": "LegacyHostIP",
                        "address": "192.168.205.14"
                    },
                    {
                        "type": "InternalIP",
                        "address": "192.168.205.14"
                    }
                ],
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10250
                    }
                },
                "nodeInfo": {
                    "machineID": "ae25864e16674b9a95afc6a5cf054621",
                    "systemUUID": "AE25864E-1667-4B9A-95AF-C6A5CF054621",
                    "bootID": "102e9302-1516-44d2-8f99-fd36cb50310d",
                    "kernelVersion": "3.10.0-327.22.2.el7.x86_64",
                    "osImage": "CentOS Linux 7 (Core)",
                    "containerRuntimeVersion": "docker://1.11.2",
                    "kubeletVersion": "v1.4.3",
                    "kubeProxyVersion": "v1.4.3",
                    "operatingSystem": "linux",
                    "architecture": "amd64"
                },
                "images": [
                    {
                        "names": [
                            "calico/node:v0.22.0"
                        ],
                        "sizeBytes": 91482858
                    }
                ]
            }
        }
    ]
}
`

func TestNamespaceStatus(t *testing.T) {
    ko := OCOutput{
        Success:     true,
        CombinedOut: SampleNamespaceResponse,
        RawOut:      []byte(SampleNamespaceResponse),
    }

    if namespaceStatus := ko.NamespaceStatus(); namespaceStatus != "Active" {
        t.Errorf("Wrong namespace status, expeted `Active`, got %d", namespaceStatus)
    }
}

const SampleNamespaceResponse = `
{
    "kind": "Namespace",
    "apiVersion": "v1",
    "metadata": {
        "name": "some-namespace",
        "selfLink": "/api/v1/namespaces/some-namespace",
        "uid": "6c0cde5e-ac03-11e6-8d03-e6eb6a2840bc",
        "resourceVersion": "79200",
        "creationTimestamp": "2016-11-16T13:48:57Z"
    },
    "spec": {
        "finalizers": [
            "kubernetes"
        ]
    },
    "status": {
        "phase": "Active"
    }
}
`
