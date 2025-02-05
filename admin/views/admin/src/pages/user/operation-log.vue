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
      <el-table-column prop="user_name" :label="$t('用户名')"></el-table-column>
      <el-table-column prop="real_name" :label="$t('姓名')">
        <template #default="scope">
          {{ scope.row.real_name || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="url" label="URL">
        <template #default="scope">
          <el-input :placeholder="$t('请输入内容')" v-model="scope.row.url" :disabled="true">
            <template v-if="false" #append>
              <el-button icon="Link" @click="handleCopy(scope.row.url)"></el-button>
            </template>
          </el-input>
        </template>
      </el-table-column>
      <el-table-column prop="title" :label="$t('操作内容')">
        <template #default="scope">
          {{ $t(scope.row.title) }}
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP"></el-table-column>
      <el-table-column prop="browser" :label="$t('来源')"></el-table-column>
      <el-table-column prop="create_time" :label="$t('操作时间')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="120">
        <template #default="scope">
          <el-button v-if="hasPermission(['admin_admin.optlog_detail'])" @click="openDetail(scope.row)" type="primary" link>{{ $t('详情') }}</el-button>
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination
      :disabled="isFetching"
      :total="data?.data?.list?.total"
      :currentPage="data?.data?.list?.current_page"
      :pageSize="data?.data?.list?.per_page"
      @change="handlePageSizeChange"
    />
    <!-- 详情 -->
    <el-dialog width="640" :title="$t('详情')" v-model="detailShow" :close-on-press-escape="false" @close="detailShow = false">
      <el-form size="" :model="detailItem" label-position="top" label-width="auto">
        <el-form-item :label="$t('标题：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ $t(detailItem.title || '') }}</div>
        </el-form-item>
        <el-form-item label="id：">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.id }}</div>
        </el-form-item>
        <el-form-item :label="$t('用户名：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.user_name || '-' }}</div>
        </el-form-item>
        <el-form-item :label="$t('真实姓名：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.real_name || '-' }}</div>
        </el-form-item>
        <el-form-item label="url：">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.url }}</div>
        </el-form-item>
        <el-form-item :label="$t('内容：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.content }}</div>
        </el-form-item>
        <el-form-item label="ip：">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.ip }}</div>
        </el-form-item>
        <el-form-item :label="$t('来源：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.agent }}</div>
        </el-form-item>
        <el-form-item :label="$t('添加时间：')">
          <div class="rounded py-2 px-3 bg-[#f6f8fb] w-full">{{ detailItem.create_time }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailShow = false">{{ $t('关闭') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useQuery } from 'vue-query';
  import copy from 'copy-to-clipboard';
  import { OperationLogData, OperationLogType, getOperationLog } from '@/api/user/operation-log';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { storeToRefs } from 'pinia';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const { hasPermission } = useUserInfoStore();
  const { roleListPath } = storeToRefs(useUserInfoStore());
  const searchParams = ref<OperationLogType>({
    username: '',
    page: 1,
    list_rows: 10,
  });
  const detailShow = ref(false);
  const detailItem = ref<OperationLogData>({});

  const { isFetching, data, refetch } = useQuery(['getOperationLog'], () => getOperationLog(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleCopy = (url: string) => {
    copy(url);
    message.success($t('复制成功'));
  };

  const openDetail = (row: OperationLogData) => {
    detailItem.value = row;
    detailShow.value = true;
  };

  const handleSearch = () => {
    searchParams.value.page = 1;
    if (hasPermission(['admin_admin.optlog_index'])) {
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

<style lang="scss" scoped></style>
