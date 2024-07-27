package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestA(t *testing.T) {
	var str = "unpacking docker.io/cesign/aria2-pro:latest (sha256:1fe80b1565031459ad2a2ed9f5f4d8887b508beef86777d10332246c86df96fb)...done"

	if strings.Contains(str, "(sha256:") {
		str = strings.Split(str, "(sha256:")[0]
	}

	fmt.Println("---1---", str)
}
