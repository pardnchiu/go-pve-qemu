package service

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"guthub.com/pardnchiu/go-qemu/internal/model"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Gateway string
}

func NewService(gateway string) *Service {
	return &Service{
		Gateway: gateway,
	}
}

func (s *Service) getStorages() map[string]bool {
	mapStorages := make(map[string]bool)

	// list all storages
	cmd := exec.Command("pvesm", "status")
	cmd.Stderr = nil
	// get command output for filtering
	output, err := cmd.Output()
	if err != nil {
		return mapStorages
	}

	/**
	 * pvesm status output example:
	 * Name 				Type 		Status 		Total 	Used 	Available 	%
	 * local-zfs 		zfspool active 		100G 		10G 	90G 				10%
	 * local 				dir 		active 		200G 		50G 	150G 				25%
	 * nfs-storage 	nfs 		inactive 	500G 		100G 	400G 				20%
	 */
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// skip header and empty lines
		if strings.Contains(line, "Name") || strings.TrimSpace(line) == "" {
			continue
		}

		// split by whitespace
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			name := fields[0]
			storageType := fields[1]
			status := fields[2]

			// list only active storages of type dir, zfspool, lvmthin, nfs
			if status == "active" && (storageType == "dir" || storageType == "zfspool" || storageType == "lvmthin" || storageType == "nfs") {
				mapStorages[name] = true
			} else {
				mapStorages[name] = false
			}
		}
	}

	/**
	 * return value example	:
	 * {
	 *  "local-zfs": 		true,
	 *  "local": 				true,
	 *  "nfs-storage": 	false
	 * }
	 */
	return mapStorages
}

func (s *Service) SetDefaults(config *model.Config) error {
	if config.Name == "" {
		config.Name = fmt.Sprintf("%d", config.ID)
	} else {
		config.Name += "-" + fmt.Sprintf("%d", config.ID)
	}
	if config.CPU == 0 {
		config.CPU = 2
	}
	if config.RAM == 0 {
		config.RAM = 2048
	}
	if config.Disk == "" {
		config.Disk = "16G"
	}

	if config.User == "" {
		switch config.OS {
		case "debian":
			config.User = "debian"
		case "rockylinux":
			config.User = "rocky"
		case "ubuntu":
			config.User = "ubuntu"
		}
	}
	if config.Passwd == "" {
		config.Passwd = "passwd"
	}

	return nil
}

func (s *Service) SSE(c *gin.Context, step string, status string, message string) {
	// 檢查連線是否斷開
	select {
	case <-c.Request.Context().Done():
		// 連線已斷開，記錄但不中斷程序
		return
	default:
	}

	// 嘗試發送 SSE，如果失敗就忽略
	defer func() {
		if r := recover(); r != nil {
			// SSE 發送失敗，忽略錯誤
		}
	}()

	msg := model.SSE{
		Step:    step,
		Status:  status,
		Message: message,
	}
	data := fmt.Sprintf("data: {\"step\":\"%s\",\"status\":\"%s\",\"message\":\"%s\"}\n\n",
		msg.Step, msg.Status, msg.Message)

	if flusher, ok := c.Writer.(http.Flusher); ok {
		c.Writer.WriteString(data)
		flusher.Flush()
	}
}
