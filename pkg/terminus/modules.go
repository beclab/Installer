package terminus

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/etcd"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/utils"
	"time"
)

type InstallWizardDownloadModule struct {
	common.KubeModule
	Version        string
	DownloadCdnUrl string
}

func (m *InstallWizardDownloadModule) Init() {
	m.Name = "DownloadInstallWizard"
	download := &task.LocalTask{
		Name: "DownloadInstallWizard",
		Action: &Download{
			Version:        m.Version,
			DownloadCdnUrl: m.DownloadCdnUrl,
		},
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		download,
	}
}

type PreparedModule struct {
	common.KubeModule
}

func (m *PreparedModule) Init() {
	m.Name = "PrepareFinished"

	prepareFinished := &task.LocalTask{
		Name:   "PrepareFinished",
		Action: &PrepareFinished{},
	}

	m.Tasks = []task.Interface{
		prepareFinished,
	}
}

type CheckPreparedModule struct {
	common.KubeModule
	Force bool
}

func (m *CheckPreparedModule) Init() {
	m.Name = "CheckPrepared"

	checkPrepared := &task.LocalTask{
		Name:   "CheckPrepared",
		Action: &CheckPrepared{Force: m.Force},
	}

	m.Tasks = []task.Interface{
		checkPrepared,
	}
}

type CheckInstalledModule struct {
	common.KubeModule
	Force bool
}

func (m *CheckInstalledModule) Init() {
	m.Name = "CheckInstalled"

	checkPrepared := &task.LocalTask{
		Name:   "CheckInstalled",
		Action: &CheckInstalled{Force: m.Force},
	}

	m.Tasks = []task.Interface{
		checkPrepared,
	}
}

type OlaresUninstallScriptModule struct {
	common.KubeModule
}

func (m *OlaresUninstallScriptModule) Init() {
	m.Name = "GenerateOlaresUninstallScript"

	uninstallScript := &task.LocalTask{
		Name:   "GenerateOlaresUninstallScript",
		Action: &GenerateOlaresUninstallScript{},
	}

	m.Tasks = []task.Interface{
		uninstallScript,
	}
}

type InstallComponentsInClusterModule struct {
	common.KubeModule
}

type GetNATGatewayIPModule struct {
	common.KubeModule
}

func (m *GetNATGatewayIPModule) Init() {
	m.Name = "GetNATGatewayIP"

	getNATGatewayIP := &task.LocalTask{
		Name:   "GetNATGatewayIP",
		Action: new(GetNATGatewayIP),
	}

	m.Tasks = []task.Interface{
		getNATGatewayIP,
	}
}

func GenerateTerminusComponentsModules(runtime connector.Runtime, manifestMap manifest.InstallationManifest) []module.Module {
	var modules []module.Module
	baseModules := []module.Module{
		&GetNATGatewayIPModule{},
		&InstallAccountModule{},
		&InstallSettingsModule{},
		&InstallOsSystemModule{},
		&InstallLauncherModule{},
		&InstallAppsModule{},
	}
	modules = append(modules, baseModules...)

	si := runtime.GetSystemInfo()
	switch {
	case si.IsDarwin():
	default:
		modules = append(
			modules,
			&InstallVeleroModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.GetBaseDir(),
				},
			})
	}

	modules = append(modules, &WelcomeModule{})

	return modules
}

type InstalledModule struct {
	common.KubeModule
}

func (m *InstalledModule) Init() {
	m.Name = "InstallFinished"

	installedState := &task.LocalTask{
		Name:   "InstallFinished",
		Action: &InstallFinished{},
	}

	m.Tasks = []task.Interface{
		installedState,
	}
}

type DeleteWizardFilesModule struct {
	common.KubeModule
}

func (d *DeleteWizardFilesModule) Init() {
	d.Name = "DeleteWizardFiles"

	deleteWizardFiles := &task.LocalTask{
		Name:   "DeleteWizardFiles",
		Action: &DeleteWizardFiles{},
	}

	d.Tasks = []task.Interface{
		deleteWizardFiles,
	}
}

type ChangeIPModule struct {
	common.KubeModule
}

