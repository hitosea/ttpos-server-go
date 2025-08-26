<template>
  <div class="product-list">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item :label="$t('商品类型')">
          <Aselect size="small" v-model:value="material_type" clearable :placeholder="$t('全部类型')" @change="onSearch">
            <el-option :label="$t('全部类型')" value=" "></el-option>
            <el-option :label="$t('材料')" value="20"></el-option>
            <el-option :label="$t('套餐')" value="30"></el-option>
            <el-option :label="$t('成品')" value="10"></el-option>
          </Aselect>
        </el-form-item>
        <el-form-item :label="$t('商品分类')">
          <Acascader
            :options="categoryList"
            :props="{ checkStrictly: true, expandTrigger: 'hover' }"
            v-model:value="searchForm.category_id"
            :placeholder="$t('请选择分类')"
            @change="onSearch('1')"
          >
          </Acascader>
        </el-form-item>
        <el-form-item :label="$t('商品库存')">
          <Aselect size="small" v-model:value="stock" :placeholder="$t('全部库存')" @change="onSearch">
            <el-option :label="$t('全部')" value=" "></el-option>
            <el-option :label="$t('库存低于10')" value="10"></el-option>
            <el-option :label="$t('库存低于20')" value="20"></el-option>
            <el-option :label="$t('库存低于50')" value="50"></el-option>
          </Aselect>
        </el-form-item>
        <el-form-item :label="$t('商品状态')">
          <Aselect size="small" v-model:value="activeName" :placeholder="$t('商品状态')" @change="onSearch">
            <el-option :label="$t('全部')" value="all"></el-option>
            <el-option :label="$t('上架中')" value="sell"></el-option>
            <el-option :label="$t('下架中')" value="lower"></el-option>
          </Aselect>
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
              <el-dropdown-item @click="openBatch(5)" v-if="erp_is_open == 0">
                {{ $t('导入商品') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(1)" v-if="erp_is_open == 0">
                {{ $t('上传图片') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(2)">
                {{ $t('修改分类') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(3)" v-if="userInfo.isOpenTax == '1' && erp_is_open == 0">
                {{ $t('修改税类') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(6)">
                {{ $t('批量修改整单折扣商品') }}
              </el-dropdown-item>
              <el-dropdown-item @click="openBatch(4)" v-if="erp_is_open == 0">
                {{ $t('删除') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button class="flex-r-b" size="small" :disabled="erp_is_open == 1" type="primary" icon="Plus" v-auth="'/product/store/product/add'" @click="addClick"
          >{{ $t('添加商品') }}
        </el-button>
      </div>
    </div>
    <!--添加产品-->
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="category.path_name_text" :label="$t('类型')">
            <template #default="scope">
              {{ typeText(scope.row.type) }}
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
                  <div class="price"> {{ $t('销售价：') }}{{ $formatPrice(scope.row.product_price) }} </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="category.path_name_text" :label="$t('分类')"></el-table-column>
          <el-table-column prop="sales_actual" :label="$t('实际销量')"></el-table-column>
          <el-table-column prop="product_stock" :label="$t('库存')">
            <template #default="scope">
              <template v-if="scope.row.type == 10 || scope.row.type == 30">
                {{ scope.row.is_open_stock == 1 ? scope.row.product_stock : '-' }}
              </template>
              <template v-else>
                {{ scope.row.product_material_stock }}
              </template>
            </template>
          </el-table-column>
          <el-table-column prop="product_status.text" :label="$t('状态')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!proxy.$filter.isAuth('/product/store/product/state') || erp_is_open == 1"
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
              <el-button @click="deleteClick(scope.row)" :disabled="scope.row.is_material_used == 1 || erp_is_open == 1" link type="primary" size="small" v-auth="'/product/store/product/delete'">{{
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

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance, nextTick } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { ElMessageBox, ElMessage } from 'element-plus';
  import PorductApi from '@/api/product.js';
  import ProductSelector from '@/components/product/Selector.vue';
  import Aselect from '@/components/a-select/index.vue';
  import Acascader from '@/components/a-cascader/index.vue';
  import { useUserStore } from '@/store/index';
  import { languageStore } from '@/store/model/language';
  import defaultImg from '@/assets/img/default.png';

  const router = useRouter();
  const route = useRoute();
  const { computedSupplier, userInfo, erp_is_open } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  const { proxy } = getCurrentInstance();

  // 响应式数据
  const activeName = ref('');
  const stock = ref('');
  const material_type = ref('');
  const loading = ref(true);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const searchForm = reactive({
    product_name: '',
    category_id: '',
  });
  const tableData = ref([]);
  const categoryList = ref([]);
  const product_count = ref({});
  const searchLoading = ref(null);
  const batch_type = ref('');
  const openProductSelector = ref(false);

  // 初始化参数
  onMounted(async () => {
    let params = languageStore().getPageParams().pageParams;
    if (params.value && params.value.page) {
      searchForm.category_id = params.value.category_id;
      searchForm.product_name = params.value.product_name;
      activeName.value = params.value.type;
      stock.value = params.value.stock;
      curPage.value = params.value.page;
      pageSize.value = params.value.list_rows;
      material_type.value = params.value.material_type;
      languageStore().setPageParams({});
    }

    if (route.query.inventory) {
      stock.value = '10';
      material_type.value = '10';
      // 清除查询参数
      router.replace({ query: {} });
    }
    getData();
  });

  // 获取列表
  const getData = async () => {
    let Params = {
      ...searchForm,
      page: curPage.value,
      list_rows: pageSize.value,
      type: activeName.value,
      stock: stock.value,
      material_type: material_type.value,
    };
    if (typeof Params.category_id === 'object' && Params.category_id) {
      Params.category_id = Number(Params.category_id[Params.category_id.length - 1]);
    }
    loading.value = true;
    try {
      const data = await PorductApi.storeProductList(Params, true);
      tableData.value = data.data.list.data;
      tableData.value.forEach((item) => {
        if (item.image.length > 0) {
          item.image.forEach((items) => {
            items.imageLoading = false;
          });
        }
      });
      categoryList.value = [];
      data.data.category.forEach((item, index) => {
        categoryList.value.push({
          value: item.category_id,
          label: item.name_text,
          children: [],
        });
        item.child.forEach((items, indexs) => {
          categoryList.value[index].children.push({
            value: items.category_id,
            label: items.name_text,
          });
        });
      });
      totalDataNumber.value = data.data.list.total;
      product_count.value = data.data.product_count;
    } catch (error) {
      // 错误处理
    } finally {
      loading.value = false;
    }
  };

  // 搜索查询
  const onSearch = (e) => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getData();
    }, 200);
  };

  // 选择第几页
  const handleCurrentChange = (val) => {
    loading.value = true;
    curPage.value = val;
    getData();
  };

  // 每页多少条
  const handleSizeChange = (val) => {
    pageSize.value = val;
    getData();
  };

  // 切换菜单
  const handleClick = (tab, event) => {
    curPage.value = 1;
    getData();
  };

  // 打开添加
  const addClick = () => {
    let pageParams = {
      ...searchForm,
      page: curPage.value,
      list_rows: pageSize.value,
      type: activeName.value,
      stock: stock.value,
      material_type: material_type.value,
    };
    if (typeof pageParams.category_id === 'object' && pageParams.category_id) {
      pageParams.category_id = Number(pageParams.category_id[pageParams.category_id.length - 1]);
    }
    languageStore().setPageParams(pageParams);
    router.push('/' + app_id + '/product/store/product/add');
  };

  // 打开编辑
  const editClick = (row) => {
    let pageParams = {
      ...searchForm,
      page: curPage.value,
      list_rows: pageSize.value,
      type: activeName.value,
      stock: stock.value,
      material_type: material_type.value,
    };
    if (typeof pageParams.category_id === 'object' && pageParams.category_id) {
      pageParams.category_id = Number(pageParams.category_id[pageParams.category_id.length - 1]);
    }
    languageStore().setPageParams(pageParams);
    router.push({
      path: '/' + app_id + '/product/store/product/edit',
      query: {
        product_id: row.product_id,
        scene: 'edit',
      },
    });
  };

  // 强制下架上架
  const undercarriage = (row, state) => {
    if (!proxy.$filter.isAuth('/product/store/product/state')) {
      return;
    }
    let war = '';
    let war_ = '';
    if (state == 20) {
      war = $t('确认要强制下架吗?');
      war_ = $t('下架');
    } else if (state == 10) {
      war = $t('确认要重新上架吗?');
      war_ = $t('上架');
    }
    ElMessageBox.confirm(war, $t('提示'), {
      type: 'warning',
    }).then(async () => {
      try {
        await PorductApi.storeStateProduct({
          product_id: row.product_id,
          state,
        });
        ElMessage({
          message: war_ + $t('成功'),
          type: 'success',
        });
        getData();
      } catch (error) {
        // 错误处理
      }
    });
  };

  // 删除
  const deleteClick = async (row) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    }).then(async () => {
      try {
        await PorductApi.storeDelProduct({
          product_id: row.product_id,
        });
        ElMessage({
          message: $t('删除成功'),
          type: 'success',
        });
        getData();
      } catch (error) {
        ElMessage({
          message: $t('删除失败'),
          type: 'error',
        });
      }
    });
  };

  // 批量操作
  const openBatch = (e) => {
    batch_type.value = e;
    switch (e) {
      case 1:
      case 2:
      case 3:
      case 5:
      case 6:
        router.push({ path: '/' + app_id + '/product/store/product/batch', query: { type: batch_type.value } });
        break;
      case 4:
        openProductSelector.value = true;
        break;
    }
  };

  const deleteArr = async (data) => {
    openProductSelector.value = false;
    if (data && data.length > 0) {
      const product_ids = data.map((item) => item.product_id).join(',');
      try {
        await PorductApi.storeDelProduct({
          product_id: product_ids,
        });
        ElMessage({
          message: $t('删除成功'),
          type: 'success',
        });
      } catch (error) {
        // 错误处理
        ElMessage({
          message: $t('删除失败'),
          type: 'error',
        });
      }
      getData();
    }
  };

  const typeText = (type) => {
    switch (type) {
      case 10:
        return $t('成品');
      case 20:
        return $t('材料');
      case 30:
        return $t('套餐');
    }
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
