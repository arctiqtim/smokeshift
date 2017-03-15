package smokeshift

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/opencredo/smokeshift/pkg/config"
	"github.com/opencredo/smokeshift/pkg/util"
	"strings"
)

const (
	runPrefix         = "smokeshift-"
	bbDeploymentName  = runPrefix + "busybox"
	ngDeploymentName  = runPrefix + "nginx"
	deploymentTimeout = 300 * time.Second
	httpTimeout       = 1000 * time.Millisecond
)

// CheckOpenshift runs checks against a cluster. It expects to find
// a configured `oc` binary in the path.
func CheckOpenshift(skipCleanup bool) error {
	out := os.Stdout
	ngServiceName := nginxServiceName()
	success := true
	registryURL := ""
	if config.RegistryURL != "" {
		registryURL = config.RegistryURL + "/"
	}

	// Make sure we have all we need
	if !checkPreconditions(out) {
		return errors.New("Pre-conditions failed")
	}

	if !skipCleanup {
		defer powerDown(ngServiceName)
	}

	printUserDetail(out)

	//Create a project in which to deploy the workloads for running the checks
	if !initProject(out) {
		return errors.New("Failed to create Project: "+config.Namespace)
	}

	// Deploy the workloads required for running checks
	if !deployTestWorkloads(registryURL, out, ngServiceName) {
		return errors.New("Failed to deploy test workloads")
	}

	// Get IPs of all nginx pods
	podIPs := []string{}
	if ko := RunOCinNamespace("get", "pods", "-l", "run=smokeshift-nginx", "-o", "json"); ko.Success {
		podIPs = ko.PodIPs()
		util.PrettyPrintOk(out, "Grab nginx pod ip addresses")
	} else {
		util.PrettyPrintErr(out, "Grab nginx pod ip addresses")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Get the service IP of the nginx service
	var serviceIP string
	if ko := RunGetService(ngServiceName); ko.Success {
		serviceIP = ko.ServiceCluserIP()
		util.PrettyPrintOk(out, "Grab nginx service ip address")
	} else {
		util.PrettyPrintErr(out, "Grab nginx service ip address")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Get the name of the busybox pod
	var busyboxPodName string
	if ko := RunOCinNamespace("get", "pods", "-l", "run=smokeshift-busybox", "-o", "json"); ko.Success {
		busyboxPodName = ko.FirstPodName()
		util.PrettyPrintOk(out, "Grab BusyBox pod name")
	} else {
		util.PrettyPrintErr(out, "Grab BusyBox pod name")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Gate on successful acquisition of all the required names / IPs
	if !success {
		return errors.New("Failed to get required information from cluster")
	}

	// The following checks verify the pod network and the ability for
	// pods to talk to each other.
	// 1. Access nginx service via service IP from another pod
	var kubeOut OCOutput
	util.PrettyPrintInfo(out,"Trying to access Nginx service at "+serviceIP+" from BusyBox")
	ok := retry(3, func() bool {
		kubeOut = RunOCinNamespace("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP)
		return kubeOut.Success
	})
	if ok {
		util.PrettyPrintOk(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
	} else {
		printFailureDetail(out, kubeOut.CombinedOut)
		util.PrettyPrintErr(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
		success = false
	}

	// 2. Access nginx service via service name (DNS) from another pod

	nginxSvc := ngServiceName+"."+config.Namespace
	util.PrettyPrintInfo(out, "Trying to access Nginx service via DNS "+nginxSvc+" from BusyBox")
	ok = retry(3, func() bool {
		kubeOut = RunOCinNamespace("exec", busyboxPodName, "--", "wget", "-qO-", nginxSvc)
		return kubeOut.Success
	})
	if ok {
		util.PrettyPrintOk(out, "Accessed Nginx service via DNS "+nginxSvc+" from BusyBox")
	} else {
		util.PrettyPrintErr(out, "Accessed Nginx service via DNS "+nginxSvc+" from BusyBox")
		printFailureDetail(out, kubeOut.CombinedOut)
		success = false
	}

	// 3. Access all nginx pods by IP
	util.PrettyPrintInfo(out, "Trying to access all nginx pods by IP")
	for _, podIP := range podIPs {
		ok = retry(3, func() bool {
			kubeOut = RunOCinNamespace("exec", busyboxPodName, "--", "wget", "-qO-", podIP)
			return kubeOut.Success
		})
		if ok {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
		} else {
			util.PrettyPrintErr(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
			printFailureDetail(out, kubeOut.CombinedOut)
			success = false
		}
	}

	// 4. Check internet connectivity from pod
	if ko := RunOCinNamespace("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(out, "Accessed Google.com from BusyBox")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from BusyBox")
	}

	client := http.Client{
		Timeout: httpTimeout,
	}
	// 5. Check connectivity from current machine to all nginx pods
	for _, podIP := range podIPs {
		if _, err := client.Get("http://" + podIP); err == nil {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from this node")
		} else {
			util.PrettyPrintErrorIgnored(out, "Accessed Nginx pod at "+podIP+" from this node")
		}
	}

	// 6. Check internet connectivity from current machine
	if _, err := client.Get("http://google.com/"); err == nil {
		util.PrettyPrintOk(out, "Accessed Google.com from this node")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from this node")
	}

	if !success {
		return errors.New("One or more required steps failed")
	}
	return nil
}

func deployTestWorkloads(registryURL string, out io.Writer, ngServiceName string) bool {
	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunOCinNamespace("run", bbDeploymentName, fmt.Sprintf("--image=%sbusybox:1", registryURL), "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(out, "Issued BusyBox start request")
		printFailureDetail(out, ko.CombinedOut)
		return false

	}
	util.PrettyPrintOk(out, "Issued BusyBox start request")

	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(ngDeploymentName, fmt.Sprintf("%snginx:stable-alpine", registryURL), nginxCount); !ko.Success {
		util.PrettyPrintErr(out, "Issued Nginx start request")
		printFailureDetail(out, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, "Issued Nginx start request")

	// Add service
	if ko := RunOCinNamespace("expose", "dc", ngDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
		util.PrettyPrintErr(out, "Issued expose Nginx service request")
		printFailureDetail(out, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, "Issued expose Nginx service request")

	// Wait until deployments are ready
	return waitForDeployments(busyboxCount, nginxCount)
}

func initProject(out io.Writer) bool {
	ocOut := RunGetProject(config.Namespace)
	progressMsg := "Issued delete "+config.Namespace + " project request"
	if ocOut.Success {
		//smokeshift project exists so delete it
		if ocDelOut := RunDeleteProject(config.Namespace); !ocDelOut.Success {
			util.PrettyPrintErr(out, progressMsg)
			printFailureDetail(out, ocDelOut.CombinedOut)
			return false
		}
		util.PrettyPrintOk(out, progressMsg)
	}

	return createProject(out)
}

func createProject(out io.Writer) bool {
	progressMsg := "Issued create "+config.Namespace + " project request"
	if ocOut := RunCreateProject(config.Namespace); !ocOut.Success {
		util.PrettyPrintErr(out, progressMsg)
		printFailureDetail(out, ocOut.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, progressMsg)

	user := "system:serviceaccount:"+config.Namespace+":default"
	progressMsg = "Enable containers with any user id to be launched in project "+config.Namespace
	if ocOut := RunEnablePolicy("add-scc-to-user", "anyuid", user); !ocOut.Success {
		util.PrettyPrintErr(out, progressMsg)
		printFailureDetail(out, ocOut.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, progressMsg)

	return true
}

func checkPreconditions(out io.Writer) bool {
	ok := true
	if !precheckOC(out) {
		return false // don't bother doing anything if oc isn't configured
	}

	if !precheckAuthenticated(out) {
		return false
	}

	return ok
}

func precheckOC(out io.Writer) bool {
	progressMsg := "Configured OC CLI exists"
	if ko := RunOCinNamespace("version"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, progressMsg)
		printFailureDetail(os.Stdout, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, progressMsg)
	return true
}

func precheckAuthenticated(out io.Writer) bool {
	progressMsg := "User authenticated to cluster"
	ocOut := RunOCinNamespace("whoami")
	if !ocOut.Success {
		util.PrettyPrintErr(os.Stdout, progressMsg)
		printFailureDetail(os.Stdout, ocOut.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, progressMsg)
	return true
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
	for time.Since(start) < deploymentTimeout {
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
	powerDownResource("Nginx service ("+nginxServiceName+")", "delete", "service", nginxServiceName)

	// Power down bb
	powerDownResource("Busybox deployment ("+bbDeploymentName+")", "delete", "dc", bbDeploymentName)

	// Power down nginx
	powerDownResource("Nginx deployment ("+ ngDeploymentName + ")", "delete", "dc", ngDeploymentName)

	//Remove Project
	progressMsg := "Deleted "+config.Namespace + " project"
	if ocOut := RunDeleteProject(config.Namespace); ocOut.Success {
		util.PrettyPrintOk(os.Stdout, progressMsg)
	} else {
		util.PrettyPrintErr(os.Stdout, progressMsg)
		printFailureDetail(os.Stdout, ocOut.CombinedOut)
	}
}

func powerDownResource (resourceName string, args ...string) {
	progressMsg := "Powered down " + resourceName
	if ocOut := RunOCinNamespace(args...); ocOut.Success {
		util.PrettyPrintOk(os.Stdout, progressMsg)
	} else {
		util.PrettyPrintErr(os.Stdout, progressMsg)
		printFailureDetail(os.Stdout, ocOut.CombinedOut)
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

func printUserDetail(out io.Writer) {
	ocOut := RunOCinNamespace("whoami")
	user := strings.Replace(ocOut.CombinedOut, "\n", "", -1)
	ocOut = RunOCinNamespace("whoami", "--show-server")
	server := strings.Replace(ocOut.CombinedOut, "\n", "", -1)
	util.PrettyPrintInfo(out, "Accessing "+ server + " as user "+user)
}
