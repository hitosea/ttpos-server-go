import { defineStore } from 'pinia';

export const useCloudStore = defineStore({
  id: 'CloudBasic',
  state: () => ({
    cloudBasic: {
      base: {},
    },
    isCloudDeploy: false,
  }),
  getters: {
    getCloudBasic: (state) => state.cloudBasic.base,
    getIsCloudDeploy: (state) => state.isCloudDeploy,
  },
  actions: {
    setCloudBasic(data) {
      this.cloudBasic = data;
    },
    setIsCloudDeploy(data) {
      this.isCloudDeploy = data;
    },
  },
  persist: {
    key: 'cloud',
    storage: localStorage,
  },
});
