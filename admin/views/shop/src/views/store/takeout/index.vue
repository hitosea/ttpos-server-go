<template>
  <div class="user">
    <!--搜索表单-->
    <div class="common-search-wrap">
      <el-form size="small" :inline="true" :model="searchForm" class="demo-form-inline">
        <el-form-item label="">
          <el-radio-group v-model="searchForm.date_range" class="radio-search" @change="timeTypeChange">
            <el-radio-button :value="0">{{ $t('今天') }}</el-radio-button>
            <el-radio-button :value="1">{{ $t('昨天') }}</el-radio-button>
            <el-radio-button :value="2">{{ $t('本周') }}</el-radio-button>
            <el-radio-button :value="-1">{{ $t('全部') }}</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="$t('外送序号/订单号')">
            <div class="flex-box">
            <el-select class="time-select" size="small" v-model="search_type" :placeholder="$t('请选择')" @change="()=>{ searchForm.serial_no = ''; searchForm.order_no = '';onSearch(); }">
              <el-option :label="$t('外送序号')" :value="1"></el-option>
              <el-option :label="$t('订单号')" :value="2"></el-option>
            </el-select>
            <el-input size="small" v-model="searchForm.serial_no" :placeholder="$t('外送序号')" @input="onSearch" v-if="search_type == 1"></el-input>
            <el-input size="small" v-model="searchForm.order_no" :placeholder="$t('订单号')" @input="onSearch" v-if="search_type == 2"></el-input>
          </div>
        </el-form-item>

        <el-form-item :label="$t('起始时间')">
          <div class="flex-box">
            <el-select class="time-select" size="small" v-model="searchForm.time_type" :placeholder="$t('请选择')" @change="onSearch">
              <el-option :label="$t('下单时间')" :value="1"></el-option>
              <el-option :label="$t('支付时间')" :value="2"></el-option>
            </el-select>
            <el-date-picker
              size="small"
              v-model="time"
              @change="createTimeChange"
              type="daterange"
              value-format="YYYY-MM-DD"
              range-separator="~"
              :start-placeholder="$t('开始日期')"
              :end-placeholder="$t('结束日期')"
              :disabled-date="(time) => time.getTime() > Date.now()"
            ></el-date-picker>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
            {{ $t('查询') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button v-auth="'/store/takeout/export'" size="small" type="primary" @click="onExport">{{ $t('导出') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
    <!--内容-->
    <div class="product-content">
      <div class="table-wrap">
        <el-tabs size="small" v-model="activeName" @tab-change="handleClick">
          <el-tab-pane :label="$t('全部订单')" name="">
            <template #label>
              <span>
                {{ $t('全部订单') }}
                <el-tag size="" class="ml-4">{{ order_count.total_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('待付款')" name="unpaid">
            <template #label>
              <span>
                {{ $t('待付款') }}
                <el-tag size="" class="ml-4">{{ order_count.unpaid_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('待接单')" name="unaccept">
            <template #label>
              <span>
                {{ $t('待接单') }}
                <el-tag size="" class="ml-4">{{ order_count.unaccept_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('备餐中')" name="accept">
            <template #label>
              <span>
                {{ $t('备餐中') }}
                <el-tag size="" class="ml-4">{{ order_count.accept_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('待配送')" name="undelivery">
            <template #label>
              <span>
                {{ $t('待配送') }}
                <el-tag size="" class="ml-4">{{ order_count.undelivery_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="$t('配送中')" name="delivery">
            <template #label>
              <span>
                {{ $t('配送中') }}
                <el-tag size="" class="ml-4">{{ order_count.delivery_num }}</el-tag>
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
          <el-tab-pane :label="$t('已完成')" name="completed">
            <template #label>
              <span>
                {{ $t('已完成') }}
                <el-tag size="" class="ml-4"> {{ order_count.complete_num }}</el-tag>
              </span>
            </template>
          </el-tab-pane>
        </el-tabs>
        <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
          <el-table-column prop="serial_number" :label="$t('外送序号')"> </el-table-column>
          <el-table-column prop="order_no" :label="$t('订单号')"></el-table-column>
          <el-table-column prop="status" :label="$t('状态')">
            <template #default="scope">
              {{ statusMap(scope.row.status) }}
            </template>
          </el-table-column>
          <el-table-column prop="create_time" :label="$t('下单时间')">
            <template #default="scope">
              {{ DTime(scope.row.create_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="pay_time" :label="$t('支付时间')">
            <template #default="scope">
              {{ DTime(scope.row.pay_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="origin_amount" :label="$t('订单金额')" width="140" show-overflow-tooltip>
            <template #default="scope">
              <div style="line-height: 24px">
                <main-currency>
                    {{ formatPrice(scope.row.origin_amount) }}
                </main-currency>
                <p class="gray98">
                  <sub-currency>
                    {{ formatPrice((Number(scope.row.origin_amount) * Number(currency.vices?.unit_rate)).toFixed(2)) }}
                  </sub-currency>
                </p>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="pay_amount" :label="$t('实付金额')" show-overflow-tooltip>
            <template #default="scope">
                <div class="orange">
                  <main-currency>
                    {{ formatPrice(scope.row.pay_amount) }}
                  </main-currency>
                </div>
            </template>
          </el-table-column>
          <el-table-column prop="" :label="$t('用户信息')" show-overflow-tooltip>
            <template #default="scope">
              <p class="gray9">{{ scope.row.contact.name || '-' }}</p>
              <p class="gray9">{{ scope.row.contact.phone || '-' }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="delivery_fee" :label="$t('配送费')" show-overflow-tooltip>
            <template #default="scope">
              <main-currency>
                {{ formatPrice(scope.row.delivery_fee) }}
              </main-currency>
            </template>
          </el-table-column>

          <el-table-column prop="pay_type" :label="$t('支付方式')" show-overflow-tooltip>
            <template #default="scope">
              <span class="gray9">{{ scope.row.pay_type || '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column fixed="right" :label="$t('操作')" width="180">
            <template #default="scope">
              <div>
                <el-button @click="detailClick(scope.row)" type="primary" link size="small" v-auth="'/store/order/detail'">{{ $t('详情') }} </el-button>
                <el-button v-if="scope.row.status == 5 || scope.row.status == 6" @click="contactClick(scope.row)" type="danger" link size="small"
                  >{{ $t('联系骑手') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_refund" @click="refundClick(scope.row)" type="danger" link size="small" v-auth="'/store/takeout/refund'"
                  >{{ $t('退款') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_reject" @click="rejectClick(scope.row)" type="danger" link size="small" v-auth="'/store/takeout/reject'"
                  >{{ $t('拒单') }}
                </el-button>
                <el-button v-if="scope.row.extra.is_cell_cancel" @click="cancelClick(scope.row)" type="danger" link size="small" v-auth="'/store/takeout/cancel'"
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
    <Cancel v-if="open_cancel" :open_cancel="open_cancel" :order_no="order_no" :member_sale_order_uuid="member_sale_order_uuid" @closeDialog="closeDialogFunc($event, 'edit')">
    </Cancel>
    <!--处理-->
    <refund
      v-if="open_refund"
      :open_refund="open_refund"
      :member_sale_order_uuid="member_sale_order_uuid"
      :pay_price="pay_price"
      @closeDialog="closerefundDialogFunc($event, 'edit')"
    >
    </refund>
  </div>
</template>

<script setup>
  import { ref, reactive, onMounted, watch, nextTick, getCurrentInstance } from 'vue';
  import { useRouter } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { ElMessageBox, ElMessage } from 'element-plus';
  import OrderApi from '@/api/order.js';
  import Cancel from './dialog/cancel.vue';
  import refund from './dialog/refund.vue';
  import qs from 'qs';
  import { useUserStore } from '@/store';
  import { DTime } from '@/utils/DateTime.js';
  import { languageStore } from '@/store/model/language';

  // 获取当前实例
  const { proxy } = getCurrentInstance();

  // 使用路由和国际化
  const router = useRouter();
  const { t: $t } = useI18n();

  // 使用store
  const { token, currency, computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;

  // 格式化价格函数
  const formatPrice = (price) => {
    return proxy.$formatPrice(price);
  };

  // 响应式数据
  const activeName = ref('');
  const loading = ref(true);
  const tableData = ref([]);
  const pageSize = ref(10);
  const totalDataNumber = ref(0);
  const curPage = ref(1);
  const search_type = ref(1);
  const searchForm = reactive({
    order_no: '',
    serial_no: '',
    style_id: ' ',
    date_range: 0,
    time_type: 1,
    query_start_time: 0,
    query_end_time: 0,
  });

  const time = ref('');

  const order_count = reactive({
    cancel_num: 0,
    complete_num: 0,
    page_no: 1,
    page_size: 10,
    total: 0,
    total_num: 0,
    unpaid_num: 0,
    accept_num: 0,
    unaccept_num: 0,
    undelivery_num: 0,
    delivery_num: 0,
  });

  const open_cancel = ref(false);
  const open_refund = ref(false);
  const order_no = ref(0);
  const pay_price = ref(0);
  const searchLoading = ref('');
  const member_sale_order_uuid = ref(0);
  // 监听器
  watch(
    () => searchForm.time_type,
    (newVal, oldVal) => {
      nextTick(() => {
        if (newVal.length == 0) {
          searchForm.time_type = oldVal;
        }
      });
    },
    { deep: true }
  );

  // 生命周期
  onMounted(() => {
    let params = languageStore().getPageParams().pageParams;

    if (params.value.page) {
      Object.assign(searchForm, {
        order_no: params.value.order_no,
        serial_no: params.value.serial_no,
        style_id: params.value.style_id,
        date_range: params.value.date_range,
        time_type: params.value.time_type,
        query_start_time: params.value.query_start_time,
        query_end_time: params.value.query_end_time,
        page_no: params.value.page_no,
        page_size: params.value.page_size,
      });
      time.value = params.value.time;
      activeName.value = params.value.dataType;
      curPage.value = params.value.page;
      pageSize.value = params.value.list_rows;
      languageStore().setPageParams({});
    }

    // 获取列表
    getData();
  });

  // 方法定义
  const arraySpanMethod = (row) => {
    if (row.rowIndex % 2 == 0) {
      if (row.columnIndex === 0) {
        return [1, 8];
      }
    }
  };

  const handleCurrentChange = (val) => {
    curPage.value = val;
    getData();
  };

  const handleSizeChange = (val) => {
    curPage.value = 1;
    pageSize.value = val;
    getData();
  };

  const handleClick = (tab, event) => {
    curPage.value = 1;
    getData();
  };

  const timeTypeChange = () => {
    searchForm.time = '';
    onSearch();
  };

  const createTimeChange = () => {
    if (time.value && time.value.length > 0) {
      // 转为时间戳
      searchForm.query_start_time = new Date(time.value[0]).getTime() / 1000;
      searchForm.query_end_time = new Date(time.value[1]).getTime() / 1000 + 86399;
    } else {
      searchForm.query_start_time = 0;
      searchForm.query_end_time = 0;
    }
    onSearch();
  };

  const getData = () => {
    let Params = { ...searchForm };
    Params.status = activeName.value;
    Params.page_no = curPage.value;
    Params.page_size = pageSize.value;
    loading.value = true;

    OrderApi.postTakeoutOrderList(Params, true)
      .then((res) => {
        tableData.value = res.data.list;
        totalDataNumber.value = res.data.meta.total;
        Object.assign(order_count, res.data.meta);
        loading.value = false;
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const detailClick = (row) => {
    let pageParams = { ...searchForm };
    pageParams.dataType = activeName.value;
    pageParams.time = time.value;
    pageParams.page = curPage.value;
    pageParams.list_rows = pageSize.value;
    languageStore().setPageParams(pageParams);

    // 如果没有拆单, 或者查看主单详情, 则sale_order_uuid = 0;
    // 反之查看子单详情, 则sale_order_uuid = 为子单sale_order_uuid
    const saleOrderUuid = row.is_split === undefined ? row.sale_order_uuid : 0;
    router.push({
      path: '/' + app_id + '/store/takeout/detail',
      query: {
        member_sale_order_uuid: row.member_sale_order_uuid,
      },
    });
  };

  const onSearch = () => {
    clearTimeout(searchLoading.value);
    searchLoading.value = setTimeout(() => {
      curPage.value = 1;
      tableData.value = [];
      getData();
    }, 200);
  };

  const onExport = () => {
    searchForm.token = token;
    searchForm.dataType = activeName.value;
    OrderApi.postTakeoutOrderExport(
      {
        ...searchForm,
        request_type: 1,
      },
      true
    )
      .then((data) => {
        loading.value = false;
        const baseUrl = window.location.protocol + '//' + window.location.host;
        const url = baseUrl + '/index.php/shop/store.MemberOrder/export?' + qs.stringify(searchForm) + '&language=' + languageStore().language;
        window.open(url, '_blank');
      })
      .catch((error) => {
        loading.value = false;
      });
  };

  const cancelClick = (item) => {
    member_sale_order_uuid.value = item.member_sale_order_uuid;
    order_no.value = item.order_no;
    open_cancel.value = true;
  };

  const refundClick = (item) => {
    member_sale_order_uuid.value = item.member_sale_order_uuid;
    pay_price.value = 0;
    open_refund.value = true;
  };

  const rejectClick = (item) => {
    member_sale_order_uuid.value = item.member_sale_order_uuid;
    ElMessageBox.confirm($t('是否取消此订单?'), $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      type: 'warning',
    })
      .then(() => {
        rejectSubmit();
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: $t('已取消'),
        });
      });
  };

  const contactClick = (item) => {
    member_sale_order_uuid.value = item.member_sale_order_uuid;
    ElMessageBox.confirm(`${$t('骑手：')}${item.rider.name}<br/>${$t('联系电话：')}${item.rider.phone}`, $t('提示'), {
      confirmButtonText: $t('确定'),
      cancelButtonText: $t('取消'),
      dangerouslyUseHTMLString: true,
    })
      .then(() => {})
      .catch(() => {});
  };

  const rejectSubmit = async () => {
    loading.value = true;
    try {
      const res = await OrderApi.postTakeoutOrderReject(
        {
          member_sale_order_uuid: member_sale_order_uuid.value,
        },
        true
      );
      ElMessage({
        message: $t('操作成功'),
        type: 'success',
      });
      getData();
    } catch (error) {
      ElMessage({
        message: $t('操作失败'),
        type: 'error',
      });
    } finally {
      loading.value = false;
    }
  };

  const closeDialogFunc = (e, f) => {
    if (f == 'edit') {
      open_cancel.value = e.openDialog;
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

  const statusMap = (status) => {
    switch (status) {
      case 0:
        return $t('选购中');
      case 1:
        return $t('待付款');
      case 2:
        return $t('待商家接单');
      case 3:
        return $t('商家备餐中');
      case 4:
        return $t('待骑手接单');
      case 5:
        return $t('骑手正在赶往商家');
      case 6:
        return $t('骑手配送中');
      case 7:
        return $t('已完成');
      case 8:
        return $t('已取消');
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
