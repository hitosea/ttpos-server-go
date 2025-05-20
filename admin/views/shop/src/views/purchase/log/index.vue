<template>
  <div class="log-list">
    <!--搜索表单-->
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('所有类型')">
          <a-select size="small" v-model:value="searchForm.type" :placeholder="$t('所有类型')" @change="onSearch">
            <el-option :label="$t('所有类型')" value=""></el-option>
            <el-option :label="$t('采购入库')" value="10"></el-option>
            <el-option :label="$t('调整入库')" value="20"></el-option>
            <el-option :label="$t('添加入库')" value="21"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('起始时间')">
          <div class="block">
            <span class="demonstration"></span>
            <el-date-picker
              size="small"
              v-model="searchForm.date"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              clearable
              @change="onSearch"
            ></el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="number" :label="$t('编号')" width="180"></el-table-column>
          <el-table-column prop="category.path_name_text" :label="$t('类型')">
            <template #default="scope">
              {{ typeS(scope.row.type) }}
            </template>
          </el-table-column>
          <el-table-column prop="number" :label="$t('采购编号')" width="180">
            <template #default="scope">
              {{ scope.row.purchaseOrder?.number || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product_name" :label="$t('采购/商品名称')" width="240">
            <template #default="scope">
              {{ scope.row.product ? scope.row.product?.product_name_text || '-' : scope.row.purchaseOrder?.name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product_sku_name_text" :label="$t('规格')">
            <template #default="scope">
              {{ scope.row.product_sku_name_text || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="num" :label="$t('商品数量')"></el-table-column>
          <el-table-column prop="remark" :label="$t('备注')" width="160" show-overflow-tooltip>
            <template #default="scope">
              {{ scope.row.remark || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" :label="$t('状态')" width="100">
            <template #default="scope">
              {{ statusJudgment(scope.row.status) }}
            </template>
          </el-table-column>
          <el-table-column prop="operator.real_name" :label="$t('操作人')"></el-table-column>
          <el-table-column prop="create_time" :label="$t('时间')" width="220">
            <template #default="scope">
              <div style="line-height: 20px">
                {{ $t('入库：') + dateReturn(scope.row.in_time) }}
              </div>
              <div v-if="scope.row.status == '30'" style="line-height: 20px">
                {{ $t('撤销：') + dateReturn(scope.row.revoke_time) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="100">
            <template #default="scope">
              <el-button
                @click="cancelClick(scope.row)"
                :disabled="scope.row.is_show_in_cancel == '0'"
                v-if="scope.row.status != 30"
                type="primary"
                link
                size="small"
                v-auth="'/purchase/log/cancel'"
                >{{ $t('撤销') }}</el-button
              >
              <el-button @click="deleteClick(scope.row)" v-if="scope.row.status == 30" type="primary" link size="small" v-auth="'/purchase/log/delete'">{{ $t('删除') }}</el-button>
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
  </div>
</template>
<script>
  import PurchaseApi from '@/api/purchase.js';
  import Aselect from '@/components/a-select/index.vue';
  import dayjs from '@/utils/dayjs';
  export default {
    components: {
      Aselect,
    },
    data() {
      return {
        /*是否正在加载*/
        loading: false,
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        /*列表数据*/
        tableData: [],
        searchForm: {
          type: '',
          date: '',
        },
        tableData: [],
        searchLoading: '',
      };
    },

    async mounted() {
      // js获取当天时间 日期格式 YYYY-MM-DD
      this.searchForm.date = [dayjs(), dayjs()];
      await this.$nextTick();
      this.getData();
    },
    methods: {
      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.loading = true;
        self.curPage = val;
        self.getData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.pageSize = val;
        this.getData();
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getData();
        }, 200);
      },

      // 获取
      getData() {
        let self = this;
        let Params = {};
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.type = self.searchForm.type;
        Params.date = self.searchForm.date;
        PurchaseApi.getErpInventoryRecordIn(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },
      /*撤销*/
      cancelClick(row) {
        let self = this;
        ElMessageBox.confirm($t('确定要撤销吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            PurchaseApi.cancelErpInventoryRecordIn(
              {
                erp_inventory_id: row.id,
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
            PurchaseApi.deleteErpInventoryRecordIn(
              {
                erp_inventory_id: row.id,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('删除成功'),
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

      typeS(e) {
        let result = '';
        if (e == '10') return (result = $t('采购入库'));
        if (e == '20') return (result = $t('调整入库'));
        if (e == '21') return (result = $t('添加入库'));
        if (e == '30') return (result = $t('销售出库'));
        if (e == '40') return (result = $t('调整出库'));
      },

      statusJudgment(e) {
        let result = '';
        if (e == '10') return (result = $t('已入库'));
        if (e == '20') return (result = $t('已出库'));
        if (e == '30') return (result = $t('已撤销'));
      },
      dateReturn(e) {
        let formattedDate = '';
        let date = new Date(e * 1000); // 创建 Date 对象并将时间戳转换成毫秒值
        let year = date.getFullYear(); // 获取年份
        let month = String(date.getMonth() + 1).padStart(2, '0'); // 获取月份（注意月份是从 0 开始计数，需要加 1）
        let day = String(date.getDate()).padStart(2, '0'); // 获取日期
        let hours = String(date.getHours()).padStart(2, '0'); // 获取小时
        let minutes = String(date.getMinutes()).padStart(2, '0'); // 获取分钟
        let seconds = String(date.getSeconds()).padStart(2, '0'); // 获取秒钟
        formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
        return formattedDate;
      },
    },
  };
</script>
<style lang="scss" scoped></style>
