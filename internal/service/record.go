package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

var maxTry = 3
var nowTry = 0

func (s *Service) GetOSUser(vmid int) (string, error) {
	data, err := os.ReadFile("osRecord.json")
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		return "", err
	}

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

	if osUser == "rockylinux" {
		osUser = "rocky"
	}

	return osUser, nil
}

func (s *Service) RecordOSUser(vmid int, osUser string) error {
	data, err := os.ReadFile("osRecord.json")
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
	}

	var osRecord map[string]string
	if err == nil {
		err = json.Unmarshal(data, &osRecord)
		if err != nil {
			fmt.Printf("failed to unmarshal JSON: %v\n", err)
			osRecord = make(map[string]string)
		}
	} else {
		osRecord = make(map[string]string)
	}

	osRecord[strconv.Itoa(vmid)] = osUser
	updatedData, err := json.MarshalIndent(osRecord, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal JSON: %v\n", err)
		nowTry++
		if nowTry < maxTry {
			return s.RecordOSUser(vmid, osUser)
		} else {
			nowTry = 0
		}
		return err
	}

	err = os.WriteFile("osRecord.json", updatedData, 0644)
	if err != nil {
		fmt.Printf("failed to write file: %v\n", err)
		nowTry++
		if nowTry < maxTry {
			return s.RecordOSUser(vmid, osUser)
		} else {
			nowTry = 0
		}
		return err
	}

	return nil
}

func (s *Service) DeleteOSUser(vmid int) {
	data, err := os.ReadFile("osRecord.json")
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
	}

	var osRecord map[string]string
	if err == nil {
		err = json.Unmarshal(data, &osRecord)
		if err != nil {
			fmt.Printf("failed to unmarshal JSON: %v\n", err)
			osRecord = make(map[string]string)
		}
	} else {
		osRecord = make(map[string]string)
	}

	delete(osRecord, strconv.Itoa(vmid))
	updatedData, err := json.MarshalIndent(osRecord, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal JSON: %v\n", err)
		nowTry++
		if nowTry < maxTry {
			s.DeleteOSUser(vmid)
		} else {
			nowTry = 0
		}
		return
	}

	err = os.WriteFile("osRecord.json", updatedData, 0644)
	if err != nil {
		fmt.Printf("failed to write file: %v\n", err)
		nowTry++
		if nowTry < maxTry {
			s.DeleteOSUser(vmid)
		} else {
			nowTry = 0
		}
		return
	}
}