func (m *ChangeIPModule) Init() {
	m.Name = "ChangeIP"

	prepared, _ := m.PipelineCache.GetMustBool(common.CachePreparedState)
	if !prepared {
		logger.Info("the Olares OS is not prepared, will only try to update /etc/hosts")
	}
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "UpdateHosts",
			Action: new(UpdateKubeKeyHosts),
		},
	}

	installed, _ := m.PipelineCache.GetMustBool(common.CacheInstalledState)
	if !installed && prepared {
		logger.Info("the Olares OS is not installed, will only try to update prepared base components")
	}

	if installed {
		stopKubeTask := &task.LocalTask{
			Name:  "StopKubernetes",
			Retry: 3,
		}
		stopKubeAction := &SystemctlCommand{
			Command: "stop",
		}
		if m.KubeConf.Arg.Kubetype == common.K3s {
			stopKubeAction.UnitNames = []string{"k3s"}
		} else {
			stopKubeAction.UnitNames = []string{"kubelet"}
		}
		if util.IsExist("/etc/systemd/system/etcd.service") {
			stopKubeAction.UnitNames = append(stopKubeAction.UnitNames, "etcd", "backup-etcd")
		}
		stopKubeTask.Action = stopKubeAction
		m.Tasks = append(m.Tasks, stopKubeTask)
	}

	if prepared {
		m.addStorageTasks()
	}
	if installed {
		m.addEtcdTasks()
		m.addKubernetesTasks()
		m.addRestartTasks()
	}
	if m.KubeConf.Arg.MasterHost != "" {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "SaveMasterHostConfig",
				Action: new(SaveMasterHostConfig),
			})
	}
}

func (m *ChangeIPModule) addStorageTasks() {
	var storageComponents []string
	juiceFSExists := util.IsExist(storage.JuiceFsServiceFile)
	if juiceFSExists {
		storageComponents = append(storageComponents, "juicefs")
	} else {
		logger.Info("JuiceFS is not installed, no storage component needs IP reconfiguration.")
		return
	}
	redisExists := util.IsExist(storage.RedisServiceFile)
	if redisExists {
		storageComponents = append(storageComponents, "redis-server")
	}
	minioExists := util.IsExist(storage.MinioServiceFile)
	if minioExists {
		storageComponents = append(storageComponents, "minio")
	}
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name: "StopStorageComponents",
			Action: &SystemctlCommand{
				Command:   "stop",
				UnitNames: storageComponents,
			},
			Retry: 3,
		},

		&task.RemoteTask{
			Name:   "GetRedisConfig",
			Hosts:  m.Runtime.GetHostsByRole(common.Master),
			Action: new(storage.GetOrSetRedisConfig),
			Retry:  1,
		},
	)
	if redisExists {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "ReGenerateRedisService",
				Action: new(storage.GenerateRedisService),
			},
			&task.LocalTask{
				Name:   "EnableRedisService",
				Action: new(storage.EnableRedisService),
				Retry:  3,
			},
			&task.LocalTask{
				Name:   "CheckRedisState",
				Action: new(storage.CheckRedisServiceState),
				Retry:  20,
			},
		)
	}

	if minioExists {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "GetOrSetMinIOPassword",
				Action: new(storage.GetOrSetMinIOPassword),
			},
			&task.LocalTask{
				Name:   "ReConfigureMinio",
				Action: new(storage.ConfigMinio),
			},
			&task.LocalTask{
				Name:   "EnableMinioService",
				Action: new(storage.EnableMinio),
			},
			&task.LocalTask{
				Name:   "CheckMinioState",
				Action: new(storage.CheckMinioState),
				Retry:  20,
			},
			&task.LocalTask{
				Name:   "ConfigJuiceFSMetaDB",
				Action: new(storage.ConfigJuiceFsMetaDB),
			},
		)
	}

	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "EnableJuiceFsService",
			Action: new(storage.EnableJuiceFsService),
		},

		&task.LocalTask{
			Name:   "CheckJuiceFsState",
			Action: new(storage.CheckJuiceFsState),
			Retry:  20,
		},
	)
}

func (m *ChangeIPModule) addEtcdTasks() {
	if !util.IsExist("/etc/systemd/system/etcd.service") {
		return
	}
	m.Tasks = append(m.Tasks,
		&task.RemoteTask{
			Name:   "GetETCDStatus",
			Action: new(etcd.GetStatus),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "GenerateETCDAccessAddress",
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
			Action: new(etcd.GenerateAccessAddress),
		},
		&task.LocalTask{
			Name:   "PrepareETCDFiles",
			Action: new(PrepareFilesForETCDIPChange),
		},
		&task.LocalTask{
			Name:   "RegenerateETCDCerts",
			Action: new(etcd.GenerateCerts),
		},
		&task.RemoteTask{
			Name:   "SyncETCDCerts",
			Action: new(etcd.SyncCertsFile),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "RefreshETCDConfig",
			Action: new(etcd.RefreshConfig),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "RestartETCD",
			Action: new(etcd.RestartETCD),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "ETCDHealthCheck",
			Action: new(etcd.HealthCheck),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
			Retry:  20,
		},
		&task.RemoteTask{
			Name:   "RefreshBackupETCDScript",
			Action: new(etcd.BackupETCD),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		})
}

