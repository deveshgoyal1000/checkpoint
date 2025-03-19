// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// NetworkStatusFile is the file containing network information in Podman checkpoints
	NetworkStatusFile = "network.status"
)

// PodmanNetworkStatus represents the network configuration from a Podman checkpoint
type PodmanNetworkStatus struct {
	Podman struct {
		Interfaces map[string]NetworkInterface `json:"interfaces"`
	} `json:"podman"`
}

// NetworkInterface represents a network interface configuration in a Podman checkpoint
type NetworkInterface struct {
	Subnets    []NetworkSubnet `json:"subnets"`
	MacAddress string          `json:"mac_address"`
}

// NetworkSubnet represents IP and gateway information for a network interface
type NetworkSubnet struct {
	IPNet   string `json:"ipnet"`
	Gateway string `json:"gateway"`
}

// ReadNetworkStatus reads and parses the network status from a checkpoint
func ReadNetworkStatus(checkpointDirectory string) (*PodmanNetworkStatus, string, error) {
	var networkStatus PodmanNetworkStatus
	networkStatusFile, err := ReadJSONFile(&networkStatus, checkpointDirectory, NetworkStatusFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read network status: %w", err)
	}

	return &networkStatus, networkStatusFile, nil
}

// FormatNetworkInfo returns a formatted string representation of network information
func FormatNetworkInfo(networkStatus *PodmanNetworkStatus) string {
	if networkStatus == nil || len(networkStatus.Podman.Interfaces) == 0 {
		return "No network interfaces found"
	}

	var result string
	for ifName, iface := range networkStatus.Podman.Interfaces {
		result += fmt.Sprintf("Interface: %s\n", ifName)
		result += fmt.Sprintf("  MAC Address: %s\n", iface.MacAddress)
		
		for _, subnet := range iface.Subnets {
			result += fmt.Sprintf("  IP/Subnet: %s\n", subnet.IPNet)
			result += fmt.Sprintf("  Gateway: %s\n", subnet.Gateway)
		}
		result += "\n"
	}
	
	return result
} 