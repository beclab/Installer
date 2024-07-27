package plugins

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

// ~ GenerateKubeSphereToken
type GenerateKubeSphereToken struct {
	common.KubeAction
}

func (t *GenerateKubeSphereToken) Execute(runtime connector.Runtime) error {
	var random, err = utils.GeneratePassword(32)
	if err != nil {
		logger.Errorf("failed to generate password: %v", err)
		return err
	}

	token, err := util.EncryptToken(random)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create kubesphere token failed")
	}

	var cmd = fmt.Sprintf("/usr/local/bin/kubectl get secrets -n %s --no-headers", common.NamespaceKubesphereSystem)
	stdout, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
	if strings.Contains(stdout, "kubesphere-secret") {
		cmd = fmt.Sprintf("/usr/local/bin/kubectl delete secrets -n %s kubesphere-secret", common.NamespaceKubesphereSystem)
		runtime.GetRunner().SudoCmd(cmd, false, true)
	}

	cmd = fmt.Sprintf("/usr/local/bin/kubectl create secret generic kubesphere-secret --from-literal=token=%s --from-literal=secret=%s -n %s",
		token, random, common.NamespaceKubesphereSystem)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create kubesphere token failed")
	}

	t.PipelineCache.Set(common.CacheJwtSecret, random)

	return nil
}

// +++++

// ~ CreateKubeSphereSecretModule
type CreateKubeSphereSecretModule struct {
	common.KubeModule
}

func (m *CreateKubeSphereSecretModule) Init() {
	m.Name = "CreateKubeSphereSecret"

	generateKubeSphereToken := &task.RemoteTask{
		Name:  "GenerateKubeSphereToken",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(GenerateKubeSphereToken),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{generateKubeSphereToken}
}
