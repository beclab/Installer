package pipelines

import (
	"errors"
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(opts *options.CliPrepareSystemOptions) error {

	var terminusVersion, _ = phase.GetTerminusVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	var arg = common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetMinikubeProfile(opts.MinikubeProfile)
	arg.SetTerminusVersion(opts.Version)
	arg.SetRegistryMirrors(opts.RegistryMirrors)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetTokenMaxAge()
	arg.SetReverseProxy()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %w", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")

	runtime.Arg.SetManifest(manifest)

	var p = system.PrepareSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}

func getStorageValueFromEnv() *common.Storage {
	storageType := os.Getenv(common.ENV_STORAGE)
	switch storageType {
	case "":
		storageType = common.ManagedMinIO
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
