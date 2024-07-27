import { defineStore } from 'pinia';
import axios from 'axios';
import {
  InstallRequest,
  InstallStatusResponse,
  InstallMessageItem,
} from 'src/types';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

const url = 'http://127.0.0.1:30080';

export type InstallState = {
  msg: InstallMessageItem[];
  percent: string; // total progress in percent
  install_status: 'Not_Started' | 'Download' | 'Install' | 'Fail' | 'Success';
};

export const useInstallStore = defineStore('install', {
  state: () => {
    return {
      msg: [],
      percent: '',
      install_status: 'Not_Started',
    } as InstallState;
  },
  actions: {
    async install(req: InstallRequest) {
      try {
        await axios.post(url + '/api/webserver/v1/install', req);

        BtNotify.show({
          type: NotifyDefinedType.SUCCESS,
          message: 'Install Started',
        });

        this.install_status = 'Download';
      } catch (error) {
        BtNotify.show({
          type: NotifyDefinedType.FAILED,
          message: 'Network error, please try again later',
        });
      } finally {
      }
    },
    async status() {
      try {
        let time = 0;
        if (this.msg.length > 0) {
          time = parseInt(this.msg[this.msg.length - 1].time);
        }
        const data: InstallStatusResponse = await axios.get(
          url + '/api/webserver/v1/status?time=' + time
        );

        console.log(data);

        if (time == 0) {
          this.msg = data.msg;
        } else {
          this.msg = this.msg.concat(data.msg);
        }
        this.percent = data.percent;
        this.install_status = data.status;
      } catch (error) {
        BtNotify.show({
          type: NotifyDefinedType.FAILED,
          message: 'Network error, please try again later',
        });
      } finally {
      }
    },
  },
});
