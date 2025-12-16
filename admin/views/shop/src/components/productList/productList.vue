<template>
  <el-dialog :title="$t('选择商品')" width="80%" v-model="dialogVisible" @close="dialogFormVisible" append-to-body :close-on-click-modal="false" :close-on-press-escape="false">
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('商品状态')">
          <a-select size="small" v-model:value="searchForm.type" :placeholder="$t('商品状态')">
            <el-option :label="$t('全部')" value="all"></el-option>
            <el-option :label="$t('开启')" value="sell"></el-option>
            <el-option :label="$t('关闭')" value="lower"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品分类')">
          <a-cascader :options="categoryList" :props="{ checkStrictly: true, expandTrigger: 'hover' }" v-model:value="searchForm.category_id" :placeholder="$t('请选择分类')">
            <template #default="{ data }">
              <span class="span-click" @click="handleValue(data)">{{ data.label }}</span>
            </template>
          </a-cascader>
        </el-form-item>
        <el-form-item :label="$t('商品名称')"><el-input size="small" v-model="searchForm.product_name" :placeholder="$t('商品名称')"></el-input></el-form-item>
        <el-form-item>
          <el-button class="search-button" size="small" type="primary" icon="Search" @click="onSubmit">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table
          v-if="multiple_choice == 0"
          size="small"
          ref="multipleTable"
          :data="tableData"
          border
          style="width: 100%"
          v-loading="loading"
          @selection-change="handleSelectionChange"
          :row-key="getRowKey"
        >
          <el-table-column prop="product_name" :label="$t('商品名称')" width="300px">
            <template #default="scope">
              <div class="product-info">
                <div class="pic"><img v-img-url="scope.row.image[0]?.file_path" alt="" /></div>
                <div class="info">
                  <div class="name">{{ scope.row.product_name_text }}</div>
                  <div class="price">{{ $t('销售价：') }}{{ this.$formatPrice(scope.row.product_price) }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="category.path_name_text" width="200" :label="$t('分类名称')"></el-table-column>
          <el-table-column prop="sales_actual" :label="$t('实际销量')"></el-table-column>
          <!-- <el-table-column prop="product_stock" :label="$t('库存')">
            <template #default="scope">
              {{ scope.row.type == 10 ? scope.row.product_stock : scope.row.product_material_stock }}
            </template>
          </el-table-column> -->

          <el-table-column prop="product_status.text" :label="$t('状态')" width="100">
            <template #default="scope">
              {{ scope.row.product_status.value == 10 ? $t('开启') : $t('关闭') }}
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')" width="180">
            <template #default="scope">
              <p class="create-time">{{ scope.row.create_time.split(' ')[0] || '-' }}</p>
              <p class="create-time">{{ scope.row.create_time.split(' ')[1] || '' }}</p>
            </template>
          </el-table-column>
          <el-table-column fixed="right" type="selection" :selectable="selectable" width="40" :reserve-selection="true"></el-table-column>
        </el-table>
        <el-table
          v-if="multiple_choice == 1"
          size="small"
          ref="multipleTable"
          :data="tableData"
          border
          style="width: 100%"
          v-loading="loading"
          highlight-current-row
          @current-change="handleClick"
        >
          <el-table-column prop="product_name" :label="$t('商品名称')" width="300px">
            <template #default="scope">
              <div class="product-info">
                <div class="pic"><img v-img-url="scope.row.image[0]?.file_path" alt="" /></div>
                <div class="info">
                  <div class="name">{{ scope.row.product_name_text }}</div>
                  <div class="price">{{ $t('销售价：') }}{{ this.$formatPrice(scope.row.product_price) }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="category.path_name_text" width="200" :label="$t('分类名称')"></el-table-column>
          <el-table-column prop="sales_actual" :label="$t('实际销量')"></el-table-column>
          <el-table-column prop="product_stock" :label="$t('库存')">
            <template #default="scope">
              {{ scope.row.type == 10 ? scope.row.product_stock : scope.row.product_material_stock }}
            </template>
          </el-table-column>

          <el-table-column prop="product_status.text" :label="$t('状态')" width="100">
            <template #default="scope">
              {{ scope.row.product_status.value == 10 ? $t('开启') : $t('关闭') }}
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')" width="180">
            <template #default="scope">
              <p class="create-time">{{ scope.row.create_time.split(' ')[0] || '-' }}</p>
              <p class="create-time">{{ scope.row.create_time.split(' ')[1] || '' }}</p>
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
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script>
  import PorductApi from '@/api/product.js';
  export default {
    data() {
      return {
        searchForm: {
          type: '',
          product_name: '',
          category_id: '',
        },

        /*一页多少条*/
        pageSize: 5,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        tableData: [],
        /*全部分类*/
        categoryList: [],
        multipleSelection: [],
      };
    },
    props: {
      open_product: {
        type: Boolean,
        default: false,
      },
      //限制选择的id数组
      limit_ids: {
        type: String,
        default: '',
      },
      //select全选  limit限选
      selectType: {
        type: String,
        default: '',
      },
      //可选index
      index: {
        type: Number,
        default: 0,
      },
      // 编辑时候返回的限制
      multiple_selection: {
        type: Array,
        default: [],
      },
      //类型 10 成品 20 材料
      material_type: {
        type: [String, Number],
        default: '',
      },
      // table单选 多选
      multiple_choice: {
        type: [String, Number],
        default: 0,
      },
    },
    created() {
      this.dialogVisible = this.open_product;
      this.getData();
    },
    methods: {
      getData() {
        let self = this;
        let Params = self.searchForm;
        Params.product_ids = this.selectType == 'limit' ? this.limit_ids : '';
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.material_type = self.material_type;
        if (typeof Params.category_id == 'object' && Params.category_id) {
          Params.category_id = Number(Params.category_id[Params.category_id.length - 1]);
        }
        self.loading = true;
        PorductApi.storeProductList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.categoryList = [];
            data.data.category.map((item, index) => {
              self.categoryList.push({
                value: item.category_id,
                label: item.name_text,
                children: [],
              });
              item.child.map((items, indexs) => {
                self.categoryList[index].children.push({
                  value: items.category_id,
                  label: items.name_text,
                });
              });
            });
            if (this.multiple_selection.length > 0) {
              // 判断是否存在勾选过的数据
              this.tableData.forEach((row, index) => {
                // 获取数据列表接口请求到的数据
                this.multiple_selection.forEach((item) => {
                  // 勾选到的数据
                  if (row.product_id == item.product_id) {
                    this.$nextTick(() => {
                      this.$refs.multipleTable.toggleRowSelection(this.tableData[index], true); // 若有重合，则回显该条数据
                    });
                    this.tableData[index].select_open = 1;
                  }
                });
              });
            }
          })
          .catch((error) => {
            console.log(error);
          });
      },

      selectable(row, index) {
        if (row.select_open != undefined || row.product_status.value != 10) {
          if (row.select_open == 1 || row.product_status.value != 10) {
            return false;
          }
        } else {
          return true;
        }
      },

      /*搜索查询*/
      onSubmit() {
        this.curPage = 1;
        this.getData();
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

      /*关闭弹窗*/
      dialogFormVisible(e) {
        if (e) {
          this.$emit('closeDialogFunc', {
            type: 'success',
            openDialog: false,
          });
        } else {
          this.$emit('closeDialogFunc', {
            type: 'error',
            openDialog: false,
          });
        }
      },

      submit() {
        if (this.selectType == 'limit') {
          this.$emit('closeDialogFunc', {
            type: 'limit',
            openDialog: false,
            data: this.multipleSelection,
          });
        } else {
          this.$emit('closeDialogFunc', {
            type: 'select',
            openDialog: false,
            index: this.index,
            data: this.multipleSelection,
          });
        }
      },

      getRowKey(row) {
        return row.product_id;
      },

      handleSelectionChange(e) {
        this.multipleSelection = e;
      },

      handleValue(data) {
        this.searchForm.category_id = [];
        this.searchForm.category_id = data.value;
      },

      handleClick(e) {
        this.multipleSelection = [];
        this.multipleSelection.push(e);
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

  .create-time {
    line-height: 24px !important;
  }
  .span-click {
    width: 100%;
    display: block;
  }
</style>
