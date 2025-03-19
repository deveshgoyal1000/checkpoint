// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadNetworkStatus(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "network-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test network.status file
	testData := `{
		"podman": {
			"interfaces": {
				"eth0": {
					"subnets": [
						{
							"ipnet": "10.88.0.9/16",
							"gateway": "10.88.0.1"
						}
					],
					"mac_address": "f2:99:8d:fb:5a:57"
				}
			}
		}
	}`

	networkFile := filepath.Join(tmpDir, NetworkStatusFile)
	if err := os.WriteFile(networkFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test ReadNetworkStatus
	networkStatus, file, err := ReadNetworkStatus(tmpDir)
	if err != nil {
		t.Fatalf("ReadNetworkStatus failed: %v", err)
	}

	if file != networkFile {
		t.Errorf("Expected file path %s, got %s", networkFile, file)
	}

	// Verify the parsed data
	if len(networkStatus.Podman.Interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(networkStatus.Podman.Interfaces))
	}

	iface, ok := networkStatus.Podman.Interfaces["eth0"]
	if !ok {
		t.Fatal("eth0 interface not found")
	}

	if iface.MacAddress != "f2:99:8d:fb:5a:57" {
		t.Errorf("Expected MAC f2:99:8d:fb:5a:57, got %s", iface.MacAddress)
	}

	if len(iface.Subnets) != 1 {
		t.Errorf("Expected 1 subnet, got %d", len(iface.Subnets))
	}

	if iface.Subnets[0].IPNet != "10.88.0.9/16" {
		t.Errorf("Expected IP 10.88.0.9/16, got %s", iface.Subnets[0].IPNet)
	}

	if iface.Subnets[0].Gateway != "10.88.0.1" {
		t.Errorf("Expected gateway 10.88.0.1, got %s", iface.Subnets[0].Gateway)
	}
}

func TestFormatNetworkInfo(t *testing.T) {
	testCases := []struct {
		name     string
		input    *PodmanNetworkStatus
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "No network interfaces found",
		},
		{
			name: "empty interfaces",
			input: &PodmanNetworkStatus{
				Podman: struct {
					Interfaces map[string]NetworkInterface `json:"interfaces"`
				}{
					Interfaces: make(map[string]NetworkInterface),
				},
			},
			expected: "No network interfaces found",
		},
		{
			name: "single interface",
			input: &PodmanNetworkStatus{
				Podman: struct {
					Interfaces map[string]NetworkInterface `json:"interfaces"`
				}{
					Interfaces: map[string]NetworkInterface{
						"eth0": {
							MacAddress: "f2:99:8d:fb:5a:57",
							Subnets: []NetworkSubnet{
								{
									IPNet:   "10.88.0.9/16",
									Gateway: "10.88.0.1",
								},
							},
						},
					},
				},
			},
			expected: "Interface: eth0\n  MAC Address: f2:99:8d:fb:5a:57\n  IP/Subnet: 10.88.0.9/16\n  Gateway: 10.88.0.1\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatNetworkInfo(tc.input)
			if result != tc.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
} 