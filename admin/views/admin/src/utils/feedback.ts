import { ElMessage, ElLoading } from 'element-plus';
import type { MessageParams } from 'element-plus';

let loadingInstance: any = null;
let messageInstance: any = null;

const resetMessage: any = (options: MessageParams) => {
  if (messageInstance) messageInstance.close();
  messageInstance = ElMessage(options);
};

['error', 'success', 'info', 'warning'].forEach((type) => {
  (resetMessage as any)[type] = (options: MessageParams) => {
    if (typeof options === 'string') {
      options = {
        message: options,
      };
    }
    (options as any).type = type;
    return resetMessage(options);
  };
});

export const message = resetMessage;

// 打开全局loading
export const globalLoading = (msg?: string) => {
  loadingInstance = ElLoading.service({
    lock: true,
    text: msg,
  });
};

// 关闭全局loading
export const closeGlobalLoading = () => {
  loadingInstance?.close();
};
