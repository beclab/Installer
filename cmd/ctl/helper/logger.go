package helper

import (
	"path"

	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
)

func InitLog() {
	logDir := path.Join(constants.WorkDir, "logs")
	logger.InitLog(logDir)
}
