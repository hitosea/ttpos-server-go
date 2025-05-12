<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-seach-wrap">
      <el-form size="small" :inline="true" :model="formInline" class="demo-form-inline">
        <el-form-item :label="$t('状态')">
          <a-select v-model:value="formInline.status" :placeholder="$t('全部')" @change="onSearch">
            <el-option :label="$t('全部')" value=""></el-option>
            <el-option :label="$t('过期')" :value="0"></el-option>
            <el-option :label="$t('有效')" :value="1"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('会员卡名称')">
          <el-input v-model="formInline.card_name" :placeholder="$t('请输入会员卡名称')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('领取时间')">
          <div class="block">
            <el-date-picker
              v-model="formInline.date"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              clearable
              @change="onSearch"
            >
            </el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="order_id" label="ID"></el-table-column>
          <el-table-column prop="user.nickName" :label="$t('昵称')"></el-table-column>
          <el-table-column prop="user.mobile" :label="$t('手机号')"> </el-table-column>
          <el-table-column prop="user.user_id" :label="$t('会员ID')"> </el-table-column>
          <el-table-column prop="card.card_name" :label="$t('会员卡名称')"> </el-table-column>
          <el-table-column prop="expire_time_text" :label="$t('有效期')"> </el-table-column>
          <el-table-column prop="discount" :label="$t('折扣')">
            <template #default="scope">
              <span v-if="scope.row.discount">{{ scope.row.discount }}%</span>
              <span v-else>{{ $t('无') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="pay_price" :label="$t('价格')">
            <template #default="scope">
              {{ this.$formatPrice(scope.row.pay_price) }}
            </template>
          </el-table-column>

          <el-table-column prop="pay_time_text" :label="$t('领取时间')"></el-table-column>
          <el-table-column fixed="right" :label="$t('操作')" width="160">
            <template #default="scope">
              <el-button
                v-auth="'/card/card/record/adjust'"
                @click="putClick(scope.row)"
                type="primary"
                link
                size="small"
                :disabled="scope.row.is_delete == 1"
                v-if="scope.row.expire_time >= 0"
                >{{ $t('调整有效期') }}</el-button
              >
              <el-button
                v-auth="'/card/card/record/cancel'"
                @click="cancel(scope.row)"
                type="primary"
                link
                size="small"
                :disabled="scope.row.is_delete == 1 || scope.row.is_used == 1"
                v-if="scope.row.pay_type == 30"
                >{{ $t('撤销') }}</el-button
              >
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
          :page-size="pageSize"
          layout="total, prev, pager, next, jumper"
          :total="totalDataNumber"
        >
        </el-pagination>
      </div>
    </div>
    <!--发卡-->
    <expire v-if="open_edit" :open_edit="open_edit" :form="userModel" @closeDialog="closeDialogFunc($event, 'edit')"></expire>
  </div>
</template>

<script>
  import Aselect from '@/components/a-select/index.vue';
  import CardApi from '@/api/card.js';
  import expire from '../dialog/expire.vue';
  import dayjs from '@/utils/dayjs';
  export default {
    components: {
      expire,
      Aselect,
    },
    data() {
      return {
        /*是否加载完成*/
        loading: true,
        /*列表数据*/
        tableData: [],
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        /*横向表单数据模型*/
        formInline: {
          card_name: '',
          status: '',
          date: '',
        },
        open_edit: false,
        userModel: {},
        searchLoading: '',
      };
    },
    mounted() {
      // js获取当天时间 日期格式 YYYY-MM-DD
      this.formInline.date = [dayjs(), dayjs()];
      /*获取列表*/
      this.getTableList();
    },
    methods: {
      /*换行*/
      keepTextStyle(val) {
        let str = val.replace(/(\\r\\n)/g, '<br/>');
        return str;
      },

      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.curPage = val;
        self.loading = true;
        self.getTableList();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getTableList();
      },

      /*获取列表*/
      getTableList() {
        let self = this;
        let Params = self.formInline;
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        self.loading = true;
        CardApi.recordlist(Params, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list.data;
            self.totalDataNumber = data.data.list.total;
          })
          .catch((error) => {});
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getTableList();
        }, 200);
      },

      /*打开添加*/
      addClick() {
        this.$router.push('/card/card/add');
      },
      /*打开编辑*/
      editClick(item) {
        this.$router.push({
          path: '/card/card/edit',
          query: {
            card_id: item.card_id,
          },
        });
      },
      /*打开编辑*/
      putClick(item) {
        this.userModel = item;
        this.open_edit = true;
      },
      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getTableList();
          }
        }
      },
      /*删除用户*/
      cancel(row) {
        let self = this;
        ElMessageBox.confirm($t('此操作将撤销已发会员卡，是否继续？'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            CardApi.cancelcard(
              {
                order_id: row.order_id,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: data.msg,
                    type: 'success',
                  });
                  self.getTableList();
                } else {
                  ElMessage.error(data.msg);
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },
    },
  };
</script>
