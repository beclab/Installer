export interface InstallRequest {
  config: {
    terminus_os_domainname: string;
    terminus_os_username: string;
    kube_type: 'k8s' | 'k3s';
    vendor: 'private' | 'aws' | 'aliyun';
    gpu_enable: number;
    gpu_share: number;
    version: string;
  };
}

export interface InstallMessageItem {
  info: string;
  time: string;
}

export interface InstallStatusResponse {
  msg: InstallMessageItem[];
  percent: string; // 总进度
  status: 'Not_Started' | 'Download' | 'Install' | 'Fail' | 'Success';
}
