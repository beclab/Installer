package pipelines

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(opts *options.CliPrepareSystemOptions) error {
	// ksVersion, _, exists := kubernetes.CheckKubeExists()
	// if exists {
	// 	return fmt.Errorf("Kubernetes %s is already installed", ksVersion)
	// }
	var terminusVersion, _ = phase.GetTerminusVersion()
	if terminusVersion != "" {
		fmt.Printf("Terminus is already installed, please uninstall it first.")
		return nil
	}

	var arg = common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetTerminusVersion(opts.Version)
	arg.SetRegistryMirrors(opts.RegistryMirrors)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetReverseProxy()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	home := runtime.GetHomeDir()
	if manifest == "" {
		manifest = home + "/.terminus/installation.manifest"
	}

	runtime.Arg.SetManifest(manifest)

	var p = system.PrepareSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}

func getStorageValueFromEnv() *common.Storage {
	storageType := os.Getenv("STORAGE")
	switch storageType {
	case common.S3, common.OSS:
	default:
		storageType = common.Minio
	}

	return &common.Storage{
		StorageType:         storageType,
		StorageBucket:       os.Getenv(common.ENV_S3_BUCKET),
		StoragePrefix:       os.Getenv(common.ENV_BACKUP_KEY_PREFIX),
		StorageAccessKey:    os.Getenv(common.ENV_AWS_ACCESS_KEY_ID_SETUP),
		StorageSecretKey:    os.Getenv(common.ENV_AWS_SECRET_ACCESS_KEY_SETUP),
		StorageToken:        os.Getenv(common.ENV_AWS_SESSION_TOKEN_SETUP),
		StorageClusterId:    os.Getenv(common.ENV_CLUSTER_ID),
		StorageSyncSecret:   os.Getenv(common.ENV_BACKUP_SECRET),
		StorageVendor:       os.Getenv(common.ENV_TERMINUS_IS_CLOUD_VERSION),
		BackupClusterBucket: os.Getenv(common.ENV_BACKUP_CLUSTER_BUCKET),
	}
}
