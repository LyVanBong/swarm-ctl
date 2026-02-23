package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner chạy Ansible playbooks
type Runner struct {
	PlaybooksDir string
	Inventory    string
	ExtraVars    map[string]string
	Verbose      bool
}

// NewRunner tạo Ansible runner mới
func NewRunner(playbooksDir string) *Runner {
	return &Runner{
		PlaybooksDir: playbooksDir,
		ExtraVars:    make(map[string]string),
	}
}

// WithHost tạo temporary inventory cho 1 host
func (r *Runner) WithHost(ip, user, keyPath string) *Runner {
	r.Inventory = fmt.Sprintf("%s ansible_user=%s ansible_ssh_private_key_file=%s ansible_ssh_common_args='-o StrictHostKeyChecking=no'",
		ip, user, keyPath)
	return r
}

// WithVar thêm extra variable
func (r *Runner) WithVar(key, value string) *Runner {
	r.ExtraVars[key] = value
	return r
}

// RunPlaybook chạy 1 playbook cụ thể
func (r *Runner) RunPlaybook(playbook string) error {
	playbookPath := filepath.Join(r.PlaybooksDir, playbook)

	args := []string{playbookPath}

	// Inventory
	if r.Inventory != "" {
		// Tạo temp inventory file
		tmpFile, err := os.CreateTemp("", "swarm-inventory-*.ini")
		if err != nil {
			return fmt.Errorf("cannot create temp inventory: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := fmt.Fprintf(tmpFile, "[nodes]\n%s\n", r.Inventory); err != nil {
			return err
		}
		tmpFile.Close()
		args = append(args, "-i", tmpFile.Name())
	}

	// Extra vars
	if len(r.ExtraVars) > 0 {
		var vars []string
		for k, v := range r.ExtraVars {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		args = append(args, "--extra-vars", strings.Join(vars, " "))
	}

	// Verbose mode
	if r.Verbose {
		args = append(args, "-v")
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playbook '%s' failed: %w", playbook, err)
	}
	return nil
}

// IsInstalled kiểm tra Ansible có được cài không
func IsInstalled() bool {
	_, err := exec.LookPath("ansible-playbook")
	return err == nil
}

// InstallAnsible cài Ansible nếu chưa có
func InstallAnsible() error {
	// Kiểm tra pip3
	if _, err := exec.LookPath("pip3"); err != nil {
		return fmt.Errorf("pip3 not found — please install Python3")
	}

	cmd := exec.Command("pip3", "install", "ansible", "--quiet")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
