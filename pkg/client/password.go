package client

import (
	"fmt"

	"github.com/howeyc/gopass"
)

const (
	minPassLen = 8
)

func EnsurePasswordLength(pass string) error {
	if len([]rune(pass)) < minPassLen {
		return fmt.Errorf("minimum password length is %d", minPassLen)
	}
	return nil
}

func GetMaskedPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	p, err := gopass.GetPasswdMasked()
	if err != nil && err != gopass.ErrInterrupted {
		return "", err
	}
	fmt.Print("\r")
	return string(p), nil
}