func (m *ChangeIPModule) addKubernetesTasks() {
	if m.KubeConf.Arg.Kubetype == common.K3s {
		cluster := k3s.NewK3sStatus()
		m.PipelineCache.GetOrSet(common.ClusterStatus, cluster)
		if !m.Runtime.GetLocalHost().IsRole(common.Master) {
			m.Tasks = append(m.Tasks,
				&task.RemoteTask{
					Name:     "GetClusterStatus(k3s)",
					Desc:     "Get k3s cluster status",
					Hosts:    m.Runtime.GetHostsByRole(common.Master),
					Action:   new(k3s.GetClusterStatus),
					Parallel: false,
				},
			)
		}
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "RegenerateK3sService",
				Action: new(k3s.GenerateK3sService),
			},
			&task.LocalTask{
				Name:   "RegenerateK3sServiceEnv",
				Action: new(k3s.GenerateK3sServiceEnv),
			},
			&task.LocalTask{
				Name:   "EnableK3sService",
				Desc:   "Enable k3s service",
				Action: new(k3s.EnableK3sService),
			},
		)
	} else {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "RegenerateKubeletServiceEnv",
				Action: new(kubernetes.GenerateKubeletEnv),
			})
		// worker node, no need to reconfigure control-plane components
		if !util.IsExist("/etc/kubernetes/manifests/kube-apiserver.yaml") {
			m.Tasks = append(m.Tasks,
				&task.LocalTask{
					Name: "RestartKubelet",
					Action: &SystemctlCommand{
						Command:             "start",
						UnitNames:           []string{"kubelet"},
						DaemonReloadPreExec: true,
					},
				})
			return
		}
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "PrepareK8sFiles",
				Action: new(PrepareFilesForK8sIPChange),
			},
			&task.RemoteTask{
				Name: "RegenerateKubeadmConfig",
				Action: &kubernetes.GenerateKubeadmConfig{
					IsInitConfiguration:     true,
					WithSecurityEnhancement: m.KubeConf.Arg.SecurityEnhancement,
				},
				Hosts: m.Runtime.GetHostsByRole(common.Master),
			},
			&task.LocalTask{
				Name:   "RegenerateK8sFilesWithKubeadm",
				Action: new(RegenerateFilesForK8sIPChange),
			},
			&task.LocalTask{
				Name:   "CopyNewKubeConfig",
				Action: new(kubernetes.CopyKubeConfigForControlPlane),
			},
		)
	}
}

func (m *ChangeIPModule) addRestartTasks() {
	restartPodsTasks := []task.Interface{
		&task.LocalTask{
			Name:   "RestartAllPods",
			Action: &DeleteAllPods{Node: m.Runtime.GetLocalHost().GetName()},
			Retry:  5,
			Delay:  15 * time.Second,
		},
	}
	if !utils.IsExist(storage.JuiceFsServiceFile) {
		restartPodsTasks = []task.Interface{
			&task.LocalTask{
				Name:   "RestartPodsUsingHostIP",
				Action: new(DeletePodsUsingHostIP),
				Delay:  3 * time.Second,
				Retry:  5,
			},
			// the app controllers in k3s have significant delay detecting state change
			// and there's a gap between a pod being deleted and recreated
			&task.LocalTask{
				Name:   "WaitForPodsUsingHostIPRecreate",
				Action: new(WaitForPodsUsingHostIPRecreate),
				Delay:  15 * time.Second,
				Retry:  20,
			},
		}
	}
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "WaitForKubeAPIServerUp",
			Action: new(precheck.GetKubernetesNodesStatus),
			Retry:  20,
		})
	m.Tasks = append(m.Tasks, restartPodsTasks...)

	m.Tasks = append(m.Tasks, &task.LocalTask{
		Name:   "EnsurePodsUpAndRunningAgain",
		Action: &CheckKeyPodsRunning{Node: m.Runtime.GetLocalHost().GetName()},
		Delay:  10 * time.Second,
		Retry:  60,
	})
}

type ChangeHostIPModule struct {
	common.KubeModule
}

func (m *ChangeHostIPModule) Init() {
	m.Name = "ChangeHostIP"

	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "CheckOlaresStateInHost",
			Action: new(CheckTerminusStateInHost),
		},
		&task.LocalTask{
			Name:   "GetNATGatewayIP",
			Action: new(GetNATGatewayIP),
		},
		&task.LocalTask{
			Name:   "UpdateNATGatewayIPForUser",
			Action: new(UpdateNATGatewayForUser),
		},
	)
}

type GetMasterInfoModule struct {
	common.KubeModule
	Print bool
}

func (m *GetMasterInfoModule) Init() {
	m.Name = "GetMasterInfo"
	m.Tasks = append(m.Tasks,
		&task.RemoteTask{
			Name:   "GetMasterInfo",
			Action: &GetMasterInfo{Print: m.Print},
			Hosts:  m.Runtime.GetHostsByRole(common.Master),
		},
		&task.LocalTask{
			Name:   "AddNodePrecheck",
			Action: new(AddNodePrecheck),
		},
	)
}

type SaveMasterHostConfigModule struct {
	common.KubeModule
}

func (m *SaveMasterHostConfigModule) Init() {
	m.Name = "SaveMasterHostConfig"
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "SaveMasterHostConfig",
			Action: new(SaveMasterHostConfig),
		})
}
