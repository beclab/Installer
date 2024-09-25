package terminus

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type InstallVelero struct {
	common.KubeAction
}

func (i *InstallVelero) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s install --crds-only --retry 10 --delay 5", velero), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "install velero failed")
	}

	return nil
}

type InstallVeleroModule struct {
	common.KubeModule
}

func (i *InstallVeleroModule) Init() {

}

type CreateBackupLocation struct {
	common.KubeAction
}

// backup-location
func (c *CreateBackupLocation) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	var ns = "os-system"
	var provider = "terminus"
	var storage = "terminus-cloud"

	var cmd = fmt.Sprintf("%s backup-location get -n %s -l 'name=%s'", velero, ns, storage)
	if res, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err == nil && res != "" {
		return nil
	}

	cmd = fmt.Sprintf("%s backup-location create %s --provider %s --namespace %s --prefix '' --bucket %s --labels name=%s",
		velero, storage, provider, ns, storage, storage)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create backup-location failed")
	}

	// TODO test

	return nil
}

type InstallPlugin struct {
	common.KubeAction
}

func (i *InstallPlugin) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	var ns = "os-system"
	var cmd = fmt.Sprintf("%s plugin get -n %s |grep 'velero.io/terminus' |wc -l", velero, ns)
	pluginCounts, _ := runtime.GetRunner().Host.SudoCmd(cmd, false, true)
	if counts := utils.ParseInt(pluginCounts); counts > 0 {
		return nil
	}

	var args string
	var veleroVersion = "v1.11.3"
	var veleroPluginVersion = "v1.0.2"
	if i.KubeConf.Arg.IsRaspbian() {
		args = " --retry 30 --delay 5"
	}

	cmd = fmt.Sprintf("%s install --no-default-backup-location --namespace %s --image beclab/velero:%s --use-volume-snapshots=false --no-secret --plugins beclab/velero-plugin-for-terminus:%s --velero-pod-cpu-request=10m --velero-pod-cpu-limit=200m --node-agent-pod-cpu-request=10m --node-agent-pod-cpu-limit=200m --wait --wait-minute 30 %s", velero, ns, veleroVersion, veleroPluginVersion, args)

	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "velero install plugin error")
	}

	cmd = fmt.Sprintf("%s plugin add beclab/velero-plugin-for-terminus:%s -n %s", velero, veleroPluginVersion, ns)
	if stdout, _ := runtime.GetRunner().Host.SudoCmd(cmd, false, true); stdout != "" && !strings.Contains(stdout, "Duplicate") {
		logger.Debug(stdout)
	}

	return nil
}

type VeleroPatch struct {
	common.KubeAction
}

func (v *VeleroPatch) Execute(runtime connector.Runtime) error {
	velero, err := util.GetCommand(common.CommandVelero)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "velero not found")
	}

	var ns = "os-system"
	var patch = `[{"op":"replace","path":"/spec/template/spec/volumes","value": [{"name":"plugins","emptyDir":{}},{"name":"scratch","emptyDir":{}},{"name":"terminus-cloud","hostPath":{"path":"/terminus/rootfs/k8s-backup", "type":"DirectoryOrCreate"}}]},{"op": "replace", "path": "/spec/template/spec/containers/0/volumeMounts", "value": [{"name":"plugins","mountPath":"/plugins"},{"name":"scratch","mountPath":"/scratch"},{"mountPath":"/data","name":"terminus-cloud"}]},{"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"privileged": true, "runAsNonRoot": false, "runAsUser": 0}}]`

	if stdout, _ := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s patch deploy velero -n %s --type='json' -p='%s'", velero, ns, patch), false, true); stdout != "" && !strings.Contains(stdout, "patched") {
		logger.Errorf("velero plugin patched error %s", stdout)
	}

	return nil
}

type InstallVeleroPluginModule struct {
	common.KubeModule
}

func (i *InstallVeleroPluginModule) Init() {

}
