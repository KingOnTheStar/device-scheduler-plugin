package main

import (
	"encoding/json"
	"fmt"
)

const (
	Resourcename = "nvidia.com/gpu"
)

var nodeConnectGraphs map[string]ConnectGraph
var nodeHealthyDevices map[string]UsableDeviceMap
var nodeIdleDevices map[string]UsableDeviceMap

func HelloDLL(input string) {
	fmt.Printf("Hello DLL : %s\n", input)
	fmt.Printf("Test DLL : %s\n", nodeIdleDevices)
	fmt.Printf("Test2 DLL : %s\n", nodeHealthyDevices)
}

func Init() {
	nodeConnectGraphs = make(map[string]ConnectGraph)
	nodeHealthyDevices = make(map[string]UsableDeviceMap)
	nodeIdleDevices = make(map[string]UsableDeviceMap)
}

func GetResourceName() string {
	return Resourcename
}

func OnAddNode(nodeName string, annotation map[string]string) {
	// Read Topology
	connectGraph := make(ConnectGraph)
	err := json.Unmarshal([]byte(annotation["node.dm.alpha.kubernetes.io/Topology"]), &connectGraph)
	if err != nil {
		fmt.Printf("AddNode:: Unmarshal Fail - %v\n", err)
		return
	}

	nodeConnectGraphs[nodeName] = connectGraph

	// Read device health status
	healthyDevice := make(UsableDeviceMap)
	err = json.Unmarshal([]byte(annotation["node.dm.alpha.kubernetes.io/UsableDeviceMap"]), &healthyDevice)
	if err != nil {
		fmt.Printf("AddNode:: Unmarshal Fail - %v\n", err)
		return
	}

	nodeHealthyDevices[nodeName] = healthyDevice

	// Set device status, Deep copy
	for k, v := range nodeHealthyDevices {
		nodeIdleDevices[k] = v.DeepCopy()
	}

	fmt.Printf("AddNode:: Success: %s\n", healthyDevice)
}

func OnUpdateNode(nodeName string, annotation map[string]string) {
	// Read Topology
	if _, ok := nodeConnectGraphs[nodeName]; !ok {
		connectGraph := make(ConnectGraph)
		err := json.Unmarshal([]byte(annotation["node.dm.alpha.kubernetes.io/Topology"]), &connectGraph)
		if err != nil {
			fmt.Printf("UpdateNode:: Unmarshal Fail - %v\n", err)
			return
		}

		nodeConnectGraphs[nodeName] = connectGraph
		fmt.Printf("UpdateNode:: Success: %s\n", connectGraph)
	}

	// Read device health status
	healthyDevice := make(UsableDeviceMap)
	err := json.Unmarshal([]byte(annotation["node.dm.alpha.kubernetes.io/UsableDeviceMap"]), &healthyDevice)
	if err != nil {
		fmt.Printf("UpdateNode:: Unmarshal Fail - %v\n", err)
		return
	}

	nodeHealthyDevices[nodeName] = healthyDevice

	// Set device status
	idleDevice := nodeIdleDevices[nodeName]
	for k, v := range healthyDevice {
		// If device doesn't exsit or is unhealthy
		if _, ok := idleDevice[k]; !ok || !v {
			idleDevice[k] = v
		}
	}

	fmt.Printf("UpdateNode:: Success: %s\n", healthyDevice)
}

func OnDeleteNode(nodeName string) {
	delete(nodeConnectGraphs, nodeName)
	delete(nodeHealthyDevices, nodeName)
	delete(nodeIdleDevices, nodeName)
}

func AssessTaskAndNode(nodeName string, requireNum int) (int, map[string]string) {
	podAnnotation := make(map[string]string)
	// Scheduler
	score, selectedIDs := schedulerStub(nodeName, requireNum)
	// Marshal
	data, err := json.Marshal(selectedIDs)
	if err != nil {
		fmt.Printf("AssessTaskAndNode:: Marshal Fail - %v\n", err)
		return score, nil
	}
	podAnnotation["node.dm.alpha.kubernetes.io/SelectedIDs"] = string(data)
	return score, podAnnotation
}

func OnAddTask(nodeName string, annotation map[string]string) {
	var selectedIDs []string
	ids, ok := annotation["node.dm.alpha.kubernetes.io/SelectedIDs"]
	if !ok {
		return
	}
	err := json.Unmarshal([]byte(ids), &selectedIDs)
	if err != nil {
		fmt.Printf("AddTask:: Unmarshal Fail - %v\n", err)
		return
	}
	for _, id := range selectedIDs {
		nodeIdleDevices[nodeName][id] = false
	}
}

func OnRemoveTask(nodeName string, annotation map[string]string) {
	var selectedIDs []string
	ids, ok := annotation["node.dm.alpha.kubernetes.io/SelectedIDs"]
	if !ok {
		return
	}
	err := json.Unmarshal([]byte(ids), &selectedIDs)
	if err != nil {
		fmt.Printf("RemoveTask:: Unmarshal Fail - %v\n", err)
		return
	}
	for _, id := range selectedIDs {
		// If healthy
		if nodeHealthyDevices[nodeName][id] {
			nodeIdleDevices[nodeName][id] = true
		} else {
			nodeIdleDevices[nodeName][id] = false
		}
	}
}
