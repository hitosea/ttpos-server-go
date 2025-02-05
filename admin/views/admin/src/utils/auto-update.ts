// 记录上一次script地址
let preScript: string[] | undefined;

// 获取script标签正则
const scriptReg = /<script.*?src=['"](.*?)['"].*?><\/script>/g;

// 获取最新的script页面链接
const extractNewScripts = async () => {
  try {
    const html = await fetch(`/?timestamp=${Date.now()}`).then((res) => res.text());
    // 重置位置
    scriptReg.lastIndex = 0;
    // src 集合
    const result: string[] = [];
    let match;
    while ((match = scriptReg.exec(html)) !== null) {
      result.push(match[0]);
    }
    return result;
  } catch (error) {
    //
  }
};

// 判断是否更新
const needUpdate = async () => {
  const newScript = await extractNewScripts();
  if (!preScript) {
    preScript = newScript;
    return false;
  }
  // 判断长度
  if (preScript.length !== newScript?.length) {
    preScript = newScript;
    return true;
  }
  let isUpdate = false;
  // 循环判断
  for (let i = 0; i < preScript.length; i++) {
    if (preScript[i] !== newScript[i]) {
      isUpdate = true;
      break;
    }
  }
  return isUpdate;
};

// 定时检测页面是否有更新
const autoRefresh = () => {
  setTimeout(async () => {
    const willUpdate = await needUpdate();
    // 检测到页面有更新
    if (willUpdate) {
      window.location.reload();
    }
    autoRefresh();
  }, 10 * 1000);
};

// 判断环境
if (['production', 'prod'].includes(process.env.NODE_ENV || '')) {
  autoRefresh();
}
