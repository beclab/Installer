package constants

var Logo = `
 _____                    _                 
/__   \___ _ __ _ __ ___ (_)_ __  _   _ ___ 
  / /\/ _ \ '__| '_ ` + "`" + ` _ \| | '_ \| | | / __|
 / / |  __/ |  | | | | | | | | | | |_| \__ \
 \/   \___|_|  |_| |_| |_|_|_| |_|\__,_|___/
                                            
`

var (
	HostName             string
	HostId               string
	OsType               string
	OsPlatform           string
	OsVersion            string
	OsArch               string
	OsKernel             string
	VirtualizationRole   string
	VirtualizationSystem string
	CpuModel             string
	CpuLogicalCount      int
	CpuPhysicalCount     int
	MemTotal             uint64
	MemFree              uint64
	DiskTotal            uint64
	DiskFree             uint64

	CgroupCpuEnabled    int
	CgroupMemoryEnabled int

	CloudVendor string

	LocalIp  string
	PublicIp []string

	InstalledKubeVersion string
)

var (
	PkgManager string
)

var (
	WorkDir                string
	ApiServerListenAddress string
	Proxy                  string
	CurrentUser            string
)
