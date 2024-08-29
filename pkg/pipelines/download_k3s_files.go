package pipelines

func PreloadImages(kubeType string, registryMirrors string) error {
	// todo
	// var ksVersion, err = getNodeVersion(kubeType, false)
	// if err != nil {
	// 	return err
	// }

	// arg := common.Argument{
	// 	KsEnable:          true,
	// 	KsVersion:         common.DefaultKubeSphereVersion,
	// 	InstallPackages:   false,
	// 	SKipPushImages:    false,
	// 	ContainerManager:  common.Containerd,
	// 	KubernetesVersion: ksVersion,
	// 	RegistryMirrors:   registryMirrors,
	// }

	// runtime, err := common.NewKubeRuntime(common.AllInOne, arg)
	// if err != nil {
	// 	return err
	// }

	// p := pa.InitPhase(arg, runtime)

	// if err := p.Start(); err != nil {
	// 	logger.Errorf("preload images failed %v", err)
	// 	return err
	// }

	// if runtime.Arg.InCluster {
	// 	if err := ctrl.UpdateStatus(runtime); err != nil {
	// 		logger.Errorf("failed to update status: %v", err)
	// 		return err
	// 	}
	// 	kkConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.ObjName))
	// 	if config, err := ioutil.ReadFile(kkConfigPath); err != nil {
	// 		logger.Errorf("failed to read kubeconfig: %v", err)
	// 		return err
	// 	} else {
	// 		runtime.Kubeconfig = base64.StdEncoding.EncodeToString(config)
	// 		if err := ctrl.UpdateKubeSphereCluster(runtime); err != nil {
	// 			logger.Errorf("failed to update kubesphere cluster: %v", err)
	// 			return err
	// 		}
	// 		if err := ctrl.SaveKubeConfig(runtime); err != nil {
	// 			logger.Errorf("failed to save kubeconfig: %v", err)
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}
