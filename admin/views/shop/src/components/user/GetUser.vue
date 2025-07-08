<template>
  <el-dialog :title="$t('选择用户')" v-model="dialogVisible" @close="cancelFunc" append-to-body :close-on-click-modal="false" :close-on-press-escape="false" width="800px">
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('等级')">
          <el-select v-model="formInline.grade_id" :placeholder="$t('请选择')" style="width: 120px">
            <el-option :label="$t('全部')" value="0"></el-option>
            <el-option v-for="(item, index) in gradeList" :key="index" :label="item.name" :value="item.grade_id"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('性别')">
          <el-select v-model="formInline.gender" :placeholder="$t('请选择')" style="width: 120px">
            <el-option :label="$t('全部')" value="-1"></el-option>
            <el-option v-for="(item, index) in sex" :key="index" :label="item" :value="index"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('昵称/手机号/ID')"><el-input :placeholder="$t('昵称/手机号/ID')" v-model="formInline.keyword"></el-input></el-form-item>
        <el-form-item style="margin-right: 0">
          <el-button class="search-button" type="primary" icon="Search" @click="search">{{ $t('查询') }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <div v-if="isShowTips" class="tips">{{ $t('注：如选择已有会员卡用户，将会根据最新操作更换其会员卡') }}</div>
        <el-table
          ref="multipleTable"
          :data="tableData"
          size="small"
          border
          style="width: 100%"
          v-loading="loading"
          @selection-change="handleSelectionChange"
          @row-click="handleRowClick"
          :highlight-current-row="is_single"
        >
          <!-- <el-table-column prop="" label="微信头像" width="70">
            <template #default="scope">
              <img :src="scope.row.avatarUrl" class="radius" :width="30" :height="30" />
            </template>
          </el-table-column> -->
          <el-table-column prop="nickName" :label="$t('昵称')"></el-table-column>
          <el-table-column prop="balance" :label="$t('主账户余额')" width="120">
            <template #default="scope">
              <span class="orange">{{ this.$formatPrice(scope.row.balance) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="balance" :label="$t('赠送账户余额')" width="120">
            <template #default="scope">
              <span class="orange">{{ this.$formatPrice(scope.row.gift_balance) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="" :label="$t('会员卡')">
            <template #default="scope">
              <span v-if="scope.row.card_id == 0">-</span>
              <span v-else>{{ scope.row.card?.card_name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="grade.name" :label="$t('会员等级')"></el-table-column>
          <el-table-column prop="pay_money" :label="$t('累积消费金额')" width="120">
            <template #default="scope">
              <span>{{ this.$formatPrice(scope.row.pay_money) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="gender" :label="$t('性别')" width="50">
            <template #default="scope">
              <span v-if="scope.row.gender == 1">{{ $t('男') }}</span>
              <span v-else-if="scope.row.gender == 0">{{ $t('女') }}</span>
              <span v-else-if="scope.row.gender == 2">{{ $t('保密') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('注册时间')" width="140"></el-table-column>

          <el-table-column v-if="!is_single" type="selection" width="45" fixed="right"></el-table-column>

          <el-table-column v-else type="radio" width="45" fixed="right">
            <template #default="scope">
              <div class="radio-box">
                <el-radio v-model="selectedRow" :label="scope.row" @change="handleRadioChange(scope.row)">&nbsp;</el-radio>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!--分页-->
      <div class="pagination">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          background
          :current-page="curPage"
          :page-sizes="[2, 10, 20, 50, 100]"
          :page-size="pageSize"
          layout="total, prev, pager, next, jumper"
          :total="totalDataNumber"
        ></el-pagination>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button size="small" @click="dialogVisible = false">{{ $t('取消') }}</el-button>
        <el-button size="small" v-if="!is_single" type="primary" @click="confirmFunc">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import DataApi from '@/api/data.js';
  export default {
    data() {
      return {
        /*是否加载完成*/
        loading: true,
        /*当前是第几页*/
        curPage: 1,
        /*一页多少条*/
        pageSize: 15,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*搜索表单对象*/
        formInline: {
          /*等级*/
          grade_id: '',
          /*昵称*/
          keyword: '',
          /*性别*/
          gender: '',
        },
        /*会员等级列表*/
        gradeList: [],
        /*会员列表*/
        tableData: [],
        /*性别列表*/
        sex: [$t('女'), $t('男'), $t('保密')],
        /*选中的*/
        multipleSelection: [],
        /*是否显示*/
        dialogVisible: false,
        /*单选模式*/
        is_single: false,
        /*选中的行*/
        selectedRow: null,
      };
    },
    props: {
      isShowTips: {
        type: Boolean,
        default: false,
      },
      detailSelection: [],
      is_open: Boolean,
      is_single: {
        type: Boolean,
        default: false,
      }, //是否单选
      exclude_user_id: {
        type: [Number, String],
        default: '',
      }, //排除的用户ID
    },
    watch: {
      is_open: function (n, o) {
        if (n != o) {
          this.dialogVisible = n;
          if (n) {
            // 打开弹窗时重置选择状态
            this.selectedRow = null;
            this.multipleSelection = [];
            this.getTableList();
          }
        }
      },
      // 监听is_single变化
      is_single: {
        immediate: true,
        handler(val) {
          // 切换模式时重置选择状态
          this.selectedRow = null;
          this.multipleSelection = [];
        },
      },
    },
    created() {
      // 初始化时设置单选/多选模式
      this.is_single = this.$props.is_single;
    },
    methods: {
      /*选择第几页*/
      handleCurrentChange(val) {
        this.curPage = val;
        this.getTableList();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getTableList();
      },

      /*获取数据*/
      getTableList() {
        let self = this;
        self.loading = true;
        let params = self.formInline;
        params.page = self.curPage;
        params.list_rows = self.pageSize;
        params.exclude_user_id = self.exclude_user_id;
        DataApi.getUser(params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
            self.gradeList = data.data.grade;

            // 默认选中ID
            if (this.detailSelection.length > 0) {
              this.$nextTick(() => {
                this.tableData.forEach((row) => {
                  if (this.detailSelection.includes(row.id)) {
                    this.$refs.multipleTable.toggleRowSelection(row, true);
                    this.multipleSelection.push(row);
                  }
                });
              });
            }
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*查询*/
      search() {
        this.curPage = 1;
        this.tableData = [];
        this.getTableList();
      },

      /*点击确定*/
      confirmFunc() {
        let params;
        if (this.is_single) {
          // 单选模式
          if (this.selectedRow) {
            params = [this.selectedRow];
          } else {
            // 如果没有选择，不关闭弹窗
            this.$message.warning(this.$t('请选择用户'));
            return;
          }
        } else {
          // 多选模式
          params = this.multipleSelection;
          if (params.length === 0) {
            this.$message.warning(this.$t('请选择用户'));
            return;
          }
        }
        this.emitFunc(params);
      },

      /*关闭弹窗*/
      cancelFunc() {
        // 清空选择
        this.selectedRow = null;
        this.multipleSelection = [];
        this.emitFunc();
      },

      /*发送事件*/
      emitFunc(e) {
        this.dialogVisible = false;
        if (e && typeof e != 'undefined') {
          this.$emit('close', {
            type: 'success',
            params: e,
          });
        } else {
          this.$emit('close', {
            type: 'error',
          });
        }
      },

      /*选择用户*/
      handleSelectionChange(e) {
        if (!this.is_single) {
          this.multipleSelection = e;
        }
      },

      /*处理行点击事件*/
      handleRowClick(row) {
        if (this.is_single) {
          this.selectedRow = row;
          this.confirmFunc();
        }
      },

      /*处理单选按钮变化事件*/
      handleRadioChange(row) {
        this.selectedRow = row;
        this.confirmFunc();
      },
    },
  };
</script>
<style scoped lang="scss">
  :deep(.el-select--small) {
    .el-select__wrapper {
      min-width: auto;
    }
  }
  .radio-box {
    display: flex;
    align-items: center;
    justify-content: center;
    padding-left: 14px;
    width: 100%;
    height: 100%;
  }
</style>
