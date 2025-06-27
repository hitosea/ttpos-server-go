<template>
  <div class="p-4 bg-white rounded min-h-full">
    <!-- 查询 -->
    <div class="flex justify-between">
      <el-form :inline="true">
        <el-form-item>
          <el-select v-model="searchParams.channel" :placeholder="$t('外送渠道')" clearable style="min-width: 200px">
            <el-option value=" " :label="$t('全部渠道')" />
            <el-option v-for="channel in channels" :key="channel.value" :value="channel.value" :label="channel.name" />
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
        <el-button class="!mr-0" v-permission="['admin_delivery_company']" type="primary" icon="plus" @click="handleAdd">{{ $t('添加商家') }}</el-button>
      </div>
    </div>
    <!-- 表格 -->
    <el-table :data="tableData" border v-loading="isFetching">
      <el-table-column prop="uuid" :label="$t('商家ID')" min-width="200"></el-table-column>
      <el-table-column prop="name" :label="$t('商家名称')" min-width="200" :show-overflow-tooltip="{ popperClass: 'max-w-[220px]' }"></el-table-column>
      <el-table-column prop="channel_names" :label="$t('外送渠道')" min-width="200">
        <template #default="scope">
          {{ scope.row.channel_names.join(',') }}
        </template>
      </el-table-column>
      <el-table-column prop="channel_config_types" :label="$t('参数设置')" min-width="200">
        <template #default="scope">
          <!-- auto_sync 自动同步 -->
          {{ scope.row.channel_config_types.map((item: string) => (item == 'auto_sync' ? $t('自动同步') : $t('手动同步'))).join(',') }}
        </template>
      </el-table-column>
      <el-table-column prop="delivery_status" :label="$t('状态')" min-width="120">
        <template #default="scope">
          <el-switch
            v-permission="['admin_shop_updateStatus']"
            :model-value="scope.row.delivery_status"
            :loading="editLoading"
            :active-value="1"
            :inactive-value="0"
            @change="handleStatus(scope.row)"
          ></el-switch>
        </template>
      </el-table-column>
      <el-table-column fixed="right" :label="$t('操作')" width="150">
        <template #default="scope">
          <el-button type="primary" v-permission="['admin_delivery_company']" link @click="handleEdit(scope.row)">
            {{ $t('编辑') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <!-- 分页 -->
    <ti-pagination :total="total" :currentPage="currentPage" :pageSize="pageSize" @change="handlePageChange"></ti-pagination>

    <add-edit-shop v-model:show="addEditShow" :has-edit="hasEdit" :edit-data="editData" @getList="handleSearch" />
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { $t } from '@/i18n';
  import AddEditShop from './components/add-edit-shop.vue';
  import { getChannels, getCompanyList, updateStatus, type TakeoutShopData, type TakeoutShopListDataItem } from '@/api/takeout/shop';
  import { message } from '@/utils/feedback';

  const searchParams = ref({
    channel: '',
    keyword: '',
  });

  const isFetching = ref(false);
  const editLoading = ref(false);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);
  const addEditShow = ref(false);
  const hasEdit = ref(false);
  const editData = ref({});

  const channels = ref<TakeoutShopData[]>([]);

  const tableData = ref<TakeoutShopListDataItem[]>([]);

  const handleSearch = () => {
    isFetching.value = true;
    getCompanyListData();
  };

  const handleAdd = () => {
    addEditShow.value = true;
    hasEdit.value = false;
  };

  const handleEdit = (row: any) => {
    editData.value = row;
    hasEdit.value = true;
    addEditShow.value = true;
  };

  const handleStatus = async (row: any) => {
    try {
      editLoading.value = true;
      const res = await updateStatus(row.uuid);
      message.success(res.msg);
      handleSearch();
    } catch (error) {
      message.error($t('操作失败'));
    } finally {
      editLoading.value = false;
    }
  };

  const handlePageChange = (page: number) => {
    currentPage.value = page;
    handleSearch();
  };

  const getChannelsData = async () => {
    try {
      const res = await getChannels({ configured: 0 });
      channels.value = res.data || [];
    } catch (error) {
      console.error('获取渠道失败:', error);
    }
  };

  const getCompanyListData = async () => {
    isFetching.value = true;
    try {
      const params = {
        channel: searchParams.value.channel,
        keyword: searchParams.value.keyword,
        page: currentPage.value,
        list_rows: pageSize.value,
      };
      const res = await getCompanyList(params);
      tableData.value = res.data?.list?.data || [];
      total.value = res.data?.list?.total || 0;
    } catch (error) {
      console.error('获取渠道失败:', error);
    } finally {
      isFetching.value = false;
    }
  };

  onMounted(() => {
    getChannelsData();
    getCompanyListData();
  });
</script>
