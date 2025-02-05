<template>
  <div class="wastage-list">
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('类型')">
          <a-select size="small" v-model:value="searchForm.type" @change="onSearch" :placeholder="$t('所有类型')">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option :label="$t('丢失')" value="1"></el-option>
            <el-option :label="$t('损耗')" value="2"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('起始时间')">
          <div class="block">
            <span class="demonstration"></span>
            <el-date-picker size="small" v-model="searchForm.date" clearable type="month" value-format="YYYY-MM" @change="onSearch" :placeholder="$t('选择月份')"></el-date-picker>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button class="search-button" size="small" type="primary" icon="Search" @click="onSearch">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
      <el-button size="small" type="primary" icon="Plus" v-auth="'/inventory/wastage/add'" @click="addClick">{{ $t('添加') }}</el-button>
    </div>

    <div class="wastage-main" v-if="echartsVueShow">
      <echartsVue :chart_list="chart_list"></echartsVue>
    </div>

    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="number" :label="$t('编号')" width="200"></el-table-column>
          <el-table-column :label="$t('报损类型')" width="120">
            <template #default="scope">
              {{ scope.row.type == 1 ? $t('丢失') : $t('损耗') }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('类型')" width="120">
            <template #default="scope">
              {{ scope.row.sku.product?.type == 10 ? $t('成品') : $t('材料') }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('商品名称')" width="260">
            <template #default="scope">
              {{ scope.row.sku.product?.product_name_text || '-' }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('规格')" width="140">
            <template #default="scope">
              {{ scope.row.sku?.spec_name_text || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="barcode" :label="$t('报损数量')" width="120">
            <template #default="scope">
              {{ scope.row.num || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="spec_name_text" :label="$t('备注')" width="120" show-overflow-tooltip>
            <template #default="scope">
              {{ scope.row.remark || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="supplier.name" :label="$t('状态')" width="120">
            <template #default="scope">
              {{ reviewStatus(scope.row.review_status) }}
            </template>
          </el-table-column>
          <el-table-column prop="product_price" :label="$t('操作人')" width="100">
            <template #default="scope">
              {{ scope.row.operator.real_name || scope.row.operator.user_name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="update_time" :label="$t('时间')" width="300">
            <template #default="scope">
              <div>{{ $t('报损：') + scope.row.create_time || '-' }}</div>
              <div v-if="scope.row.review_status == 1">
                {{ $t('通过：') + scope.row.approved_time || '-' }}
              </div>
              <div v-if="scope.row.review_status == 2">
                {{ $t('驳回：') + scope.row.rejected_time || '-' }}
              </div>
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="200">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" class="button-p" v-if="scope.row.review_status == 0" type="primary" link size="small" v-auth="'/inventory/wastage/edit'">{{
                $t('编辑')
              }}</el-button>
              <el-button
                @click="passClick(scope.row)"
                class="button-p"
                v-if="scope.row.review_status == 0 && (sType != '20' || level != '2')"
                type="primary"
                link
                size="small"
                v-auth="'/inventory/wastage/examine'"
                >{{ $t('通过') }}</el-button
              >
              <el-button
                @click="turnDownClick(scope.row)"
                class="button-p"
                v-if="scope.row.review_status == 0 && (sType != '20' || level != '2')"
                type="primary"
                link
                size="small"
                v-auth="'/inventory/wastage/examine'"
                >{{ $t('驳回') }}</el-button
              >
              <el-button
                @click="turnDownClick(scope.row)"
                class="button-p"
                v-if="scope.row.review_status == 2"
                type="primary"
                link
                size="small"
                v-auth="'/inventory/wastage/examine'"
                >{{ $t('驳回原因') }}</el-button
              >
              <el-button
                @click="deleteClick(scope.row)"
                class="button-p"
                type="primary"
                v-if="scope.row.review_status != 1"
                link
                size="small"
                v-auth="'/inventory/wastage/delete'"
                >{{ $t('删除') }}</el-button
              >
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
    <addEdit v-if="open_dialog" :open_dialog="open_dialog" :title="title" :editData="editData" @closeDialog="closeDialog"></addEdit>
    <turnDown v-if="dialogVisible" :dialogVisible="dialogVisible" :turnDownRow="turnDownRow" @closeDialog="closeTurnDown"></turnDown>
  </div>
</template>
<script>
  import echartsVue from './echartsVue.vue';
  import addEdit from './addEdit.vue';
  import InventoryApi from '@/api/inventory.js';
  import turnDown from './turnDown.vue';
  import { useUserStore } from '@/store';
  import Aselect from '@/components/a-select/index.vue';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const sType = supplier.value.store_type;
  const level = supplier.value.level;
  const date = new Date();
  const year = date.getFullYear();
  const month = date.getMonth() + 1; // Add 1 because getMonth() returns 0-11
  const formattedMonth = `${year}-${month.toString().padStart(2, '0')}`;

  export default {
    components: { echartsVue, addEdit, turnDown, Aselect },
    data() {
      return {
        searchForm: {
          type: '',
          date: formattedMonth,
        },
        open_dialog: false,

        echartsVueShow: false,
        loading: false,
        title: '',
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        tableData: [],

        editData: '',

        turnDownRow: '',
        dialogVisible: false,
        chart_list: [],
        sType: sType,
        level: level,
        searchLoading: '',
      };
    },
    mounted() {
      this.getData();
    },
    methods: {
      getData() {
        let self = this;
        let Params = {};
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.type = self.searchForm.type;
        Params.date = self.searchForm.date;
        self.loading = true;
        self.echartsVueShow = false;
        InventoryApi.erpDamagedProductRecordList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.chart_list = data.data.chart_list;
            if (self.chart_list.length > 0) {
              self.echartsVueShow = true;
            }
          })
          .catch((error) => {});
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getData();
        }, 200);
      },

      reviewStatus(status) {
        let result = '';
        if (status == 0) {
          result = $t('待审核');
        }
        if (status == 1) {
          result = $t('已通过');
        }
        if (status == 2) {
          result = $t('已拒绝');
        }
        return result;
      },

      //编辑
      editClick(data) {
        this.editData = data;
        this.open_dialog = true;
        this.title = $t('编辑');
      },

      /*通过*/
      passClick(row) {
        let self = this;
        ElMessageBox.confirm($t('确定审核通过吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            InventoryApi.erpDamagedProductRecordReview(
              {
                id: row.id,
                review_status: 1,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('操作成功'),
                    type: 'success',
                  });
                  //刷新页面
                  self.getData();
                } else {
                  self.loading = false;
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },

      /*删除*/
      deleteClick(row) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            InventoryApi.erpDamagedProductRecordDelete(
              {
                id: row.id,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: data.msg,
                    type: 'success',
                  });
                  self.getData();
                } else {
                  ElMessage.error($t('删除失败'));
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },

      /*驳回*/
      turnDownClick(row) {
        this.dialogVisible = true;
        this.turnDownRow = row;
      },

      closeTurnDown(e) {
        this.dialogVisible = false;
        this.turnDownRow = '';
        if ((e.type = 'success')) {
          this.getData();
        }
      },

      /*选择第几页*/
      handleCurrentChange(val) {
        this.curPage = val;
        this.getData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getData();
      },

      addClick() {
        this.open_dialog = true;
        this.title = $t('报损');
      },

      closeDialog(e) {
        this.open_dialog = false;
        this.editData = '';
        if ((e.type = 'success')) {
          this.getData();
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .common-seach-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .wastage-main {
    width: 100%;
    height: 400px;
    position: relative;
    margin-bottom: 24px;
  }

  .button-p {
    color: var(--el-color-primary);

    &:focus {
      color: var(--el-color-primary);
    }
  }

  .el-table {
    :deep(.cell) {
      -webkit-line-clamp: initial !important;
    }
  }
  :deep(.el-popper) {
    max-width: 500px;
  }
</style>
