import type { DirectiveBinding } from 'vue';
import { useUserInfoStore } from '@/stores/userInfo';
import { isArray } from 'lodash-es';

export const VPermission = {
  // mounted是指令的一个生命周期
  mounted(el: Element, binding: DirectiveBinding<string[]>) {
    // value 获取用户使用自定义指令绑定的内容
    const { value } = binding;
    if (!value) return;
    // 判断用户使用自定义指令，是否使用正确了
    if (isArray(value)) {
      // 用户权限
      const { hasPermission } = useUserInfoStore();
      // 找不到返回false，使用自定义指令的钩子函数，操作dom元素删除该节点
      if (!hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    } else {
      throw new Error(`使用方式： v-permission="['admin','editor']"`);
    }
  },
};
