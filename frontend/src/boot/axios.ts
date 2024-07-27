import { boot } from 'quasar/wrappers';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
const api = axios.create({ baseURL: 'https://api.example.com' });
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $axios: AxiosInstance;
  }
}

// Be careful when using SSR for cross-request state pollution
// due to creating a Singleton instance here;
// If any client changes this (global) instance, it might be a
// good idea to move this instance creation inside of the
// "export default () => {}" function below (which runs individually
// for each client)
//const api = axios.create({ baseURL: 'https://api.example.com' });

export default boot(({ app, router }) => {
  app.config.globalProperties.$axios = axios;

  app.config.globalProperties.$api = api;

  app.config.globalProperties.$axios.interceptors.response.use((response) => {
    if (!response || response.status !== 200 || !response.data) {
      BtNotify.show({
        type: NotifyDefinedType.FAILED,
        message: 'Network error, please try again later',
      });
      return response;
    }

    const data = response.data;

    if (data.code !== 0) {
      BtNotify.show({
        type: NotifyDefinedType.FAILED,
        message: data.message || 'Something Wrong. Please try again!',
      });
      throw new Error(data.message);
    }

    return data.data;
  });
});

export { api };
