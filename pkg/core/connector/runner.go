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
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

type Runner struct {
	Conn  Connection
	Debug bool
	Host  Host
	Index int
}

func (r *Runner) Exec(cmd string, printOutput bool) (string, int, error) {
	if r.Conn == nil {
		return "", 1, errors.New("no ssh connection available")
	}

	stdout, code, err := r.Conn.Exec(cmd, r.Host)
	logger.Debugf("command: [%s]\n%s", r.Host.GetName(), cmd)
	if stdout != "" {
		logger.Debugf("stdout: [%s]\n%s", r.Host.GetName(), stdout)
	}

	if err != nil {
		logger.Errorf("[exec] %s CMD: %s, ERROR: %s", r.Host.GetName(), cmd, err)
	}

	if printOutput {
		if stdout != "" {
			fmt.Printf("stdout: [%s]\n%s\n", r.Host.GetName(), stdout)
		}
	}
	return stdout, code, err
}

func (r *Runner) Cmd(cmd string, printOutput bool) (string, error) {
	stdout, _, err := r.Exec(cmd, printOutput)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (r *Runner) SudoExec(cmd string, printOutput bool) (string, int, error) {
	return r.Exec(cmd, printOutput)
}

func (r *Runner) SudoCmd(cmd string, printOutput bool) (string, error) {
	return r.Cmd(SudoPrefix(cmd), printOutput)
}

func (r *Runner) Fetch(local, remote string, printOutput bool) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Fetch(local, remote, r.Host); err != nil {
		logger.Errorf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Infof("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Scp(local, remote, r.Host); err != nil {
		logger.Debugf("scp local file %s to remote %s failed: %v", local, remote, err)
		return err
	}
	logger.Debugf("scp local file %s to remote %s success", local, remote)
	return nil
}

func (r *Runner) SudoScp(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	// scp to tmp dir
	remoteTmp := filepath.Join(common.TmpDir, remote)
	//remoteTmp := remote
	if err := r.Scp(local, remoteTmp); err != nil {
		return err
	}

	baseRemotePath := remote
	if !util.IsDir(local) {
		baseRemotePath = filepath.Dir(remote)
	}
	if err := r.Conn.MkDirAll(baseRemotePath, "", r.Host); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false); err != nil {
		return err
	}
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
	out, _, err := r.Conn.Exec(cmd, r.Host)
	if err != nil {
		logger.Errorf("count remote %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}
