<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-input v-model="searchParams.keyword" :placeholder="$t('商家名称/ID')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.status" :placeholder="$t('外送渠道')" clearable style="min-width: 200px">
            <el-option value=" " :label="$t('全部渠道')" />
            <el-option :value="1" label="SKootar" />
            <el-option :value="2" label="Grab" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-date-picker v-model="searchParams.time" type="month" :placeholder="$t('统计月份')" style="min-width: 200px"></el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button class="!mr-0" type="primary" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button v-permission="" icon="download" @click="handleExport">{{ $t('导出报表') }}</el-button>
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
          <p class="text-lg font-bold text-green-500">是</p>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';

  const searchParams = ref({
    status: '',
    keyword: '',
    time: '',
  });

  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);
  const isFetching = ref(false);

  const handleSearch = () => {
    isFetching.value = true;
  };

  const handleExport = () => {};
</script>
