package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func (s *Service) checkUserPubkey() (bool, error) {
	dirHome, err := os.UserHomeDir()
	if err != nil {
		return true, err
	}

	dirSSH := filepath.Join(dirHome, ".ssh")
	list := []string{
		"id_rsa.pub",
		"id_ed25519.pub",
		"id_ecdsa.pub",
	}

	for _, e := range list {
		path := filepath.Join(dirSSH, e)
		if _, err := os.Stat(path); err == nil {
			return false, nil
		}
	}

	return false, fmt.Errorf("failed to find user SSH public key")
}

func (s *Service) createUserSSHKeyPair() error {
	dirHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirSSH := filepath.Join(dirHome, ".ssh")
	if err := os.MkdirAll(dirSSH, 0700); err != nil {
		return err
	}

	keyPath := filepath.Join(dirSSH, "id_ed25519")
	hostname, _ := os.Hostname()
	currentUser := os.Getenv("USER")

	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-N", "", "-C", fmt.Sprintf("%s@%s", currentUser, hostname))
	return cmd.Run()
}
