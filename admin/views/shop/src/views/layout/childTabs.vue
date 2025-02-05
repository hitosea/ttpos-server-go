<template>
  <div class="el-tabs-container" v-if="tabList.length > 0">
    <el-tabs :model-value="activeValue" @tab-click="tabClick">
      <el-tab-pane :label="item.value" :name="item.key" v-for="(item, index) in tabList" :key="index"></el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup>
  import { ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { useUserStore } from '../../store/index';

  const { bus_emit, bus_on, computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = ref(supplier.value?.app_id || 0);
  const activeValue = ref(0);
  const tab_type = ref('');
  const router = useRouter();
  const tabList = ref([]);

  bus_on('tabData', (res) => {
    tabList.value = res.list;
    activeValue.value = res.active;
    tab_type.value = res.tab_type;
  });

  bus_on('activeValue', (res) => {
    if (res && res.params) {
      activeValue.value = res.params;
    } else {
      activeValue.value = res;
    }
  });
  bus_on('noTarget', (res) => {
    activeValue.value = res;
  });
  const tabClick = (event) => {
    let e = event.props;
    /*店内商品*/
    if (tab_type.value == 'storeproduct') {
      router.push({
        path: '/' + app_id.value + '/product/store/index',
        query: {
          type: e.name,
        },
      });
    }

    /*会员管理*/
    if (tab_type.value == 'uesrmanage') {
      router.push({
        path: '/' + app_id.value + '/card/user/index',
        query: {
          type: e.name,
        },
      });
    }
    /*会员卡管理*/
    if (tab_type.value == 'cardmanage') {
      router.push({
        path: '/' + app_id.value + '/card/card/index',
        query: {
          type: e.name,
        },
      });
    }

    /*商品扩展*/
    if (tab_type.value == 'expand') {
      router.push({
        path: '/' + app_id.value + '/product/expand/index',
        query: {
          type: e.name,
        },
      });
    }

    /*自助餐*/
    if (tab_type.value == 'buffetproduct') {
      router.push({
        path: '/' + app_id.value + '/product/buffet/index',
        query: {
          type: e.name,
        },
      });
    }

    /*商品扩展*/
    if (tab_type.value == 'printing') {
      router.push({
        path: '/' + app_id.value + '/supplier/printing/index',
        query: {
          type: e.name,
        },
      });
    }
    /*商品扩展*/
    if (tab_type.value == 'business') {
      router.push({
        path: '/' + app_id.value + '/supplier/business/index',
        query: {
          type: e.name,
        },
      });
    }

    /*桌位管理*/
    if (tab_type.value == 'tablemanage') {
      router.push({
        path: '/' + app_id.value + '/supplier/table/index',
        query: {
          type: e.name,
        },
      });
    }

    activeValue.value = e.name;
    bus_emit('activeValue', e.name);
  };
</script>
<style lang="scss" scoped>
  .el-tabs__header {
    margin-bottom: 16px;
  }
</style>
