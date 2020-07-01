package tokens

import (
	"log"
	"testing"
)

func TestGenerateVerificationCode(t *testing.T) {
	t.Run("with 4 digits", func(t *testing.T) {
		code, err := GenerateVerificationCode(4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(code) != 4 {
			t.Errorf("unexpected length: %d", len(code))
			log.Println(code)
		}
	})

	t.Run("with 6 digits", func(t *testing.T) {
		code, err := GenerateVerificationCode(6)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(code) != 6 {
			t.Errorf("unexpected length: %d", len(code))
		}
	})
}
