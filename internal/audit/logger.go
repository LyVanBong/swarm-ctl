package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Logger lưu vết các lệnh đã thực thi
func Log(command []string) {
	// Không log các lệnh chỉ mang tính chất đọc/giám sát
	readOnlyCmds := map[string]bool{
		"list": true, "ls": true, "status": true, "version": true,
		"dashboard": true, "logs": true, "doctor": true,
	}

	if len(command) == 0 {
		return
	}

	// Lọc các subcommands
	subCmd := command[len(command)-1]
	if len(command) > 1 {
		subCmd = command[1]
	}

	if readOnlyCmds[subCmd] || subCmd == "help" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	auditDir := filepath.Join(home, ".swarm-ctl")
	os.MkdirAll(auditDir, 0755)
	
	auditFile := filepath.Join(auditDir, "audit.log")
	
	f, err := os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	// Che mật khẩu/secret values nếu có trong lệnh secret add/rotate
	sanitizedCmd := filterSensitiveArgs(command)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] USER=%s COMMAND=\"swarm-ctl %s\"\n", timestamp, user, strings.Join(sanitizedCmd, " "))
	
	f.WriteString(logEntry)
}

func filterSensitiveArgs(args []string) []string {
	res := make([]string, len(args))
	copy(res, args)

	if len(res) > 2 && res[0] == "secret" {
		if res[1] == "add" || res[1] == "rotate" {
			// Cấu trúc: secret add <name> <value> 
			// Hoặc:    secret rotate <name> <new-value>
			if len(res) >= 3 {
				res[len(res)-1] = "********"
			}
		}
	}
	// Che thông số --key nếu có
	for i, arg := range res {
		if (arg == "--key" || arg == "-k") && i+1 < len(res) {
			res[i+1] = "********"
		}
	}
	
	return res
}
