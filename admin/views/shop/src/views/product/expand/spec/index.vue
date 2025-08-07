<template>
  <div class="product-list">
    <!--添加规格-->
    <div class="common-level-rail">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item>
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('规格名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div>
        <el-button size="small" type="primary" icon="Plus" v-auth="'/product/expand/spec/add'" @click="addClick"> {{ $t('添加规格') }}</el-button>
        <el-button size="small" v-auth="'/product/expand/spec/batch_delete'" :disabled="multipleSelection.length == 0" @click="deleteBatch">{{ $t('批量删除') }}</el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading" @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="45"></el-table-column>
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="spec_name_text" :label="$t('规格名称')"></el-table-column>
          <el-table-column prop="product_ids" :label="$t('关联商品数量')" width="120">
            <template #default="scope">
              {{ scope.row.product_ids?.length ?? 0 }}
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="240">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/expand/spec/edit'">{{ $t('编辑') }} </el-button>
              <el-button @click="relatedProductClick(scope.row)" v-auth="'/product/expand/spec/relatedProduct'" type="primary" link size="small">{{ $t('关联商品') }} </el-button>
              <el-button @click="editPriceClick(scope.row)" v-auth="'/product/expand/spec/batchPrice'" type="primary" link size="small">{{ $t('修改价格') }} </el-button>
              <el-button
                @click="deleteClick(scope.row.spec_id)"
                type="primary"
                link
                size="small"
                v-auth="'/product/expand/spec/delete'"
                :disabled="scope.row.product_ids?.length > 0"
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
    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialogFunc($event, 'add')"></Add>
    <!--修改-->
    <Edit v-if="open_edit" :open_edit="open_edit" :editform="model" @closeDialog="closeDialogFunc($event, 'edit')"> </Edit>

    <!-- 商品选择器 -->
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      :isLoading="loading"
      :selectedProductIds="model?.product_ids ?? []"
    >
    </ProductSelector>

    <ProductSpecPrice
      v-if="openProductSpecPrice"
      :open="openProductSpecPrice"
      @close="handleProductSpecPriceClose"
      :specId="model?.spec_id ?? -1"
      :title="`【${model?.spec_name_text ?? $t('规格')}】${$t('价格')}`"
    >
    </ProductSpecPrice>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import { ElMessageBox, ElMessage } from 'element-plus';
  import ProductApi from '@/api/product.js';
  import Add from './add.vue';
  import Edit from './edit.vue';
  import ProductSelector from '@/components/product/Selector.vue';
  import ProductSpecPrice from '@/components/product/SpecPrice.vue';

  // 获取当前实例
  const { proxy } = getCurrentInstance();

  // 响应式数据
  const activeName = ref('sell');
  const activeIndex = ref('0');
  const loading = ref(true);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const model = ref({});
  const open_edit = ref(false);
  const open_add = ref(false);
  const tableData = ref([]);
  const multipleSelection = ref([]);
  const searchForm = reactive({
    name: '',
  });
  const searchLoading = ref('');
  const openProductSelector = ref(false);
  const openProductSpecPrice = ref(false);

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

  // 搜索查询
  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getData();
    }, 200);
  };

  // 获取列表
  const getData = async () => {
    const params = {
      page: curPage.value,
      list_rows: pageSize.value,
      spec_name: searchForm.name,
    };
    loading.value = true;

    try {
      const data = await ProductApi.SpecList(params, true);
      loading.value = false;
      tableData.value = data.data.list.data;
      totalDataNumber.value = data.data.list.total;
    } catch (error) {
      loading.value = false;
    }
  };

  // 关闭弹窗
  const closeDialogFunc = (e, f) => {
    if (f == 'add') {
      open_add.value = e.openDialog;
      if (e.type == 'success') {
        getData();
      }
    }
    if (f == 'edit') {
      open_edit.value = e.openDialog;
      if (e.type == 'success') {
        getData();
      }
    }
  };

  // 打开添加
  const addClick = () => {
    open_add.value = true;
  };

  // 删除单个
  const deleteClick = async (id) => {
    try {
      await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
        type: 'warning',
      });

      await ProductApi.deleteSpec({
        spec_id: id,
      });

      ElMessage({
        message: $t('删除成功'),
        type: 'success',
      });
      getData();
    } catch (error) {
      // 用户取消删除或删除失败
    }
  };

  // 批量删除
  const deleteBatch = async () => {
    const arr = [];
    multipleSelection.value.forEach((item) => {
      arr.push(item.spec_id);
    });
    const spec_id = arr.join(',');

    try {
      await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
        type: 'warning',
      });

      await ProductApi.deleteSpec({
        spec_id: spec_id,
      });

      ElMessage({
        message: $t('删除成功'),
        type: 'success',
      });
      getData();
    } catch (error) {
      // 用户取消删除或删除失败
    }
  };

  // 选择变化
  const handleSelectionChange = (e) => {
    multipleSelection.value = e;
  };

  // 打开编辑
  const editClick = (row) => {
    model.value = row;
    open_edit.value = true;
  };

  // 关联商品
  const relatedProductClick = (row) => {
    model.value = row;
    openProductSelector.value = true;
  };

  // 修改价格
  const editPriceClick = (row) => {
    model.value = row;
    openProductSpecPrice.value = true;
  };

  // 商品选择器关闭
  const handleProductSelectorClose = async (list) => {
    if (Array.isArray(list)) {
      try {
        loading.value = true;
        await ProductApi.relateBySpec(
          {
            spec_id: model.value.spec_id,
            product_ids: list.map((item) => item.product_id),
          },
          false
        );

        ElMessage({
          message: $t('关联成功'),
          type: 'success',
        });
        getData();
        loading.value = false;
      } catch (error) {
        // 处理错误
        loading.value = false;
      }
    }
    model.value = {};
    openProductSelector.value = false;
  };

  // 规格价格关闭
  const handleProductSpecPriceClose = () => {
    model.value = {};
    openProductSpecPrice.value = false;
  };

  // 序号方法
  const indexMethod = (index) => {
    return index + 1 + (curPage.value - 1) * pageSize.value;
  };

  // 组件挂载时获取数据
  onMounted(() => {
    getData();
  });
</script>

<style scoped>
  .common-level-rail {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>
