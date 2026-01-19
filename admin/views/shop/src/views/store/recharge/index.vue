<template>
  <div class="recharge">
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

        <el-form-item :label="$t('订单号')">
          <el-input size="small" v-model="searchForm.order_no" :placeholder="$t('订单号')" @input="onSearch"></el-input>
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
              <el-option :label="$t('添加时间')" :value="0"></el-option>
              <el-option :label="$t('支付时间')" :value="1"></el-option>
              <template #tag>
                <span v-if="searchForm.time_mode.includes(0)">{{ $t('添加时间') }}</span>
                <span v-if="searchForm.time_mode.includes(1) && !searchForm.time_mode.includes(0)">{{ $t('支付时间') }}</span>
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
          <el-button v-auth="'/store/recharge/export'" size="small" type="primary" @click="onExport">{{ $t('导出') }}</el-button>
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
                <el-tag size="" class="ml-4">{{ order_count.unpaid_num + order_count.cancel_num + order_count.complete_num }}</el-tag>
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
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column :label="$t('订单类型')">
            <template #default="scope">
              <div>{{ $t('充值订单') }}</div>
            </template>
          </el-table-column>

          <el-table-column prop="order_no" :label="$t('订单号')"></el-table-column>
          <el-table-column prop="status" :label="$t('状态')">
            <template #default="scope">
              <div>
                {{ scope.row.status == 0 ? $t('待付款') : scope.row.status == 2 ? $t('已取消') : $t('已完成') }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="payment_time" :label="$t('支付时间')"></el-table-column>
          <el-table-column prop="amount" :label="$t('订单金额')" width="140" show-overflow-tooltip>
            <template #default="scope">
              <div style="line-height: 24px">
                <main-currency>
                  {{ proxy.$formatPrice(scope.row.recharge_amount) }}
                </main-currency>
                <p class="gray98">
                  <sub-currency>
                    {{ proxy.$formatPrice((Number(scope.row.recharge_amount) * Number(currency.vices?.unit_rate)).toFixed(2)) }}
                  </sub-currency>
                </p>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="amount" :label="$t('实付金额')" show-overflow-tooltip>
            <template #default="scope">
              <div>
                <div class="orange" v-if="scope.row.status == 1">
                  <main-currency>
                    {{ proxy.$formatPrice(scope.row.amount || 0) }}
                  </main-currency>
                </div>
                <div v-else>-</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="" :label="$t('会员')" show-overflow-tooltip>
            <template #default="scope">
              <template v-if="scope.row.member_uuid">
                <span class="gray9">{{ $t('会员：') }}&nbsp;{{ scope.row.member_uuid }}</span>
              </template>
              <p v-else>-</p>
            </template>
          </el-table-column>

          <el-table-column prop="pay_type.text" :label="$t('支付方式')" show-overflow-tooltip>
            <template #default="scope">
              <span class="gray9">{{ scope.row.status == 1 ? patType(scope.row.payment_methods) : '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column fixed="right" :label="$t('操作')" width="160">
            <template #default="scope">
              <div>
                <el-button @click="detailClick(scope.row)" type="primary" link size="small" v-auth="'/store/recharge/detail'">{{ $t('详情') }} </el-button>

                <el-button v-if="scope.row.extra.is_cell_refund" @click="refundClick(scope.row)" type="danger" link size="small" v-auth="'/store/recharge/refund'"
                  >{{ $t('退款') }}
                </el-button>

                <el-button v-if="scope.row.extra.is_cell_cancel" @click="cancelClick(scope.row)" type="danger" link size="small" v-auth="'/store/recharge/cancel'"
                  >{{ $t('取消') }}
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
    <Cancel v-if="open_edit" :open_edit="open_edit" :order_no="order_no" :id="id" @closeDialog="closeDialogFunc($event, 'edit')"> </Cancel>
    <!--处理-->
    <refund v-if="open_refund" :open_edit="open_refund" :id="id" @closeDialog="closerefundDialogFunc($event, 'edit')"> </refund>
  </div>
</template>
<script setup>
  import { nextTick, ref, watch, getCurrentInstance, onMounted } from 'vue';
  import qs from 'qs';
  import { languageStore } from '@/store/model/language';
  import { useUserStore } from '@/store';
  import Cancel from './dialog/cancel.vue';
  import Refund from './dialog/refund.vue';
  import OrderApi from '@/api/order.js';
  import { useRouter } from 'vue-router';
  import { message } from '@/utils/message.js';

  const { token, currency, computedSupplier } = useUserStore();
  const { proxy } = getCurrentInstance();
  const supplier = ref(computedSupplier().supplier);
  const app_id = ref(supplier.value?.app_id || 0);

  const router = useRouter();
  const searchForm = ref({
    order_no: '',
    time: '',
    time_type: 1,
    time_mode: [0],
  });
  const searchLoading = ref('');
  const tableData = ref([]);
  const curPage = ref(1);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const loading = ref(false);
  const activeName = ref('all');
  const order_count = ref({
    total: 0,
    unpaid_num: 0,
    cancel_num: 0,
    complete_num: 0,
  });

  const open_edit = ref(false);
  const order_no = ref('');
  const id = ref('');

  const open_refund = ref(false);

  watch(
    () => searchForm.value.time_mode,
    (newVal, oldVal) => {
      nextTick(() => {
        if (newVal.length == 0) {
          searchForm.value.time_mode = oldVal;
        }
      });
    },
    { deep: true }
  );

  const onExport = () => {
    searchForm.value.token = token;
    const baseUrl = window.location.protocol + '//' + window.location.host;
    const url = baseUrl + '/index.php/shop/store.UserRechargeOrder/export?' + qs.stringify(searchForm.value) + '&language=' + languageStore().language;
    window.open(url, '_blank');
  };

  const getData = async () => {
    loading.value = true;
    try {
      let Params = searchForm.value;
      Params.data_type = activeName.value;
      Params.page = curPage.value;
      Params.list_rows = pageSize.value;
      const res = await OrderApi.getRechargeOrder(Params, true);
      tableData.value = res.data.list;
      totalDataNumber.value = res.data.meta.total;
      order_count.value = res.data.meta;
      loading.value = false;
    } catch (error) {
      loading.value = false;
    }
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      tableData.value = [];
      getData();
    }, 200);
  };

  /*切换菜单*/
  const handleClick = () => {
    curPage.value = 1;
    getData();
  };

  /*选择第几页*/
  const handleCurrentChange = (val) => {
    curPage.value = val;
    getData();
  };

  /*每页多少条*/
  const handleSizeChange = (val) => {
    curPage.value = 1;
    pageSize.value = val;
    getData();
  };

  /*切换周*/
  const timeTypeChange = () => {
    searchForm.value.time = '';
    onSearch();
  };

  /*切换时间段*/
  const createTimeChange = () => {
    searchForm.value.time_type = '';
    onSearch();
  };

  /*关闭弹窗*/
  const closeDialogFunc = (e, f) => {
    if (f == 'edit') {
      open_edit.value = e.openDialog;
      if (e.type == 'success') {
        getData();
      }
    }
  };

  /*关闭弹窗*/
  const closerefundDialogFunc = (e, f) => {
    if (f == 'edit') {
      open_refund.value = e.openDialog;
      if (e.type == 'success') {
        getData();
      }
    }
  };

  const patType = (arr) => {
    let result = '-';
    if (arr && arr.length > 0) {
      let nameArr = [];
      (arr || []).map((item) => {
        nameArr.push(item);
      });
      result = nameArr.join(',');
    }
    return result;
  };

  // 详情页
  const detailClick = (row) => {
    let params = row.uuid;

    let pageParams = searchForm.value;
    pageParams.data_type = activeName.value;
    pageParams.page = curPage.value;
    pageParams.list_rows = pageSize.value;
    languageStore().setPageParams(pageParams);
    router.push({
      path: '/' + app_id.value + '/store/recharge/detail',
      query: {
        id: params,
      },
    });
  };

  // 退款
  const refundClick = (row) => {
    id.value = row.uuid;
    open_refund.value = true;
  };

  /*打开取消*/
  const cancelClick = (item) => {
    id.value = item.uuid;
    ElMessageBox.confirm($t('是否取消此订单?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(async () => {
        await cancelSubmit();
        await getData();
      })
      .catch(() => {
        this.$ElMessage({
          type: 'info',
          message: $t('已取消'),
        });
      });
  };

  /*取消订单提交*/
  const cancelSubmit = async () => {
    if (loading.value) {
      return;
    }
    loading.value = true;
    try {
      const params = {
        id: id.value,
      };
      const res = await OrderApi.getRechargeOrderCancel(params, true);
      const type = res.code == 1 ? 'success' : 'error';
      loading.value = false;
      message({
        message: res.msg,
        type: type,
      });
    } catch (error) {
      loading.value = false;
    }
  };

  onMounted(() => {
    let params = languageStore().getPageParams().pageParams;

    if (params.value.page) {
      searchForm.value = {
        order_no: params.value.order_no,
        time: params.value.time,
        time_type: params.value.time_type,
        time_mode: params.value.time_mode,
      };
      activeName.value = params.value.data_type;
      curPage.value = params.value.page;
      pageSize.value = params.value.list_rows;
      languageStore().setPageParams({});
    }
    getData();
  });
</script>
<style lang="scss" scoped>
  .radio-search {
    :deep(.el-radio-button) {
      margin-right: -3px;
      margin-bottom: 0;

      .el-radio-button__inner {
        padding: 8px 11px !important;
      }
    }
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
</style>
