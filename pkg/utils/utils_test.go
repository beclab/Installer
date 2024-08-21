package utils

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestA(t *testing.T) {
	var a = "aaa.tar.gz"
	var b = filepath.Base(a)
	fmt.Print(b)

}
