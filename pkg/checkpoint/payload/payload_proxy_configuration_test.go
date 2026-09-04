package payload

import "testing"

func TestGetServePathTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "template with leading slash",
			template: "/data/v1/payloads/%s/%s",
			expected: "/data/v1/payloads/%s/%s",
		},
		{
			name:     "template without leading slash",
			template: "data/v1/payloads/%s/%s",
			expected: "/data/v1/payloads/%s/%s",
		},
		{
			name:     "empty template",
			template: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RequestPayloadProxyConfiguration{
				ServePathTemplate: tt.template,
			}
			if got := config.GetServePathTemplate(); got != tt.expected {
				t.Errorf("GetServePathTemplate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
