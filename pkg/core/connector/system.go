package connector

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/util"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type UbuntuVersion string
type DebianVersion string

func (u UbuntuVersion) String() string {
	switch u {
	case UbuntuAbove20:
		return "20."
	case UbuntuAbove22:
		return "22."
	case UbuntuAbove24:
		return "24."
	}
	return ""
}

func (d DebianVersion) String() string {
	switch d {
	case DebianAbove11:
		return "11"
	case DebianAbove12:
		return "12"
	}
	return ""
}

const (
	UbuntuAbove20 UbuntuVersion = "20."
	UbuntuAbove22 UbuntuVersion = "22."
	UbuntuAbove24 UbuntuVersion = "24."

	DebianAbove11 DebianVersion = "11"
	DebianAbove12 DebianVersion = "12"
)

type Systems interface {
	IsSupport() error

	IsDarwin() bool
	IsWsl() bool
	IsPve() bool
	IsRaspbian() bool
	IsLinux() bool

	IsUbuntu() bool
	IsDebian() bool

	IsUbuntuVersionAbove(ver UbuntuVersion) bool
	IsDebianVersionAbove(ver UbuntuVersion) bool

	SetHostname(v string)
	GetHostname() string
	GetOsType() string
	GetOsArch() string
	GetOsVersion() string
	GetPkgManager() string

	GetOsPlatformFamily() string

	GetLocalIp() string

	CgroupCpuEnabled() bool
	CgroupMemoryEnabled() bool
	GetFsType() string
	GetDefaultZfsPrefixName() string

	String() string
}

type SystemInfo struct {
	HostInfo   *HostInfo       `json:"host"`
	CpuInfo    *CpuInfo        `json:"cpu"`
	DiskInfo   *DiskInfo       `json:"disk"`
	MemoryInfo *MemoryInfo     `json:"memory"`
	FsInfo     *FileSystemInfo `json:"filesystem"`
	CgroupInfo *CgroupInfo     `json:"cgroup,omitempty"`
	LocalIp    string          `json:"local_ip"`
	PkgManager string          `json:"pkg_manager"`
}

func (s *SystemInfo) IsSupport() error {
	if !s.IsLinux() && !s.IsDarwin() {
		return fmt.Errorf("unsupported os type '%s', exit ...", s.GetOsType())
	}

	if s.GetOsArch() == "" {
		return fmt.Errorf("unsupported arch '%s', exit ...", s.GetOsArch())
	}

	return nil
}

func (s *SystemInfo) GetLocalIp() string {
	return s.LocalIp
}

func (s *SystemInfo) GetHostname() string {
	return s.HostInfo.HostName
}

func (s *SystemInfo) SetHostname(v string) {
	s.HostInfo.HostName = v
}

func (s *SystemInfo) GetOsType() string {
	return s.HostInfo.OsType
}

func (s *SystemInfo) GetOsArch() string {
	return s.HostInfo.OsArch
}

func (s *SystemInfo) GetOsVersion() string {
	return s.HostInfo.OsVersion
}

func (s *SystemInfo) GetOsPlatformFamily() string {
	return s.HostInfo.OsPlatformFamily
}

func (s *SystemInfo) String() string {
	str, _ := json.Marshal(s)
	return string(str)
}

func (s *SystemInfo) IsDarwin() bool {
	return s.HostInfo.OsPlatform == common.Darwin
}

func (s *SystemInfo) IsPve() bool {
	return strings.Contains(s.HostInfo.OsKernel, "-pve")
}

func (s *SystemInfo) IsWsl() bool {
	return s.HostInfo.OsPlatform == common.WSL
}

func (s *SystemInfo) IsRaspbian() bool {
	return s.HostInfo.OsPlatform == common.Raspbian
}

func (s *SystemInfo) IsLinux() bool {
	return s.HostInfo.OsPlatform == common.Linux
}

func (s *SystemInfo) IsUbuntu() bool {
	return s.HostInfo.OsPlatformFamily == common.Ubuntu
}

func (s *SystemInfo) IsDebian() bool {
	return s.HostInfo.OsPlatformFamily == common.Debian
}

func (s *SystemInfo) IsUbuntuVersionAbove(ver UbuntuVersion) bool {
	return strings.Contains(s.HostInfo.OsVersion, ver.String())
}

func (s *SystemInfo) IsDebianVersionAbove(ver UbuntuVersion) bool {
	return strings.Contains(s.HostInfo.OsVersion, ver.String())
}

func (s *SystemInfo) CgroupCpuEnabled() bool {
	return s.CgroupInfo.CpuEnabled >= 1
}

func (s *SystemInfo) CgroupMemoryEnabled() bool {
	return s.CgroupInfo.MemoryEnabled >= 1
}

func (s *SystemInfo) GetFsType() string {
	return s.FsInfo.Type
}

func (s *SystemInfo) GetDefaultZfsPrefixName() string {
	return s.FsInfo.DefaultZfsPrefixName
}

func (s *SystemInfo) GetPkgManager() string {
	return s.PkgManager
}

func GetSystemInfo() *SystemInfo {
	var si = new(SystemInfo)
	si.HostInfo = getHost()
	si.CpuInfo = getCpu()
	si.DiskInfo = getDisk()
	si.MemoryInfo = getMem()
	si.FsInfo = getFs()
	si.LocalIp = util.LocalIP()

	if si.IsLinux() {
		si.CgroupInfo = getCGroups()
	}

	switch si.GetOsPlatformFamily() {
	case common.Ubuntu, common.Debian:
		si.PkgManager = "apt-get"
	case common.Fedora:
		si.PkgManager = "dnf"
	case common.CentOs, common.RHEL:
		si.PkgManager = "yum"
	default:
		si.PkgManager = "apt-get"
	}

	return si
}

