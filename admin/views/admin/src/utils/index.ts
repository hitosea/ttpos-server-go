/**
 * 随机字符
 * @param len
 * @returns {string}
 */
export function randomString(len: number = 32) {
  const $chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678oOLl9gqVvUuI1';
  const maxPos = $chars.length;
  let pwd = '';
  for (let i = 0; i < len; i++) {
    pwd += $chars.charAt(Math.floor(Math.random() * maxPos));
  }
  return pwd;
}

export function replaceLinkIcon(link: string = '') {
  if (!link) return;
  // 获取所有的link标签
  const linkTags = document.getElementsByTagName('link');
  // 定义一个正则表达式
  const regex = /icon/i; // 忽略大小写匹配
  let foundIcon = false;
  // 遍历link标签
  for (let i = 0; i < linkTags.length; i++) {
    const linkTag = linkTags[i];
    const rel = linkTag.getAttribute('rel');
    // 使用正则表达式匹配包含icon的link标签
    if (rel && regex.test(rel)) {
      // 替换icon链接
      linkTag.setAttribute('href', link);
      foundIcon = true;
      break;
    }
  }
  // 如果页面中没有包含icon的link标签，则新增一个
  if (!foundIcon) {
    const newLinkTag = document.createElement('link');
    newLinkTag.setAttribute('rel', 'icon');
    newLinkTag.setAttribute('href', link);
    document.head.appendChild(newLinkTag);
  }
}
