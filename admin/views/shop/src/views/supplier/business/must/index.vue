<template>
  <div class="must-list">
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('状态')">
          <a-select size="small" v-model:value="searchForm.status" :placeholder="$t('方案状态')" @change="onSearch">
            <el-option :label="$t('开启')" value="1"></el-option>
            <el-option :label="$t('关闭')" value="0"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('方案名称')">
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('请输入方案名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <el-button size="small" type="primary" icon="Plus" v-auth="'/supplier/business/must/add'" @click="addClick">{{ $t('添加') }}</el-button>
    </div>

    <div class="common-table-wrap">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="name_text" :label="$t('方案名称')" width="200" />
          <el-table-column prop="use_channel" :label="$t('使用渠道')" width="180">
            <template #default="scope">
              <span v-if="scope.row.use_channel.indexOf('10') != -1">{{ $t('点餐方式') }}</span>
              <span v-if="scope.row.use_channel.indexOf('10') != -1 && scope.row.use_channel.indexOf('20') != -1">、</span>
              <span v-if="scope.row.use_channel.indexOf('20') != -1">{{ $t('桌台方式') }}</span>
              <span v-if="(scope.row.use_channel || []).length == 0">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="must_type" :label="$t('必点类型')" width="180">
            <template #default="scope">
              {{ scope.row.must_type == 1 ? $t('每人必点1份') : $t('每笔订单必点1份') }}
            </template>
          </el-table-column>
          <el-table-column prop="must_rule" :label="$t('必点规则')">
            <template #default="scope">
              {{ scope.row.must_rule == 1 ? $t('固定商品') : $t('可选商品') }}
            </template>
          </el-table-column>
          <el-table-column prop="attribute_name_text" :label="$t('状态')" width="120">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/supplier/business/must/status')"
                :model-value="scope.row.status == 1 ? true : false"
                @click="handleChange(scope.row)"
                :loading="loading"
              />
            </template>
          </el-table-column>
          <el-table-column prop="area_names" :label="$t('桌台区域')" show-overflow-tooltip>
            <template #default="scope">
              {{ scope.row.area_names || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')" />
          <el-table-column :label="$t('操作')" fixed="right" width="160">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/supplier/business/must/edit'">{{ $t('编辑') }}</el-button>
              <el-button @click="deleteClick(scope.row)" type="primary" link size="small" v-auth="'/supplier/business/must/del'">{{ $t('删除') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    <div class="pagination">
      <el-pagination
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
        background
        :current-page="pagination.page"
        :page-size="pagination.pageSize"
        layout="total, prev, pager, next, jumper"
        :total="pagination.total"
      ></el-pagination>
    </div>
    <addEdit v-if="addEditOpen" :open="addEditOpen" :editData="editData" :area_list="tableAreaList" @close="handleClose"></addEdit>
  </div>
</template>
<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import addEdit from './addEdit.vue';
  import SettingApi from '@/api/setting.js';
  const { proxy } = getCurrentInstance();
  const searchForm = ref({
    status: '',
    name: '',
  });
  const pagination = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
  });
  const tableData = ref([]);
  const tableAreaList = ref([]);
  const loading = ref(false);
  const addEditOpen = ref(false);
  const editData = ref({});
  const searchLoading = ref(null);

  const getData = () => {
    const params = {
      page: pagination.page,
      list_rows: pagination.pageSize,
      ...searchForm.value,
    };
    loading.value = true;
    SettingApi.orderSchemeList(params, true)
      .then((data) => {
        loading.value = false;
        tableData.value = data.data.list.data;
        tableAreaList.value = data.data.table_area_list.data;
        pagination.total = data.data.list.total;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const handleChange = (row) => {
    if (!proxy.$filter.isAuth('/supplier/business/must/status')) {
      return;
    }
    let war = $t('确认要禁用吗?');
    if (row.status == 0) {
      war = $t('确认要启用吗?');
    }
    ElMessageBox.confirm(war, $t('提示'), { type: 'warning' }).then(() => {
      loading.value = true;
      let Params = {};
      Params.id = row.id;
      Params.status = row.status == 1 ? 0 : 1;
      SettingApi.orderSchemeStatus(Params, true)
        .then((data) => {
          proxy.$ElMessage({
            message: $t('操作成功'),
            type: 'success',
          });
          loading.value = false;
          getData();
        })
        .catch((error) => {
          loading.value = false;
        });
    });
  };

  const editClick = (row) => {
    editData.value = row;
    addEditOpen.value = true;
  };

  const deleteClick = (row) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(() => {
        loading.value = true;
        SettingApi.orderSchemeDel(
          {
            id: row.id,
          },
          true
        )
          .then((data) => {
            loading.value = false;
            proxy.$ElMessage({
              message: $t('操作成功'),
              type: 'success',
            });
            getData();
          })
          .catch((error) => {
            loading.value = false;
          });
      })
      .catch(() => {});
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      pagination.page = 1;
      getData();
    }, 200);
  };

  const addClick = () => {
    addEditOpen.value = true;
  };

  const indexMethod = (index) => {
    return index + 1 + (pagination.page - 1) * pagination.pageSize;
  };

  const handlePageSizeChange = (size) => {
    pagination.page = 1;
    pagination.pageSize = size;
    getData();
  };

  const handlePageChange = (page) => {
    pagination.page = page;
    getData();
  };

  const handleClose = (isRefresh) => {
    addEditOpen.value = false;
    editData.value = {};
    if (isRefresh) {
      onSearch();
    }
  };

  onMounted(() => {
    getData();
  });
</script>
<style lang="scss" scoped>
  .must-list {
    overflow: auto;
    height: 100%;
  }
  .common-seach-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>
