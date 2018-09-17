package main

import "fmt"

func schedulerStub(nodeName string, requireNum int) (int, []string) {
	var selectedIDs []string
	score := 0
	for i := 0; i < requireNum; i++ {
		switch i {
		case 0:
			selectedIDs = append(selectedIDs, "GPU-49a1eb24-cedf-74a0-a03f-c714ce3f9aac")
			break
		case 1:
			selectedIDs = append(selectedIDs, "GPU-67e491a3-e289-fb76-7ec2-f444503730b6")
			break
		}
	}
	score = 1
	return score, selectedIDs
}

func bestFit(nodeName string, requireNum int) (int, []string) {
	var selectedIDs []string
	score := 0
	connectGraphs := nodeConnectGraphs[nodeName]
	idleDevices := nodeIdleDevices[nodeName]
	// Search each node
	for nodeID, _ := range connectGraphs {
		// This node is idle
		if idleDevices[nodeID] {
			var selectNode string
			localSelectedNodes := make(map[string]Empty)
			localScore := LinkSpeed(0)
			maxSpeed := LinkSpeed(0)
			// Select it (Add to selectedNodes), and start selecting (requireNum-1) devices base on it (suppose it is the start point)
			selectNode = nodeID
			for i := 0; i < requireNum; i++ {
				localSelectedNodes[selectNode] = Empty{}
				localScore += maxSpeed
				maxSpeed = LinkSpeed(0)
				selectNode = ""
				// for each selected node
				for selected, _ := range localSelectedNodes {
					// Search its neighbor
					for neig, speed := range connectGraphs[selected].ConnectedDevice {
						_, ok := localSelectedNodes[neig]
						// This node is idle and haven't been chosen
						if idleDevices[neig] && !ok {
							// Find the neighbor who has the quickest speed
							if speed > maxSpeed {
								selectNode = neig
								maxSpeed = speed
							}
						}
					}
				}
				if selectNode == "" {
					fmt.Printf("bestFit:: node has no neighbor\n")
				}
			}
			if localScore > LinkSpeed(score) {
				score = int(localScore)
				for k, _ := range localSelectedNodes {
					selectedIDs = append(selectedIDs, k)
				}
			}
			// Search done when suppose this node as start point
		}
	}
	fmt.Printf("bestFit:: Success\n")
	return score, selectedIDs
}
