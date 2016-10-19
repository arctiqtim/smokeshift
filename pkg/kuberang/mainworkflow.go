package kuberang

import (
	"net/http"
	"os"
	"time"

	"github.com/apprenda/kuberang/pkg/util"
)

const RunPrefix = "kuberang-"
const NGServiceName = RunPrefix + "nginx"
const BBDeploymentName = RunPrefix + "busybox"
const NGDeploymentName = RunPrefix + "nginx"
const Timeout = 300 //seconds
const HTTP_Timeout = 1000 * time.Millisecond

func CheckKubernetes() error {
	if !PrecheckKubectl() ||
		!PrecheckServices() ||
		!PrecheckDeployments() {
		PowerDown()
		return nil
	}

	// Scale out busybox

	if ko := RunKubectl("run", BBDeploymentName, "--image=busybox", "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued BusyBox start request")
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued BusyBox start request")
	}
	// Scale out nginx
	if ko := RunPod(NGDeploymentName, "nginx", 2); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued Nginx start request")
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued Nginx start request")
	}

	// Check for both
	if !WaitForDeployments() {
		return nil
	}

	// Add service
	if ko := RunKubectl("expose", "deployment", NGDeploymentName, "--name="+NGServiceName, "--port=80"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued expose Nginx service request")
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued expose Nginx service request")
	}

	// Get pod & service IPs
	var podIPs []string
	var serviceIP string
	var busyboxPodName string
	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-nginx", "-o", "json"); ko.Success {
		podIPs = ko.PodIPs()
		util.PrettyPrintOk(os.Stdout, "Grab nginx pod ip addresses")
	} else {
		podIPs = make([]string, 0)
		util.PrettyPrintErr(os.Stdout, "Grab nginx pod ip addresses")
	}

	if ko := RunGetService(NGServiceName); ko.Success {
		serviceIP = ko.ServiceCluserIP()
		util.PrettyPrintOk(os.Stdout, "Grab nginx service ip address")
	} else {
		serviceIP = ""
		util.PrettyPrintErr(os.Stdout, "Grab nginx service ip address")
	}

	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-busybox", "-o", "json"); ko.Success {
		busyboxPodName = ko.FirstPodName()
		util.PrettyPrintOk(os.Stdout, "Grab BusyBox pod name")
	} else {
		busyboxPodName = ""
		util.PrettyPrintErr(os.Stdout, "Grab BusyBox pod name")
	}

	// Check connectivity between pods (using busybox)
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service at "+serviceIP+" from BusyBox")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Nginx service at "+serviceIP+" from BusyBox", ko.CombinedOut, busyboxPodName)
	}
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", NGServiceName); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service via DNS "+NGServiceName+" from BusyBox")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Nginx service via DNS "+NGServiceName+" from BusyBox")
	}
	for _, podIP := range podIPs {
		if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", podIP); busyboxPodName == "" || ko.Success {
			util.PrettyPrintOk(os.Stdout, "Accessed Nginx pod at "+podIP+" from BusyBox")
		} else {
			util.PrettyPrintErr(os.Stdout, "Accessed Nginx pod at "+podIP+" from BusyBox")
		}
	}

	// Check connectivity with internet
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Google.com from BusyBox")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Google.com from BusyBox")
	}

	// Check connectivity from current machine (using curl or wget)
	// Set Timeout or it could wait forever
	client := http.Client{
		Timeout: HTTP_Timeout,
	}
	if _, err := client.Get(NGServiceName); err == nil {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service via DNS "+NGServiceName+" from this node")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Nginx service via DNS "+NGServiceName+" from this node")
	}
	for _, podIP := range podIPs {
		if _, err := client.Get(podIP); err == nil {
			util.PrettyPrintOk(os.Stdout, "Accessed Nginx pod at "+podIP+" from this node")
		} else {
			util.PrettyPrintErr(os.Stdout, "Accessed Nginx pod at "+podIP+" from this node")
		}
	}
	if _, err := client.Get("http://google.com/"); err == nil {
		util.PrettyPrintOk(os.Stdout, "Accessed Google.com from this node")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Google.com from this node")
	}

	PowerDown()

	return nil
}

func PrecheckKubectl() bool {
	ret := true
	if ko := RunKubectl("version"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Configured kubectl exists")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Configured kubectl exists")
	}
	return ret
}

func PrecheckServices() bool {
	ret := true
	if ko := RunGetService(NGServiceName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func PrecheckDeployments() bool {
	ret := true
	if ko := RunGetDeployment(BBDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "BusyBox service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "BusyBox service does not already exist")
	}
	if ko := RunGetDeployment(NGDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func CheckDeployments() bool {
	ret := true
	if ko := RunGetDeployment(BBDeploymentName); !ko.Success || ko.ObservedReplicaCount() != 1 {
		ret = false
	}
	if ko := RunGetDeployment(NGDeploymentName); !ko.Success || ko.ObservedReplicaCount() != 2 {
		ret = false
	}
	return ret
}

func WaitForDeployments() bool {
	for i := 0; i < Timeout; i++ {
		if CheckDeployments() {
			util.PrettyPrintOk(os.Stdout, "Both deployments completed successfully within timeout")
			return true
		}
		time.Sleep(1 * time.Second)
	}
	util.PrettyPrintErr(os.Stdout, "Both deployments completed successfully within timeout")
	return false
}

func PowerDown() {
	// Power down service
	if ko := RunKubectl("delete", "service", NGServiceName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx service")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx service")
	}
	// Power down bb
	if ko := RunKubectl("delete", "deployments", BBDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Busybox deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Busybox deployment")
	}
	// Power down nginx
	if ko := RunKubectl("delete", "deployments", NGDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx deployment")
	}
}
