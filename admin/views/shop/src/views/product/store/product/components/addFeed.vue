<template>
  <el-dialog
    class="feed-selector"
    @close="dialogFormVisible"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :title="$t('选择加料')"
    append-to-body
  >
    <div class="feed-selector__body" v-loading="loading">
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
      <el-table :data="tableData" style="width: 100%" ref="multipleTable" max-height="480" :row-key="getRowKey">
        <el-table-column width="40" :reserve-selection="true">
          <template #header>
            <el-checkbox 
              v-model="selectAll" 
              :indeterminate="isIndeterminate"
              @change="handleSelectAll"
            >
            </el-checkbox>
          </template>
          <template #default="scope">
            <el-checkbox 
              :model-value="isRowSelected(scope.row)"
              @change="(checked) => handleRowSelect(scope.row, checked)"
              :disabled="!checkSelectable(scope.row)"
            >
            </el-checkbox>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="$t('序号')" width="90" align="center">
          <template #default="scope">
            {{ scope.$index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="feed_name" :label="$t('加料名称')">
          <template #default="scope">
            {{ JSON.parse(scope.row.feed_name)[languageKey] || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="$t('关联材料')" width="120">
          <template #default="scope">
            {{ scope.row.material?.length ?? 0 }}
          </template>
        </el-table-column>
      </el-table>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button class="dialog-add" type="primary" size="small" icon="Plus" @click="handleAdd">{{ $t('新增加料') }}</el-button>
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
  <!--添加-->
  <Add v-if="open_add" :open_add="open_add" :addform="model" @closeDialog="closeDialogFunc($event, 'add')"></Add>
</template>

<script setup>
  import { ref, reactive, watch, nextTick, onMounted } from 'vue';
  import { ElMessage, ElCheckbox } from 'element-plus';
  import ProductApi from '@/api/product.js';
  import Add from '../../../expand/feed/add.vue';
  import { languageStore } from '@/store/model/language.js';

  // 获取语言配置
  const languageKey = languageStore().getLanguageKey().language.value;

  // 定义props
  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
    feed_ids: {
      type: Array,
      default: () => [],
    },
    // 最大选择数量
    maxSelect: {
      type: Number,
      default: Infinity,
    },
  });

  // 定义emits
  const emit = defineEmits(['close']);

  // 响应式数据
  const dialogVisible = ref(false);
  const loading = ref(false);
  const tableData = ref([]);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const pageSize = ref(1000);
  const searchForm = reactive({
    name: '',
  });
  const selectedProductsTmp = ref([]);
  const open_add = ref(false);
  const searchLoading = ref('');
  const multipleTable = ref(null);
  const model = ref({});
  const selectAll = ref(false);
  const isIndeterminate = ref(false);

  // 获取列表
  const getData = async () => {
    loading.value = true;
    try {
      const Params = {
        page: curPage.value,
        list_rows: pageSize.value,
        feed_name: searchForm.name,
      };
      const data = await ProductApi.FeedList(Params, true);
      loading.value = false;
      tableData.value = data.data.list.data;
      totalDataNumber.value = data.data.list.total;

      // 判断是否存在勾选过的数据
      if (props.feed_ids.length > 0) {
        // 根据传入的feed_ids设置已选中的项目
        const preSelectedItems = tableData.value.filter(row => props.feed_ids.includes(row.feed_id));
        selectedProductsTmp.value = preSelectedItems;
      }
      
      // 更新全选状态
      nextTick(() => {
        updateSelectAllState();
      });
    } catch (error) {
      loading.value = false;
    }
  };

  // 监听open属性变化
  watch(
    () => props.open,
    (val) => {
      dialogVisible.value = val;
      if (val) {
        getData();
      }
    },
    { immediate: true }
  );

  // 搜索查询
  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      getData();
    }, 200);
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
    return row.feed_id;
  };

  // 检查行是否被选中
  const isRowSelected = (row) => {
    return selectedProductsTmp.value.some((item) => item.feed_id === row.feed_id);
  };

  // 处理单行选择
  const handleRowSelect = (row, checked) => {
    if (checked) {
      // 检查是否超出最大选择数量
      if (selectedProductsTmp.value.length >= props.maxSelect) {
        ElMessage({
          message: $t('最多只能选择') + props.maxSelect + $t('个'),
          type: 'warning',
        });
        return;
      }
      // 添加到选中列表
      selectedProductsTmp.value.push(row);
    } else {
      // 从选中列表移除
      const index = selectedProductsTmp.value.findIndex((item) => item.feed_id === row.feed_id);
      if (index > -1) {
        selectedProductsTmp.value.splice(index, 1);
      }
    }
    updateSelectAllState();
  };

  // 处理全选
  const handleSelectAll = (checked) => {
    if (checked) {
      // 全选逻辑：选中前 maxSelect 个项目
      const currentSelected = [...selectedProductsTmp.value];
      const currentSelectedIds = currentSelected.map((item) => item.feed_id);
      
      // 找到未选中的项目
      const unselectedItems = tableData.value.filter((row) => !currentSelectedIds.includes(row.feed_id));
      
      // 计算还能选择多少个
      const remainingSlots = props.maxSelect - currentSelected.length;
      
      if (remainingSlots > 0) {
        // 从未选中的项目中按顺序选择剩余数量
        const itemsToAdd = unselectedItems.slice(0, remainingSlots);
        selectedProductsTmp.value = [...currentSelected, ...itemsToAdd];
      }
      
      if (selectedProductsTmp.value.length === props.maxSelect && unselectedItems.length > remainingSlots) {
        ElMessage({
          message: $t('最多只能选择') + props.maxSelect + $t('个'),
          type: 'info',
        });
      }
    } else {
      // 取消全选
      selectedProductsTmp.value = [];
    }
    updateSelectAllState();
  };

  // 更新全选状态
  const updateSelectAllState = () => {
    const selectedCount = selectedProductsTmp.value.length;
    const totalCount = tableData.value.length;
    
    if (selectedCount === 0) {
      selectAll.value = false;
      isIndeterminate.value = false;
    } else if (selectedCount === totalCount || selectedCount === props.maxSelect) {
      selectAll.value = true;
      isIndeterminate.value = false;
    } else {
      selectAll.value = false;
      isIndeterminate.value = true;
    }
  };

  // 检查行是否可选择
  const checkSelectable = (row) => {
    // 如果当前行已经被选中，允许取消选择
    const isCurrentRowSelected = selectedProductsTmp.value.some((item) => item.feed_id === row.feed_id);
    if (isCurrentRowSelected) {
      return true;
    }

    // 如果已选择的数量达到最大限制，禁止选择新的行
    if (selectedProductsTmp.value.length >= props.maxSelect) {
      return false;
    }

    return true;
  };


  // 关闭对话框
  const dialogFormVisible = () => {
    emit('close', false);
  };

  // 新增加料
  const handleAdd = () => {
    open_add.value = true;
  };

  // 确定选择
  const handleClick = () => {
    emit('close', selectedProductsTmp.value);
  };

  onMounted(() => {
    dialogVisible.value = props.open;
    if (props.open) {
      getData();
    }
  });
</script>

<style lang="scss" scoped>
  .dialog-add {
    float: left;
  }
</style>
