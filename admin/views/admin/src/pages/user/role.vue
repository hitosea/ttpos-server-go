<template>
  <div class="p-4 bg-white rounded min-h-full">
    <div class="flex justify-end pb-4">
      <el-button v-permission="['admin_admin.role_add']" type="primary" icon="Plus" :loading="!roleId && roleLoading" @click="handleAdd">{{ $t('添加角色') }} </el-button>
    </div>
    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list" row-key="id" border>
      <el-table-column prop="id" label="ID"></el-table-column>
      <el-table-column prop="role_name" :label="$t('角色名称')"></el-table-column>
      <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="150">
        <template #default="scope">
          <el-button
            v-permission="['admin_admin.role_edit']"
            :loading="(editLoading && adminId == scope.row.id) || (scope.row.id == roleId && roleLoading)"
            type="primary"
            link
            @click="handleEdit(scope.row)"
          >
            {{ $t('编辑') }}
          </el-button>
          <el-button v-permission="['admin_admin.role_delete']" :disabled="scope.row.is_used == 1" type="primary" link @click="handleDelete(scope.row)">
            {{ $t('删除') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <!-- <ti-pagination
      :disabled="isFetching"
      :total="data?.data?.list?.total"
      :currentPage="data?.data?.list?.current_page"
      :pageSize="data?.data?.list?.per_page"
      @change="handlePageSizeChange"
    /> -->
    <!-- 新增、编辑 -->
    <role-form v-model:show="formShow" :detail="formDetail" :roleChecked="roleChecked" :roleList="roleList" @change="initData" />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useQuery } from 'vue-query';
  import { ElMessageBox } from 'element-plus';
  import { message } from '@/utils/feedback';
  import { RoleListType, RoleListData, getRoleList, fetchDelete, getRoleAdd, getRoleEdit } from '@/api/user/role';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { $t } from '@/i18n';
  //
  import RoleForm from './components/role-form.vue';

  const { hasPermission } = useUserInfoStore();
  const { roleListPath } = storeToRefs(useUserInfoStore());
  const searchParams = ref<RoleListType>({
    keyword: '',
    page: 1,
    list_rows: 10,
  });
  const adminId = ref();
  const editLoading = ref(false);
  const formShow = ref(false);
  const formDetail = ref<RoleListData>();
  //
  const roleId = ref<number>();
  const roleLoading = ref(false);
  const roleList = ref([]);
  const roleChecked = ref([]);

  const { isFetching, data, refetch } = useQuery(['getRoleList'], () => getRoleList(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    if (hasPermission(['admin_admin.role_index'])) {
      refetch.value();
    }
  };

  const handleAdd = async () => {
    try {
      roleId.value = undefined;
      roleLoading.value = true;
      const res = await getRoleAdd();
      roleList.value = res.data?.menu || [];
      (roleList.value || []).map((item: any) => {
        item.name = $t(item.name);
        (item.children || []).map((items: any) => {
          items.name = $t(items.name);
          (items.children || []).map((child: any) => {
            child.name = $t(child.name);
            (child.children || []).map((children: any) => {
              children.name = $t(children.name);
            });
          });
        });
      });
      roleChecked.value = [];
      formDetail.value = {};
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      roleLoading.value = false;
    }
  };

  const handleEdit = async (row: RoleListData) => {
    try {
      roleId.value = row?.id;
      roleLoading.value = true;
      const res = await getRoleEdit(row?.id);
      roleList.value = res.data?.menu || [];
      (roleList.value || []).map((item: any) => {
        item.name = $t(item.name);
        (item.children || []).map((items: any) => {
          items.name = $t(items.name);
          (items.children || []).map((child: any) => {
            child.name = $t(child.name);
            (child.children || []).map((children: any) => {
              children.name = $t(children.name);
            });
          });
        });
      });
      roleChecked.value = res.data?.select_menu || [];
      formDetail.value = row;
      formShow.value = true;
    } catch (error) {
      //
    } finally {
      roleLoading.value = false;
    }
  };

  const handleDelete = (row: RoleListData) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
      beforeClose: async (action: string, instance: any, done: () => void) => {
        if (action === 'confirm') {
          try {
            instance.confirmButtonLoading = true;
            //
            const res = await fetchDelete(row.id);
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

  // const handlePageSizeChange = (page: number, list_rows: number) => {
  //   searchParams.value.page = page || 1;
  //   searchParams.value.list_rows = list_rows || 15;
  //   refetch.value();
  // };

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
