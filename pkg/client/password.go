package client

import (
	"fmt"

	"github.com/howeyc/gopass"
)

func GetMaskedPassword() (string, error) {
	fmt.Print("Password: ")
	p, err := gopass.GetPasswdMasked()
	if err != nil && err != gopass.ErrInterrupted {
		return "", err
	}
	return string(p), nil
}
