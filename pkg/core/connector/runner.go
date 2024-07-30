/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package connector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type Runner struct {
	Conn  Connection
	Debug bool
	Host  Host
	Index int
}

func (r *Runner) Exec(cmd string, printOutput bool, printLine bool) (string, int, error) {
	if !r.Host.GetMinikube() {
		if r.Conn == nil {
			return "", 1, errors.New("no ssh connection available")
		}
	}

	var stdout string
	var code int
	var err error

	if r.Host.GetMinikube() {
		stdout, code, err = r.Host.Exec(cmd, printOutput, printLine)
	} else {
		stdout, code, err = r.Conn.Exec(SudoPrefix(cmd), r.Host, printLine)
	}

	if err != nil {
		logger.Errorf("[exec] %s CMD: %s, ERROR: %s", r.Host.GetName(), cmd, err)
	}

	if printOutput {
		logger.Debugf("[exec] %s CMD: %s, OUTPUT: \n%s", r.Host.GetName(), cmd, stdout)
	}

	logger.Infof("[exec] %s CMD: %s, OUTPUT: %s", r.Host.GetName(), cmd, stdout)

	return stdout, code, err
}

func (r *Runner) Cmd(cmd string, printOutput bool, printLine bool) (string, error) {
	stdout, _, err := r.Exec(cmd, printOutput, printLine)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

// ~ Extension
func (r *Runner) CmdExt(cmd string, printOutput bool, printLine bool) (string, error) {
	if !r.Host.GetMinikube() {
		if r.Conn == nil {
			return "", errors.New("no ssh connection available")
		}
	}

	var stdout string
	var err error
	if r.Host.GetMinikube() {
		stdout, _, err = r.Host.Exec(cmd, printOutput, printLine)
	} else {
		stdout, _, err = r.Conn.Exec(cmd, r.Host, printLine)
	}

	if printOutput {
		logger.Debugf("[exec] %s CMD: %s, OUTPUT: \n%s", r.Host.GetName(), cmd, stdout)
	}

	logger.Infof("[exec] %s CMD: %s, OUTPUT: %s", r.Host.GetName(), cmd, stdout)

	return stdout, err
}

func (r *Runner) SudoExec(cmd string, printOutput bool, printLine bool) (string, int, error) {
	return r.Exec(cmd, printOutput, printLine)
}

func (r *Runner) SudoCmd(cmd string, printOutput bool, printLine bool) (string, error) {
	return r.Cmd(cmd, printOutput, printLine)
}

// ~ Extension
func (r *Runner) SudoCmdExt(cmd string, printOutput bool, printLine bool) (string, error) {
	if !r.Host.GetMinikube() {
		if r.Conn == nil {
			return "", errors.New("no ssh connection available")
		}
	}

	var stdout string
	var err error

	if r.Host.GetMinikube() {
		// stdout, _, err = util.Exec(SudoPrefix(cmd), printOutput, printLine)
		stdout, err = r.Host.CmdExt(cmd, printOutput, printLine)
	} else {
		stdout, _, err = r.Conn.Exec(SudoPrefix(cmd), r.Host, printLine)
	}

	if printOutput {
		logger.Debugf("[exec] %s CMD: %s, OUTPUT: \n%s", r.Host.GetName(), cmd, stdout)
	}

	logger.Infof("[exec] %s CMD: %s, OUTPUT: %s", r.Host.GetName(), cmd, stdout)

	return stdout, err
}

func (r *Runner) Fetch(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Fetch(local, remote, r.Host); err != nil {
		logger.Debugf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Debugf("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(local, remote string) error {
	if !r.Host.GetMinikube() {
		if r.Conn == nil {
			return errors.New("no ssh connection available")
		}
	}

	var err error
	if r.Host.GetMinikube() {
		err = r.Host.Scp(local, remote)
	} else {
		err = r.Conn.Scp(local, remote, r.Host)
	}

	if err != nil {
		logger.Debugf("scp local file %s to remote %s failed: %v", local, remote, err)
		return err
	}
	logger.Infof("scp local file %s to remote %s success", local, remote)
	return nil
}

func (r *Runner) SudoScp(local, remote string) error {
	if !r.Host.GetMinikube() {
		if r.Conn == nil {
			return errors.New("no ssh connection available")
		}
	}

	// ! remote             /etc/kubernetes/addons/clusterconfigurations.yaml
	// ! remoteTmp          /tmp/kubekey/etc/kubernetes/addons/clusterconfigurations.yaml
	// scp to tmp dir
	remoteTmp := filepath.Join(common.TmpDir, remote)

	// remoteTmp := remote
	if err := r.Scp(local, remoteTmp); err != nil { // ~ copy
		return err
	}

	// ! local              /Users/admin/my/build_1/install-wizard-v1.7.0-6659/pkg/admindeMBP-2/kubesphere.yaml
	// ! baseRemotePath     /etc/kubernetes/addons
	baseRemotePath := remote
	if !util.IsDir(local) {
		baseRemotePath = filepath.Dir(remote)
	}
	if !r.Host.GetMinikube() {
		if err := r.Conn.MkDirAll(baseRemotePath, "", r.Host); err != nil {
			return err
		}
	}

	var remoteDir = filepath.Dir(remote)
	if !util.IsExist(remoteDir) {
		util.Mkdir(remoteDir)
	}

	if !r.Host.GetMinikube() {
		if _, err := r.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false, false); err != nil {
			return err
		}

		if _, err := r.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false, false); err != nil {
			return err
		}
	}

	// if r.Host.GetMinikube() {
	// 	if _, err := r.Host.CmdExt(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false, false); err != nil {
	// 		return err
	// 	}

	// 	if _, err := r.Host.CmdExt(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false, false); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	if _, err := r.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false, false); err != nil {
	// 		return err
	// 	}

	// 	if _, err := r.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false, false); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (r *Runner) FileExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok := r.Conn.RemoteFileExist(remote, r.Host)
	logger.Debugf("check remote file exist: %v", ok)
	return ok, nil
}

func (r *Runner) DirExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok, err := r.Conn.RemoteDirExist(remote, r.Host)
	if err != nil {
		logger.Debugf("check remote dir exist failed: %v", err)
		return false, err
	}
	logger.Debugf("check remote dir exist: %v", ok)
	return ok, nil
}

func (r *Runner) MkDir(path string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.MkDirAll(path, "", r.Host); err != nil {
		logger.Errorf("make remote dir %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) Chmod(path string, mode os.FileMode) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Chmod(path, mode); err != nil {
		logger.Errorf("chmod remote path %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) FileMd5(path string) (string, error) {
	if r.Conn == nil {
		return "", errors.New("no ssh connection available")
	}

	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", path)
	out, _, err := r.Conn.Exec(cmd, r.Host, false)
	if err != nil {
		logger.Errorf("count remote %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}
