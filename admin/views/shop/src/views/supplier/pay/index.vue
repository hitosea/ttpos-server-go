<template>
  <div class="supplier">
    <!--内容-->
    <div class="supplier-content">
      <div class="common-level-rail">
        <el-button size="small" type="primary" @click="addClick" v-auth="'/supplier/pay/add'" :disabled="erp_is_open == 1 && erp_site_code == '2'">
          {{ $t('添加') }}
        </el-button>
      </div>
      <div class="table-wrap" v-if="payment.is_other == '1'">
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column type="index" width="45" :label="$t('序号')" header-align="center" align="center"></el-table-column>
          <el-table-column prop="remark" :label="$t('名称')">
            <template #default="scope">
              <div class="flex-box">
                <img v-if="scope.row.img" :src="scope.row.img" style="width: 48px" />
                <img v-else src="@/assets/img/ja_pay.png" style="width: 48px" />
                {{ scope.row.remark }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="name" :label="$t('支付方式')"></el-table-column>
          <el-table-column prop="source_text" :label="$t('渠道提供方')"> </el-table-column>
          <el-table-column prop="sort" :label="$t('排序')" width="80"></el-table-column>
          <el-table-column prop="" :label="$t('状态')" width="100">
            <template #default="scope">
              <el-switch
                :disabled="!this.$filter.isAuth('/supplier/pay/state')"
                :model-value="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @click="changeStatus(scope.row)"
              ></el-switch>
            </template>
          </el-table-column>
          <el-table-column prop="is_show_checkout" :label="$t('显示终端')">
            <template #default="scope">
              <span v-if="scope.row.is_show_checkout.indexOf('10') != -1">{{ $t('收银机') }}</span>
              <template v-if="scope.row.is_show_checkout.indexOf('10') != -1 && scope.row.is_show_checkout.indexOf('20') != -1">、</template>
              <span v-if="scope.row.is_show_checkout.indexOf('20') != -1">{{ $t('点餐助手') }}</span>
              <template v-if="scope.row.is_show_checkout.indexOf('10') == -1 && scope.row.is_show_checkout.indexOf('20') == -1">-</template>
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('添加时间')" width="180"></el-table-column>
          <el-table-column prop="" :label="$t('操作')" width="120" fixed="right">
            <template #default="scope">
              <el-button @click="editClick(scope.row)" type="primary" link size="small" v-auth="'/supplier/pay/edit'">
                {{ $t('编辑') }}
              </el-button>
              <el-button
                v-if="scope.row.source == 1"
                @click="deleteClick(scope.row)"
                type="primary"
                link
                size="small"
                v-auth="'/supplier/pay/delete'"
                :disabled="scope.row.status == 2 || erp_is_open == 1"
              >
                {{ $t('删除') }}
              </el-button>
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
          layout=" prev, pager, next, jumper"
          :total="totalDataNumber"
        >
        </el-pagination>
      </div>
    </div>
    <addEdit v-if="open" :title="title" :open="open" :editItem="editItem" @close="handleClose" :list="tableData" :erp_is_open="erp_is_open" :erp_site_code="erp_site_code">
    </addEdit>
  </div>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import addEdit from './addEdit.vue';
  import { useUserStore } from '@/store';
  const { computedSupplier, erp_is_open, erp_site_code } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_member = supplier.value?.is_open_member || 0;

  export default {
    components: { addEdit },
    data() {
      return {
        /*是否加载完成*/
        loading: false,
        /*列表数据*/
        tableData: [],
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,

        /*是否打开添加弹窗*/
        open: false,
        /*是否打开编辑弹窗*/
        title: '',
        payment: {
          is_balance: '1',
          is_cash: '1',
          is_other: '1',
        },
        editItem: '',
        is_open_member: is_open_member,
        erp_is_open: erp_is_open,
        erp_site_code: erp_site_code,
      };
    },
    created() {
      this.getData();
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
        this.curPage = 1;
        this.pageSize = val;
        this.getData();
      },
      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.getPaytype({}, true)
          .then((data) => {
            self.loading = false;
            self.tableData = data.data.list;
            self.payment = data.data.payment;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      changeStatus(row) {
        let next = 0;
        this.tableData.map((item) => {
          if (item.status == '1') {
            next++;
          }
        });
        if (Number(this.payment.is_cash) + Number(this.payment.is_balance) + Number(this.payment.is_other) == 1 && row.status == '1' && next == 1) {
          this.$ElMessage({
            message: $t('至少需要开启一种支付方式'),
            type: 'error',
          });
          return;
        }
        let self = this;
        let params = {
          id: row.id,
          name: row.name,
          remark: row.remark,
          sort: row.sort,
          img: row.img,
          qrcode: row.qrcode,
          status: row.status == 1 ? 0 : 1,
        };
        let text = '';
        text = row.status == 1 ? $t('禁用') : $t('启用');
        ElMessageBox.confirm($t('确定') + text + $t('这个支付方式?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            SettingApi.setPaytype(params, true)
              .then((data) => {
                self.loading = false;
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },

      changePay(key, status) {
        if (!this.$filter.isAuth('/supplier/pay/state')) {
          return;
        }
        let self = this;
        if (Number(self.payment.is_cash) + Number(self.payment.is_balance) + Number(self.payment.is_other) == 1 && status == '1') {
          this.$ElMessage({
            message: $t('至少需要开启一种支付方式'),
            type: 'error',
          });
          return;
        }
        let params = {
          is_cash: key == 'is_cash' ? (status == 1 ? 0 : 1) : self.payment.is_cash,
          is_balance: key == 'is_balance' ? (status == 1 ? 0 : 1) : self.payment.is_balance,
          is_other: key == 'is_other' ? (status == 1 ? 0 : 1) : self.payment.is_other,
        };
        let text = '';
        text = status == 1 ? $t('禁用') : $t('启用');
        ElMessageBox.confirm($t('确定') + text + $t('这个支付方式?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            SettingApi.editPayment(params, true)
              .then((data) => {
                self.loading = false;
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {});
      },

      /*打开添加*/
      addClick() {
        this.title = $t('添加支付方式');
        this.open = true;
      },

      /*打开编辑*/
      editClick(item) {
        this.editItem = item;
        this.title = $t('编辑支付方式');
        this.open = true;
      },

      handleClose(e) {
        this.open = false;
        if (e == 1) {
          this.getData();
        }
        this.editItem = '';
      },

      /*删除用户*/
      deleteClick(row) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            self.loading = true;
            SettingApi.deletePayment(
              {
                id: row.id,
              },
              true
            )
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: data.msg,
                  type: 'success',
                });
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {
            self.loading = false;
          });
      },
    },
  };
</script>
<style lang="scss" scoped>
  .switch-box {
    display: flex;
    align-items: center;
    margin: 16px 0;
  }

  .switch-main {
    flex: 1;
    display: flex;
    align-items: center;
  }

  .el-button--primary.is-link {
    color: var(--el-color-primary);
  }

  .el-button--primary.is-link:focus {
    color: var(--el-color-primary);
  }
  .flex-box {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>
