package helper

import (
	"path"

	"bytetrade.io/web3os/installer/pkg/core/logger"
)

func InitLog(workDir string) error {
	logDir := path.Join(workDir, "logs")
	logger.InitLog(logDir)
	return nil
}
