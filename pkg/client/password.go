package client

import (
	"fmt"

	"github.com/howeyc/gopass"
)

const (
	minPassLen = 8
)

func checkPassword(pass string) error {
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
	if err := checkPassword(string(p)); err != nil {
		return "", err
	}
	return string(p), nil
}
