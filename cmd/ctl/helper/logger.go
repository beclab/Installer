package helper

import (
	"os"
	"path"

	"bytetrade.io/web3os/installer/pkg/core/logger"
)

func InitLog(baseDir string) {
	if baseDir == "" {
		baseDir = path.Join(os.Getenv("HOME"), ".terminus")
	}
	logger.InitLog(path.Join(baseDir, "logs"))
}
