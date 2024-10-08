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
	arg.SetTokenMaxAge()
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
	case "s3", "oss":
	default:
		storageType = "minio"
	}

	return &common.Storage{
		StorageType:       storageType,
		StorageDomain:     os.Getenv("S3_BUCKET"),
		StorageBucket:     os.Getenv("S3_BUCKET"), // os.Getenv("BACKUP_CLUSTER_BUCKET"),
		StoragePrefix:     os.Getenv("BACKUP_KEY_PREFIX"),
		StorageAccessKey:  os.Getenv("AWS_ACCESS_KEY_ID_SETUP"),
		StorageSecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY_SETUP"),
		StorageToken:      os.Getenv("AWS_SESSION_TOKEN_SETUP"),
		StorageClusterId:  os.Getenv("CLUSTER_ID"),
		StorageSyncSecret: os.Getenv("BACKUP_SECRET"),
		StorageVendor:     os.Getenv("TERMINUS_IS_CLOUD_VERSION"),
	}
}
