<template>
  <el-dialog class="attribute-group-selector" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('管理属性组')">
    <div class="attribute-group-selector-content">
      <div class="attribute-group-selector-form">
        <el-table :data="tableData" size="small" border style="width: 100%" v-loading="loading">
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center" :index="indexMethod"></el-table-column>
          <el-table-column prop="attribute_name_text" :label="$t('属性组名称')" />
          <el-table-column prop="children" :label="$t('关联属性值')" width="120">
            <template #default="scope">
              {{ scope.row.children.length }}
            </template>
          </el-table-column>

          <el-table-column fixed="right" :label="$t('操作')" width="240">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/product/expand/attr/edit'">{{ $t('编辑') }}</el-button>
              <el-button
                @click="deleteClick(scope.row.attribute_id)"
                type="primary"
                link
                size="small"
                v-auth="'/product/expand/attr/delete'"
                :disabled="scope.row.children?.length > 0"
                >{{ $t('删除') }}</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="pagination">
      <el-pagination
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
        background
        :current-page="pagination.page"
        :page-size="pagination.pageSize"
        layout="total, prev, pager, next, jumper"
        :total="pagination.total"
      ></el-pagination>
    </div>

    <Edit v-if="openEdit" :open_edit="openEdit" @closeDialog="handleCloseEdit" :editform="currentEdit" />
  </el-dialog>
</template>

<script setup>
  defineOptions({
    name: 'ProductAttributeGroupSelectorDialog',
  });

  import { ref, reactive, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product.js';
  import Edit from './edit.vue';
  import { ElMessageBox } from 'element-plus';

  const { proxy } = getCurrentInstance();

  const props = defineProps({
    open: {
      type: Boolean,
      default: false,
    },
  });

  const emit = defineEmits(['close']);

  const dialogVisible = ref(props.open);

  const loading = ref(false);

  const tableData = ref([]);

  const openEdit = ref(false);

  const currentEdit = ref(null);

  const pagination = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPage: 1,
  });

  const indexMethod = (index) => {
    return index + 1 + (pagination.page - 1) * pagination.pageSize;
  };

  const handlePageSizeChange = (size) => {
    pagination.page = 1;
    pagination.pageSize = size;
    getData();
  };

  const handlePageChange = (page) => {
    pagination.page = page;
    getData();
  };

  const getData = async () => {
    loading.value = true;
    try {
      const { data } = await ProductApi.AttributeList(
        {
          page: pagination.page,
          pageSize: pagination.pageSize,
          type: 1,
        },
        true
      );

      pagination.page = data.list.current_page;
      pagination.pageSize = data.list.per_page;
      pagination.total = data.list.total;
      pagination.totalPage = data.list.last_page;

      tableData.value = data.list.data;
    } catch (error) {
      //
    } finally {
      loading.value = false;
    }
  };

  const editClick = (row) => {
    currentEdit.value = {
      attribute_id: row.attribute_id,
      attribute_name: JSON.parse(row.attribute_name),
      attribute_name_text: row.attribute_name_text,
      sort: row.sort,
    };
    openEdit.value = true;
  };

  const deleteClick = async (attribute_id) => {
    ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    }).then(() => {
      ProductApi.deleteAttribute({
        attribute_id,
      }).then(() => {
        proxy.$ElMessage({
          message: $t('删除成功'),
          type: 'success',
        });
        getData();
      });
    });
  };

  const handleCloseEdit = () => {
    openEdit.value = false;
    currentEdit.value = null;
    getData();
  };

  const handleClose = () => {
    emit('close');
  };

  getData();
</script>

<style lang="scss" scoped>
  .attribute-group-selector-content {
    display: flex;
    justify-content: flex-start;
    align-items: stretch;
    gap: 8px;

    .attribute-group-selector-tree {
      width: 240px;
      flex-shrink: 0;
      overflow-x: hidden;
    }

    .attribute-group-selector-divider {
      margin: 0 4px;
      width: 2px;
      flex-shrink: 0;
      background-color: #f0f2f5;
    }

    .attribute-group-selector-form {
      flex-grow: 1;
    }
  }
</style>
