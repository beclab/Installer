package utils

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestA(t *testing.T) {
	var a = "/home/ubuntu/.terminus/versions/v1.8.0-20240928/wizard/config/apps/argo"
	var b = filepath.Base(a)
	fmt.Println("---b---", b)
}
