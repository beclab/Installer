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

	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type BaseRuntime struct {
	ObjName         string
	connector       Connector
	runner          *Runner
	storage         storage.Provider
	rootDir         string
	workDir         string
	packageDir      string
	filesDir        string
	verbose         bool
	ignoreErr       bool
	allHosts        []Host
	roleHosts       map[string][]Host
	deprecatedHosts map[string]string
	cmdSed          string
}

func NewBaseRuntime(name string, connector Connector, verbose bool, ignoreErr bool, sqlProvider storage.Provider) BaseRuntime {
	base := BaseRuntime{
		ObjName:         name,
		connector:       connector,
		storage:         sqlProvider,
		verbose:         verbose,
		ignoreErr:       ignoreErr,
		allHosts:        make([]Host, 0, 0),
		roleHosts:       make(map[string][]Host),
		deprecatedHosts: make(map[string]string),
		cmdSed:          util.FormatSed(constants.OsType == common.Darwin),
	}
	if err := base.GenerateWorkDir(); err != nil {
		fmt.Printf("[ERRO]: Failed to create KubeKey work dir: %s\n", err)
		os.Exit(1)
	}
	// if err := base.InitLogger(); err != nil {
	// 	fmt.Printf("[ERRO]: Failed to init KubeKey log entry: %s\n", err)
	// 	os.Exit(1)
	// }

	return base
}

func (b *BaseRuntime) GetObjName() string {
	return b.ObjName
}

func (b *BaseRuntime) SetObjName(name string) {
	b.ObjName = name
}

func (b *BaseRuntime) GetRunner() *Runner {
	return b.runner
}

func (b *BaseRuntime) SetRunner(r *Runner) {
	b.runner = r
}

func (b *BaseRuntime) GetStorage() storage.Provider {
	return b.storage
}

func (b *BaseRuntime) SetStorage(s storage.Provider) {
	b.storage = s
}

func (b *BaseRuntime) GetConnector() Connector {
	return b.connector
}

func (b *BaseRuntime) SetConnector(c Connector) {
	b.connector = c
}

func (b *BaseRuntime) GenerateWorkDir() error {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "get current dir failed")
	}
	b.rootDir = currentDir

	rootPath := filepath.Join(currentDir, common.Pkg)
	if err := util.CreateDir(rootPath); err != nil {
		return errors.Wrap(err, "create work dir failed")
	}
	b.workDir = rootPath

	pkgsPath := filepath.Join(currentDir, common.Packages)
	if err := util.CreateDir(pkgsPath); err != nil {
		return errors.Wrap(err, "create package dir failed")
	}
	b.packageDir = pkgsPath

	filesPath := filepath.Join(currentDir, common.FilesDir)
	if err := util.CreateDir(filesPath); err != nil {
		return errors.Wrap(err, "create files dir failed")
	}
	b.filesDir = filesPath

	for i := range b.allHosts {
		subPath := filepath.Join(rootPath, b.allHosts[i].GetName())
		if err := util.CreateDir(subPath); err != nil {
			return errors.Wrap(err, "create work dir failed")
		}
	}
	return nil
}

func (b *BaseRuntime) GetHostWorkDir() string {
	return filepath.Join(b.workDir, b.RemoteHost().GetName())
}

func (b *BaseRuntime) GetRootDir() string {
	return b.rootDir
}

func (b *BaseRuntime) GetWorkDir() string {
	return b.workDir
}

func (b *BaseRuntime) GetPackageDir() string {
	return b.packageDir
}

func (b *BaseRuntime) GetFilesDir() string {
	return b.filesDir
}

func (b *BaseRuntime) GetIgnoreErr() bool {
	return b.ignoreErr
}

func (b *BaseRuntime) GetAllHosts() []Host {
	hosts := make([]Host, 0, 0)
	for i := range b.allHosts {
		if b.allHosts[i] == nil || b.HostIsDeprecated(b.allHosts[i]) {
			continue
		}
		hosts = append(hosts, b.allHosts[i])
	}
	return hosts
}

func (b *BaseRuntime) SetAllHosts(hosts []Host) {
	b.allHosts = hosts
}

func (b *BaseRuntime) GetHostsByRole(role string) []Host {
	if _, ok := b.roleHosts[role]; ok {
		return b.roleHosts[role]
	} else {
		return []Host{}
	}
}

func (b *BaseRuntime) RemoteHost() Host {
	return b.GetRunner().Host
}

func (b *BaseRuntime) DeleteHost(host Host) {
	i := 0
	for j := range b.allHosts {
		if b.allHosts[j].GetName() != host.GetName() {
			b.allHosts[i] = b.allHosts[j]
			i++
		}
	}
	b.allHosts[i] = nil
	b.allHosts = b.allHosts[:i]
	b.RoleMapDelete(host)
	b.deprecatedHosts[host.GetName()] = ""
}

func (b *BaseRuntime) HostIsDeprecated(host Host) bool {
	if _, ok := b.deprecatedHosts[host.GetName()]; ok {
		return true
	}
	return false
}

// func (b *BaseRuntime) InitLogger() error {
// 	if b.GetWorkDir() == "" {
// 		if err := b.GenerateWorkDir(); err != nil {
// 			return err
// 		}
// 	}
// 	logDir := filepath.Join(b.GetWorkDir(), "logs")
// 	logger.InitLog(logDir, b.verbose)
// 	return nil
// }

func (b *BaseRuntime) GetCommandSed() string {
	return b.cmdSed
}

func (b *BaseRuntime) Copy() Runtime {
	runtime := *b
	return &runtime
}

func (b *BaseRuntime) GenerateRoleMap() {
	for i := range b.allHosts {
		b.AppendRoleMap(b.allHosts[i])
	}
}

func (b *BaseRuntime) AppendHost(host Host) {
	b.allHosts = append(b.allHosts, host)
}

func (b *BaseRuntime) AppendRoleMap(host Host) {
	for _, r := range host.GetRoles() {
		if hosts, ok := b.roleHosts[r]; ok {
			hosts = append(hosts, host)
			b.roleHosts[r] = hosts
		} else {
			first := make([]Host, 0, 0)
			first = append(first, host)
			b.roleHosts[r] = first
		}
	}
}

func (b *BaseRuntime) RoleMapDelete(host Host) {
	for role, hosts := range b.roleHosts {
		i := 0
		for j := range hosts {
			if hosts[j].GetName() != host.GetName() {
				hosts[i] = hosts[j]
				i++
			}
		}
		hosts = hosts[:i]
		b.roleHosts[role] = hosts
	}
}
