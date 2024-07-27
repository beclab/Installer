package storage

import (
	"fmt"
	"strings"
	"testing"
)

func TestOss(t *testing.T) {
	var bkt = "https://terminus-os-cn-hongkong.oss-cn-hongkong-internal.aliyuncs.com"
	b, a, f := strings.Cut(bkt, "://")
	fmt.Println("---b---", b)
	fmt.Println("---a---", a)
	fmt.Println("---f---", f)
}
