package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func (s *Service) GetOSUser(vmid int) (string, error) {
	data, err := os.ReadFile("osRecord.json")
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return "", err
	}

	// 解析 JSON 到 map
	var osRecord map[string]string
	err = json.Unmarshal(data, &osRecord)
	if err != nil {
		fmt.Printf("failed to unmarshal JSON: %v\n", err)
		return "", err
	}

	osUser := osRecord[strconv.Itoa(vmid)]
	if osUser == "" {
		return "", fmt.Errorf("no OS user found for VMID %d", vmid)
	}

	return osUser, nil
}
