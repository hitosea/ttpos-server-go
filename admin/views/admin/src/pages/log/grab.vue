<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-select
            v-model="searchParams.company_uuid"
            type="text"
            filterable
            remote
            :remote-method="remoteMethod"
            :placeholder="$t('请选择门店')"
            clearable
            style="min-width: 200px"
          >
            <el-option v-for="item in shopList" :key="item.id" :value="item.app_id?.toString()" :label="item.shop_supplier_name"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.import_type" :placeholder="$t('类型')" clearable style="min-width: 200px">
            <el-option :value="0" :label="$t('全部')" />
            <el-option :value="1" :label="$t('TTPOS推送到平台')" />
            <el-option :value="2" :label="$t('平台推送到TTPOS')" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.status" :placeholder="$t('状态')" clearable style="min-width: 200px">
            <el-option :value="-1" :label="$t('全部')" />
            <el-option :value="0" :label="$t('进行中')" />
            <el-option :value="1" :label="$t('成功')" />
            <el-option :value="2" :label="$t('失败')" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list" row-key="uuid" border>
      <el-table-column prop="import_type" :label="$t('类型')" width="180">
        <template #default="scope">
          {{ scope.row.import_type === 1 ? $t('TTPOS推送到平台') : scope.row.import_type === 2 ? $t('平台推送到TTPOS') : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="company_name" :label="$t('门店')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }">
        <template #default>
          {{ getSelectedShopName() || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="platform" :label="$t('名称')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }">
        <template #default="scope">
          {{ scope.row.platform || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('状态')" width="100">
        <template #default="scope">
          <el-tag v-if="scope.row.status === 0" type="warning">{{ $t('进行中') }}</el-tag>
          <el-tag v-else-if="scope.row.status === 1" type="success">{{ $t('成功') }}</el-tag>
          <el-tag v-else-if="scope.row.status === 2" type="danger">{{ $t('失败') }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="create_time" :label="$t('时间')" width="180">
        <template #default="scope">
          {{ formatTime(scope.row.create_time) }}
        </template>
      </el-table-column>
      <el-table-column prop="error_message" :label="$t('原因')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }">
        <template #default="scope">
          {{ scope.row.error_message || '-' }}
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination
      :disabled="isFetching"
      :total="data?.data?.page_info?.total"
      :currentPage="data?.data?.page_info?.page_no"
      :pageSize="data?.data?.page_info?.page_size"
      @change="handlePageSizeChange"
    />
  </div>
</template>
<script setup lang="ts">
  import { getLogList, logListType } from '@/api/log';
  import { getShopList, ShopListData } from '@/api/merchant';
  import { ref } from 'vue';
  import { useQuery } from 'vue-query';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { $t } from '@/i18n';
  import { message } from '@/utils/feedback';

  const { hasPermission } = useUserInfoStore();

  const searchParams = ref<logListType>({
    page: 1,
    page_size: 10,
  });

  const shopList = ref<ShopListData[]>([]);

  const { isFetching, data, refetch } = useQuery(['getLogList'], () => getLogList(searchParams.value), {
    enabled: false,
    keepPreviousData: true,
  });

  // 远程搜索门店
  const remoteMethod = (keyword: string) => {
    setTimeout(() => {
      getShopListData(keyword);
    }, 200);
  };

  // 获取门店列表
  const getShopListData = async (keyword: string = '') => {
    try {
      const res = await getShopList({ keyword: keyword, list_rows: 100, shop_type: 0, status: 0 });
      const newShops = res.data?.list?.data || [];
      // 合并新数据，避免重复
      if (keyword) {
        shopList.value = newShops;
      } else {
        // 首次加载或清空搜索时，合并所有门店
        const existingIds = new Set(shopList.value.map((s) => s.app_id));
        const uniqueShops = newShops.filter((s) => !existingIds.has(s.app_id));
        shopList.value = [...shopList.value, ...uniqueShops];
      }
    } catch (error) {
      console.error('获取门店列表失败:', error);
    }
  };

  // 根据当前选中的门店获取名称
  const getSelectedShopName = () => {
    if (!searchParams.value.company_uuid) return '';
    const shop = shopList.value.find((item) => item.app_id?.toString() === searchParams.value.company_uuid?.toString());
    return shop?.shop_supplier_name || '';
  };

  // 格式化时间
  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '-';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // 查询
  const handleSearch = () => {
    if (!searchParams.value.company_uuid) {
      message.warning($t('请选择门店'));
      return;
    }

    searchParams.value.page = 1;
    if (hasPermission(['admin_takeout_logs'])) {
      refetch.value();
    }
  };

  // 分页变化
  const handlePageSizeChange = (page: number, pageSize: number) => {
    searchParams.value.page = page || 1;
    searchParams.value.page_size = pageSize || 10;
    refetch.value();
  };

  // 初始化数据
  //   const initData = () => {
  //     searchParams.value = {
  //       page: 1,
  //       page_size: 10,
  //     };
  //     getShopListData();
  //     handleSearch();
  //   };

  //   onMounted(() => {
  //     if (roleListPath.value && roleListPath.value.length > 0) {
  //       initData();
  //     }
  //   });
</script>
<style lang="scss" scoped></style>
