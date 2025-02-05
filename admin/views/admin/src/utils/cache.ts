export const prefix_key = 'cloud_admin';

export const setCache = (key: string, value: any, expire?: string) => {
  key = getCacheKey(key);
  let data: any = {
    expire: expire ? currentTime() + expire : '',
    value,
  };

  if (typeof data === 'object') {
    data = JSON.stringify(data);
  }
  try {
    localStorage.setItem(key, data);
  } catch (e) {
    return null;
  }
};

export const getCache = (key: string) => {
  key = getCacheKey(key);
  try {
    const data = localStorage.getItem(key);
    if (!data) {
      return null;
    }
    const { value, expire } = JSON.parse(data);
    if (expire && expire < currentTime()) {
      localStorage.removeItem(key);
      return null;
    }
    return value;
  } catch (e) {
    return null;
  }
};

export const removeCache = (key: string) => {
  key = getCacheKey(key);
  localStorage.removeItem(key);
};

const currentTime = () => {
  return Math.round(new Date().getTime() / 1000);
};

const getCacheKey = (key: string) => {
  return `${prefix_key}_${key}`.toUpperCase();
};
