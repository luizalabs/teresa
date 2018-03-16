package app

import "strings"

func IsCronJob(processType string) bool {
	return strings.HasPrefix(processType, ProcessTypeCronPrefix)
}
