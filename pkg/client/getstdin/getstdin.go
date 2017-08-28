package getstdin

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetStdin(msg string) (string, error) {
	fmt.Print(msg)
	resp, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimRight(resp, "\r\n")), nil
}
