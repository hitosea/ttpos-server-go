<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-select v-model="searchParams.status" :placeholder="$t('商家状态')" clearable style="min-width: 200px">
            <el-option value=" " :label="$t('全部状态')" />
            <el-option :value="1" :label="$t('开启')" />
            <el-option :value="2" :label="$t('关闭')" />
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
        <el-button v-permission="['admin_shop_add']" type="primary" icon="plus" @click="handleAdd">{{ $t('添加商家') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table v-loading="isFetching" :data="data?.data?.list?.data" row-key="app_id" :default-expand-all="false" :tree-props="{ children: 'suppliers' }" border>
      <el-table-column prop="app_id" width="200" :label="$t('商家ID')"></el-table-column>
      <el-table-column prop="shop_supplier_name" :label="$t('商家名称')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>

      <el-table-column prop="status" :label="$t('状态')">
        <template #default="scope">
          <el-switch
            v-if="hasPermission(['admin_shop_updateStatus'])"
            :model-value="scope.row.status"
            :loading="statusLoading && shopId == scope.row.id"
            :active-value="1"
            :inactive-value="0"
            @click="handleStatus(scope.row)"
          ></el-switch>
          <span v-else>{{ scope.row.status == 1 ? $t('开启') : $t('关闭') }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="sale_stock" :label="$t('进销存')">
        <template #default="scope">
          <el-switch
            v-if="hasPermission(['admin_shop_updateSaleStock'])"
            :model-value="scope.row.sale_stock"
            :loading="saleStockLoading && shopId == scope.row.id"
            :active-value="1"
            :inactive-value="0"
            @click="handleSaleStock(scope.row)"
          ></el-switch>
          <span v-else>{{ scope.row.status == 1 ? $t('开启') : $t('关闭') }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="link_phone" :label="$t('联系电话')"></el-table-column>
      <el-table-column prop="user_name" :label="$t('邮箱')"></el-table-column>
      <el-table-column prop="address" :label="$t('所在地')" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>
      <el-table-column prop="expire_time" :label="$t('过期时间')">
        <template #default="scope">
          <span v-if="scope.row.expire_time == 0">{{ $t('永不过期') }}</span>
          <span v-else>{{ scope.row.expire_time_text }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="create_time" :label="$t('添加時間')"></el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="200">
        <template #default="scope">
          <el-button v-permission="['admin_shop_payment']" :loading="editLoading && shopId == scope.row.id" type="primary" link @click="handlePay(scope.row)">
            {{ $t('支付') }}
          </el-button>
          <el-button v-permission="['admin_shop_edit']" :loading="editLoading && shopId == scope.row.id" type="primary" link @click="handleEdit(scope.row)">
            {{ $t('编辑') }}
          </el-button>
          <el-button type="primary" link @click="handleDelete(scope.row)">
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
    <!-- 授权码 -->
    <dialog-license v-model:show="licenseShow" :detail="licenseDetail" />
    <!-- 新增、编辑 -->
    <dialog-edit v-model:show="editShow" :detail="editDetail" @change="handleSearch()" @close="handleClose()" />
    <!-- 支付 -->
    <dialog-pay v-model:show="payShow" :detail="payDetail" />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import { useQuery } from 'vue-query';
  import { ElMessageBox } from 'element-plus';
  import { ShopListType, ShopListData, getShopList, fetchShopDelete, fetchUpdateStatus, fetchUpdateSaleStock, getShowEdit } from '@/api/merchant';
  import { storeToRefs } from 'pinia';
  import { useUserInfoStore } from '@/stores/userInfo';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';
  //
  import DialogLicense from './components/dialog-license.vue';
  import DialogEdit from './components/dialog-edit.vue';
  import DialogPay from './components/dialog-pay.vue';
  //

  const { hasPermission } = useUserInfoStore();
  const { roleListPath } = storeToRefs(useUserInfoStore());
  const route = useRoute();
  const searchParams = ref<ShopListType>({
    keyword: '',
    shop_type: 0,
    status: 0,
    page: 1,
    list_rows: 10,
  });
  const shopId = ref();
  const statusLoading = ref(false);
  const editLoading = ref(false);
  const saleStockLoading = ref(false);
  //   const reserveLoading = ref(false);
  const licenseShow = ref(false);
  const licenseDetail = ref();
  const editShow = ref(false);
  const payShow = ref(false);
  const editDetail = ref();
  const payDetail = ref();
  const { isFetching, data, refetch } = useQuery(['getShopList'], () => getShopList(searchParams.value), {
    enabled: false,
    // 一般情况分页查询，表格分页切换这种场景会用到，保证换页时不会出现一个空列表
    keepPreviousData: true,
  });

  const handleSearch = () => {
    searchParams.value.page = 1;
    // 判断权限
    if (hasPermission(['admin_shop_index'])) {
      refetch.value();
    }
  };

  const handleClose = () => {
    // 判断权限
    if (hasPermission(['admin_shop_index'])) {
      refetch.value();
    }
  };

  const handleAdd = () => {
    editDetail.value = {};
    editShow.value = true;
  };

  const handleEdit = async (row: ShopListData) => {
    try {
      editLoading.value = true;
      const res = await getShowEdit(row.app_id);
      editDetail.value = res.data.model;
      editShow.value = true;
    } catch (error) {
      //
    } finally {
      editLoading.value = false;
    }
  };

  const handlePay = async (row: ShopListData) => {
    payDetail.value = row;
    payShow.value = true;
  };

  const handleDelete = (row: ShopListData) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
      beforeClose: async (action: string, instance: any, done: () => void) => {
        if (action === 'confirm') {
          try {
            instance.confirmButtonLoading = true;
            //
            const res = await fetchShopDelete(row.id);
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

  const handleStatus = async (row: ShopListData) => {
    try {
      statusLoading.value = true;
      shopId.value = row.id;
      const res = await fetchUpdateStatus(row.id);
      message.success(res.msg);
      //
      refetch.value();
    } catch (error) {
      //
    } finally {
      statusLoading.value = false;
    }
  };

  const handleSaleStock = async (row: ShopListData) => {
    try {
      saleStockLoading.value = true;
      shopId.value = row.id;
      const res = await fetchUpdateSaleStock(row.id);
      message.success(res.msg);
      //
      refetch.value();
    } catch (error) {
      //
    } finally {
      saleStockLoading.value = false;
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
    //
    handleSearch();
  };

  watch(
    () => route?.query?.type,
    () => {
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

<style lang="scss" scoped></style>
