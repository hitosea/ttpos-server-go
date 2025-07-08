<template>
  <!--
      描述：属性库
  -->
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
      <ProductSelector
        maxCount="10"
        v-if="openProductSelector"
        :open="openProductSelector"
        @close="handleProductSelectorClose"
        selectorType="all"
        :selectedProductIds="model?.product_ids ?? []"
      >
      </ProductSelector>
      <GroupManager v-if="openGroupManager" :open="openGroupManager" @close="handleGroupManagerClose" />
    </div>
  </div>
</template>

<script>
  import ProductApi from '@/api/product.js';
  import Add from './add.vue';
  import Edit from './edit.vue';
  import ProductSelector from '@/components/product/Selector.vue';
  import GroupManager from './group.vue';
  import { ElMessageBox } from 'element-plus';

  export default {
    name: 'ProductExpandAttrIndex',
    components: {
      Add,
      Edit,
      ProductSelector,
      GroupManager,
    },
    data() {
      return {
        /*切换菜单*/
        activeName: 'sell',
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
        /*当前编辑的对象*/
        model: {},
        open_edit: false,
        open_add: false,
        open_manage_group: false,
        /*列表数据*/
        tableData: [],
        multipleSelection: [],
        //
        searchForm: {
          name: '',
        },
        searchLoading: '',
        openProductSelector: false,
        openGroupManager: false,
        treeData: [],
        attributeTreeCurrentKey: 0,
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
      attributeTableData() {
        const data = this.tableData ?? [];
        const treeData = this.treeData ?? [];
        return (
          data
            // .filter((value) => value.parent_id !== 0)
            .map((value) => ({
              attribute_id: value.attribute_id,
              attribute_name: JSON.parse(value.attribute_name || '{}'),
              attribute_name_text: value.attribute_name_text,
              group_id: value.parent_id,
              group_name_text: treeData.find((group) => group.attribute_id === value.parent_id)?.attribute_name_text,
              product_ids: value.product_ids,
              parent_id: value.parent_id,
            }))
        );
      },
    },
    created() {
      /*获取列表*/
      this.getData();
      this.getGroupData();
    },
    methods: {
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

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
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
        Params.parent_ids = self.attributeTreeCurrentKey === 0 ? '' : self.attributeTreeCurrentKey;
        self.loading = true;
        ProductApi.AttributeList(Params, true)
          .then((data) => {
            self.loading = false;
            if (typeof Params.parent_ids === 'number' && Params.parent_ids !== self.attributeTreeCurrentKey) return;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch(() => {
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
        });
      },
      attrjoin(e) {
        if (e) {
          return e.join('|');
        } else {
          return '';
        }
      },
      /*搜索查询*/
      onSubmit() {
        this.curPage = 1;
        this.getData();
      },

      /*打开添加*/
      addClick() {
        this.open_add = true;
      },
      deleteClick(id) {
        let self = this;
        ElMessageBox.confirm(self.$t('删除后不可恢复，确认删除吗?'), self.$t('提示'), {
          type: 'warning',
        }).then(() => {
          ProductApi.deleteAttribute({
            attribute_id: id,
          }).then(() => {
            self.$ElMessage({
              message: self.$t('删除成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
      deleteBatch() {
        let self = this;
        let arr = [];
        this.multipleSelection.forEach((item) => {
          arr.push(item.attribute_id);
        });
        let attribute_id = arr.join(',');
        ElMessageBox.confirm(self.$t('删除后不可恢复，确认删除吗?'), self.$t('提示'), {
          type: 'warning',
        }).then(() => {
          ProductApi.deleteAttribute({
            attribute_id: attribute_id,
          }).then(() => {
            self.$ElMessage({
              message: self.$t('删除成功'),
              type: 'success',
            });
            self.getData();
            self.getGroupData();
          });
        });
      },
      handleSelectionChange(e) {
        this.multipleSelection = e;
      },
      /*打开编辑*/
      editClick(row) {
        this.model = row;
        this.open_edit = true;
      },
      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success') {
            this.getData();
            this.getGroupData();
          }
        }
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getData();
            this.getGroupData();
          }
        }
        this.model = {};
      },
      handleOpenGroupManager() {
        this.openGroupManager = true;
      },
      handleGroupManagerClose() {
        this.openGroupManager = false;
        this.getData();
        this.getGroupData();
      },
      relatedProductClick(row) {
        this.model = row;
        this.openProductSelector = true;
      },
      handleProductSelectorClose(list) {
        if (Array.isArray(list)) {
          ProductApi.relateByAttr(
            {
              attribute_id: this.model.attribute_id,
              product_ids: list.map((item) => item.product_id),
            },
            false
          )
            .then(() => {
              this.$ElMessage({
                message: this.$t('关联成功'),
                type: 'success',
              });
              this.getData();
            })
            .catch();
        }
        this.model = {};
        this.openProductSelector = false;
      },
      handleAttributeTreeCurrentChange({ id }) {
        this.attributeTreeCurrentKey = id;
        this.getData();
      },
      indexMethod(index) {
        return index + 1 + (this.curPage - 1) * this.pageSize;
      },
    },
  };
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
