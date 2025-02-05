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
<script>
  import ProductApi from '@/api/product.js';
  import Add from '../../../expand/feed/add.vue';
  import { languageStore } from '@/store/model/language.js';
  const languageKey = languageStore().getLanguageKey().language.value;
  export default {
    name: 'AddFeed',
    components: { Add },
    props: {
      open: {
        type: Boolean,
        default: false,
      },
      feed_ids: {
        type: Array,
        default: () => [],
      },
    },

    created() {
      this.dialogVisible = this.open;
      this.getData();
    },
    data() {
      return {
        languageKey: languageKey,
        dialogVisible: false,
        loading: false,
        tableData: [],
        totalDataNumber: 0,
        curPage: 1,
        pageSize: 1000,
        searchForm: {
          name: '',
        },
        selectedProductsTmp: [],
        open_add: false,
        searchLoading: '',
      };
    },
    methods: {
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
        Params.feed_name = self.searchForm.name;
        self.loading = true;
        ProductApi.FeedList(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            if (this.feed_ids.length > 0) {
              // 判断是否存在勾选过的数据
              this.tableData.map((row, index) => {
                // 获取数据列表接口请求到的数据
                if (this.feed_ids.includes(row.feed_id)) {
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
        return row.feed_id;
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
    },
  };
</script>
<style lang="scss" scoped>
  .dialog-add {
    float: left;
  }
</style>
