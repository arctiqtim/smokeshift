package kuberang

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/apprenda/kuberang/pkg/config"
	"github.com/apprenda/kuberang/pkg/util"
)

const RunPrefix = "kuberang-"
const BBDeploymentName = RunPrefix + "busybox"
const NGDeploymentName = RunPrefix + "nginx"
const Timeout = 300 //seconds
const HTTP_Timeout = 1000 * time.Millisecond

func CheckKubernetes() error {
	ngServiceName := nginxServiceName()
	if !PrecheckKubectl() ||
		!PrecheckNamespace() ||
		!PrecheckServices(ngServiceName) ||
		!PrecheckDeployments() {
		PowerDown(ngServiceName)
		return errors.New("Pre-conditions failed; must clean up before we can smoke test")
	}

	success := true

	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunKubectl("run", BBDeploymentName, "--image=busybox", "--image-pull-policy=IfNotPresent", "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued BusyBox start request")
		success = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued BusyBox start request")
	}
	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(NGDeploymentName, "nginx", nginxCount); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued Nginx start request")
		success = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued Nginx start request")
	}

	// Check for both
	if !WaitForDeployments(busyboxCount, nginxCount) {
		return nil
	}

	// Add service
	if ko := RunKubectl("expose", "deployment", NGDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued expose Nginx service request")
		success = false
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
		success = false
	}

	if ko := RunGetService(ngServiceName); ko.Success {
		serviceIP = ko.ServiceCluserIP()
		util.PrettyPrintOk(os.Stdout, "Grab nginx service ip address")
	} else {
		serviceIP = ""
		util.PrettyPrintErr(os.Stdout, "Grab nginx service ip address")
		success = false
	}

	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-busybox", "-o", "json"); ko.Success {
		busyboxPodName = ko.FirstPodName()
		util.PrettyPrintOk(os.Stdout, "Grab BusyBox pod name")
	} else {
		busyboxPodName = ""
		util.PrettyPrintErr(os.Stdout, "Grab BusyBox pod name")
		success = false
	}

	// Check connectivity between pods (using busybox)
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service at "+serviceIP+" from BusyBox")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Nginx service at "+serviceIP+" from BusyBox")
		success = false
	}
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", ngServiceName); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
	} else {
		util.PrettyPrintErr(os.Stdout, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
		success = false
	}

	for _, podIP := range podIPs {
		if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", podIP); busyboxPodName == "" || ko.Success {
			util.PrettyPrintOk(os.Stdout, "Accessed Nginx pod at "+podIP+" from BusyBox")
		} else {
			util.PrettyPrintErr(os.Stdout, "Accessed Nginx pod at "+podIP+" from BusyBox")
			success = false
		}
	}

	// Check connectivity with internet
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(os.Stdout, "Accessed Google.com from BusyBox")
	} else {
		util.PrettyPrintErrorIgnored(os.Stdout, "Accessed Google.com from BusyBox")
	}

	// Check connectivity from current machine (using curl or wget)
	// Set Timeout or it could wait forever
	client := http.Client{
		Timeout: HTTP_Timeout,
	}
	if _, err := client.Get("http://" + ngServiceName); err == nil {
		util.PrettyPrintOk(os.Stdout, "Accessed Nginx service via DNS "+ngServiceName+" from this node")
	} else {
		util.PrettyPrintErrorIgnored(os.Stdout, "Accessed Nginx service via DNS "+ngServiceName+" from this node")
	}
	for _, podIP := range podIPs {
		if _, err := client.Get("http://" + podIP); err == nil {
			util.PrettyPrintOk(os.Stdout, "Accessed Nginx pod at "+podIP+" from this node")
		} else {
			util.PrettyPrintErrorIgnored(os.Stdout, "Accessed Nginx pod at "+podIP+" from this node")
		}
	}
	if _, err := client.Get("http://google.com/"); err == nil {
		util.PrettyPrintOk(os.Stdout, "Accessed Google.com from this node")
	} else {
		util.PrettyPrintErrorIgnored(os.Stdout, "Accessed Google.com from this node")
	}

	PowerDown(ngServiceName)

	if success {
		return nil
	}
	return errors.New("One or more required steps failed")
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

func PrecheckServices(nginxServiceName string) bool {
	ret := true
	if ko := RunGetService(nginxServiceName); ko.Success {
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

func PrecheckNamespace() bool {
	ret := true
	if config.Namespace != "" {
		if ko := RunGetNamespace(config.Namespace); !ko.Success || ko.NamespaceStatus() != "Active" {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `" + config.Namespace + "` exists")
			ret = false
		} else {
			util.PrettyPrintOk(os.Stdout, "Configured kubernetes namespace `" + config.Namespace + "` exists")
		}
	}
	return ret
}

func CheckDeployments(busbyboxCount, nginxCount int64) bool {
	ret := true
	if ko := RunGetDeployment(BBDeploymentName); !ko.Success || ko.ObservedReplicaCount() != busbyboxCount {
		ret = false
	}
	if ko := RunGetDeployment(NGDeploymentName); !ko.Success || ko.ObservedReplicaCount() != nginxCount {
		ret = false
	}
	return ret
}

func WaitForDeployments(busbyboxCount, nginxCount int64) bool {
	for i := 0; i < Timeout; i++ {
		if CheckDeployments(busbyboxCount, nginxCount) {
			util.PrettyPrintOk(os.Stdout, "Both deployments completed successfully within timeout")
			return true
		}
		time.Sleep(1 * time.Second)
	}
	util.PrettyPrintErr(os.Stdout, "Both deployments completed successfully within timeout")
	return false
}

func PowerDown(nginxServiceName string) {
	// Power down service
	if ko := RunKubectl("delete", "service", nginxServiceName); ko.Success {
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

func nginxServiceName() string {
	return fmt.Sprintf("%s-%d", RunPrefix+"nginx", time.Now().UnixNano())
}
