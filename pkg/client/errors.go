package client

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	"google.golang.org/grpc/status"
)

func GetErrorMsg(err error) string {
	stat, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}
	return stat.Message()
}

func PrintErrorAndExit(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, color.RedString(format, args...))
	os.Exit(1)
}

func PrintConnectionErrorAndExit(err error) {
	PrintErrorAndExit(
		"A connection error occoured.\n" +
		"Please, try again.\n\n" +
		"Server details: %s", GetErrorMsg(err),
	)
}
