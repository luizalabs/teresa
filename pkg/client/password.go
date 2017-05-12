package client

import (
	"fmt"

	"github.com/howeyc/gopass"
)

func GetMaskedPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	p, err := gopass.GetPasswdMasked()
	if err != nil && err != gopass.ErrInterrupted {
		return "", err
	}
	fmt.Print("\r")
	return string(p), nil
}
