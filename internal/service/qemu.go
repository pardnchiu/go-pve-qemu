package service

import (
	"os/exec"
	"strconv"
	"strings"
)

func (s *Service) GetVMStatus(vmid int) (string, error) {
	cmd := exec.Command("qm", "status", strconv.Itoa(vmid))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(output))
	if parts := strings.Split(result, ": "); len(parts) == 2 {
		return parts[1], nil
	}

	return "unknown", nil
}
