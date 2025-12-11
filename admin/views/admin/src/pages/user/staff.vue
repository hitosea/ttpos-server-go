<template>
  <div class="p-4 bg-white rounded min-h-full">
    <div class="flex justify-between">
      <!-- 查询 -->
      <el-form :inline="true">
        <el-form-item>
          <el-input v-model="searchParams.keyword" :placeholder="$t('邮箱/手机号/员工ID')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
      <div class="pb-4">
        <el-button type="primary" icon="Plus" :loading="formLoading" @click="handleAdd" :disabled="true">{{ $t('添加账号') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list?.data" row-key="uuid" border>
      <el-table-column prop="uuid" label="ID"></el-table-column>
      <el-table-column prop="email" :label="$t('邮箱')"></el-table-column>
      <el-table-column prop="phone" :label="$t('手机号')">
        <template #default="scope">
          {{ scope.row.phone || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="real_name" :label="$t('姓名')"></el-table-column>
      <el-table-column prop="company_list" :label="$t('关联门店和角色')" width="400">
        <template #default="scope">
          <div v-if="scope.row?.company_list && scope.row.company_list.length > 0" class="flex flex-col gap-2">
            <div v-for="(company, index) in scope.row.company_list" :key="index" class="flex items-center gap-2">
              <el-tag type="primary" size="small">{{ company.company_name }}</el-tag>
              <div v-if="company.roles && company.roles.length > 0" class="flex flex-wrap gap-1">
                <el-tag v-for="(role, roleIndex) in company.roles" :key="roleIndex" type="info" size="small">
                  {{ role.role_name }}
                </el-tag>
              </div>
              <span v-else class="text-gray-400 text-xs">{{ $t('无角色') }}</span>
            </div>
          </div>
          <div v-else>-</div>
        </template>
      </el-table-column>
      <el-table-column prop="is_disable" :label="$t('状态')">
        <template #default="scope">
          <el-switch
            :loading="statusLoading && staffId == scope.row.uuid"
            :disabled="true"
            :model-value="scope.row.is_disable === 0 ? 1 : 0"
            :active-value="1"
            :inactive-value="0"
            @change="handleStatus(scope.row)"
          ></el-switch>
        </template>
      </el-table-column>
      <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="150">
        <template #default="scope">
          <el-button type="primary" link :disabled="true" :loading="formLoading && staffId == scope.row.uuid" @click="handleEdit(scope.row)">
            {{ $t('编辑') }}
          </el-button>
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
    <!-- 新增、编辑 -->
    <staff-form :show="formShow" :detail="formDetail" :companyList="companyList" :roleList="roleList" @update:show="formShow = $event" @change="initData" />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useQuery } from 'vue-query';
  import { message } from '@/utils/feedback';
  import { StaffListData, StaffListType, getStaffList, fetchUpdateStaffStatus } from '@/api/user/staff';
  import { getRoleList } from '@/api/user/role';
  import { getShopList } from '@/api/merchant/index';
  import { $t } from '@/i18n';
  //
  import StaffForm from './components/staff-form.vue';

  const searchParams = ref<StaffListType>({
    keyword: '',
    page: 1,
    list_rows: 10,
  });
  const staffId = ref<number>();
  const statusLoading = ref(false);
  const formShow = ref(false);
  const formDetail = ref<StaffListData>();
  const formLoading = ref<boolean>(false);
  const companyList = ref<any[]>([]);
  const roleList = ref<any[]>([]);

  const { isFetching, data, refetch } = useQuery(['getStaffList'], () => getStaffList(searchParams.value), {
    enabled: false,
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    refetch.value();
  };

  const handleStatus = async (row: StaffListData) => {
    try {
      statusLoading.value = true;
      staffId.value = row.uuid;
      const res = await fetchUpdateStaffStatus(row.uuid);
      message.success(res.msg);
      refetch.value();
    } catch (error) {
      //
    } finally {
      statusLoading.value = false;
    }
  };

  const handleAdd = async () => {
    try {
      formLoading.value = true;
      staffId.value = undefined;
      
      // 获取门店列表
      const shopRes = await getShopList({ page: 1, list_rows: 1000 });
      companyList.value = shopRes.data?.list?.data || [];
      
      // 获取角色列表（这里需要获取所有门店的角色，暂时使用第一个门店的角色）
      // 注意：实际应该根据选择的门店动态获取角色列表
      const roleRes = await getRoleList({ page: 1, list_rows: 1000 });
      roleList.value = roleRes.data?.list?.data || [];
      
      formDetail.value = {};
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      formLoading.value = false;
    }
  };

  const handleEdit = async (row: StaffListData) => {
    try {
      formLoading.value = true;
      staffId.value = row.uuid;
      
      // 获取门店列表
      const shopRes = await getShopList({ page: 1, list_rows: 1000 });
      companyList.value = shopRes.data?.list?.data || [];
      
      // 获取角色列表
      const roleRes = await getRoleList({ page: 1, list_rows: 1000 });
      roleList.value = roleRes.data?.list?.data || [];
      
      formDetail.value = row;
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      formLoading.value = false;
    }
  };

  const handlePageSizeChange = (page: number, list_rows: number) => {
    searchParams.value.page = page || 1;
    searchParams.value.list_rows = list_rows || 15;
    refetch.value();
  };

  const initData = () => {
    searchParams.value = {
      keyword: '',
      page: 1,
      list_rows: 10,
    };
    handleSearch();
  };

  // 初始化加载数据
  handleSearch();
</script>

<style lang="scss" scoped></style>
