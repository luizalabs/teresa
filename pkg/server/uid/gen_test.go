package uid

import "testing"

func TestGen(t *testing.T) {
	generatedIds := make(map[string]bool)
	for i := 0; i < 10; i++ {
		gId := New()
		if _, found := generatedIds[gId]; found {
			t.Fatal("collision detected")
		}
		generatedIds[gId] = true
	}
}
