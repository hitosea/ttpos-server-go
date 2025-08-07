<template>
  <div class="product-wrapper">
    <div class="product-tree">
      <el-tree-v2
        ref="attributeTreeRef"
        :height="480"
        :data="attributeTreeData"
        node-key="id"
        highlight-current
        :current-node-key="attributeTreeCurrentKey"
        @current-change="handleAttributeTreeCurrentChange"
        auto-expand-parent
        :expand-on-click-node="false"
        :default-expanded-keys="[0]"
        :props="{ children: 'children', label: 'label' }"
      />
    </div>
    <div class="product-list">
      <!--添加属性-->
      <div class="common-level-rail">
        <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
          <el-form-item>
            <el-input size="small" v-model="searchForm.name" :placeholder="$t('属性名称')" @input="onSearch"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
              {{ $t('查询') }}
            </el-button>
          </el-form-item>
        </el-form>
        <div>
          <el-button size="small" type="primary" v-auth="'/product/expand/attr/edit'" @click="handleOpenGroupManager">{{ $t('管理属性组') }}</el-button>
          <el-button size="small" v-auth="'/product/expand/attr/batch_delete'" :disabled="multipleSelection.length == 0" @click="deleteBatch">{{ $t('批量删除') }}</el-button>
          <el-button size="small" type="primary" icon="Plus" v-auth="'/product/expand/attr/add'" @click="addClick">{{ $t('添加属性') }}</el-button>
        </div>
      </div>
      <!--内容-->
      <div class="product-content">
        <div class="table-wrap">
          <el-table size="small" :data="attributeTableData" border style="width: 100%" v-loading="loading" @selection-change="handleSelectionChange">
            <el-table-column type="selection" width="45"></el-table-column>
            <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
            <el-table-column prop="group_name_text" :label="$t('属性组')" width="400px"></el-table-column>
            <el-table-column prop="attribute_name_text" :label="$t('属性值')"></el-table-column>
            <!-- <el-table-column prop="sort" :label="$t('排序')"></el-table-column> -->
            <el-table-column prop="product_ids" :label="$t('关联商品数量')" width="120">
              <template #default="scope">
                {{ scope.row.product_ids?.length ?? 0 }}
              </template>
            </el-table-column>
            <el-table-column fixed="right" :label="$t('操作')" width="240">
              <template #default="scope">
                <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/expand/attr/edit'">{{ $t('编辑') }}</el-button>
                <el-button @click="relatedProductClick(scope.row)" v-auth="'/product/expand/attr/relatedProduct'" type="primary" link size="small">{{ $t('关联商品') }} </el-button>
                <el-button
                  @click="deleteClick(scope.row.attribute_id)"
                  type="primary"
                  link
                  size="small"
                  v-auth="'/product/expand/attr/delete'"
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
      <ProductSelector v-if="openProductSelector" :open="openProductSelector" @close="handleProductSelectorClose" selectorType="all" :isLoading="loading" :selectedProductIds="model?.product_ids ?? []">
      </ProductSelector>
      <GroupManager v-if="openGroupManager" :open="openGroupManager" @close="handleGroupManagerClose" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, getCurrentInstance } from 'vue';
import { ElMessageBox, ElMessage } from 'element-plus';
import ProductApi from '@/api/product.js';
import Add from './add.vue';
import Edit from './edit.vue';
import ProductSelector from '@/components/product/Selector.vue';
import GroupManager from './group.vue';

// 获取当前实例
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
const searchForm = reactive({
  name: '',
});
const searchLoading = ref('');
const openProductSelector = ref(false);
const openGroupManager = ref(false);
const treeData = ref([]);
const attributeTreeCurrentKey = ref(0);

// 计算属性
const attributeTreeData = computed(() => {
  const data = treeData.value ?? [];
  return [
    {
      id: 0,
      label: $t('全部'),
      children: data
        .map((group) => ({
          id: group.attribute_id,
          label: group.attribute_name_text,
        })),
    },
  ];
});

const attributeTableData = computed(() => {
  const data = tableData.value ?? [];
  const treeDataValue = treeData.value ?? [];
  return data.map((value) => ({
    attribute_id: value.attribute_id,
    attribute_name: JSON.parse(value.attribute_name || '{}'),
    attribute_name_text: value.attribute_name_text,
    group_id: value.parent_id,
    group_name_text: treeDataValue.find((group) => group.attribute_id === value.parent_id)?.attribute_name_text,
    product_ids: value.product_ids,
    parent_id: value.parent_id,
  }));
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
    attribute_name: searchForm.name,
    type: 2,
    parent_ids: attributeTreeCurrentKey.value === 0 ? '' : attributeTreeCurrentKey.value,
  };
  loading.value = true;
  
  try {
    const data = await ProductApi.AttributeList(params, true);
    loading.value = false;
    if (typeof params.parent_ids === 'number' && params.parent_ids !== attributeTreeCurrentKey.value) return;
    tableData.value = data.data.list.data;
    totalDataNumber.value = data.data.list.total;
  } catch (error) {
    loading.value = false;
  }
};

// 获取组数据
const getGroupData = async () => {
  try {
    const data = await ProductApi.AttributeList(
      {
        page: 1,
        list_rows: 10000,
        type: 1,
      },
      true
    );
    treeData.value = data.data.list.data;
  } catch (error) {
    console.error('获取组数据失败', error);
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
    
    await ProductApi.deleteAttribute({
      attribute_id: id,
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
    arr.push(item.attribute_id);
  });
  const attribute_id = arr.join(',');
  
  try {
    await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    });
    
    await ProductApi.deleteAttribute({
      attribute_id: attribute_id,
    });
    
    ElMessage({
      message: $t('删除成功'),
      type: 'success',
    });
    getData();
    getGroupData();
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

// 关闭弹窗
const closeDialogFunc = (e, f) => {
  if (f == 'add') {
    open_add.value = e.openDialog;
    if (e.type == 'success') {
      getData();
      getGroupData();
    }
  }
  if (f == 'edit') {
    open_edit.value = e.openDialog;
    if (e.type == 'success') {
      getData();
      getGroupData();
    }
  }
  model.value = {};
};

// 打开组管理器
const handleOpenGroupManager = () => {
  openGroupManager.value = true;
};

// 组管理器关闭
const handleGroupManagerClose = () => {
  openGroupManager.value = false;
  getData();
  getGroupData();
};

// 关联商品
const relatedProductClick = (row) => {
  model.value = row;
  openProductSelector.value = true;
};

// 商品选择器关闭
const handleProductSelectorClose = async (list) => {
  if (Array.isArray(list)) {
    try {
      loading.value = true;
      await ProductApi.relateByAttr(
        {
          attribute_id: model.value.attribute_id,
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

// 属性树当前变化
const handleAttributeTreeCurrentChange = ({ id }) => {
  attributeTreeCurrentKey.value = id;
  getData();
};

// 序号方法
const indexMethod = (index) => {
  return index + 1 + (curPage.value - 1) * pageSize.value;
};

// 组件挂载时获取数据
onMounted(() => {
  getData();
  getGroupData();
});
</script>

<style lang="scss" scoped>
  .common-level-rail {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .product-wrapper {
    display: flex;
    justify-content: flex-start;
    align-items: stretch;
    gap: 8px;

    .product-tree {
      width: 240px;
      // height: 480px;
      flex-shrink: 0;
      overflow-x: hidden;
      overflow-y: auto;
    }

    .product-list {
      flex-grow: 1;
      overflow: hidden;
    }
  }
</style>
