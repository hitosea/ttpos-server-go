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
      <el-table :data="tableData" style="width: 100%" @selection-change="handleSelectionChange" ref="multipleTable" max-height="480" :row-key="getRowKey">
        <el-table-column type="selection" width="40" :reserve-selection="true" />
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
        await Promise.resolve().then(() => {
          tableData.value.map((row, index) => {
            if (props.feed_ids.includes(row.feed_id)) {
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

  // 选择变化处理
  const handleSelectionChange = (val) => {
    selectedProductsTmp.value = val;
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
