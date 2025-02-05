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
<script>
  import ProductApi from '@/api/product.js';
  import Add from '../../../expand/attr/add.vue';
  export default {
    name: 'AddFeed',
    components: { Add },
    props: {
      open: {
        type: Boolean,
        default: false,
      },
      select_arr: {
        type: Array,
        default: () => [],
      },
    },

    created() {
      this.dialogVisible = this.open;
      /*获取列表*/

      this.getGroupData();

      this.$nextTick(() => {
        this.$refs.attributeTreeRef.setCheckedKeys(this.checkedGroupIds);
      });
    },
    data() {
      return {
        dialogVisible: false,
        loading: false,
        totalDataNumber: 0,
        curPage: 1,
        pageSize: 1000,
        searchForm: {
          name: '',
        },
        selectedProductsTmp: [],
        open_add: false,
        tableData: [],
        checkedGroupIds: [0],
        treeData: [],
        searchTimer: '',
      };
    },
    computed: {
      attributeTreeData() {
        const data = this.treeData ?? [];
        return [
          {
            id: 0,
            label: this.$t('全部'),
            children: data
              // .filter((group) => group.parent_id === 0)
              .map((group) => ({
                id: group.attribute_id,
                label: group.attribute_name_text,
              })),
          },
        ];
      },
    },

    methods: {
      /*获取列表*/
      onDebounceSearch() {
        clearTimeout(this.searchTimer);
        this.searchTimer = setTimeout(() => {
          this.getData();
        }, 200);
      },
      /*获取列表*/
      getData() {
        let self = this;
        let Params = {};
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        Params.attribute_name = self.searchForm.name;
        Params.type = 2;
        Params.parent_ids = self.checkedGroupIds.includes(0) ? '' : self.checkedGroupIds.join(',');
        self.loading = true;
        ProductApi.AttributeList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.treeData.map((group) => {
              self.tableData.map((row) => {
                if (group.attribute_id == row.parent_id) {
                  row.group_name_text = group.attribute_name_text;
                }
              });
            });
            if (this.select_arr.length > 0) {
              // 判断是否存在勾选过的数据
              this.tableData.map((row, index) => {
                // 获取数据列表接口请求到的数据
                if (this.select_arr.includes(row.attribute_id)) {
                  this.$nextTick(() => {
                    this.$refs.multipleTable.toggleRowSelection(this.tableData[index], true); // 若有重合，则回显该条数据
                  });
                  this.tableData[index].select_open = 1;
                }
              });
            }
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      getGroupData() {
        ProductApi.AttributeList(
          {
            page: 1,
            list_rows: 10000,
            type: 1,
          },
          true
        ).then((data) => {
          this.treeData = data.data.list.data;
          this.getData();
        });
      },
      handleGroupCheckChange(_, { checkedKeys }) {
        this.checkedGroupIds = checkedKeys;
        this.getData();
      },
      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
      },

      getRowKey(row) {
        return row.attribute_id;
      },

      selectable(row) {
        if (row.select_open != undefined && row.select_open == 1) {
          return false;
        } else {
          return true;
        }
      },

      handleSelectionChange(val) {
        this.selectedProductsTmp = val;
      },

      dialogFormVisible() {
        this.$emit('close', false);
      },
      handleAdd() {
        this.open_add = true;
      },
      handleClick() {
        this.$emit('close', this.selectedProductsTmp);
      },
      indexMethod(index) {
        return index + 1 + (this.curPage - 1) * this.pageSize;
      },
    },
  };
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
