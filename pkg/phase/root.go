package phase

import (
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func GetTerminusVersion() (string, error) {
	var terminusTask = &terminus.GetTerminusVersion{}
	return terminusTask.Execute()
}

func GetKubeVersion() (string, string, error) {
	var kubeTask = &kubernetes.GetKubeVersion{}
	return kubeTask.Execute()
}
