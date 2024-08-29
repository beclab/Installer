package artifact

// func InitPhase(args common.Argument, runtime *common.KubeRuntime) *pipeline.Pipeline {
// 	noArtifact := runtime.Arg.Artifact == ""
// 	skipPushImages := runtime.Arg.SKipPushImages || noArtifact || (!noArtifact && runtime.Cluster.Registry.PrivateRegistry == "")
// 	skipLocalStorage := true
// 	if runtime.Arg.DeployLocalStorage != nil {
// 		skipLocalStorage = !*runtime.Arg.DeployLocalStorage
// 	} else if runtime.Cluster.KubeSphere.Enabled {
// 		skipLocalStorage = false
// 	}

// 	m := []module.Module{
// 		&precheck.GreetingsModule{},
// 		&precheck.GetSysInfoModel{},
// 		&plugins.CopyEmbed{},
// 		//
// 		&artifact.UnArchiveModule{Skip: noArtifact},                            // skip
// 		&os.RepositoryModule{Skip: noArtifact || !runtime.Arg.InstallPackages}, // skip
// 		&binaries.K3sNodeBinariesModule{},
// 		&os.ConfigureOSModule{},
// 		&k3s.StatusModule{},
// 		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
// 		&etcd.CertsModule{},
// 		&etcd.InstallETCDBinaryModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
// 		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
// 		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
// 		&loadbalancer.K3sKubevipModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
// 		&k3s.InstallKubeBinariesModule{},
// 		&k3s.InitClusterModule{}, // +
// 		&k3s.StatusModule{},
// 		&k3s.JoinNodesModule{},
// 		&images.CopyImagesToRegistryModule{Skip: skipPushImages},
// 		&loadbalancer.K3sHaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
// 		&network.DeployNetworkPluginModule{},
// 		&kubernetes.ConfigureKubernetesModule{},
// 		&filesystem.ChownModule{},
// 		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
// 		&k3s.SaveKubeConfigModule{},
// 		&addons.AddonsModule{}, // relative ks-installer
// 		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
// 		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
// 	}

// 	return &pipeline.Pipeline{
// 		Name:    "Preload Images",
// 		Modules: m,
// 		Runtime: runtime,
// 	}
// }
