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
              {{ scope.row.finish_time || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="order_amount" :label="$t('订单金额')" width="140" show-overflow-tooltip>
            <template #default="scope">
              <div style="line-height: 24px">
                <main-currency>
                  {{ $formatPrice(scope.row.order_amount) }}
                </main-currency>
                <p class="gray98">
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
              <span v-if="scope.row.consumer_uuids" class="gray9">{{ $t('会员：') }}&nbsp;{{ scope.row.consumer_uuids }}</span>
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
                <el-button @click="() => addClick(scope.row)" type="primary" link size="small" v-auth="'/store/order/detail'">{{ $t('详情') }} </el-button>
                <el-button v-if="scope.row.extra.is_cell_refund" @click="() => refundClick(scope.row)" type="danger" link size="small" v-auth="'/store/operate/refund'"
                  >{{ $t('退款') }}
                </el-button>

                <el-button v-if="scope.row.extra.is_cell_cancel" @click="() => cancelClick(scope.row)" type="danger" link size="small" v-auth="'/store/operate/order_cancel'"
                  >{{ $t('取消') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_delete" @click="() => delClick(scope.row)" type="danger" link size="small" v-auth="'/store/order/delete'"
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
        />
      </div>
    </div>
    <!--处理-->
    <Cancel v-if="openEdit" :open_edit="openEdit" :order_no="orderNo" :order_id="orderId" @closeDialog="e => closeDialogFunc(e, 'edit')" />
    <!--处理-->
    <refund
      v-if="openRefund"
      :open_edit="openRefund"
      :order_id="orderId"
      :sub_order_id="subOrderId"
      :pay_price="payPrice"
      @closeDialog="e => closeRefundDialogFunc(e, 'edit')"
    />
  </div>
</template>

<script setup>
// 引入Vue3组合式API
import { ref, reactive, onMounted, watch, nextTick } from 'vue';
// 引入Element Plus弹窗和消息
import { ElMessageBox, ElMessage } from 'element-plus';
// 引入路由
import { useRouter } from 'vue-router';
// 引入API
import OrderApi from '@/api/order.js';
// 引入组件
import Cancel from './dialog/cancel.vue';
import refund from './dialog/refund.vue';
import Aselect from '@/components/a-select/index.vue';
// 引入工具
import qs from 'qs';
// 引入store
import { useUserStore } from '@/store';
import { languageStore } from '@/store/model/language';

// 获取路由实例
const router = useRouter();

// 获取store数据
const { token, currency, computedSupplier } = useUserStore();
const supplier = computedSupplier().supplier;
const appId = supplier.value?.app_id || 0;

// 切换菜单
const activeName = ref('all');
// 是否加载完成
const loading = ref(true);
// 列表数据
const tableData = ref([]);
// 每页多少条
const pageSize = ref(10);
// 总数据条数
const totalDataNumber = ref(0);
// 当前页码
const curPage = ref(1);
// 横向表单数据模型
const searchForm = reactive({
  order_no: '',
  style_id: ' ',
  time: '',
  time_type: 1,
  order_source: ' ',
  time_mode: [0],
});
// 配送方式
const exStyle = ref([]);
// 门店列表
const shopList = ref([]);
// 时间
const time = ref('');
// 统计
const order_count = reactive({
  cancel_num: 0,
  complete_num: 0,
  page_no: 1,
  page_size: 10,
  total: 0,
  total_num: 0,
  unpaid_num: 0,
});
// 是否打开编辑弹窗
const openEdit = ref(false);
const openRefund = ref(false);
// 当前编辑的对象
const orderNo = ref(0);
const orderId = ref(0);
const subOrderId = ref(0);
const payPrice = ref(0);
const searchLoading = ref('');

// 获取数据
async function getData() {
  const params = { ...searchForm };
  params.dataType = activeName.value;
  params.page = curPage.value;
  params.list_rows = pageSize.value;
  
  loading.value = true;
  try {
    const res = await OrderApi.storeOrderlist(params, true);
    tableData.value = res.data.list;
    tableData.value.map((item) => {
      if (item.sale_orders.length > 0) {
        item.children = item.sale_orders;
      }
      item.unique_key = item.order_no + item.serial_no;
    });
    totalDataNumber.value = res.data.meta.total;
    exStyle.value = res.data.ex_style;
    Object.assign(order_count, res.data.meta);
  } catch (error) {
    // 错误处理
  } finally {
    loading.value = false;
  }
}

// 分页-切换页码
function handleCurrentChange(val) {
  curPage.value = val;
  getData();
}

// 分页-切换每页条数
function handleSizeChange(val) {
  curPage.value = 1;
  pageSize.value = val;
  getData();
}

// 切换菜单
function handleClick(tab, event) {
  curPage.value = 1;
  getData();
}

// 切换周
function timeTypeChange() {
  searchForm.time = '';
  onSearch();
}

// 切换时间段
function createTimeChange() {
  searchForm.time_type = '';
  onSearch();
}

// 打开详情
function addClick(row) {
  const pageParams = { ...searchForm };
  pageParams.dataType = activeName.value;
  pageParams.page = curPage.value;
  pageParams.list_rows = pageSize.value;
  languageStore().setPageParams(pageParams);
  
  // 如果没有拆单, 或者查看主单详情, 则sale_order_uuid = 0;
  // 反之查看子单详情, 则sale_order_uuid = 为子单sale_order_uuid
  const saleOrderUuid = row.is_split === undefined ? row.sale_order_uuid : 0;
  router.push({
    path: '/' + appId + '/store/order/detail',
    query: {
      sale_bill_uuid: row.sale_bill_uuid,
      sale_order_uuid: saleOrderUuid,
    },
  });
}

// 获取物流信息
async function getLogistics(row) {
  const params = {
    order_id: row.order_id,
  };
  
  loading.value = true;
  try {
    const res = await OrderApi.queryLogistics(params, true);
    // logisticsData.value = res.data.express.list;
    // openLogistics();
  } catch (error) {
    // 错误处理
  } finally {
    loading.value = false;
  }
}

// 搜索查询
function onSearch() {
  clearTimeout(searchLoading.value);
  searchLoading.value = setTimeout(() => {
    curPage.value = 1;
    tableData.value = [];
    getData();
  }, 200);
}

// 导出
async function onExport() {
  searchForm.token = token;
  searchForm.dataType = activeName.value;
  try {
    await OrderApi.storeExport(
      {
        ...searchForm,
        request_type: 1,
      },
      true
    );
    
    const baseUrl = window.location.protocol + '//' + window.location.host;
    const url = baseUrl + '/index.php/shop/store.operate/export?' + qs.stringify(searchForm) + '&language=' + languageStore().language;
    window.open(url, '_blank');
  } catch (error) {
    // 错误处理
  }
}

// 打开取消
function cancelClick(item) {
  orderNo.value = item.order_no;
  orderId.value = item.sale_bill_uuid;
  openEdit.value = true;
}

// 打开退款
function refundClick(item) {
  orderNo.value = item.order_no;
  orderId.value = item.sale_bill_uuid;
  subOrderId.value = item.sale_order_uuid;
  payPrice.value = 0;
  openRefund.value = true;
}

// 删除
async function delClick(item) {
  try {
    await ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
      type: 'warning',
    });
    
    const saleOrderUuid = item.is_split === undefined ? item.sale_order_uuid : 0;
    await OrderApi.storedelete({
      sale_bill_uuid: item.sale_bill_uuid,
      sale_order_uuid: saleOrderUuid,
    });
    
    ElMessage({
      message: $t('删除成功'),
      type: 'success',
    });
    getData();
  } catch (error) {
    // 用户取消或出错
  }
}

// 关闭弹窗
function closeDialogFunc(e, f) {
  if (f === 'edit') {
    openEdit.value = e.openDialog;
    if (e.type === 'success') {
      getData();
    }
  }
}

// 关闭退款弹窗
function closeRefundDialogFunc(e, f) {
  if (f === 'edit') {
    openRefund.value = e.openDialog;
    if (e.type === 'success') {
      getData();
    }
  }
}

// 监听搜索表单时间模式变化
watch(() => searchForm.time_mode, (newVal, oldVal) => {
  nextTick(() => {
    if (newVal.length === 0) {
      searchForm.time_mode = oldVal;
    }
  });
}, { deep: true });

// 页面挂载时初始化
onMounted(() => {
  const params = languageStore().getPageParams().pageParams;

  if (params.value.page) {
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

  // 获取列表
  getData();
});
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
