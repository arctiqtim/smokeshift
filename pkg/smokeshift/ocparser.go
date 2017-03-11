package smokeshift

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/opencredo/smokeshift/pkg/config"
)

type OCOutput struct {
	Success     bool
	CombinedOut string
	RawOut      []byte
}

func RunOCinNamespace(args ...string) OCOutput {
	if config.Namespace != "" {
		args = append([]string{"--namespace=" + config.Namespace}, args...)
	}

	return RunOC(args...)
}

func RunOC(args ...string) OCOutput {
	OCCmd := exec.Command("oc", args...)
	bytes, err := OCCmd.CombinedOutput()
	if err != nil {
		return OCOutput{
			Success:     false,
			CombinedOut: string(bytes),
			RawOut:      bytes,
		}
	}
	return OCOutput{
		Success:     true,
		CombinedOut: string(bytes),
		RawOut:      bytes,
	}
}

func RunGetService(svcName string) OCOutput {
	return RunOCinNamespace("get", "service", svcName, "-o", "json")
}

func RunGetPodByImage(name string) OCOutput {
	return RunOCinNamespace("get", "deployment", name, "-o", "json")
}

func RunGetDeployment(name string) OCOutput {
	return RunOCinNamespace("get", "dc", name, "-o", "json")
}

func RunGetProject(name string) OCOutput {
	return RunOCinNamespace("get", "project", name, "-o", "json")
}

func RunCreateProject(name string) OCOutput {
	return RunOC("new-project", name, "--skip-config-write=true")
}

func RunDeleteProject(name string) OCOutput {
	return RunOC("delete", "project", name)
}

func RunEnablePolicy(args ...string) OCOutput {
	args = append([]string{"adm", "policy"}, args...)
	return RunOC(args...)
}

func RunPod(name string, image string, count int64) OCOutput {
	return RunOCinNamespace("run", name, "--image="+image, "--image-pull-policy=IfNotPresent", "--replicas="+strconv.FormatInt(count, 10), "-o", "json")
}

func RunGetNodes() OCOutput {
	return RunOCinNamespace("get", "nodes", "-o", "json")
}

func (ko OCOutput) ObservedReplicaCount() int64 {
	resp := DeploymentResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Status.AvaiableReplicas
}

type DeploymentResponse struct {
	Status struct {
		AvaiableReplicas int64 `json:"availableReplicas"`
	} `json:"status"`
}

func (ko OCOutput) ServiceCluserIP() string {
	resp := ServiceResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Spec.ClusterIP
}

type ServiceResponse struct {
	Spec struct {
		ClusterIP string `json:"clusterIP"`
	} `json:"spec"`
}

func (ko OCOutput) PodIPs() []string {
	//In Scala, this code would be gorgeous. In Golang, it's a blood blister
	resp := PodsResponse{}
	if err := json.Unmarshal(ko.RawOut, &resp); err != nil {
		fmt.Println(err)
	}
	podIPs := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		podIPs[i] = item.Status.PodIP
	}
	return podIPs
}

func (ko OCOutput) FirstPodName() string {
	resp := PodsResponse{}
	if err := json.Unmarshal(ko.RawOut, &resp); err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(ko.RawOut, &resp)
	if len(resp.Items) < 1 {
		return ""
	}
	return resp.Items[0].Metadata.Name
}

type PodsResponse struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			PodIP string `json:"podIP"`
		} `json:"status"`
	} `json:"items"`
}

type NodeResponse struct {
	Items []struct {
		Spec struct {
			Unschedulable bool `json:"unschedulable,omitempty"`
		} `json:"spec"`
	} `json:"items"`
}

func (ko OCOutput) NodeCount() int {
	resp := NodeResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	count := 0
	for _, item := range resp.Items {
		if item.Spec.Unschedulable == false {
			count++
		}
	}
	return count
}

func (ko OCOutput) NamespaceStatus() string {
	resp := NamespaceResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Status.Phase
}

type NamespaceResponse struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}
