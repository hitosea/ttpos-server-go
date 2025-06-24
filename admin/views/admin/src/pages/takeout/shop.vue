<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-select v-model="searchParams.status" :placeholder="$t('外送渠道')" clearable style="min-width: 200px">
            <el-option value=" " :label="$t('全部渠道')" />
            <el-option :value="1" label="SKootar" />
            <el-option :value="2" label="Grab" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchParams.keyword" :placeholder="$t('商家名称/ID')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
      <div>
        <el-button class="!mr-0" v-permission="" type="primary" icon="plus" @click="handleAdd">{{ $t('添加商家') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table :data="[]" border>
      <el-table-column prop="" :label="$t('商家ID')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('商家名称')" min-width="200" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>
      <el-table-column prop="" :label="$t('外送渠道')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('参数设置')" min-width="200"></el-table-column>
      <el-table-column prop="" :label="$t('状态')" min-width="120">
        <template #default="scope">
          <el-switch v-model="scope.row.status" :loading="editLoading" :active-value="1" :inactive-value="0" @click="handleStatus(scope.row)"></el-switch>
        </template>
      </el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="150">
        <template #default="scope">
          <el-button type="primary" link @click="handleEdit(scope.row)">
            {{ $t('编辑') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination :total="total" :currentPage="currentPage" :pageSize="pageSize" @change="handlePageChange"></ti-pagination>

    <add-edit-shop v-model:show="addEditShow" />
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { $t } from '@/i18n';
  import AddEditShop from './components/add-edit-shop.vue';

  const searchParams = ref({
    status: '',
    keyword: '',
  });
  const isFetching = ref(false);
  const editLoading = ref(false);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);
  const addEditShow = ref(false);

  const handleSearch = () => {
    isFetching.value = true;
  };

  const handleAdd = () => {
    addEditShow.value = true;
  };

  const handleEdit = (row: any) => {
    console.log(row);
  };

  const handleStatus = (row: any) => {
    console.log(row);
  };

  const handlePageChange = (page: number) => {
    currentPage.value = page;
    handleSearch();
  };
</script>
