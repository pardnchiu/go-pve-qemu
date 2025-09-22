package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	maxTry   = 3
	nowTry   = 0
	filename = ".go_qemu_record"
)

func readRecord() (map[int]string, error) {
	list := make(map[int]string)
	data, err := os.ReadFile(filename)
	if err != nil {
		return list, err
	}

	if len(data) == 0 {
		return list, nil
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			vmid, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}
			list[vmid] = parts[1]
		}
	}
	return list, nil
}

func (s *Service) GetOSUser(vmid int) (string, error) {
	for retry := 0; retry < maxTry; retry++ {
		records, err := readRecord()
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("no record file found, please create VM first")
			}

			if retry == maxTry-1 {
				return "", fmt.Errorf("failed to read records")
			}

			continue
		}

		if osUser, exists := records[vmid]; exists {
			if osUser == "rockylinux" {
				osUser = "rocky"
			}
			return osUser, nil
		}

		return "", fmt.Errorf("no OS user found for VMID %d", vmid)
	}

	return "", fmt.Errorf("max retries reached")
}

func (s *Service) RecordOSUser(vmid int, osUser string) error {
	for retry := 0; retry < maxTry; retry++ {
		records, err := readRecord()
		if err != nil && os.IsNotExist(err) {
			fmt.Printf("file does not exist, creating new one\n")
			records = make(map[int]string)
		} else if err != nil {
			if retry == maxTry-1 {
				return err
			}
			continue
		}

		records[vmid] = osUser
		lines := []string{}
		for k, v := range records {
			lines = append(lines, fmt.Sprintf("%d:%s", k, v))
		}
		content := strings.Join(lines, "\n") + "\n"

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			if retry == maxTry-1 {
				return err
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("max retries reached")
}

func (s *Service) DeleteOSUser(vmid int) error {
	for retry := 0; retry < maxTry; retry++ {
		records, err := readRecord()
		if err != nil && os.IsNotExist(err) {
			return nil
		} else if err != nil {
			if retry == maxTry-1 {
				return err
			}
			continue
		}

		if _, exists := records[vmid]; !exists {
			return nil
		}

		delete(records, vmid)
		lines := []string{}
		for k, v := range records {
			lines = append(lines, fmt.Sprintf("%d:%s", k, v))
		}
		content := strings.Join(lines, "\n") + "\n"

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			if retry == maxTry-1 {
				return err
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("max retries reached")
}
