package packages

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type PackagesModule struct {
	common.KubeModule
	common.KubeConf
}

func (m *PackagesModule) Init() {
	m.Name = "DownloadInstaller"
	m.Desc = "Download installer packages"

	download := &task.LocalTask{
		Name:   "Download",
		Desc:   "Download installer packages",
		Action: new(PackageDownload),
		Retry:  0,
	}

	// untar := &task.LocalTask{
	// 	Name:   "Decompress",
	// 	Desc:   "Decompress installer package",
	// 	Action: new(PackageUntar),
	// 	Retry:  0,
	// }

	m.Tasks = []task.Interface{
		download,
		//untar,
	}
}
