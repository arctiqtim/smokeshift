package kuberang

import (
	"fmt"
	"io"
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
	deploymentTimeoutSeconds = 300 * time.Second
	httpTimeout              = 1000 * time.Millisecond
)

func CheckKubernetes() error {
	out := os.Stdout
	ngServiceName := nginxServiceName()
	if !precheckKubectl() ||
		!precheckNamespace() ||
		!precheckServices(ngServiceName) ||
		!precheckDeployments() {
		powerDown(ngServiceName)
		return errors.New("Pre-conditions failed; must clean up before we can smoke test")
	}

	success := true
	registryURL := ""
	if config.RegistryURL != "" {
		registryURL = config.RegistryURL + "/"
	}

	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunKubectl("run", bbDeploymentName, fmt.Sprintf("--image=%sbusybox:latest", registryURL), "--image-pull-policy=IfNotPresent", "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(out, "Issued BusyBox start request")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	} else {
		util.PrettyPrintOk(out, "Issued BusyBox start request")
	}
	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(ngDeploymentName, fmt.Sprintf("%snginx:stable-alpine", registryURL), nginxCount); !ko.Success {
		util.PrettyPrintErr(out, "Issued Nginx start request")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	} else {
		util.PrettyPrintOk(out, "Issued Nginx start request")
	}

	// Check for both
	if !waitForDeployments(busyboxCount, nginxCount) {
		return nil
	}

	// Add service
	if ko := RunKubectl("expose", "deployment", ngDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
		util.PrettyPrintErr(out, "Issued expose Nginx service request")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	} else {
		util.PrettyPrintOk(out, "Issued expose Nginx service request")
	}

	// Get pod & service IPs
	var podIPs []string
	var serviceIP string
	var busyboxPodName string
	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-nginx", "-o", "json"); ko.Success {
		podIPs = ko.PodIPs()
		util.PrettyPrintOk(out, "Grab nginx pod ip addresses")
	} else {
		podIPs = make([]string, 0)
		util.PrettyPrintErr(out, "Grab nginx pod ip addresses")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	if ko := RunGetService(ngServiceName); ko.Success {
		serviceIP = ko.ServiceCluserIP()
		util.PrettyPrintOk(out, "Grab nginx service ip address")
	} else {
		serviceIP = ""
		util.PrettyPrintErr(out, "Grab nginx service ip address")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-busybox", "-o", "json"); ko.Success {
		busyboxPodName = ko.FirstPodName()
		util.PrettyPrintOk(out, "Grab BusyBox pod name")
	} else {
		busyboxPodName = ""
		util.PrettyPrintErr(out, "Grab BusyBox pod name")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Check connectivity between pods (using busybox)
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
	} else {
		util.PrettyPrintErr(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", ngServiceName); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(out, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
	} else {
		util.PrettyPrintErr(out, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	for _, podIP := range podIPs {
		if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", podIP); busyboxPodName == "" || ko.Success {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
		} else {
			util.PrettyPrintErr(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
			printFailureDetail(out, ko.CombinedOut)
			success = false
		}
	}

	// Check connectivity with internet
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(out, "Accessed Google.com from BusyBox")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from BusyBox")
	}

	// Check connectivity from current machine (using curl or wget)
	// Set Timeout or it could wait forever
	client := http.Client{
		Timeout: httpTimeout,
	}
	if _, err := client.Get("http://" + ngServiceName); err == nil {
		util.PrettyPrintOk(out, "Accessed Nginx service via DNS "+ngServiceName+" from this node")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Nginx service via DNS "+ngServiceName+" from this node")
	}
	for _, podIP := range podIPs {
		if _, err := client.Get("http://" + podIP); err == nil {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from this node")
		} else {
			util.PrettyPrintErrorIgnored(out, "Accessed Nginx pod at "+podIP+" from this node")
		}
	}
	if _, err := client.Get("http://google.com/"); err == nil {
		util.PrettyPrintOk(out, "Accessed Google.com from this node")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from this node")
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
		printFailureDetail(os.Stdout, ko.CombinedOut)
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
		printFailureDetail(os.Stdout, ko.CombinedOut)
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
		printFailureDetail(os.Stdout, ko.CombinedOut)
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "BusyBox service does not already exist")
	}
	if ko := RunGetDeployment(ngDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func precheckNamespace() bool {
	ret := true
	if config.Namespace != "" {
		ko := RunGetNamespace(config.Namespace)
		if !ko.Success {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
			printFailureDetail(os.Stdout, ko.CombinedOut)
			ret = false
		} else if ko.NamespaceStatus() != "Active" {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
			ret = false
		} else {
			util.PrettyPrintOk(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
		}
	}
	return ret
}

func checkDeployments(busyboxCount, nginxCount int64) bool {
	ret := true
	ko := RunGetDeployment(bbDeploymentName)
	if !ko.Success {
		ret = false
	} else if ko.ObservedReplicaCount() != busyboxCount {
		ret = false
	}
	ko = RunGetDeployment(ngDeploymentName)
	if !ko.Success {
		ret = false
	} else if ko.ObservedReplicaCount() != nginxCount {
		ret = false
	}
	return ret
}

func waitForDeployments(busyboxCount, nginxCount int64) bool {
	start := time.Now()
	for time.Since(start) < deploymentTimeoutSeconds {
		if checkDeployments(busyboxCount, nginxCount) {
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
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
	// Power down bb
	if ko := RunKubectl("delete", "deployments", bbDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Busybox deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Busybox deployment")
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
	// Power down nginx
	if ko := RunKubectl("delete", "deployments", ngDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx deployment")
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
}

func nginxServiceName() string {
	return fmt.Sprintf("%s-%d", runPrefix+"nginx", time.Now().UnixNano())
}

func printFailureDetail(out io.Writer, detail string) {
	fmt.Fprintln(out, "-------- OUTPUT --------")
	fmt.Fprintf(out, detail)
	fmt.Fprintln(out, "------------------------")
	fmt.Fprintln(out)
}
