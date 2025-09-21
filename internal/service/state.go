package service

import (
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func (s *Service) Stop(vmid int) error {
	cmd := exec.Command("qm", "stop", strconv.Itoa(vmid))
	return cmd.Run()
}

func (s *Service) Shutdown(vmid int) error {
	cmd := exec.Command("qm", "shutdown", strconv.Itoa(vmid))
	return cmd.Run()
}

func (s *Service) Destroy(vmid int) error {
	cmd := exec.Command("qm", "destroy", strconv.Itoa(vmid))
	return cmd.Run()
}

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
