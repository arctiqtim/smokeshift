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

const (
	runPrefix                = "kuberang-"
	bbDeploymentName         = runPrefix + "busybox"
	ngDeploymentName         = runPrefix + "nginx"
	deploymentTimeoutSeconds = 300 //seconds
	httpTimeoutMillis        = 1000 * time.Millisecond
)

func CheckKubernetes() error {
	ngServiceName := nginxServiceName()
	if !precheckKubectl() ||
		!precheckNamespace() ||
		!precheckServices(ngServiceName) ||
		!precheckDeployments() {
		powerDown(ngServiceName)
		return errors.New("Pre-conditions failed; must clean up before we can smoke test")
	}

	success := true

	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunKubectl("run", bbDeploymentName, "--image=busybox", "--image-pull-policy=IfNotPresent", "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued BusyBox start request")
		fmt.Fprintf(os.Stdout, "- error: %v", ko.CombinedOut)
		success = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued BusyBox start request")
	}
	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(ngDeploymentName, "nginx", nginxCount); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Issued Nginx start request")
		success = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Issued Nginx start request")
	}

	// Check for both
	if !waitForDeployments(busyboxCount, nginxCount) {
		return nil
	}

	// Add service
	if ko := RunKubectl("expose", "deployment", ngDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
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
		Timeout: httpTimeoutMillis,
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

	powerDown(ngServiceName)

	if success {
		return nil
	}
	return errors.New("One or more required steps failed")
}

func precheckKubectl() bool {
	ret := true
	if ko := RunKubectl("version"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Configured kubectl exists")
		fmt.Fprintf(os.Stdout, "---\n%v\n---\n", ko.CombinedOut)
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Configured kubectl exists")
	}
	return ret
}

func precheckServices(nginxServiceName string) bool {
	ret := true
	if ko := RunGetService(nginxServiceName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func precheckDeployments() bool {
	ret := true
	if ko := RunGetDeployment(bbDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "BusyBox service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "BusyBox service does not already exist")
	}
	if ko := RunGetDeployment(ngDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func precheckNamespace() bool {
	ret := true
	if config.Namespace != "" {
		if ko := RunGetNamespace(config.Namespace); !ko.Success || ko.NamespaceStatus() != "Active" {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
			ret = false
		} else {
			util.PrettyPrintOk(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
		}
	}
	return ret
}

func checkDeployments(busbyboxCount, nginxCount int64) bool {
	ret := true
	if ko := RunGetDeployment(bbDeploymentName); !ko.Success || ko.ObservedReplicaCount() != busbyboxCount {
		ret = false
	}
	if ko := RunGetDeployment(ngDeploymentName); !ko.Success || ko.ObservedReplicaCount() != nginxCount {
		ret = false
	}
	return ret
}

func waitForDeployments(busbyboxCount, nginxCount int64) bool {
	for i := 0; i < deploymentTimeoutSeconds; i++ {
		if checkDeployments(busbyboxCount, nginxCount) {
			util.PrettyPrintOk(os.Stdout, "Both deployments completed successfully within timeout")
			return true
		}
		time.Sleep(1 * time.Second)
	}
	util.PrettyPrintErr(os.Stdout, "Both deployments completed successfully within timeout")
	return false
}

func powerDown(nginxServiceName string) {
	// Power down service
	if ko := RunKubectl("delete", "service", nginxServiceName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx service")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx service")
	}
	// Power down bb
	if ko := RunKubectl("delete", "deployments", bbDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Busybox deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Busybox deployment")
	}
	// Power down nginx
	if ko := RunKubectl("delete", "deployments", ngDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx deployment")
		fmt.Fprintf(os.Stdout, "---\n%s---\n", ko.CombinedOut)
	}
}

func nginxServiceName() string {
	return fmt.Sprintf("%s-%d", runPrefix+"nginx", time.Now().UnixNano())
}
