<template>
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
                <main-currency>
                  {{ $formatPrice(scope.row.order_amount) }}
                </main-currency>
                <p class="gray98" v-if="currency.is_open == 1">
                  <sub-currency>
                    {{ $formatPrice((Number(scope.row.order_amount) * Number(currency.vices?.unit_rate)).toFixed(2)) }}
                  </sub-currency>
                </p>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="payment_amount" :label="$t('实付金额')" show-overflow-tooltip>
            <template #default="scope">
              <div>
                <div class="orange" v-if="scope.row.status == 1 || (scope.row.sale_orders && scope.row.sale_orders.map((item) => item.status == 1).includes(true))">
                  <main-currency>
                    {{ $formatPrice(scope.row.payment_amount) }}
                  </main-currency>
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

<script setup>
// 中文注释：改为 Vue3 <script setup>，统一箭头函数与分号
import { ref, reactive, watch, computed, nextTick, getCurrentInstance, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import OrderOldApi from '@/api/orderOld.js';
import Cancel from './dialog/cancel.vue';
import refund from './dialog/refund.vue';
import qs from 'qs';
import { useUserStore } from '@/store';
import { languageStore } from '@/store/model/language';

// 组件在 <script setup> 中按需引入即可在模板使用

// 全局实例（用于 $t、$ElMessage、$formatPrice）
const { proxy } = getCurrentInstance();
const router = useRouter();

// 门店/用户信息
const userStore = useUserStore();
const { token, currency, computedSupplier } = userStore; // 避免丢失原有用法
const supplier = computed(() => supplier.value?.app_id || 0);

// 列表与查询状态
const activeName = ref('all');
const loading = ref(true);
const tableData = ref([]);
const pageSize = ref(10);
const totalDataNumber = ref(0);
const curPage = ref(1);

const searchForm = reactive({
  order_no: '',
  style_id: ' ',
  time: '',
  time_type: 1,
  order_source: ' ',
  time_mode: [0],
});

const exStyle = ref([]);
const order_count = reactive({
  cancel_num: 0,
  complete_num: 0,
  page_no: 1,
  page_size: 10,
  total: 0,
  total_num: 0,
  unpaid_num: 0,
});

const open_edit = ref(false);
const open_refund = ref(false);
const order_no = ref(0);
const order_id = ref(0);
const sub_order_id = ref(0);
const pay_price = ref(0);
const searchLoading = ref();

// 页面恢复：从语言存储读取上次的分页及查询参数
onMounted(() => {
  const params = languageStore().getPageParams().pageParams;
  if (params.value?.page) {
    Object.assign(searchForm, {
      order_no: params.value.order_no,
      style_id: params.value.style_id,
      time: params.value.time,
      time_type: params.value.time_type,
      order_source: params.value.order_source,
      time_mode: params.value.time_mode,
    });
    activeName.value = params.value.dataType;
    curPage.value = params.value.page;
    pageSize.value = params.value.list_rows;
    languageStore().setPageParams({});
  }
  getData();
});

// 防止时间模式被清空
watch(
  () => searchForm.time_mode,
  (newVal, oldVal) => {
    nextTick(() => {
      if ((newVal || []).length === 0) {
        searchForm.time_mode = oldVal || [0];
      }
    });
  },
  { deep: true }
);

// 跨多列（保留原有逻辑）
const arraySpanMethod = (row) => {
  if (row.rowIndex % 2 == 0) {
    if (row.columnIndex === 0) {
      return [1, 8];
    }
  }
};

// 分页事件
const handleCurrentChange = (val) => {
  curPage.value = val;
  getData();
};

const handleSizeChange = (val) => {
  curPage.value = 1;
  pageSize.value = val;
  getData();
};

// Tab 切换
const handleClick = () => {
  curPage.value = 1;
  getData();
};

// 周期切换
const timeTypeChange = () => {
  searchForm.time = '';
  onSearch();
};

// 时间范围变更
const createTimeChange = () => {
  searchForm.time_type = '';
  onSearch();
};

// 获取列表
const getData = async () => {
  const Params = { ...searchForm, dataType: activeName.value, page: curPage.value, list_rows: pageSize.value };
  loading.value = true;
  try {
    const res = await OrderOldApi.storeOrderlist(Params, true);
    const list = res.data.list || [];
    list.forEach((item) => {
      if (item.sale_orders && item.sale_orders.length > 0) {
        item.children = item.sale_orders;
      }
      item.unique_key = item.order_no + item.serial_no;
    });
    tableData.value = list;
    totalDataNumber.value = res.data.meta.total;
    exStyle.value = res.data.ex_style || [];
    Object.assign(order_count, res.data.meta || {});
  } catch (e) {
    // 忽略错误
  } finally {
    loading.value = false;
  }
};

// 查看详情
const addClick = (row) => {
  const pageParams = {
    ...searchForm,
    dataType: activeName.value,
    page: curPage.value,
    list_rows: pageSize.value,
  };
  languageStore().setPageParams(pageParams);
  const saleOrderUuid = row.is_split === undefined ? row.sale_order_uuid : 0;
  router.push({
    path: `/${app_id.value}/store/history_order/detail`,
    query: {
      sale_bill_uuid: row.sale_bill_uuid,
      sale_order_uuid: saleOrderUuid,
    },
  });
};

// 核销
const verifyClick = async (row) => {
  try {
    await ElMessageBox.confirm(proxy.$t('确定要核销吗?'), proxy.$t('提示'), {
      confirmButtonText: proxy.$t('确定'),
      cancelButtonText: proxy.$t('取消'),
      type: 'warning',
    });
    const extract_form = { order_id: row.order_id };
    await OrderOldApi.storeExtract(extract_form, true);
    proxy.$ElMessage({ message: proxy.$t('操作成功'), type: 'success' });
    getData();
  } catch (e) {
    // 取消或错误
    if (e !== 'cancel') {
      // do nothing
    } else {
      proxy.$ElMessage({ type: 'info', message: proxy.$t('已取消核销') });
    }
  }
};

// 查询物流
const getLogistics = async (row) => {
  const Params = { order_id: row.order_id };
  loading.value = true;
  try {
    const res = await OrderOldApi.queryLogistics(Params, true);
    // 假定有 logisticsData 与 isLogistics 控制
    // 若无视图使用，可按需删除
    // logisticsData.value = res.data.express.list;
    loading.value = false;
    openLogistics();
  } catch (e) {
    loading.value = false;
  }
};

const openLogistics = () => {
  // 若需要可定义 isLogistics 控制弹窗
  // isLogistics.value = true;
};
const closeLogistics = () => {
  // isLogistics.value = false;
};

// 搜索
const onSearch = () => {
  clearTimeout(searchLoading.value);
  searchLoading.value = setTimeout(() => {
    curPage.value = 1;
    tableData.value = [];
    getData();
  }, 200);
};

// 导出
const onExport = async () => {
  const query = { ...searchForm, token: token };
  try {
    await OrderOldApi.storeExport({ ...query, request_type: 1 }, true);
  } catch (e) {
    // 忽略错误
  }
  const baseUrl = window.location.protocol + '//' + window.location.host;
  const url = baseUrl + '/index.php/shop/store.operateOld/export?' + qs.stringify(query) + '&language=' + languageStore().language;
  window.open(url, '_blank');
};

// 取消
const cancelClick = (item) => {
  order_no.value = item.order_no;
  order_id.value = item.sale_bill_uuid;
  open_edit.value = true;
};

// 退款
const refundClick = (item) => {
  order_no.value = item.order_no;
  order_id.value = item.sale_bill_uuid;
  sub_order_id.value = item.sale_order_uuid;
  pay_price.value = item.payment_amount;
  open_refund.value = true;
};

// 删除
const delClick = async (item) => {
  try {
    await ElMessageBox.confirm(proxy.$t('删除后不可恢复，确认删除吗?'), proxy.$t('提示'), { type: 'warning' });
    const saleOrderUuid = item.is_split === undefined ? item.sale_order_uuid : 0;
    await OrderOldApi.storedelete({
      sale_bill_uuid: item.sale_bill_uuid,
      sale_order_uuid: saleOrderUuid,
    });
    proxy.$ElMessage({ message: proxy.$t('删除成功'), type: 'success' });
    getData();
  } catch (e) {
    // 取消或错误
  }
};

// 子弹窗回调
const closeDialogFunc = (e, f) => {
  if (f == 'edit') {
    open_edit.value = e.openDialog;
    if (e.type == 'success') {
      getData();
    }
  }
};

const closerefundDialogFunc = (e, f) => {
  if (f == 'edit') {
    open_refund.value = e.openDialog;
    if (e.type == 'success') {
      getData();
    }
  }
};

// 支付方式名
const patType = (arr) => {
  let result = '-';
  if (arr && arr.length > 0) {
    const nameArr = [];
    (arr || []).map((item) => {
      nameArr.push(item.name);
      return item;
    });
    result = nameArr.join(',');
  }
  return result;
};

// 桌号
const tableNo = (arr) => {
  let result = '-';
  if (arr && arr.length > 0) {
    const nameArr = [];
    (arr || []).map((item) => {
      nameArr.push(item.table_no);
      return item;
    });
    result = nameArr.join('+');
  }
  return result;
};

// 订单去重后的会员
const uniqueUsers = (e) => {
  const userIds = new Set();
  const result = (e || [])
    .filter((item) => {
      if (item.user && !userIds.has(item.user.user_id)) {
        userIds.add(item.user.user_id);
        return true;
      }
      return false;
    })
    .map((item) => item.user);
  if (result.length > 0) {
    const idArr = [];
    (result || []).map((item) => {
      idArr.push(item.user_id);
      return item;
    });
    idArr.join(',');
    return proxy.$t('会员ID') + ' (' + idArr.toString() + ')';
  } else {
    return '-';
  }
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
