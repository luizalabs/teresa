package uid

import (
	"crypto/rand"
	"fmt"
)

func New() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(err.Error()) // rand should never fail
	}
	return fmt.Sprintf("%x", b)
}
