package zipcode

import "testing"

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		cep  string
		want bool
	}{
		{"valid cep", "12345678", true},
		{"invalid cep - too short", "123", false},
		{"invalid cep - too long", "123456789", false},
		{"invalid cep - contains letters", "a1234567", false},
		{"invalid cep - contains special characters", "1234567-", false},
		{"invalid cep - empty string", "", false},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			got := IsValid(v.cep)

			if got != v.want {
				t.Errorf("IsValid(%q)=%v, want %t", v.cep, got, v.want)
			}
		})
	}
}
