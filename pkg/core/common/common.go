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

package common

const (
	// TerminusKey        = ".terminus"
	ManifestDir        = "manifest"
	ImageCacheDir      = "images"
	PackageCacheDir    = "pkg"
	BuildFilesCacheDir = "files"
	BuildDir           = "build"
	GpuDir             = "gpu"

	// KubeKey = "kubekey"
	Cli               = "cli"
	KubeKey           = "pkg"
	Pkg               = "pkg"
	InstallDir        = "install-wizard"
	ImagesDir         = "images"
	ScriptsDir        = "scripts"
	WizardDir         = "wizard"
	ComponentsDir     = "components"
	DeployDir         = "deploy"
	TerminusDir       = "terminus"
	GreetingShell     = "greeting.sh"
	PrecheckOsShell   = "precheck_os.sh"
	InstallDepsShell  = "install_deps.sh"
	UninstallOsScript = "uninstall_cmd.sh"

	DefaultDomainName = "myterminus.com"

	ManifestImage     = "images.mf"
	ManifestImageNode = "images.node.mf"
	ManifestDeps      = "dependencies.mf"

	Pipeline = "Pipeline"
	Module   = "Module"
	Task     = "Task"
	Node     = "Node"

	LocalHost = "LocalHost"

	FileMode0755 = 0755
	FileMode0644 = 0644
	FileMode0600 = 0600
	FileMode0640 = 0640

	TmpDir = "/tmp/kubekey/" // todo

	// command
	CopyCmd = "cp -r %s %s"
	MoveCmd = "mv -f %s %s"
)

const (
	StateDownload = "Download"
	StateInstall  = "Install"
	StateFail     = "Fail"
	StateSuccess  = "Success"
)

const (
	DefaultInstallSteps int64 = 32
)

const (
	Linux   = "linux"
	Darwin  = "darwin"
	Windows = "windows"
)

const (
	DownloadUrl = "https://dc3p1870nn3cj.cloudfront.net"
)

const (
	ZfsSnapshotter = "/var/lib/containerd/io.containerd.snapshotter.v1.zfs"
)
