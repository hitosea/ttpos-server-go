<template>
  <!--
      	描述：营销活动管理
      -->
  <div class="common-search-wrap">
    <!--活动管理-->
    <manage v-if="activeName == 'manage'"></manage>
  </div>
</template>

<script setup>
  import { reactive, ref, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue';
  import { useRoute } from 'vue-router';
  import { useUserStore } from '@/store';
  import manage from './manage/index.vue';

  // 获取全局属性和路由
  const { proxy } = getCurrentInstance();
  const { $t, $filter } = proxy;
  const route = useRoute();

  // 获取store
  const { bus_emit, bus_off, bus_on } = useUserStore();

  // 响应式数据
  const activeName = ref('');

  // 切换数组原始数据
  const sourceList = reactive([
    {
      key: 'manage',
      value: $t('活动管理'),
      path: '/marketing/activity/manage/index',
    },
  ]);

  // 权限筛选后的数据
  const tabList = ref([]);

  // 方法
  const authFilter = () => {
    const list = [];
    for (let i = 0; i < sourceList.length; i++) {
      const item = sourceList[i];
      if ($filter.isAuth(item.path)) {
        list.push(item);
      }
    }
    return list;
  };

  // 生命周期
  onMounted(() => {
    tabList.value = authFilter();
    if (tabList.value.length > 0) {
      activeName.value = tabList.value[0].key;
    }

    if (route.query.type != null) {
      activeName.value = route.query.type;
    }

    // 监听传插件的值
    bus_on('activeValue', (res) => {
      activeName.value = res;
    });

    // 发送类别切换
    const params = {
      active: activeName.value,
      list: tabList.value,
      tab_type: 'activitymanage',
    };
    bus_emit('tabData', params);
  });

  onBeforeUnmount(() => {
    // 发送类别切换
    bus_emit('tabData', { active: null, tab_type: 'activitymanage', list: [] });
    bus_off('activeValue');
  });
</script>
