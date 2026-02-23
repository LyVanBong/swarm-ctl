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
	fmt.Println("Đang tự động cài đặt Ansible...")
	installCmd := ""

	// Phán đoán HĐH
	if _, err := exec.LookPath("apt-get"); err == nil {
		// Update repository and install ansible via python3-pip or native apt?
		// Ubuntu 20.04+ has standard ansible pkg
		installCmd = "sudo apt-get update && sudo apt-get install -y ansible sshpass"
	} else if _, err := exec.LookPath("yum"); err == nil {
		installCmd = "sudo yum install -y epel-release && sudo yum install -y ansible sshpass"
	} else if _, err := exec.LookPath("brew"); err == nil {
		installCmd = "brew install ansible hudochenkov/sshpass/sshpass"
	}

	if installCmd != "" {
		cmd := exec.Command("sh", "-c", installCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
		// Fallback to pip3 if system package fail
	}

	// Kiểm tra pip3
	if _, err := exec.LookPath("pip3"); err != nil {
		fmt.Println("Không tìm thấy lệnh cài đặt tự động (apt/yum/brew) hoặc pip3. Vui lòng tự cài Ansible thủ công.")
		return fmt.Errorf("không thể cài tự động Ansible")
	}

	fmt.Println("Đang chạy lệnh pip3 install ansible...")
	cmd := exec.Command("pip3", "install", "ansible", "--quiet")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