type HostInfo struct {
	HostName             string `json:"hostname"`
	HostId               string `json:"hostid"`
	OsType               string `json:"os_type"`
	OsPlatform           string `json:"os_platform"`
	OsPlatformFamily     string `json:"os_platform_family"`
	OsVersion            string `json:"os_version"`
	OsArch               string `json:"os_arch"`
	VirtualizationRole   string `json:"virtualization_role"`
	VirtualizationSystem string `json:"virtualization_system"`
	OsKernel             string `json:"os_kernel"`
	OsInfo               string `json:"os_info"`
}

func getHost() *HostInfo {
	hostInfo, err := host.Info()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("sh", "-c", "echo $(uname -a) |tr -d '\\n'")
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var _osType = hostInfo.OS
	var _osPlatform = hostInfo.Platform
	var _osPlatformFamily = hostInfo.PlatformFamily
	var _osKernel = hostInfo.KernelVersion

	return &HostInfo{
		HostName:             hostInfo.Hostname,
		HostId:               hostInfo.HostID,
		OsType:               _osType,                                           // darwin linux
		OsPlatform:           formatOsPlatform(_osType, _osPlatform, _osKernel), // darwin linux wsl raspbian pve
		OsPlatformFamily:     formatOsPlatformFamily(_osPlatform, _osPlatformFamily),
		OsVersion:            hostInfo.PlatformVersion,
		OsArch:               ArchAlias(hostInfo.KernelArch),
		VirtualizationRole:   hostInfo.VirtualizationRole,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		OsKernel:             hostInfo.KernelVersion,
		OsInfo:               string(output),
	}
}

func formatOsPlatform(osType, osPlatform, osKernel string) string {
	if osType == common.Darwin {
		return common.Darwin
	}

	if osPlatform == common.Raspbian {
		return common.Raspbian
	}

	if strings.Contains(osKernel, "pve") {
		return common.PVE
	}

	if strings.Contains(osKernel, "-WSL") {
		return common.WSL
	}

	return common.Linux
}

func formatOsPlatformFamily(osPlatform, osPlatformFamily string) string {
	if osPlatform == common.Darwin {
		return common.Darwin
	}

	if osPlatform == common.Raspbian {
		return osPlatformFamily
	}

	return osPlatform
}

type CpuInfo struct {
	CpuModel         string `json:"cpu_model"`
	CpuLogicalCount  int    `json:"cpu_logical_count"`
	CpuPhysicalCount int    `json:"cpu_physical_count"`
}

func getCpu() *CpuInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cpuInfo, _ := cpu.InfoWithContext(ctx)
	cpuLogicalCount, _ := cpu.CountsWithContext(ctx, true)
	cpuPhysicalCount, _ := cpu.CountsWithContext(ctx, false)

	var cpuModel = ""
	if cpuInfo != nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return &CpuInfo{
		CpuModel:         cpuModel,
		CpuLogicalCount:  cpuLogicalCount,
		CpuPhysicalCount: cpuPhysicalCount,
	}
}

type DiskInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func getDisk() *DiskInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usageInfo, _ := disk.UsageWithContext(ctx, "/")

	var total uint64
	var free uint64
	if usageInfo != nil {
		total = usageInfo.Total
		free = usageInfo.Free
	}

	return &DiskInfo{
		Total: total,
		Free:  free,
	}
}

type MemoryInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func getMem() *MemoryInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	memInfo, _ := mem.VirtualMemoryWithContext(ctx)

	var total uint64
	var free uint64
	if memInfo != nil {
		total = memInfo.Total
		free = memInfo.Free
	}

	return &MemoryInfo{
		Total: total,
		Free:  free,
	}
}

type FileSystemInfo struct {
	Type                 string `json:"type"`
	DefaultZfsPrefixName string `json:"default_zfs_prefix_name"`
}

func getFs() *FileSystemInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var fsType = "overlayfs"
	var zfsPrefixName = ""

	ps, _ := disk.PartitionsWithContext(ctx, true)
	if ps != nil && len(ps) > 0 {
		for _, p := range ps {
			if p.Mountpoint == "/var/lib" && p.Fstype == "zfs" {
				fsType = "zfs"
				zfsPrefixName = p.Device
				break
			}
		}
	}

	return &FileSystemInfo{
		Type:                 fsType,
		DefaultZfsPrefixName: zfsPrefixName,
	}
}

type CgroupInfo struct {
	CpuEnabled    int `json:"cpu_enabled"`
	MemoryEnabled int `json:"memory_enabled"`
}

func getCGroups() *CgroupInfo {

	file, err := os.Open("/proc/cgroups")
	if err != nil {
		return nil
	}
	defer file.Close()

	var cpuEnabled int64
	var memoryEnabled int64

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
			cpuEnabled, _ = strconv.ParseInt(fields[3], 10, 64)
		case "memory":
			memoryEnabled, _ = strconv.ParseInt(fields[3], 10, 64)
		default:
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return &CgroupInfo{
		CpuEnabled:    int(cpuEnabled),
		MemoryEnabled: int(memoryEnabled),
	}
}

func ArchAlias(arch string) string {
	switch arch {
	case "aarch64", "armv7l", "arm64", "arm":
		return "arm64"
	case "x86_64", "amd64":
		fallthrough
	case "ppc64le":
		fallthrough
	case "s390x":
		return "amd64"
	default:
		return ""
	}
}
