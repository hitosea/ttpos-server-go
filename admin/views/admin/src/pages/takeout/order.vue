<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-select v-model="searchParams.uuid" type="text" filterable remote :remote-method="remoteMethod" :placeholder="$t('请选择商家')" clearable>
            <el-option class="!h-auto" v-for="item in companyList" :key="item.uuid" :value="item.uuid" :label="item.name">
              <div>
                <p class="">{{ item.name }}</p>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.channel" :placeholder="$t('外送渠道')" clearable style="min-width: 200px">
            <el-option value=" " :label="$t('全部渠道')" />
            <el-option v-for="channel in channels" :key="channel.value" :value="channel.value" :label="channel.name" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-date-picker v-model="searchParams.month" type="month" :placeholder="$t('统计月份')" style="min-width: 200px"></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button class="!mr-0" type="primary" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button icon="download" @click="handleExport">{{ $t('导出报表') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <!-- 表格 -->
    <el-table :data="[]" border>
      <el-table-column prop="" :label="$t('商家名称')" min-width="200" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>
      <el-table-column prop="" :label="$t('外送渠道')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('订单编号')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('订单配送费')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('距离(公里)')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('基础服务费')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('起步配送费')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('距离单价')" min-width="200"></el-table-column>
    </el-table>
    <!-- 本月统计汇总 -->
    <el-card class="w-full mt-4" shadow="hover">
      <h3 class="text-lg font-bold mb-4">{{ $t('本月统计汇总') }}</h3>
      <el-row :gutter="24">
        <el-col :span="8" class="mb-2">
          <p class="text-sm">{{ $t('总订单数') }}</p>
          <p class="text-lg font-bold text-primary">0</p>
        </el-col>
        <el-col :span="8" class="mb-2">
          <p class="text-sm">{{ $t('总配送费') }}</p>
          <p class="text-lg font-bold text-primary">0</p>
        </el-col>
        <el-col :span="8" class="mb-2">
          <p class="text-sm">{{ $t('SKootar渠道') }}</p>
          <p class="text-lg font-bold text-primary">0</p>
        </el-col>
        <el-col :span="8" class="mb-2">
          <p class="text-sm">{{ $t('Grab渠道') }}</p>
          <p class="text-lg font-bold text-primary">0</p>
        </el-col>
        <el-col :span="8" class="mb-2">
          <p class="text-sm">{{ $t('结清状态') }}</p>
          <p class="text-lg font-bold text-green-500"> {{ $t('是') }}</p>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { getTakeoutOrder } from '@/api/takeout/order';
  import { getCompanySelect, type TakeoutShopSelectData } from '@/api/takeout/shop';
  import { getChannels, type TakeoutShopData } from '@/api/takeout/shop';

  const searchParams = ref({
    uuid: null,
    channel: '',
    month: '',
  });

  const isFetching = ref(false);

  const channels = ref<TakeoutShopData[]>([]);
  const companyList = ref<TakeoutShopSelectData[]>([]);

  // 查询
  const handleSearch = () => {
    getTakeoutOrderData();
  };

  // 获取订单数据
  const getTakeoutOrderData = async () => {
    isFetching.value = true;
    try {
      const res = await getTakeoutOrder({ uuid: Number(searchParams.value.uuid), channel: searchParams.value.channel, month: searchParams.value.month });
      console.log(res);
    } catch (error) {
      console.error('获取渠道失败:', error);
    } finally {
      isFetching.value = false;
    }
  };

  // 远程搜索
  const remoteMethod = (keyword: string) => {
    setTimeout(() => {
      getCompanyListData(keyword);
    }, 200);
  };

  // 获取商家列表
  const getCompanyListData = async (keyword: string) => {
    try {
      const res = await getCompanySelect({ keyword: keyword, list_rows: 50, configured: 1 });
      companyList.value = res.data?.list?.data || [];
    } catch (error) {
      console.error('获取渠道失败:', error);
    }
  };

  // 获取渠道列表
  const getChannelsData = async () => {
    try {
      const res = await getChannels({ configured: 0 });
      channels.value = res.data || [];
    } catch (error) {
      console.error('获取渠道失败:', error);
    }
  };

  // 导出
  const handleExport = () => {};

  onMounted(() => {
    getCompanyListData('');
    getChannelsData();
  });
</script>
