<template>
  <div class="buffet-list">
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('状态')">
          <a-select size="small" v-model:value="searchForm.status" :placeholder="$t('全部状态')" clearable @change="onSearch">
            <el-option :label="$t('全部状态')" value=""></el-option>
            <el-option :label="$t('开启')" value="1"></el-option>
            <el-option :label="$t('关闭')" value="0"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('自助餐名称')">
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('请输入自助餐名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <el-button size="small" type="primary" icon="Plus" v-auth="'/product/buffet/list/add'" @click="addClick">{{ $t('添加自助餐') }}</el-button>
    </div>

    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="name_text" :label="$t('自助餐名称')" width="400px">
            <template #default="scope">
              <div class="product-info">
                <div class="info">
                  <div class="name">{{ scope.row.name_text }}</div>
                  <div class="price"> {{ $t('销售价：') }}{{ proxy.$formatPrice(scope.row.buffetCustomerType[0]?.price || 0) }} </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="sale_num" :label="$t('实际销量')"></el-table-column>
          <el-table-column prop="time_limit" :label="$t('用餐时间')">
            <template #default="scope">
              <div class="name">
                {{ scope.row.time_limit == 0 ? $t('不限制') : scope.row.time_limit }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="open_overall_discount" :label="$t('整单折扣')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/buffet/list/overallDiscount')"
                :model-value="scope.row.open_overall_discount == 1 ? true : false"
                @click="handleOpenOverallDiscount(scope.row)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="is_comb" :label="$t('组合')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/buffet/list/assembly')"
                :model-value="scope.row.is_comb == 1 ? true : false"
                @click="handleComb(scope.row, scope.row.is_comb == 1 ? 0 : 1)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="status" :label="$t('状态')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/buffet/list/status')"
                :model-value="scope.row.status == 1 ? true : false"
                @click="handleStatus(scope.row)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')" width="180"></el-table-column>
          <el-table-column prop="sort" :label="$t('排序')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="120">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/buffet/list/edit'">{{ $t('编辑') }}</el-button>
              <el-button @click="deleteClick(scope.row)" :disabled="scope.row.can_delete == 0" type="primary" link size="small" v-auth="'/product/buffet/list/delete'">
                {{ $t('删除') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    <!--分页-->
    <div class="pagination">
      <el-pagination
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        background
        :current-page="curPage"
        :page-size="pageSize"
        layout="total, prev, pager, next, jumper"
        :total="totalDataNumber"
      ></el-pagination>
    </div>
    <!--添加-->
    <addEdit v-if="open_dialog" :title="title" :open_dialog="open_dialog" :editData="editData" @closeDialog="closeDialogFunc($event)"> </addEdit>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import addEdit from './addEdit.vue';
  import PorductApi from '@/api/product.js';

  // 获取实例
  const { proxy } = getCurrentInstance();

  // 响应式数据
  const loading = ref(false);
  const searchForm = reactive({
    status: '',
    name: '',
  });
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const tableData = ref([]);
  const open_dialog = ref(false);
  const title = ref('');
  const editData = ref('');
  const searchLoading = ref('');

  // 搜索查询（防抖）
  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getData();
    }, 200);
  };

  // 获取数据列表
  const getData = async () => {
    loading.value = true;
    try {
      const Params = {
        name: searchForm.name,
        status: searchForm.status,
        page: curPage.value,
        list_rows: pageSize.value,
      };
      const data = await PorductApi.getBuffetList(Params, true);
      loading.value = false;
      tableData.value = data.data.list.data;
      totalDataNumber.value = data.data.list.total;
    } catch (error) {
      loading.value = false;
      console.log(error);
    }
  };

  // 新增
  const addClick = () => {
    title.value = $t('添加自助餐');
    open_dialog.value = true;
  };

  // 编辑
  const editClick = (row) => {
    title.value = $t('编辑自助餐');
    editData.value = row;
    open_dialog.value = true;
  };

  // 删除
  const deleteClick = (row) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(async () => {
        try {
          loading.value = true;
          const resp = await PorductApi.deleteBuffet({ buffet_id: row.id }, true);
          loading.value = false;
          if (resp.code == 1) {
            ElMessage({ message: resp.msg, type: 'success' });
            getData();
          } else {
            ElMessage.error($t('操作失败'));
          }
        } catch (e) {
          loading.value = false;
        }
      })
      .catch(() => {});
  };

  // 改变组合
  const handleComb = (row) => {
    if (!proxy.$filter.isAuth('/product/buffet/list/assembly')) return;
    let war = '';
    let war_ = '';
    if (row.is_comb == 1) {
      war = $t('确认要关闭组合吗?');
      war_ = $t('关闭');
    } else if (row.is_comb == 0) {
      war = $t('确认要开启组合吗?');
      war_ = $t('开启');
    }
    ElMessageBox.confirm(war, $t('提示'), { type: 'warning' }).then(async () => {
      await PorductApi.combBuffet({ buffet_id: row.id, is_comb: row.is_comb == 1 ? 0 : 1 });
      ElMessage({ message: war_ + $t('成功'), type: 'success' });
      getData();
    });
  };

  // 改变状态
  const handleStatus = (row) => {
    if (!proxy.$filter.isAuth('/product/buffet/list/status')) return;
    let war = '';
    let war_ = '';
    if (row.status == 1) {
      war = $t('确认要强制下架吗?');
      war_ = $t('下架');
    } else if (row.status == 0) {
      war = $t('确认要重新上架吗?');
      war_ = $t('上架');
    }
    ElMessageBox.confirm(war, $t('提示'), { type: 'warning' }).then(async () => {
      await PorductApi.stateBuffet({ buffet_id: row.id, state: row.status == 1 ? 0 : 1 });
      ElMessage({ message: war_ + $t('成功'), type: 'success' });
      getData();
    });
  };

  // 整单折扣
  const handleOpenOverallDiscount = (row) => {
    if (!proxy.$filter.isAuth('/product/buffet/list/overallDiscount')) return;
    let war = '';
    let war_ = '';
    if (row.open_overall_discount == 1) {
      war = $t('确认要关闭整单折扣吗?');
      war_ = $t('关闭');
    } else if (row.open_overall_discount == 0) {
      war = $t('确认要开启整单折扣吗?');
      war_ = $t('开启');
    }
    ElMessageBox.confirm(war, $t('提示'), { type: 'warning' }).then(async () => {
      await PorductApi.openOverallDiscount({ buffet_id: row.id, open_overall_discount: row.open_overall_discount == 1 ? 0 : 1 });
      ElMessage({ message: war_ + $t('成功'), type: 'success' });
      getData();
    });
  };

  /*选择第几页*/
  const handleCurrentChange = (val) => {
    curPage.value = val;
    getData();
  };

  /*每页多少条*/
  const handleSizeChange = (val) => {
    pageSize.value = val;
    curPage.value = 1;
    getData();
  };

  /*关闭弹窗*/
  const closeDialogFunc = (e) => {
    open_dialog.value = e.openDialog;
    editData.value = '';
    if (e.type == 'success') {
      getData();
    }
  };

  onMounted(() => {
    getData();
  });
</script>

<style lang="scss" scoped>
  .common-search-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .el-button--primary.is-link {
    color: var(--el-color-primary);
  }
</style>
