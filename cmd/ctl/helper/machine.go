package helper

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
)

func GetMachineInfo() {
	getWorkDir()
	getHost()
	getCpu()
	getDisk()
	getMem()
	getRepoManager()
	getCGroups()
	getLocalIp()
}

func getWorkDir() {
	// workDir, err := utils.WorkDir()
	workDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("working path error", err)
		os.Exit(1)
	}

	constants.WorkDir = workDir
}

func getHost() {
	host, err := util.GetHost()
	if err != nil {
		panic(err)
	}
	constants.HostName = host[0]
	constants.HostId = host[1]
	constants.OsType = host[2]
	constants.OsPlatform = host[3]
	constants.OsVersion = host[4]
	constants.OsArch = utils.ArchAlias(host[5]) //host[5]
	constants.VirtualizationRole = host[6]
	constants.VirtualizationSystem = host[7]
}

func getCpu() {
	cpuModel, cpuLogicalCount, cpuPhysicalCount, err := util.GetCpu()
	if err != nil {
		panic(err)
	}
	constants.CpuModel = cpuModel
	constants.CpuLogicalCount = cpuLogicalCount
	constants.CpuPhysicalCount = cpuPhysicalCount
}

func getDisk() {
	diskTotal, diskFree, err := util.GetDisk()
	if err != nil {
		panic(err)
	}
	constants.DiskTotal = diskTotal
	constants.DiskFree = diskFree
}

func getMem() {
	memTotal, memFree, err := util.GetMem()
	if err != nil {
		panic(err)
	}
	constants.MemTotal = memTotal
	constants.MemFree = memFree
}

func getRepoManager() {
	switch constants.OsPlatform {
	case common.Ubuntu, common.Debian, common.Raspbian:
		constants.PkgManager = "apt-get"
	case common.Fedora:
		constants.PkgManager = "dnf"
	case common.CentOs, common.RHEl:
		constants.PkgManager = "yum"
	default:
		constants.PkgManager = "apt-get"
	}
}

func getCGroups() {
	if constants.OsType == common.Darwin || constants.OsType == common.Windows {
		return
	}

	file, err := os.Open("/proc/cgroups")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		switch fields[0] {
		case "cpu":
			cpuEnabled, _ := strconv.ParseInt(fields[3], 10, 64)
			constants.CgroupCpuEnabled = int(cpuEnabled)
		case "memory":
			memoryEnabled, _ := strconv.ParseInt(fields[3], 10, 64)
			constants.CgroupMemoryEnabled = int(memoryEnabled)
		default:
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func getLocalIp() {
	constants.LocalIp = util.LocalIP()
}
