package precheck

import (
	"bytetrade.io/web3os/installer/pkg/core/ending"
	"bytetrade.io/web3os/installer/pkg/core/module"
)

// import (
// 	"fmt"

// 	"bytetrade.io/web3os/installer/pkg/constants"
// 	"bytetrade.io/web3os/installer/pkg/core/ending"
// 	"bytetrade.io/web3os/installer/pkg/core/logger"
// 	"bytetrade.io/web3os/installer/pkg/core/module"
// 	"bytetrade.io/web3os/installer/pkg/core/util"
// 	"bytetrade.io/web3os/installer/pkg/utils"
// )

type PrintMachineInfoHook struct {
	Module module.Module
	Result *ending.ModuleResult
}

func (h *PrintMachineInfoHook) Init(module module.Module, result *ending.ModuleResult) {
	h.Module = module
	h.Result = result
}

func (h *PrintMachineInfoHook) Try() error {
	// fmt.Printf("MACHINE, hostname: %s, cpu: %d, mem: %s, disk: %s, local-ip: %s\n",
	// 	constants.HostName, constants.CpuPhysicalCount, utils.FormatBytes(int64(constants.MemTotal)),
	// 	utils.FormatBytes(int64(constants.DiskTotal)), constants.LocalIp)
	// fmt.Printf("OS, info: %s\n", constants.OsInfo)
	// fmt.Printf("SYSTEM, os: %s, platform: %s, arch: %s, version: %s\nKERNEL: version: %s\n", constants.OsType, constants.OsPlatform, constants.OsArch, constants.OsVersion, constants.OsKernel)
	// fmt.Printf("FS, type: %s, zfsmount: %s\n", constants.FsType, constants.DefaultZfsPrefixName)
	// fmt.Printf("VIRTUAL, role: %s, system: %s\n", constants.VirtualizationRole, constants.VirtualizationSystem)
	// fmt.Printf("CGROUP, cpu-enabled: %d, memory-enabled: %d\n", constants.CgroupCpuEnabled, constants.CgroupMemoryEnabled)

	// if constants.InstalledKubeVersion != "" {
	// 	fmt.Printf("KUBE, version: %s\n", constants.InstalledKubeVersion)
	// }

	return nil
}

func (h *PrintMachineInfoHook) Catch(err error) error {
	return err
}

func (h *PrintMachineInfoHook) Finally() {
}

// // ~ hook GetSysInfoHook
// type GetSysInfoHook struct {
// 	Module module.Module
// 	Result *ending.ModuleResult
// }

// func (h *GetSysInfoHook) Init(module module.Module, result *ending.ModuleResult) {
// 	h.Module = module
// 	h.Result = result
// }

// func (h *GetSysInfoHook) Try() error {
// 	host, err := util.GetHost()
// 	if err != nil {
// 		return err
// 	}
// 	constants.HostName = host[0]
// 	constants.HostId = host[1]
// 	constants.OsType = host[2]
// 	constants.OsPlatform = host[3]
// 	constants.OsVersion = host[4]
// 	constants.OsArch = host[5]

// 	cpuModel, cpuLogicalCount, cpuPhysicalCount, err := util.GetCpu()
// 	if err != nil {
// 		return err
// 	}
// 	constants.CpuModel = cpuModel
// 	constants.CpuLogicalCount = cpuLogicalCount
// 	constants.CpuPhysicalCount = cpuPhysicalCount

// 	diskTotal, diskFree, err := util.GetDisk()
// 	if err != nil {
// 		return err
// 	}
// 	constants.DiskTotal = diskTotal
// 	constants.DiskFree = diskFree

// 	memTotal, memFree, err := util.GetMem()
// 	if err != nil {
// 		return err
// 	}
// 	constants.MemTotal = memTotal
// 	constants.MemFree = memFree

// 	logger.Debugf("[hook] GetSysInfoHook, hostname: %s, cpu: %d, mem: %d, disk: %d",
// 		constants.HostName, constants.CpuPhysicalCount, constants.MemTotal, constants.DiskTotal)

// 	logger.Infof("host info, hostname: %s, hostid: %s, os: %s, platform: %s, version: %s, arch: %s",
// 		constants.HostName, constants.HostId, constants.OsType, constants.OsPlatform, constants.OsVersion, constants.OsArch)
// 	logger.Infof("cpu info, model: %s, logical count: %d, physical count: %d",
// 		constants.CpuModel, constants.CpuLogicalCount, constants.CpuPhysicalCount)
// 	logger.Infof("disk info, total: %d, free: %d", constants.DiskTotal, constants.DiskFree)
// 	logger.Infof("mem info, total: %d, free: %d", constants.MemTotal, constants.MemFree)

// 	return nil

// }

// func (h *GetSysInfoHook) Catch(err error) error {
// 	return err
// }

// func (h *GetSysInfoHook) Finally() {
// }

// // ~ hook GetLocalIpHook
// type GetLocalIpHook struct {
// 	Module module.Module
// 	Result *ending.ModuleResult
// }

// func (h *GetLocalIpHook) Init(module module.Module, result *ending.ModuleResult) {
// 	h.Module = module
// 	h.Result = result
// }

// func (h *GetLocalIpHook) Try() error {
// 	pingCmd := fmt.Sprintf("ping -c 1 %s", constants.HostName)
// 	pingCmdRes, _, err := util.Exec(pingCmd, true, false)
// 	if err != nil {
// 		return err
// 	}

// 	pingIps, err := utils.ExtractIP(pingCmdRes)
// 	if err != nil {
// 		return err
// 	}

// 	logger.Debugf("[hook] GetLocalIpHook, local ip: %s", pingIps)
// 	constants.LocalIp = pingIps

// 	return nil
// }

// func (h *GetLocalIpHook) Catch(err error) error {
// 	return err
// }

// func (h *GetLocalIpHook) Finally() {

// }
