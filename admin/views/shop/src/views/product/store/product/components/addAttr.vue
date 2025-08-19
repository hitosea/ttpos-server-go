<template>
  <el-dialog
    class="feed-selector"
    @close="dialogFormVisible"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :title="$t('选择属性')"
    append-to-body
  >
    <div class="product-boxs">
      <div class="product-tree">
        <el-tree
          ref="attributeTreeRef"
          :data="attributeTreeData"
          node-key="id"
          show-checkbox
          default-expand-all
          @check="handleGroupCheckChange"
          :props="{ children: 'children', label: 'label' }"
        />
      </div>
      <div class="feed-selector__body">
        <el-form size="small" ref="formRef" :model="searchForm" :inline="true">
          <el-form-item>
            <el-input size="small" v-model="searchForm.name" @input="onDebounceSearch" :placeholder="$t('属性组属性值')" />
          </el-form-item>
          <el-form-item>
            <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
              {{ $t('查询') }}
            </el-button>
          </el-form-item>
        </el-form>
        <el-table :data="tableData" ref="multipleTable" style="width: 100%" @selection-change="handleSelectionChange" :row-key="getRowKey">
          <el-table-column type="selection" width="45" :selectable="selectable" :reserve-selection="true"></el-table-column>
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="group_name_text" :label="$t('属性组')" minWidth="140"></el-table-column>
          <el-table-column prop="attribute_name_text" :label="$t('属性值')" minWidth="140"></el-table-column>
        </el-table>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button class="dialog-add" type="primary" size="small" icon="Plus" @click="handleAdd">{{ $t('新增属性') }}</el-button>
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
  <!--添加-->
  <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialogFunc($event, 'add')"></Add>
</template>

<script setup>
  import { ref, reactive, computed, watch, nextTick, onMounted } from 'vue';
  import ProductApi from '@/api/product.js';
  import Add from '../../../expand/attr/add.vue';

  // 定义props
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    select_arr: {
      type: Array,
      default: () => [],
    },
  });

  // 定义emits
  const emit = defineEmits(['close']);

  // 响应式数据
  const dialogVisible = ref(false);
  const loading = ref(false);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const pageSize = ref(1000);
  const searchForm = reactive({
    name: '',
  });
  const selectedProductsTmp = ref([]);
  const open_add = ref(false);
  const tableData = ref([]);
  const checkedGroupIds = ref([0]);
  const treeData = ref([]);
  const searchTimer = ref('');
  const attributeTreeRef = ref(null);
  const multipleTable = ref(null);
  const formRef = ref(null);
  const model = ref({});

  // 计算属性：属性树数据
  const attributeTreeData = computed(() => {
    const data = treeData.value ?? [];
    return [
      {
        id: 0,
        label: $t('全部'),
        children: data.map((group) => ({
          id: group.attribute_id,
          label: group.attribute_name_text,
        })),
      },
    ];
  });

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
      await getData();

      // 设置树形组件的选中状态
      nextTick(() => {
        attributeTreeRef.value?.setCheckedKeys(checkedGroupIds.value);
      });
    } catch (error) {
      // 错误处理
    }
  };

  // 监听open属性变化
  watch(
    () => props.open,
    (val) => {
      dialogVisible.value = val;
      if (val) {
        getGroupData();
      }
    },
    { immediate: true }
  );

  // 防抖搜索
  const onDebounceSearch = () => {
    clearTimeout(searchTimer.value);
    searchTimer.value = setTimeout(() => {
      getData();
    }, 200);
  };

  // 搜索查询
  const onSearch = () => {
    getData();
  };

  // 获取列表
  const getData = async () => {
    loading.value = true;
    try {
      const Params = {
        page: curPage.value,
        list_rows: pageSize.value,
        attribute_name: searchForm.name,
        type: 2,
        parent_ids: checkedGroupIds.value.includes(0) ? '' : checkedGroupIds.value.join(','),
      };
      const data = await ProductApi.AttributeList(Params, true);
      loading.value = false;
      tableData.value = data.data.list.data;
      totalDataNumber.value = data.data.list.total;

      // 补充组名信息
      await Promise.resolve().then(() => {
        treeData.value.map((group) => {
          tableData.value.map((row) => {
            if (group.attribute_id == row.parent_id) {
              row.group_name_text = group.attribute_name_text;
            }
          });
        });
      });

      // 判断是否存在勾选过的数据
      if (props.select_arr.length > 0) {
        await Promise.resolve().then(() => {
          tableData.value.map((row, index) => {
            if (props.select_arr.includes(row.attribute_id)) {
              nextTick(() => {
                multipleTable.value?.toggleRowSelection(tableData.value[index], true);
              });
              tableData.value[index].select_open = 1;
            }
          });
        });
      }
    } catch (error) {
      loading.value = false;
    }
  };

  // 组选择变化处理
  const handleGroupCheckChange = (_, { checkedKeys }) => {
    checkedGroupIds.value = checkedKeys;
    getData();
  };

  // 关闭弹窗
  const closeDialogFunc = async (e, f) => {
    if (f == 'add') {
      open_add.value = e.openDialog;
      if (e.type == 'success') {
        await getData();
      }
    }
  };

  // 获取行键值
  const getRowKey = (row) => {
    return row.attribute_id;
  };

  // 是否可选
  const selectable = (row) => {
    if (row.select_open != undefined && row.select_open == 1) {
      return false;
    } else {
      return true;
    }
  };

  // 选择变化处理
  const handleSelectionChange = (val) => {
    selectedProductsTmp.value = val;
  };

  // 关闭对话框
  const dialogFormVisible = () => {
    emit('close', false);
  };

  // 新增属性
  const handleAdd = () => {
    open_add.value = true;
  };

  // 确定选择
  const handleClick = () => {
    emit('close', selectedProductsTmp.value);
  };

  // 序号方法
  const indexMethod = (index) => {
    return index + 1 + (curPage.value - 1) * pageSize.value;
  };

  onMounted(() => {
    dialogVisible.value = props.open;
    if (props.open) {
      getGroupData();
    }
  });
</script>

<style lang="scss" scoped>
  .dialog-add {
    float: left;
  }
  .product-boxs {
    display: flex;
    gap: 12px;
    overflow: hidden;
    .product-tree {
      flex: 1;
      overflow: hidden;
      :deep(.el-tree-node__content) {
        overflow: hidden;
        .el-tree-node__label {
          white-space: nowrap; /* 不换行 */
          overflow: hidden; /* 超出的内容隐藏 */
          text-overflow: ellipsis; /* 使用省略号显示被截断的文本 */
        }
      }
    }
    .feed-selector__body {
      flex: 2;
      overflow: hidden;
    }
  }
</style>
