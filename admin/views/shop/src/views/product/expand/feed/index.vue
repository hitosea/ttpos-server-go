<template>
  <!--
      描述：加料库
  -->
  <div class="product-list">
    <!--添加加料-->
    <div class="common-level-rail">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item>
          <el-input size="small" v-model="searchForm.name" :placeholder="$t('加料名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div>
        <el-button size="small" type="primary" icon="Plus" :disabled="erp_is_open == 1" v-auth="'/product/expand/feed/add'" @click="addClick"> {{ $t('添加加料') }}</el-button>
        <el-button size="small" v-auth="'/product/expand/feed/batch_delete'" :disabled="multipleSelection.length == 0 || erp_is_open == 1" @click="deleteBatch">{{
          $t('批量删除')
        }}</el-button>
      </div>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading" @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="45" align="center"></el-table-column>
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="feed_name_text" :label="$t('加料名称')"></el-table-column>
          <el-table-column prop="price" :label="$t('价格')" width="400px">
            <template #default="scope">
              {{ $formatPrice(scope.row.price) }}
            </template>
          </el-table-column>
          <!-- <el-table-column prop="sort" :label="$t('排序')"></el-table-column> -->
          <el-table-column prop="product_ids" :label="$t('关联商品数量')" width="120">
            <template #default="scope">
              {{ scope.row.product_ids?.length ?? 0 }}
            </template>
          </el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="240">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" :disabled="erp_is_open == 1" v-auth="'/product/expand/feed/edit'">{{
                $t('编辑')
              }} </el-button>
              <el-button @click="relatedProductClick(scope.row)" v-auth="'/product/expand/feed/relatedProduct'" type="primary" link size="small" :disabled="erp_is_open == 1">{{
                $t('关联商品')
              }} </el-button>
              <el-button
                @click="deleteClick(scope.row.feed_id)"
                type="primary"
                link
                size="small"
                v-auth="'/product/expand/feed/delete'"
                :disabled="scope.row.product_ids?.length > 0 || erp_is_open == 1"
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
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product.js';
  import Add from './add.vue';
  import Edit from './edit.vue';
  import ProductSelector from '@/components/product/Selector.vue';
  import { useUserStore } from '@/store';
  const { erp_is_open } = useUserStore();
  // 获取组件实例
  const { proxy } = getCurrentInstance();

  // 响应式数据
  const loading = ref(true);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const model = ref({});
  const open_edit = ref(false);
  const open_add = ref(false);
  const tableData = ref([]);
  const multipleSelection = ref([]);
  const searchLoading = ref('');
  const openProductSelector = ref(false);

  const searchForm = reactive({
    name: '',
  });

  // 初始化数据
  onMounted(() => {
    getData();
  });

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
      feed_name: searchForm.name,
    };

    loading.value = true;
    try {
      const data = await ProductApi.FeedList(params, true);
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

      await ProductApi.deleteFeed({
        feed_id: id,
      });

      proxy.$ElMessage({
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
      arr.push(item.feed_id);
    });
    const feed_id = arr.join(',');

    try {
      await proxy.$ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
        type: 'warning',
      });

      await ProductApi.deleteFeed({
        feed_id: feed_id,
      });

      proxy.$ElMessage({
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

  // 关联商品点击
  const relatedProductClick = (row) => {
    model.value = row;
    openProductSelector.value = true;
  };

  // 商品选择器关闭
  const handleProductSelectorClose = async (list) => {
    if (Array.isArray(list)) {
      try {
        loading.value = true;
        await ProductApi.relateByFeed(
          {
            feed_id: model.value.feed_id,
            product_ids: list.map((item) => item.product_id),
          },
          false
        );

        proxy.$ElMessage({
          message: $t('关联成功'),
          type: 'success',
        });
        getData();
        loading.value = false;
      } catch (error) {
        // 关联失败处理
        loading.value = false;
      }
    }
    model.value = {};
    openProductSelector.value = false;
  };

  // 序号方法
  const indexMethod = (index) => {
    return index + 1 + (curPage.value - 1) * pageSize.value;
  };
</script>

<style scoped>
  .common-level-rail {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }
</style>
