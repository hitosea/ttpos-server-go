<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-input v-model="searchParams.version_number" :placeholder="$t('版本号')" clearable style="min-width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.forced_update" :placeholder="$t('强制更新')" clearable style="min-width: 200px">
            <el-option value="1" :label="$t('是')" />
            <el-option value="2" :label="$t('否')" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" class="ti-search-btn" :disabled="isFetching" @click="handleSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
      <!-- <div>
        <el-button v-permission="['admin_shop_add']" type="primary" icon="plus" @click="handleAdd">{{ $t('添加') }}</el-button>
      </div> -->
    </div>

    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list?.data" row-key="id" :default-expand-all="false" border>
      <el-table-column prop="version_name" width="200" :label="$t('版本号')"></el-table-column>
      <el-table-column prop="shop_supplier_name" :label="$t('品牌')">
        <template #default="scope">
          {{ scope.row.brand == '1' ? 'TTPOS' : 'JBCレジ' }}
        </template>
      </el-table-column>
      <el-table-column prop="language_update_log" :label="$t('更新日志')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }">
        <template #default="scope">
          {{ scope.row.language_update_log ? scope.row.language_update_log : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="forced_update" :label="$t('强制更新')">
        <template #default="scope">
          {{ scope.row.forced_update == '0' ? $t('否') : $t('是') }}
        </template>
      </el-table-column>

      <el-table-column prop="create_time" :label="$t('添加時間')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="220">
        <template #default="scope">
          <el-button type="primary" link @click="handleQR(scope.row)">
            {{ $t('二维码') }}
          </el-button>
          <el-button type="primary" link @click="handleDownload(scope.row)">
            {{ $t('下载') }}
          </el-button>
          <el-button v-if="scope.row.is_publish == '1'" link @click="handleDetail(scope.row)">
            {{ $t('查看') }}
          </el-button>
          <el-button v-if="scope.row.is_publish == '0'" link @click="handleAdd(scope.row)">
            {{ $t('发布') }}
          </el-button>
          <!-- <el-button  type="primary" link @click="handleDelete(scope.row)">
            {{ $t('删除') }}
          </el-button> -->
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

    <!-- 新增 -->
    <addForm v-model:show="editShow" :title="title" :detail="editDetail" :type="'6'" @change="handleSearch()" />
    <QRcode v-model:show="qrShow" :uuid="uuid" />
  </div>
</template>
<script setup lang="ts">
  import { type cashierListType, type dataList, getClientList } from '@/api/client';
  import { useQuery } from 'vue-query';
  import { useRoute } from 'vue-router';
  import { ref, watch } from 'vue';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  import addForm from './components/add-form.vue';
  import QRcode from './components/QRcode/index.vue';
  //   import { ElMessageBox } from 'element-plus';
  //   import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';
  //
  const { hasPermission } = useUserInfoStore();
  const route = useRoute();
  const title = ref('');
  const qrShow = ref(false);
  const uuid = ref('');

  const { roleListPath } = storeToRefs(useUserInfoStore());
  const searchParams = ref<cashierListType>({
    type: '6',
    version_number: '',
    forced_update: '',
    page: 1,
    list_rows: 10,
  });
  const editShow = ref(false);
  const editDetail = ref();

  const { isFetching, data, refetch } = useQuery(['getShopList'], () => getClientList(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    // 判断权限
    if (hasPermission(['admin_admin.user_index'])) {
      refetch.value();
    }
  };

  const handleAdd = (row: any) => {
    editDetail.value = row;
    title.value = $t('发布自助点餐机');
    editShow.value = true;
  };

  const handlePageSizeChange = (page: number, list_rows: number) => {
    searchParams.value.page = page || 1;
    searchParams.value.list_rows = list_rows || 15;
    refetch.value();
  };

  const initData = () => {
    searchParams.value = {
      type: '6',
      version_number: '',
      forced_update: '',
      page: 1,
      list_rows: 10,
    };
    //
    handleSearch();
  };

  const handleQR = async (row: dataList) => {
    uuid.value = row.uuid || '';
    qrShow.value = true;
  };
  const handleDownload = async (row: dataList) => {
    let fileURL = window.location.origin + `/api/admin/client.client/download?uuid=${row.uuid}`;
    window.open(fileURL);
    handleSearch();
  };

  const handleDetail = (row: dataList) => {
    editDetail.value = row;
    title.value = $t('查看自助点餐机');
    editShow.value = true;
  };

  //   const handleDelete = (row: dataList) => {
  //     ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
  //       confirmButtonText: $t('确定'),
  //       cancelButtonText: $t('取消'),
  //       type: 'warning',
  //       beforeClose: async (action: string, instance: any, done: () => void) => {
  //         if (action === 'confirm') {
  //           try {
  //             instance.confirmButtonLoading = true;
  //             //
  //             const data = { id: '' };
  //             data.id = row.id || '';
  //             const res = await deleteClient(data);
  //             let msg = res.msg == 'success' ? $t('删除成功') : '';
  //             message.success(msg);
  //             handleSearch();
  //             //
  //             done();
  //           } catch (error) {
  //             //
  //           } finally {
  //             instance.confirmButtonLoading = false;
  //           }
  //         } else {
  //           done();
  //         }
  //       },
  //     })
  //       .then(() => {})
  //       .catch(() => {});
  //   };

  watch(
    () => route?.query?.type,
    () => {
      console.log(route?.query?.type);
      initData();
    },
  );

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
<style lang=""></style>
