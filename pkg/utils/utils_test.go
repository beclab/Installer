package utils

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	for i := 0; i < 100; i++ {
		a, _ := GeneratePassword(16)
		fmt.Println(a)
	}

}
