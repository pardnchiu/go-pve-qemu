package service

import (
	"fmt"
	"os/exec"
	"strconv"
)

func (s *Service) Stop(vmid int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)
	if isMain {
		cmd := exec.Command("qm", "stop", strconv.Itoa(vmid))
		err := cmd.Run()
		if err != nil {
			err = fmt.Errorf("[-] failed to stop VM: %v", err)
			return err
		}
	} else {
		args := []string{
			"-o", "ConnectTimeout=10",
			"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("root@%s", ip),
			"qm", "stop", strconv.Itoa(vmid),
		}
		cmd := exec.Command("ssh", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("[-] failed to stop VM via SSH: %v, output: %s", err, string(output))
			return err
		}
	}
	return nil
}
