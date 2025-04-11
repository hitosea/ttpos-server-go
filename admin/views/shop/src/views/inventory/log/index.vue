<template>
  <div class="log-list">
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('类型')">
          <a-select size="small" v-model:value="searchForm.type" :placeholder="$t('所有类型')" @change="onSearch">
            <el-option :label="$t('销售出库')" value="30"></el-option>
            <el-option :label="$t('调整出库')" value="40"></el-option>
            <el-option :label="$t('删除出库')" value="41"></el-option>
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
          <el-table-column prop="number" :label="$t('编号')" width="200"></el-table-column>
          <el-table-column prop="type" :label="$t('类型')">
            <template #default="scope">
              {{ typeReturn(scope.row.type) }}
            </template>
          </el-table-column>
          <el-table-column prop="product.product_name_text" :label="$t('商品名称')" width="300"></el-table-column>
          <el-table-column prop="num" :label="$t('商品数量')"> </el-table-column>
          <el-table-column prop="num" :label="$t('规格/单位/加料')" width="140">
            <template #default="scope">
              {{ scope.row.product?.type == 10 ? scope.row.product_sku_name_text || '-' : scope.row.product?.product_unit_text || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="remark" :label="$t('备注')">
            <template #default="scope">
              {{ scope.row.remark || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product_stock" :label="$t('状态')">
            <template #default="scope">
              {{ statusReturn(scope.row.status) }}
            </template>
          </el-table-column>
          <el-table-column prop="product_stock" :label="$t('操作人')">
            <template #default="scope">
              {{ scope.row.operator?.real_name || scope.row.operator?.user_name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product_stock" :label="$t('时间')" width="220">
            <template #default="scope">
              <div style="line-height: 20px">
                {{ $t('出库：') + dateReturn(scope.row.out_time) }}
              </div>
              <div v-if="scope.row.status == '30'" style="line-height: 20px">
                {{ $t('撤销：') + dateReturn(scope.row.revoke_time) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="120">
            <template #default="scope">
              <el-button
                class="button-p"
                @click="cancelClick(scope.row)"
                type="primary"
                v-if="scope.row.type == '40' && scope.row.status == '20' && scope.row.is_show_out_cancel == '1'"
                link
                size="small"
                v-auth="'/inventory/log/cancel'"
                >{{ $t('撤销') }}</el-button
              >
              <el-button class="button-p" @click="deleteClick(scope.row)" type="primary" v-if="scope.row.status == '30'" link size="small" v-auth="'/inventory/log/delete'">{{
                $t('删除')
              }}</el-button>
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
  import InventoryApi from '@/api/inventory.js';
  import Aselect from '@/components/a-select/index.vue';
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
          tyep: '',
          date: '',
        },
        searchLoading: '',
      };
    },
    mounted() {
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
        InventoryApi.getErpInventoryRecordOut(Params, true)
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
            InventoryApi.erpInventoryRecordOutCancel(
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
            InventoryApi.erpInventoryRecordOutDelete(
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

      statusReturn(e) {
        let result = '';
        if (e == '10') {
          result = $t('已入库');
        }
        if (e == '20') {
          result = $t('已出库');
        }
        if (e == '30') {
          result = $t('已撤销');
        }
        return result;
      },
      typeReturn(e) {
        let result = '';
        if (e == '30') {
          result = $t('销售出库');
        }
        if (e == '40') {
          result = $t('调整出库');
        }
        if (e == '41') {
          result = $t('删除出库');
        }
        return result;
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
<style lang="scss" scoped>
  .common-seach-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .button-p {
    color: var(--el-color-primary);

    &:focus {
      color: var(--el-color-primary);
    }
  }
</style>
