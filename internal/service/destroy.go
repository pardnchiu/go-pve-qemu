package service

import (
	"fmt"
	"os/exec"
	"strconv"
)

func (s *Service) Destroy(vmid int) error {
	isMain, _, ip := s.getVMIDsNode(vmid)
	if isMain {
		cmd := exec.Command("qm", "destroy", strconv.Itoa(vmid))
		err := cmd.Run()
		if err != nil {
			err = fmt.Errorf("[-] failed to destroy VM: %v", err)
			return err
		}
	} else {
		args := []string{
			"-o", "ConnectTimeout=10",
			"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("root@%s", ip),
			"qm", "destroy", strconv.Itoa(vmid),
		}
		cmd := exec.Command("ssh", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("[-] failed to destroy VM via SSH: %v, output: %s", err, string(output))
			return err
		}
	}

	err := s.DeleteOSUser(vmid)
	if err != nil {
		println("[-] failed to delete OS user record: %v", err)
	}

	return nil
}
