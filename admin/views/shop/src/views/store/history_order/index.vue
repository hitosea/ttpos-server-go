<template>
  <!--

      时间：2019-10-25
      描述：订单列表
  -->
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item label="">
          <el-radio-group v-model="searchForm.time_type" class="radio-search" @change="timeTypeChange">
            <el-radio-button :value="1">{{ $t('今天') }}</el-radio-button>
            <el-radio-button :value="2">{{ $t('昨天') }}</el-radio-button>
            <el-radio-button :value="3">{{ $t('本周') }}</el-radio-button>
            <el-radio-button :value="0">{{ $t('全部') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('订单类型')">
          <a-select size="small" v-model:value="searchForm.order_source" :placeholder="$t('请选择')" clearable @change="onSearch">
            <el-option :label="$t('全部')" value=" "></el-option>
            <el-option :label="$t('点餐订单')" :value="20"></el-option>
            <el-option :label="$t('桌台订单')" :value="10"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('订单号')">
          <el-input size="small" v-model="searchForm.order_no" :placeholder="$t('订单号')" @input="onSearch"></el-input>
        </el-form-item>
        <el-form-item :label="$t('用餐方式')">
          <a-select size="small" v-model:value="searchForm.style_id" :placeholder="$t('请选择')" @change="onSearch">
            <el-option :label="$t('全部')" value=" "></el-option>
            <el-option v-for="(item, index) in exStyle" :key="index" :label="item.name" :value="item.value"></el-option>
          </a-select>
        </el-form-item>
        <el-form-item :label="$t('起始时间')">
          <div class="flex-box">
            <el-select
              class="time-select"
              size="small"
              multiple
              v-model="searchForm.time_mode"
              :placeholder="$t('请选择')"
              @change="onSearch"
              @clear="
                () => {
                  searchForm.time_mode = [0];
                }
              "
            >
              <el-option :label="$t('开台时间')" :value="0"></el-option>
              <el-option :label="$t('支付时间')" :value="1"></el-option>
              <template #tag>
                <span v-if="searchForm.time_mode.includes(0)">{{ $t('开台时间') }}</span>
                <span v-if="searchForm.time_mode.includes(1) && !searchForm.time_mode.includes(0)">{{ $t('支付时间') }}</span>
                <span v-if="!searchForm.time_mode.includes(1) && !searchForm.time_mode.includes(0)">{{ $t('请选择') }}</span>
              </template>
            </el-select>
            <el-date-picker
              size="small"
              v-model="searchForm.time"
              @change="createTimeChange"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              :disabledDate="(time) => time.getTime() > Date.now()"
            ></el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button v-auth="'/store/operate/export'" size="small" type="primary" @click="onExport">{{ $t('导出') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-tabs size="small" v-model="activeName" @tab-change="handleClick">
          <el-tab-pane :label="$t('全部订单')" name="all">
            <template #label>
              <span>
                {{ $t('全部订单') }}
                <el-tag size="" class="ml-4">{{ order_count.total_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('待付款')" name="payment">
            <template #label>
              <span>
                {{ $t('待付款') }}
                <el-tag size="" class="ml-4">{{ order_count.unpaid_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>

          <el-tab-pane :label="$t('已取消')" name="cancel">
            <template #label>
              <span>
                {{ $t('已取消') }}
                <el-tag size="" class="ml-4">{{ order_count.cancel_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('已完成')" name="complete">
            <template #label>
              <span>
                {{ $t('已完成') }}
                <el-tag size="" class="ml-4"> {{ order_count.complete_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
        </el-tabs>
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading" row-key="unique_key">
          <el-table-column prop="serial_no" :label="$t('订单类型')">
            <template #default="scope">
              {{ scope.row.sale_orders ? (scope.row.bill_type == 1 ? $t('点餐订单') : $t('桌台订单')) : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="serial_no" :label="$t('桌号/序号')"></el-table-column>
          <el-table-column prop="order_no" :label="$t('订单号')"></el-table-column>
          <el-table-column prop="status" :label="$t('状态')">
            <template #default="scope">
              {{ scope.row.status == 0 ? $t('待付款') : scope.row.status == 2 ? $t('已取消') : $t('已完成') }}
            </template>
          </el-table-column>
          <el-table-column prop="finish_time" :label="$t('支付时间')">
            <template #default="scope">
              {{ scope.row.finish_time }}
            </template>
          </el-table-column>
          <el-table-column prop="order_amount" :label="$t('订单金额')" width="140" show-overflow-tooltip>
            <template #default="scope">
              <div style="line-height: 24px">
                <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                {{ this.$formatPrice(scope.row.order_amount) }}
                <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
                <p class="gray98" v-if="currency.is_open == 1">
                  <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
                  {{ this.$formatPrice((Number(scope.row.order_amount) * Number(currency.vices?.unit_rate)).toFixed(2))
                  }}<template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }} </template>
                </p>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="payment_amount" :label="$t('实付金额')" show-overflow-tooltip>
            <template #default="scope">
              <div>
                <div class="orange" v-if="scope.row.status == 1 || (scope.row.sale_orders && scope.row.sale_orders.map((item) => item.status == 1).includes(true))">
                  <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                  {{ this.$formatPrice(scope.row.payment_amount) }}
                  <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
                </div>
                <div v-else>-</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="" :label="$t('会员')" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.consumer_uuids" class="gray9">{{ $t('会员ID') }}&nbsp;({{ scope.row.consumer_uuids }})</span>
              <span v-else class="gray9">-</span>
            </template>
          </el-table-column>

          <el-table-column prop="pay_type_name" :label="$t('支付方式')" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.status == 1 || (scope.row.sale_orders && scope.row.sale_orders.map((item) => item.status == 1).includes(true))">{{
                scope.row.pay_type_name
              }}</span>
              <span v-else class="gray9">-</span>
            </template>
          </el-table-column>

          <el-table-column fixed="right" :label="$t('操作')" width="160">
            <template #default="scope">
              <div>
                <el-button @click="addClick(scope.row)" type="primary" link size="small" v-auth="'/store/order/detail'">{{ $t('详情') }} </el-button>
                <el-button v-if="scope.row.extra.is_cell_refund" @click="refundClick(scope.row)" type="danger" link size="small" v-auth="'/store/operate/refund'"
                  >{{ $t('退款') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_cancel" @click="cancelClick(scope.row)" type="danger" link size="small" v-auth="'/store/operate/order_cancel'"
                  >{{ $t('取消') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_delete" @click="delClick(scope.row)" type="danger" link size="small" v-auth="'/store/order/delete'"
                  >{{ $t('删除') }}
                </el-button>
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
          :page-size="pageSize"
          layout="total, prev, pager, next, jumper"
          :total="totalDataNumber"
        ></el-pagination>
      </div>
    </div>
    <!--处理-->
    <Cancel v-if="open_edit" :open_edit="open_edit" :order_no="order_no" :order_id="order_id" @closeDialog="closeDialogFunc($event, 'edit')"> </Cancel>
    <!--处理-->
    <refund
      v-if="open_refund"
      :open_edit="open_refund"
      :order_id="order_id"
      :sub_order_id="sub_order_id"
      :pay_price="pay_price"
      @closeDialog="closerefundDialogFunc($event, 'edit')"
    >
    </refund>
  </div>
</template>

<script>
  import OrderOldApi from '@/api/orderOld.js';
  import Cancel from './dialog/cancel.vue';
  import refund from './dialog/refund.vue';
  import qs from 'qs';
  import { useUserStore } from '@/store';
  import Aselect from '@/components/a-select/index.vue';
  const { token, currency, computedSupplier } = useUserStore();
  import { languageStore } from '@/store/model/language';
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  export default {
    components: {
      Cancel,
      refund,
      Aselect,
    },
    data() {
      return {
        currency: currency,
        /*切换菜单*/
        activeName: 'all',
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
        searchForm: {
          order_no: '',
          style_id: ' ',
          time: '',
          time_type: 1,
          order_source: ' ',
          time_mode: [0],
        },
        /*配送方式*/
        exStyle: [],
        /*门店列表*/
        shopList: [],
        /*时间*/
        time: '',
        /*统计*/
        order_count: {
          cancel_num: 0,
          complete_num: 0,
          page_no: 1,
          page_size: 10,
          total: 0,
          total_num: 0,
          unpaid_num: 0,
        },
        /*是否打开编辑弹窗*/
        open_edit: false,
        open_refund: false,
        /*当前编辑的对象*/
        order_no: 0,
        order_id: 0,
        sub_order_id: 0,
        pay_price: 0,
        token,
        app_id: app_id,
        searchLoading: '',
      };
    },
    created() {
      let params = languageStore().getPageParams().pageParams;

      if (params.value.page) {
        this.searchForm = {
          order_no: params.value.order_no,
          style_id: params.value.style_id,
          time: params.value.time,
          time_type: params.value.time_type,
          order_source: params.value.order_source,
          time_mode: params.value.time_mode,
        };
        this.activeName = params.value.dataType;
        this.curPage = params.value.page;
        this.pageSize = params.value.list_rows;
        languageStore().setPageParams({});
      }

      /*获取列表*/
      this.getData();
    },

    watch: {
      'searchForm.time_mode': {
        handler(newVal, oldVal) {
          this.$nextTick(() => {
            if (newVal.length == 0) {
              this.searchForm.time_mode = oldVal;
            }
          });
        },
        deep: true,
      },
    },
    methods: {
      /*跨多列*/
      arraySpanMethod(row) {
        if (row.rowIndex % 2 == 0) {
          if (row.columnIndex === 0) {
            return [1, 8];
          }
        }
      },
      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.curPage = val;
        self.getData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getData();
      },

      /*切换菜单*/
      handleClick(tab, event) {
        let self = this;
        self.curPage = 1;
        self.getData();
      },

      /*切换周*/
      timeTypeChange() {
        this.searchForm.time = '';
        this.onSearch();
      },

      /*切换时间段*/
      createTimeChange() {
        this.searchForm.time_type = '';
        this.onSearch();
      },

      /*获取列表*/
      getData() {
        let self = this;
        let Params = this.searchForm;
        Params.dataType = self.activeName;
        Params.page = self.curPage;
        Params.list_rows = self.pageSize;
        self.loading = true;
        OrderOldApi.storeOrderlist(Params, true)
          .then((res) => {
            self.tableData = res.data.list;
            self.tableData.map((item) => {
              if (item.sale_orders.length > 0) {
                item.children = item.sale_orders;
              }
              item.unique_key = item.order_no + item.serial_no;
            });
            self.totalDataNumber = res.data.meta.total;
            self.exStyle = res.data.ex_style;
            self.order_count = res.data.meta;
            self.loading = false;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*打开添加*/
      addClick(row) {
        let self = this;
        let pageParams = self.searchForm;
        pageParams.dataType = self.activeName;
        pageParams.page = self.curPage;
        pageParams.list_rows = self.pageSize;
        languageStore().setPageParams(pageParams);
        // 如果没有拆单, 或者查看主单详情, 则sale_order_uuid = 0;
        // 反之查看子单详情, 则sale_order_uuid = 为子单sale_order_uuid
        const saleOrderUuid = row.is_split === undefined ? row.sale_order_uuid : 0;
        self.$router.push({
          path: '/' + this.app_id + '/store/history_order/detail',
          query: {
            sale_bill_uuid: row.sale_bill_uuid,
            sale_order_uuid: saleOrderUuid,
          },
        });
      },
      /*核销*/
      verifyClick(row) {
        let self = this;
        let extract_form = {};
        ElMessageBox.confirm('确定要核销吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        })
          .then(() => {
            extract_form.order_id = row.order_id;
            OrderOldApi.storeExtract(extract_form, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('操作成功'),
                  type: 'success',
                });
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          })
          .catch(() => {
            this.$ElMessage({
              type: 'info',
              message: '已取消核销',
            });
          });
      },
      getLogistics(row) {
        let self = this;
        let Params = {
          order_id: row.order_id,
        };
        self.loading = true;
        OrderOldApi.queryLogistics(Params, true)
          .then((res) => {
            self.logisticsData = res.data.express.list;
            self.loading = false;
            self.openLogistics();
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      openLogistics() {
        this.isLogistics = true;
      },
      closeLogistics() {
        this.isLogistics = false;
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.tableData = [];
          this.getData();
        }, 200);
      },

      onExport: function () {
        this.searchForm.token = this.token;
        OrderOldApi.storeExport(
          {
            ...this.searchForm,
            request_type: 1,
          },
          true
        )
          .then((data) => {
            self.loading = false;
            const baseUrl = window.location.protocol + '//' + window.location.host;
            const url = baseUrl + '/index.php/shop/store.operateOld/export?' + qs.stringify(this.searchForm) + '&language=' + languageStore().language;
            window.open(url, '_blank');
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      /*打开取消*/
      cancelClick(item) {
        this.order_no = item.order_no;
        this.order_id = item.sale_bill_uuid;
        this.open_edit = true;
      },
      refundClick(item) {
        this.order_no = item.order_no;
        this.order_id = item.sale_bill_uuid;
        this.sub_order_id = item.sale_order_uuid;
        this.pay_price = item.payment_amount;

        this.open_refund = true;
      },

      delClick(item) {
        let self = this;
        ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
          type: 'warning',
        }).then(() => {
          const saleOrderUuid = item.is_split === undefined ? item.sale_order_uuid : 0;
          OrderOldApi.storedelete({
            sale_bill_uuid: item.sale_bill_uuid,
            sale_order_uuid: saleOrderUuid,
          }).then((data) => {
            this.$ElMessage({
              message: $t('删除成功'),
              type: 'success',
            });
            self.getData();
          });
        });
      },
      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'edit') {
          this.open_edit = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
      },
      /*关闭弹窗*/
      closerefundDialogFunc(e, f) {
        if (f == 'edit') {
          this.open_refund = e.openDialog;
          if (e.type == 'success') {
            this.getData();
          }
        }
      },
      patType(arr) {
        let result = '-';
        if (arr && arr.length > 0) {
          let nameArr = [];
          (arr || []).map((item) => {
            nameArr.push(item.name);
          });
          result = nameArr.join(',');
        }
        return result;
      },

      tableNo(arr) {
        let result = '-';
        if (arr && arr.length > 0) {
          let nameArr = [];
          (arr || []).map((item) => {
            nameArr.push(item.table_no);
          });
          result = nameArr.join('+');
        }
        return result;
      },

      uniqueUsers(e) {
        let userIds = new Set(); // 使用 Set 来存储唯一的 user_id
        let result = e
          .filter((item) => {
            if (item.user && !userIds.has(item.user.user_id)) {
              userIds.add(item.user.user_id);
              return true; // 保留这个用户
            }
            return false; // 过滤掉重复的用户
          })
          .map((item) => item.user); // 返回去重后的用户对象
        if (result.length > 0) {
          let idArr = [];
          (result || []).map((item) => {
            idArr.push(item.user_id);
          });
          idArr.join(',');
          return this.$t('会员ID') + ' (' + idArr.toString() + ')';
        } else {
          return '-';
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .product-info {
    padding: 10px 0;
    border-top: 1px solid #eeeeee;
  }

  .order-code {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .order-code .state-text {
    padding: 0 4px;
    border-radius: 4px;
    background: #808080;
    color: #ffffff;
    height: 24px;
    line-height: 24px;
  }

  .order-code .state-text-red {
    background: red;
  }

  .table-wrap .product-info:first-of-type {
    border-top: none;
  }

  .table-wrap .el-table__body tbody .el-table__row:nth-child(odd) {
    background: #f5f7fa;
  }

  .radio-search {
    :deep(.el-radio-button) {
      margin-right: -3px;
      margin-bottom: 0;

      .el-radio-button__inner {
        padding: 8px 11px !important;
      }
    }
  }

  .el-button--danger.is-link {
    color: var(--el-color-primary);
  }

  .el-button--danger.is-link:focus {
    color: var(--el-color-primary);
  }

  .el-tag {
    padding: 0 6px;
    height: 20px;
  }

  .gray98 {
    font-size: 12px;
    color: #999;
  }

  .gray99 {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 12px;
    color: #ff0000;
  }
  .ml-4 {
    margin-left: 4px;
  }
  .flex-box {
    display: flex;
    align-items: center;
    .time-select {
      width: 180px;
      :deep(.el-select__wrapper) {
        border-radius: 4px 0 0 4px;
        .el-select__selection {
          flex-wrap: nowrap;
          overflow: hidden;
        }
      }
    }
    :deep(.el-input__wrapper) {
      border-radius: 0 4px 4px 0;
      margin-left: -1px;
    }
  }
  :deep(.el-table__expand-column) {
    .cell {
      line-height: 23px !important;
    }
  }
</style>
