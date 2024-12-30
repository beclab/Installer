package gpu

import (
	"context"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GPUEnablePrepare struct {
	common.KubePrepare
}

func (p *GPUEnablePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsWsl() {
		return false, nil
	}

	if systemInfo.IsUbuntu() && systemInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
		return false, nil
	}
	return p.KubeConf.Arg.GPU.Enable, nil
}

type GPUSharePrepare struct {
	common.KubePrepare
}

func (p *GPUSharePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return p.KubeConf.Arg.GPU.Share || runtime.GetSystemInfo().IsWsl(), nil
}

type CudaInstalled struct {
	common.KubePrepare
	precheck.CudaCheckTask
}

func (p *CudaInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	err := p.CudaCheckTask.Execute(runtime)
	if err != nil {
		if err == precheck.ErrCudaInstalled {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

type CudaNotInstalled struct {
	common.KubePrepare
	precheck.CudaCheckTask
}

func (p *CudaNotInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	err := p.CudaCheckTask.Execute(runtime)
	if err != nil {
		if err == precheck.ErrCudaInstalled {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type K8sNodeInstalled struct {
	common.KubePrepare
}

func (p *K8sNodeInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	node, err := client.Kubernetes().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(errors.WithStack(err), "list nodes error")
	}

	if len(node.Items) == 0 {
		return false, nil
	}

	return true, nil
}
