package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func (s *Service) Start(c *gin.Context, vmid int) error {
	origin := c.Request.Header.Get("Origin")
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	var step string

	// * 1. Start VM
	stepStart := time.Now()
	isMain, _, ip := s.getVMIDsNode(vmid)
	if isMain {
		step = "starting VM"
		s.SSE(c, step, "info", "[*] starting")
		cmd := exec.Command("qm", "start", strconv.Itoa(vmid))
		if err := cmd.Run(); err != nil {
			err = fmt.Errorf("[-] failed to start VM: %v", err)
			s.SSE(c, step, "error", err.Error())
			return err
		}
	} else {
		step = "starting VM via SSH"
		s.SSE(c, step, "processing", "[*] starting VM via SSH")
		args := []string{
			"-o", "ConnectTimeout=10",
			"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("root@%s", ip),
			"qm", "start", strconv.Itoa(vmid),
		}
		cmd := exec.Command("ssh", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("[-] failed to start VM via SSH: %v, output: %s", err, string(output))
			s.SSE(c, step, "error", err.Error())
			return err
		}
	}
	elapsed := time.Since(stepStart)
	s.SSE(c, step, "success", fmt.Sprintf("[+] VM starting (%.2fs)", elapsed.Seconds()))

	// * 2. Wait for SSH connection
	stepStart = time.Now()
	step = "waiting for SSH"
	osUser, err := s.GetOSUser(vmid)
	if err != nil {
		err = fmt.Errorf("[-] failed to get OS user: %v", err)
		s.SSE(c, step, "info", err.Error())
		return nil
	}

	if err := s.CheckAlive(c, osUser, vmid); err != nil {
		err = fmt.Errorf("[-] failed to connect to VM via SSH: %w", err)
		s.SSE(c, step, "error", err.Error())
		return nil
	}

	elapsed = time.Since(stepStart)
	s.SSE(c, step, "success", fmt.Sprintf("[+] VM is ready (%.2fs)", elapsed.Seconds()))

	// * 3. Finalizing
	step = "finalizing"
	time.Sleep(5 * time.Second)
	s.SSE(c, step, "info", "[+] VM started successfully")

	return nil
}
