package audit

import (
	"testing"
)

func TestFilterSensitiveArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Lệnh bình thường (Bỏ qua)",
			input:    []string{"node", "list"},
			expected: []string{"node", "list"},
		},
		{
			name:     "Lệnh Secret Add (Che password)",
			input:    []string{"secret", "add", "db_password", "SuperS3cr3t!"},
			expected: []string{"secret", "add", "db_password", "********"},
		},
		{
			name:     "Lệnh có cờ --key (Che đường dẫn Key)",
			input:    []string{"cluster", "init", "--key", "/root/.ssh/id_rsa", "--master", "10.0.0.1"},
			expected: []string{"cluster", "init", "--key", "********", "--master", "10.0.0.1"},
		},
		{
			name:     "Secret Rotate (Che password mới)",
			input:    []string{"secret", "rotate", "api_key", "NewK3y123"},
			expected: []string{"secret", "rotate", "api_key", "********"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterSensitiveArgs(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("Độ dài kết quả không khớp: mong đợi %d, thực tế %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Lỗi tại vị trí %d: \n Mong đợi: %s\n Thực tế: %s\n Input Test: %s",
						i, tt.expected[i], result[i], tt.name)
				}
			}
		})
	}
}
