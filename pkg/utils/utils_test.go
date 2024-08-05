package utils

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	var a = MD5("beclab/notification-manager-operator-ext:v0.1.0-ext")
	fmt.Println(a)
}
