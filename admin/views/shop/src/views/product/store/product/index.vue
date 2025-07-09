<template>
  <!--
  
        时间：2019-10-24
        描述：商品管理
    -->
  <div class="product-list">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('商品类型')">
          <a-select size="small" v-model:value="material_type" clearable :placeholder="$t('全部类型')" @change="onSearch">
            <el-option :label="$t('全部类型')" value=" "></el-option>
            <el-option :label="$t('材料')" value="20"></el-option>
            <el-option :label="$t('成品')" value="10"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品分类')">
          <a-cascader
            :options="categoryList"
            :props="{ checkStrictly: true, expandTrigger: 'hover' }"
            v-model:value="searchForm.category_id"
            :placeholder="$t('请选择分类')"
            @change="onSearch('1')"
          >
          </a-cascader>
        </el-form-item>
        <el-form-item :label="$t('商品库存')">
          <a-select size="small" v-model:value="stock" :placeholder="$t('全部库存')" @change="onSearch">
            <el-option :label="$t('全部')" value=" "></el-option>
            <el-option :label="$t('库存低于10')" value="10"></el-option>
            <el-option :label="$t('库存低于20')" value="20"></el-option>
            <el-option :label="$t('库存低于50')" value="50"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品状态')">
          <a-select size="small" v-model:value="activeName" :placeholder="$t('商品状态')" @change="onSearch">
            <el-option :label="$t('全部')" value="all"></el-option>
            <el-option :label="$t('上架中')" value="sell"></el-option>
            <el-option :label="$t('下架中')" value="lower"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('商品名称')">
          <el-input size="small" v-model="searchForm.product_name" :placeholder="$t('请输入商品名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div class="flex-right">
        <el-dropdown class="mr16" v-auth="'/product/store/product/batch'">
          <el-button size="small" type="primary" icon="ArrowDown">
            {{ $t('批量') }}
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="openBatch(5)">
                {{ $t('导入商品') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(1)">
                {{ $t('上传图片') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(2)">
                {{ $t('修改分类') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(3)" v-if="userInfo.isOpenTax == '1'">
                {{ $t('修改税类') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(4)">
                {{ $t('删除') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button class="flex-r-b" size="small" type="primary" icon="Plus" v-auth="'/product/store/product/add'" @click="addClick">{{ $t('添加商品') }} </el-button>
      </div>
    </div>
    <!--添加产品-->
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="category.path_name_text" :label="$t('类型')">
            <template #default="scope">
              {{ scope.row.type == 10 ? $t('成品') : $t('材料') }}
            </template>
          </el-table-column>
          <el-table-column prop="product_name" :label="$t('商品名称')" width="400px">
            <template #default="scope">
              <div class="product-info">
                <div class="pic">
                  <img
                    v-show="scope.row.image[0]?.imageLoading"
                    v-img-url="scope.row.image[0]?.file_path"
                    alt=""
                    @load="
                      () => {
                        scope.row.image[0] ? (scope.row.image[0].imageLoading = true) : '';
                      }
                    "
                  />
                  <img v-show="!scope.row.image[0]?.imageLoading" :src="defaultImg" alt="" />
                </div>
                <div class="info">
                  <div class="name">{{ scope.row.product_name_text }}</div>
                  <div class="price"> {{ $t('销售价：') }}{{ this.$formatPrice(scope.row.product_price) }} </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="category.path_name_text" :label="$t('分类')"></el-table-column>
          <el-table-column prop="sales_actual" :label="$t('实际销量')"></el-table-column>
          <el-table-column prop="product_stock" :label="$t('库存')">
            <template #default="scope">
              {{ scope.row.type == 10 ? scope.row.product_stock : scope.row.product_material_stock }}
            </template>
          </el-table-column>
          <el-table-column prop="product_status.text" :label="$t('状态')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!this.$filter.isAuth('/product/store/product/state')"
                :model-value="scope.row.product_status.value == 10 ? true : false"
                @click="undercarriage(scope.row, scope.row.product_status.value == 10 ? 20 : 10)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')"></el-table-column>
          <el-table-column prop="product_sort" :label="$t('排序')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="120">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" link type="primary" size="small" v-auth="'/product/store/product/edit'">{{ $t('编辑') }}</el-button>
              <el-button @click="deleteClick(scope.row)" :disabled="scope.row.is_material_used == 1" link type="primary" size="small" v-auth="'/product/store/product/delete'">{{
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

    <!-- 商品选择器 -->
    <ProductSelector v-if="openProductSelector" :open="openProductSelector" @close="deleteArr" selectorType="all" type="all"> </ProductSelector>
  </div>
</template>

<script>
  import PorductApi from '@/api/product.js';
  import ProductSelector from '@/components/product/Selector.vue';
  import { useUserStore } from '@/store/index';
  import { languageStore } from '@/store/model/language';
  import defaultImg from '@/assets/img/default.png';
  import Aselect from '@/components/a-select/index.vue';
  import Acascader from '@/components/a-cascader/index.vue';
  const { computedSupplier, userInfo } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  export default {
    components: { ProductSelector, Aselect, Acascader },
    data() {
      return {
        userInfo,
        defaultImg,
        /*切换菜单*/
        activeName: '',
        stock: '',
        material_type: '',
        /*切换选中值*/
        activeIndex: '0',
        /*是否正在加载*/
        loading: true,
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        /*搜索参数*/
        searchForm: {
          product_name: '',
          category_id: '',
        },
        /*列表数据*/
        tableData: [],
        /*全部分类*/
        categoryList: [],
        /*商品统计*/
        product_count: {},
        app_id: app_id,
        searchLoading: '',

        open_import_product: false,
        batch_type: '',

        openProductSelector: false,
      };
    },
    created() {
      let params = languageStore().getPageParams().pageParams;
      if (params.value.page) {
        this.searchForm = {
          category_id: params.value.category_id,
          product_name: params.value.product_name,
        };
        this.activeName = params.value.type;
        this.stock = params.value.stock;
        this.curPage = params.value.page;
        this.pageSize = params.value.list_rows;
        this.material_type = params.value.material_type;
        languageStore().setPageParams({});
      }

      /*获取列表*/
      if (this.$route.query.inventory) {
        this.stock = '10';
        this.material_type = '10';
        this.$route.query = {};
      }
      this.getData();
    },
    methods: {
      initGetData() {
        this.curPage = 1;
        this.getData();
      },

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

      /*切换菜单*/
      handleClick(tab, event) {
        let self = this;
        self.curPage = 1;
        self.getData();
      },

      /*获取列表*/
      getData() {
        let self = this;
        let Params = self.searchForm;
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.type = self.activeName;
        Params.stock = self.stock;
        Params.material_type = self.material_type;
        if (typeof Params.category_id == 'object' && Params.category_id) {
          Params.category_id = Number(Params.category_id[Params.category_id.length - 1]);
        }
        self.loading = true;
        PorductApi.storeProductList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.tableData.map((item) => {
              if (item.image.length > 0) {
                item.image.map((items) => {
                  items.imageLoading = false;
                });
              }
            });

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
            self.totalDataNumber = data.data.list.total;
            self.product_count = data.data.product_count;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*搜索查询*/
      onSearch(e) {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getData();
        }, 200);
      },

      /*打开添加*/
      addClick() {
        let self = this;
        let pageParams = self.searchForm;
        pageParams.page = self.curPage;
        pageParams.list_rows = self.pageSize;
        pageParams.type = self.activeName;
        pageParams.stock = self.stock;
        pageParams.material_type = self.material_type;
        if (typeof pageParams.category_id == 'object' && pageParams.category_id) {
          pageParams.category_id = Number(pageParams.category_id[pageParams.category_id.length - 1]);
        }
        languageStore().setPageParams(pageParams);
        this.$router.push('/' + this.app_id + '/product/store/product/add');
      },

      /*打开编辑*/
      editClick(row) {
        let self = this;
        let pageParams = self.searchForm;
        pageParams.page = self.curPage;
        pageParams.list_rows = self.pageSize;
        pageParams.type = self.activeName;
        pageParams.stock = self.stock;
        pageParams.material_type = self.material_type;
        if (typeof pageParams.category_id == 'object' && pageParams.category_id) {
          pageParams.category_id = Number(pageParams.category_id[pageParams.category_id.length - 1]);
        }
        languageStore().setPageParams(pageParams);
        this.$router.push({
          path: '/' + this.app_id + '/product/store/product/edit',
          query: {
            product_id: row.product_id,
            scene: 'edit',
          },
        });
      },

      /* 强制下架上架*/
      undercarriage(row, state) {
        if (!this.$filter.isAuth('/product/store/product/state')) {
          return;
        }
        let self = this;
        let war = '';
        let war_ = '';
        if (state == 20) {
          (war = $t('确认要强制下架吗?')), (war_ = $t('下架'));
        } else if (state == 10) {
          (war = $t('确认要重新上架吗?')), (war_ = $t('上架'));
        }
        ElMessageBox.confirm(war, $t('提示'), {
          type: 'warning',
        }).then(() => {
          PorductApi.storeStateProduct({
            product_id: row.product_id,
            state,
          }).then((data) => {
            this.$ElMessage({
              message: war_ + $t('成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
      /*打开复制*/
      copyClick(row) {
        this.$router.push({
          path: '/product/product/edit',
          query: {
            product_id: row.product_id,
            scene: 'copy',
          },
        });
      },

      /*删除*/
      deleteClick: function (row) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          type: 'warning',
        }).then(() => {
          PorductApi.storeDelProduct({
            product_id: row.product_id,
          }).then((data) => {
            this.$ElMessage({
              message: $t('删除成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },

      openImportProduct() {
        this.$router.push({ path: '/' + this.app_id + '/product/store/product/importProduct' });
      },

      // 批量操作
      openBatch(e) {
        this.batch_type = e;
        switch (e) {
          case 1:
            // 修改图片
            this.$router.push({ path: '/' + this.app_id + '/product/store/product/batch', query: { type: this.batch_type } });
            break;
          case 2:
            // 修改分类
            this.$router.push({ path: '/' + this.app_id + '/product/store/product/batch', query: { type: this.batch_type } });
            break;
          case 3:
            // 修改税类
            this.$router.push({ path: '/' + this.app_id + '/product/store/product/batch', query: { type: this.batch_type } });
            break;
          case 4:
            // 删除
            this.openProductSelector = true;
            break;
          case 5:
            // 商品批量导入
            this.$router.push({ path: '/' + this.app_id + '/product/store/product/batch', query: { type: this.batch_type } });
            break;
        }
      },

      deleteArr(data) {
        this.openProductSelector = false;
        if (data && data.length > 0) {
          const product_ids = [];
          data.map((item) => {
            product_ids.push(item.product_id);
          });
          const product_id = product_ids.join(',');
          PorductApi.storeDelProduct({
            product_id: product_id,
          }).then((data) => {
            this.$ElMessage({
              message: $t('删除成功'),
              type: 'success',
            });
            this.getData();
          });
        }
      },
    },
  };
</script>

<style scoped>
  .common-search-wrap {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .flex-right {
    flex-shrink: 0;
  }
  .flex-r-b {
    margin-right: 0;
  }
</style>
