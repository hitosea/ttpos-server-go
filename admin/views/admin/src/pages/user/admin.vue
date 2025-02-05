<template>
  <div class="p-4 bg-white rounded min-h-full">
    <div class="flex justify-between">
      <!-- 查询 -->
      <el-form :inline="true">
        <!-- <el-form-item>
          <el-input v-model="searchParams.keyword" :placeholder="$t('用户名/姓名/用户ID')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item> -->
      </el-form>
      <div class="pb-4">
        <el-button v-permission="['admin_admin.user_add']" type="primary" icon="Plus" :loading="!roleId && roleLoading" @click="handleAdd">{{ $t('添加管理员') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list?.data" row-key="admin_user_id" border>
      <el-table-column prop="admin_user_id" label="ID"></el-table-column>
      <el-table-column prop="user_name" :label="$t('用户名')"></el-table-column>
      <el-table-column prop="real_name" :label="$t('姓名')"></el-table-column>
      <el-table-column prop="userRole" :label="$t('角色')">
        <template #default="scope">
          <div v-if="scope.row.is_super == 1">{{ $t('超级管理员') }}</div>
          <div v-else-if="scope.row?.userRole && scope.row?.userRole.length > 0" class="flex flex-wrap gap-2">
            <el-tag v-for="(item, index) in scope.row?.userRole" :key="index" type="info">{{ item?.role?.role_name }}</el-tag>
          </div>
          <div v-else>-</div>
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('状态')">
        <template #default="scope">
          <el-switch
            v-if="hasPermission(['admin_admin.user_updateStatus'])"
            :loading="statusLoading && adminId == scope.row.admin_user_id"
            :disabled="scope.row.is_super == 1"
            :model-value="scope.row.status"
            :active-value="1"
            :inactive-value="0"
            @change="handleStatus(scope.row)"
          ></el-switch>
          <span v-else>{{ scope.row.status == 1 ? $t('开启') : $t('关闭') }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="150">
        <template #default="scope">
          <el-button
            v-permission="['admin_admin.user_edit']"
            :disabled="scope.row.is_super == 1"
            :loading="roleId == scope.row.admin_user_id && roleLoading"
            type="primary"
            link
            @click="handleEdit(scope.row)"
          >
            {{ $t('编辑') }}
          </el-button>
          <el-button v-permission="['admin_admin.user_delete']" :disabled="scope.row.is_super == 1" type="primary" link @click="handleDelete(scope.row)">
            {{ $t('删除') }}
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
    <admin-form v-model:show="formShow" :detail="formDetail" :roleList="roleList" @change="initData" />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useQuery } from 'vue-query';
  import { ElMessageBox } from 'element-plus';
  import { message } from '@/utils/feedback';
  import { AdminListData, AdminListType, getAdminList, fetchUpdateStatus, fetchDelete } from '@/api/user/admin';
  import { getRoleList } from '@/api/user/role';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { storeToRefs } from 'pinia';
  import { $t } from '@/i18n';
  //
  import AdminForm from './components/admin-form.vue';

  const { hasPermission } = useUserInfoStore();
  const { roleListPath } = storeToRefs(useUserInfoStore());
  const searchParams = ref<AdminListType>({
    keyword: '',
    page: 1,
    list_rows: 10,
  });
  const adminId = ref();
  const statusLoading = ref(false);
  const formShow = ref(false);
  const formDetail = ref<AdminListData>();
  const roleId = ref<number>();
  const roleList = ref<any>([]);
  const roleLoading = ref<boolean>(false);

  const { isFetching, data, refetch } = useQuery(['getAdminList'], () => getAdminList(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    if (hasPermission(['admin_admin.user_index'])) {
      refetch.value();
    }
  };

  const handleStatus = async (row: AdminListData) => {
    try {
      statusLoading.value = true;
      adminId.value = row.admin_user_id;
      const res = await fetchUpdateStatus(row.admin_user_id);
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
      roleId.value = undefined;
      roleLoading.value = true;
      const res = await getRoleList({ page: 1, list_rows: 1000 });
      roleList.value = res.data?.list || [];
      formDetail.value = {};
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      roleLoading.value = false;
    }
  };

  const handleEdit = async (row: AdminListData) => {
    try {
      roleId.value = row.admin_user_id;
      roleLoading.value = true;
      const res = await getRoleList({ page: 1, list_rows: 1000 });
      roleList.value = res.data?.list || [];
      formDetail.value = row;
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      roleLoading.value = false;
    }
  };

  const handleDelete = (row: AdminListData) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
      beforeClose: async (action: string, instance: any, done: () => void) => {
        if (action === 'confirm') {
          try {
            instance.confirmButtonLoading = true;
            //
            const res = await fetchDelete(row.admin_user_id);
            message.success(res.msg);
            handleSearch();
            //
            done();
          } catch (error) {
            //
          } finally {
            instance.confirmButtonLoading = false;
          }
        } else {
          done();
        }
      },
    })
      .then(() => {})
      .catch(() => {});
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
