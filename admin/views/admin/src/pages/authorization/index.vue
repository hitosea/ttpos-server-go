<template>
  <div class="p-4 bg-white rounded min-h-full">
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-input v-model="searchParams.keyword" :placeholder="$t('商家名称/ID')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
      <div>
        <el-button v-permission="['admin_erpnext_add']" type="primary" icon="plus" @click="handleAdd">{{ $t('授权商家') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table :data="tableData" border v-loading="isFetching">
      <el-table-column prop="uuid" :label="$t('商家ID')" min-width="200"></el-table-column>
      <el-table-column prop="name" :label="$t('商家名称')" min-width="200" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>
      <el-table-column prop="link_phone" :label="$t('联系电话')" min-width="200">
        <template #default="scope">
          {{ scope.row.link_phone || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="real_name" :label="$t('超管用户名')" min-width="200"></el-table-column>
      <el-table-column prop="erpnext_company_abbr" :label="$t('erpnext公司')" min-width="200"></el-table-column>
      <el-table-column prop="erpnext_pos_profile_name" :label="$t('Pos Profile')" min-width="200"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="100">
        <template #default="scope">
          <el-button v-permission="['admin_erpnext_add']" type="primary" link @click="handleEdit(scope.row)">{{ $t('查看') }}</el-button>
          <!-- 支付管理功能暂时关闭, 任务ID37460 -->
          <!-- <el-button type="primary" link @click="handlePay(scope.row)">{{ $t('支付管理') }}</el-button> -->
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination :total="total" :currentPage="currentPage" :pageSize="pageSize" @change="handlePageChange"></ti-pagination>

    <add-edit v-model:show="addEditShow" :has-edit="hasEdit" :editRow="editRow" ref="addEditRef" @refresh="getList" />
    <pay-manage v-model:show="payManageShow" :editRow="editRow" ref="payManageRef" @refresh="getList" />
  </div>
</template>
<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { getAuthorizationList, authorizationListTypeItem } from '@/api/authorization';
  import AddEdit from './components/add-edit.vue';
  import PayManage from './components/pay-manage.vue';
  const addEditRef = ref<InstanceType<typeof AddEdit>>();
  const searchParams = ref({
    keyword: '',
    page: 1,
    list_rows: 10,
    configured: 1,
  });
  const tableData = ref<authorizationListTypeItem[]>([]);
  const isFetching = ref(false);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);
  const addEditShow = ref(false);
  const hasEdit = ref(false);
  const editRow = ref<authorizationListTypeItem>();
  const payManageShow = ref(false);
  const payManageRef = ref<InstanceType<typeof PayManage>>();
  const getList = async () => {
    try {
      const params = {
        keyword: searchParams.value.keyword,
        page: currentPage.value,
        list_rows: pageSize.value,
        configured: searchParams.value.configured,
      };
      isFetching.value = true;
      const res = await getAuthorizationList(params);
      tableData.value = res.data?.list?.data || [];
      total.value = res.data?.list?.total || 0;
      isFetching.value = false;
    } catch (error) {
      isFetching.value = false;
    }
  };

  const handlePageChange = (page: number) => {
    currentPage.value = page;
    getList();
  };

  const handleSearch = () => {
    currentPage.value = 1;
    getList();
  };
  const handleAdd = () => {
    addEditShow.value = true;
    hasEdit.value = false;
  };
  const handleEdit = (row: authorizationListTypeItem) => {
    addEditShow.value = true;
    hasEdit.value = true;
    editRow.value = row;
  };

  //   const handlePay = (row: authorizationListTypeItem) => {
  //     editRow.value = row;
  //     payManageShow.value = true;
  //   };

  onMounted(() => {
    getList();
  });
</script>
<style lang="scss" scoped></style>
