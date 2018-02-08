package uid

import "github.com/pborman/uuid"

func New() string {
	return uuid.New()[:8]
}
