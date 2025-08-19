<template>
  <div class="inventory-list">
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('库存排序')">
          <a-select size="small" v-model:value="searchForm.sort" :placeholder="$t('无')" @change="onSearch">
            <el-option :label="$t('无')" value=" "></el-option>
            <el-option :label="$t('从小到大')" value="asc"></el-option>
            <el-option :label="$t('从大到小')" value="desc"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('库存')">
          <a-select size="small" v-model:value="searchForm.stock_num" :placeholder="$t('全部库存')" @change="onSearch">
            <el-option :label="$t('全部')" value=" "></el-option>
            <el-option :label="$t('库存低于10')" value="10"></el-option>
            <el-option :label="$t('库存低于20')" value="20"></el-option>
            <el-option :label="$t('库存低于50')" value="50"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品名称')">
          <el-input size="small" v-model="searchForm.keyword" :placeholder="$t('请输入商品名称')" @input="onSearch"></el-input>
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
          <el-table-column prop="product.type" :label="$t('类型')">
            <template #default="scope">
              {{ scope.row.product.type == 10 ? $t('成品') : $t('材料') }}
            </template>
          </el-table-column>
          <el-table-column prop="product.product_name_text" :label="$t('商品名称')" width="300"></el-table-column>
          <el-table-column prop="barcode" :label="$t('商品条码')" width="120">
            <template #default="scope">
              {{ scope.row.barcode || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="spec_name_text" :label="$t('规格')" width="120">
            <template #default="scope">
              {{ scope.row.spec_name_text || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="supplier.name" :label="$t('供应商')" width="120">
            <template #default="scope">
              {{ scope.row.product.erpSupplier[0]?.name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="product_price" :label="$t('售价')" width="100">
            <template #default="scope">
              {{ this.$formatPrice(scope.row.product_price) }}
            </template>
          </el-table-column>
          <el-table-column prop="stock_num" :label="$t('当前库存')" width="120">
            <template #default="scope">
              {{ scope.row.product.type == 10 ? scope.row.stock_num : scope.row.material_stock }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('库存金额')" width="120">
            <template #default="scope">
              {{
                scope.row.product.type == 10
                  ? this.$formatPrice(Number(scope.row.product_price) * Number(scope.row.stock_num))
                  : this.$formatPrice(Number(scope.row.product_price) * Number(scope.row.material_stock))
              }}
            </template>
          </el-table-column>
          <el-table-column prop="history_purchase_num" :label="$t('历史进货数')" width="120">
            <template #default="scope">
              {{ scope.row.history_purchase_num }}
            </template>
          </el-table-column>
          <el-table-column prop="history_loss_num" :label="$t('历史报损数')" width="120">
            <template #default="scope">
              {{ scope.row.history_loss_num }}
            </template>
          </el-table-column>
          <el-table-column prop="product_sales" :label="$t('历史销售数')" width="120">
            <template #default="scope">
              {{ scope.row.product_sales }}
            </template>
          </el-table-column>
          <el-table-column prop="update_time" :label="$t('最后变动时间')" width="160">
            <template #default="scope">
              <div style="line-height: 20px"> {{ scope.row.update_time.split(' ')[0] || '-' }}<br />{{ scope.row.update_time.split(' ')[1] || '-' }} </div>
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
          sort: '',
          keyword: '',
          stock_num: '',
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
        Params.keyword = self.searchForm.keyword;
        Params.sort = self.searchForm.sort;
        Params.stock_num = self.searchForm.stock_num;
        Params.filter_having_package = 1;
        InventoryApi.getErpInventory(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },
    },
  };
</script>
<style lang="scss" scoped>
  .common-search-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>
