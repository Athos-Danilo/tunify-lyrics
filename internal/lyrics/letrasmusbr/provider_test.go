package letrasmusbr

import "testing"

func TestNormalizeForURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Espaços em branco",
			input:    "Red Hot Chili Peppers",
			expected: "red-hot-chili-peppers",
		},
		{
			name:     "Caracteres especiais e maiúsculas",
			input:    "Beyoncé",
			expected: "beyonce",
		},
		{
			name:     "Caracteres não alfanuméricos e parênteses",
			input:    "System Of A Down (SOAD)",
			expected: "system-of-a-down-soad",
		},
		{
			name:     "Múltiplos espaços e traços",
			input:    "  Artist  - Name  ",
			expected: "artist-name",
		},
		{
			name:     "Acentuação pesada",
			input:    "João & Maria (Ao Vivo)",
			expected: "joao-maria-ao-vivo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeForURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeForURL(%q) = %q; esperado %q", tt.input, result, tt.expected)
			}
		})
	}
}
