<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <el-form :inline="true">
      <el-form-item>
        <el-input v-model="searchParams.username" :placeholder="$t('用户名')" clearable style="min-width: 200px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
      </el-form-item>
    </el-form>
    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list?.data" border>
      <el-table-column prop="id" label="ID"></el-table-column>
      <el-table-column prop="ip" label="IP"></el-table-column>
      <el-table-column prop="result" :label="$t('登录状态')"></el-table-column>
      <el-table-column prop="username" :label="$t('用户名')"></el-table-column>
      <el-table-column prop="create_time" :label="$t('操作时间')"></el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination
      :disabled="isFetching"
      :total="data?.data?.list?.total"
      :currentPage="data?.data?.list?.current_page"
      :pageSize="data?.data?.list?.per_page"
      @change="handlePageSizeChange"
    />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useQuery } from 'vue-query';
  import { LoginLogType, getLoginLog } from '@/api/user/login-log';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';

  const { hasPermission } = useUserInfoStore();
  const { roleListPath } = storeToRefs(useUserInfoStore());
  const searchParams = ref<LoginLogType>({
    username: '',
    page: 1,
    list_rows: 10,
  });

  const { isFetching, data, refetch } = useQuery(['getLoginLog'], () => getLoginLog(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    if (hasPermission(['admin_admin.loginlog_index'])) {
      refetch.value();
    }
  };

  const handlePageSizeChange = (page: number, list_rows: number) => {
    searchParams.value.page = page || 1;
    searchParams.value.list_rows = list_rows || 15;
    refetch.value();
  };

  const initData = () => {
    searchParams.value = {
      username: '',
      page: 1,
      list_rows: 10,
    };
    //
    handleSearch();
  };

  watch(
    () => roleListPath.value,
    () => {
      if (roleListPath.value && roleListPath.value.length > 0) initData();
    },
    {
      deep: true,
      immediate: true,
    },
  );
</script>
