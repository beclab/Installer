package download

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type PackageDownloadModule struct {
	common.KubeModule
	Manifest string
	BaseDir  string
}

func (i *PackageDownloadModule) Init() {
	i.Name = "PackageDownloadModule"
	i.Desc = "Download terminus installation package"

	download := &task.LocalTask{
		Name:   i.Name,
		Desc:   i.Desc,
		Action: &PackageDownload{Manifest: i.Manifest, BaseDir: i.BaseDir},
	}

	i.Tasks = []task.Interface{
		download,
	}
}

type CheckDownloadModule struct {
	common.KubeModule
	Manifest string
	BaseDir  string
}

func (i *CheckDownloadModule) Init() {
	i.Name = "CheckDownloadModule"
	i.Desc = "Check downloaded terminus installation package"

	check := &task.LocalTask{
		Name:   i.Name,
		Desc:   i.Desc,
		Action: &CheckDownload{Manifest: i.Manifest, BaseDir: i.BaseDir},
	}

	i.Tasks = []task.Interface{
		check,
	}
}
