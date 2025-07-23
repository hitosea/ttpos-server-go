<template>
  <div class="pb50" v-loading="loading">
    <div class="product-content">
      <!--基本信息-->
      <div class="common-form">{{ $t('基本信息') }}</div>
      <div class="table-wrap">
        <el-row>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单类型：') }}</span>
              {{ $t('外送订单') }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单号：') }}</span>
              {{ detail.order_no }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('会员：') }}</span>
              <span>{{ $t('会员ID') }}&nbsp;({{ detail?.member?.id }})</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('外送序号：') }}</span>
              <span>{{ detail?.serial_no || '-' }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单状态：') }}</span>
              {{ statusMap(detail.status) }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('收银员：') }}</span>
              <span>{{ detail.cachier?.name || '-' }}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单金额：') }}</span>
              <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
              {{ formatPrice(detail.origin_amount) }}
              <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
              <span v-if="currency.is_open == 1" style="padding-left: 8px">
                <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
                {{ formatPrice(Number(detail.origin_amount) * Number(currency.vices?.unit_rate)) }}
                <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }}</template>
              </span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('会员折扣：') }}</span>
              <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
              {{ formatPrice(detail.member_discount) }}
              <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
            </div>
          </el-col>

          <el-col :span="6" v-if="detail.status == 7">
            <div class="pb16">
              <span class="gray9">{{ $t('实付金额：') }}</span>
              <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
              {{ formatPrice(Number(detail.pay_amount)) }}
              <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
            </div>
          </el-col>
          <el-col :span="6" v-if="Number(detail.refund_amount || 0) > 0">
            <div class="pb16">
              <span class="gray9">{{ $t('退款金额：') }}</span>
              <span>
                <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                {{ formatPrice(detail.refund_amount) }}
                <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 7" >
            <div class="pb16">
              <span class="gray9">{{ $t('支付方式：') }}</span>
              <span>
                {{ detail.pay_type }}
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 7">
            <div class="pb16">
              <span class="gray9">{{ $t('支付时间：') }}</span>
              <span>
                {{ DTime(detail.pay_time) }}
              </span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单时间：') }}</span>
              {{ DTime(detail.create_time) }} {{ $t('至') }} {{ DTime(detail.finish_time) }}
            </div>
          </el-col>
          <el-col :span="6" >
            <div class="pb16">
              <span class="gray9">{{ $t('备注：') }}</span>
              {{ detail.remark || '-' }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 8">
            <div class="pb16">
              <span class="gray9">{{ $t('取消原因：') }}</span>
              {{ detail.cancel_reason || '-' }}
            </div>
          </el-col>



        </el-row>
      </div>

      <div class="common-form mt16"> {{ $t('商品信息') }} </div>

      <div class="table-wrap">
        <el-table size="small" :data="detail?.product_list?.list" border style="width: 100%">
          <el-table-column prop="name_text" :label="$t('商品')" width="400">
            <template #default="scope">
              <div class="product-info">
                <div class="pic">
                  <img v-if="scope.row.is_buffet_customer" src="../../../assets/img/buffet.png" />
                  <img v-else-if="scope.row.is_delay" src="../../../assets/img/clock.png" />
                  <img v-else v-img-url="scope.row.image_url" />
                </div>
                <div class="info">
                  <div class="name">
                    <el-tag class="mr-8" type="danger" effect="light" :hit="true" size="large" v-if="scope.row.refund_reason">{{ $t('退') }}</el-tag>
                    <el-tag class="mr-8" color="#FF3300" effect="dark" :hit="true" size="large" v-if="scope.row.is_gift">{{ $t('赠') }}</el-tag>
                    {{ scope.row.locale_name[language] || '-' }}
                  </div>
                  <div class="gray9" v-if="scope.row.locale_attribute_name[language]">
                    {{ scope.row.locale_attribute_name[language] }}
                  </div>
                  <div class="price">
                    <span>
                      <template v-if="currency.unit_position == '0'">{{ currency.unit }}</template>
                      {{ formatPrice(scope.row.origin_unit_price) }}
                      <template v-if="currency.unit_position == '1'">{{ currency.unit }}</template>
                      <span v-if="currency.is_open == 1" style="padding-left: 8px">
                        <template v-if="currency.vices.vice_unit_position == '0'">{{ currency.vices?.vice_unit }}</template>
                        {{ formatPrice(Number(scope.row.origin_unit_price) * Number(currency.vices?.unit_rate)) }}
                        <template v-if="currency.vices.vice_unit_position == '1'">{{ currency.vices?.vice_unit }}</template>
                      </span>
                    </span>
                  </div>
                  <div>
                    <p class="gray9" v-if="scope.row.is_gift > 0">{{ $t('赠送原因：') }}{{ scope.row.gift_reason || '-' }}</p>
                    <p class="gray9" v-if="scope.row.refund_reason">{{ $t('退菜原因：') }}{{ scope.row.refund_reason || '-' }}</p>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="num" :label="$t('购买数量')">
            <template #default="scope">
              <p>{{ scope.row.num }}</p>
            </template>
          </el-table-column>
          <el-table-column prop="product_price" :label="$t('商品总价')">
            <template #default="scope">
              <div>
                <template v-if="scope.row.total_price != scope.row.sale_price">
                  <span class="text-line-through" v-if="scope.row.sale_price">
                    <template v-if="currency.unit_position == '0'">
                      {{ currency.unit }}
                    </template>
                    {{ formatPrice(Number(scope.row.sale_price)) }}
                    <template v-if="currency.unit_position == '1'">
                      {{ currency.unit }}
                    </template>
                  </span>
                  <span class="text-line-through" v-if="currency.is_open == 1 && scope.row.sale_price">
                    <template v-if="currency.vices.vice_unit_position == '0'">
                      {{ currency.vices?.vice_unit }}
                    </template>
                    {{ formatPrice(Number(scope.row.sale_price) * Number(currency.vices?.unit_rate)) }}
                    <template v-if="currency.vices.vice_unit_position == '1'">
                      {{ currency.vices?.vice_unit }}
                    </template>
                  </span>
                </template>

                <span>
                  <template v-if="currency.unit_position == '0'">
                    {{ currency.unit }}
                  </template>
                  {{ formatPrice(Number(scope.row.total_price)) }}
                  <template v-if="currency.unit_position == '1'">
                    {{ currency.unit }}
                  </template>
                </span>

                <span v-if="currency.is_open == 1">
                  <template v-if="currency.vices.vice_unit_position == '0'">
                    {{ currency.vices?.vice_unit }}
                  </template>
                  {{ formatPrice(Number(scope.row.total_price) * Number(currency.vices?.unit_rate)) }}
                  <template v-if="currency.vices.vice_unit_position == '1'">
                    {{ currency.vices?.vice_unit }}
                  </template>
                </span>
                <span class="tips" v-if="Number(scope.row.refund_amount) > 0"> （{{ $t('退款：') + currency.unit + formatPrice(scope.row.refund_amount) }}） </span>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="common-form mt16">{{ $t('操作记录') }}</div>

      <div class="timeline-box">
        <el-timeline class="timeline-main">
          <el-timeline-item v-for="(activity, index) in activities" placement="top" :key="index" :color="activity.color" :hollow="activity.hollow" :timestamp="activity.timestamp">
            <div class="flex">
              {{ activity.content }}
              <template v-if="(activity.pay_type || []).length > 0"
                >&nbsp; (<span class="flex" v-for="(item, itemIndex) in activity.pay_type" :key="itemIndex">
                  <span>{{ item.name }}:{{ item.unit + formatPrice(item.refund_money) }}</span>
                  <el-tag v-if="item.refund_status == '0'" class="cupon" type="danger" @click="handleRetry(item)">{{ $t('退款失败，重试') }}</el-tag>
                  <el-tooltip v-if="item.value == '90333'" placement="bottom" trigger="click">
                    <el-icon class="icon"><WarningFilled /></el-icon>
                    <template #content>
                      <div> {{ $t('银行') }}{{ $t('：') }}{{ bankName(item.bank_code) }}</div>
                      <div> {{ $t('账户名') }}{{ $t('：') }}{{ item.account_no }} </div>
                      <div> {{ $t('账号') }}{{ $t('：') }}{{ item.account_name }} </div>
                    </template>
                  </el-tooltip>
                  <span v-if="itemIndex < (activity.pay_type || []).length - 1">、</span> </span
                >)
              </template>
            </div>
          </el-timeline-item>
        </el-timeline>
      </div>

      <div class="table-wrap" v-if="detail.status == 2 && detail.cancel_reason != ''">
        <div class="common-form mt16">{{ $t('取消信息') }}</div>
        <div class="table-wrap">
          <el-row>
            <el-col :span="6">
              <div class="pb16">
                <span class="gray9">{{ $t('商家备注') }}:</span>
                {{ detail.cancel_reason }}
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="cancelFunc">{{ $t('返回') }}</el-button>
      <!-- <el-button v-if="extra?.is_cell_refund" @click="refundClick(detail)" type="danger" size="small" v-auth="'/store/operate/refund'">{{ $t('退款') }}</el-button>
      <el-button v-if="extra?.is_cell_cancel" @click="cancelClick(detail)" type="danger" size="small" v-auth="'/store/operate/order_cancel'">{{ $t('取消') }} </el-button>
      <el-button v-if="extra?.is_cell_delete" @click="delClick(detail)" type="danger" size="small" v-auth="'/store/order/delete'">{{ $t('删除') }} </el-button> -->
    </div>
    <!--处理-->
    <!-- <Cancel v-if="open_edit" :open_edit="open_edit" :order_no="order_no" :order_id="sale_bill_uuid" @closeDialog="closeDialogFunc($event, 'edit')"> </Cancel>
    <refund
      v-if="open_refund"
      :open_edit="open_refund"
      :order_no="order_no"
      :order_id="sale_bill_uuid"
      :sub_order_id="sale_order_uuid"
      :pay_price="pay_price"
      @closeDialog="closerefundDialogFunc($event, 'edit')"
    ></refund>
    <refundAgain v-if="open_refundAgain" :open_edit="open_refundAgain" :refundOrder="refundOrder" @closeDialog="closerefundAgainDialogFunc($event)"> </refundAgain> -->
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, getCurrentInstance } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ElMessageBox, ElMessage } from 'element-plus';
import { WarningFilled } from '@element-plus/icons-vue';
import OrderApi from '@/api/order.js';
// import Cancel from './dialog/cancel.vue';
// import refund from './dialog/refund.vue';
// import refundAgain from './dialog/refundAgain.vue';
import { useUserStore } from '@/store';
import { languageStore } from '@/store/model/language';
import { DTime } from '@/utils/DateTime.js';

// 获取当前实例
const { proxy } = getCurrentInstance();

// 使用路由和国际化
const route = useRoute();
const router = useRouter();
const { t: $t } = useI18n();

// 使用store
const { currency } = useUserStore();

// 格式化价格函数
const formatPrice = (price) => {
  return proxy.$formatPrice(price);
};

// 响应式数据
const language = ref('');
const loading = ref(true);

const detail = reactive({
  sale_bill_uuid: 0,
  sale_order_uuid: 0,
  pay_status: [],
  pay_type: [],
  delivery_type: [],
  user: {},
  address: [],
  product: [],
  order_status: [],
  extract: [],
  pay_types: [],
  sale_orders: [],
  supplier: {
    name: '',
  },
  status: 0,
});

// const extra = ref({});
// const open_edit = ref(false);
// const open_refund = ref(false);
const open_refundAgain = ref(false);
const refundOrder = ref({});
// const order_no = ref(0);
// const sale_bill_uuid = ref(0);
// const sale_order_uuid = ref(0);
// const pay_price = ref(0);
const pageParams = ref({});
const activities = ref([]);

const bankList = ref([
  { name: 'BANGKOK BANK (BBL)', value: '002' },
  { name: 'KASIKORNBANK (KBANK)', value: '004' },
  { name: 'KRUNG THAI BANK (KTB)', value: '006' },
  { name: 'TMBTHANACHART BANK (TTB)', value: '011' },
  { name: 'SIAM COMMERCIAL BANK (SCB)', value: '014' },
  { name: 'CITIBANK BANGKOK BRANCH (CITI)', value: '017' },
  { name: 'SUMITOMO MITSUI BANK (SMBC)', value: '018' },
  { name: 'STANDARD CHARTERED BANK THAI (SCBT)', value: '020' },
  { name: 'CIMB THAI BANK (CIMBT)', value: '022' },
  { name: 'UNITED OVERSEAS BANK THAI (UOBT)', value: '024' },
  { name: 'BANK OF AYUDHYA (BAY)', value: '025' },
  { name: 'GOVERNMENT SAVINGS BANK (GSB)', value: '030' },
  { name: 'THE HONGKONG AND SHANGHAI BANKING CORPORATION (HSBC)', value: '031' },
  { name: 'GOVERNMENT HOUSING BANK (GHB)', value: '033' },
  { name: 'BANK FOR AGRICULTURE AND AGRICULTURAL COOPERATIVES (BAAC)', value: '034' },
  { name: 'MIZUHO CORPORATE BANK (MHCB)', value: '039' },
  { name: 'ISLAMIC BANK OF THAILAND (ISBT)', value: 'ISBT' },
  { name: 'TISCO BANK (TISCO)', value: 'TISCO' },
  { name: 'KIATNAKIN BANK (KK)', value: '069' },
  { name: 'INDUSTRIAL AND COMMERCIAL BANK OF CHINA (ICBC THAI)', value: '070' },
  { name: 'THAI CREDIT RETAIL BANK (TCRB)', value: '071' },
  { name: 'LAND AND HOUSES BANK (LH BANK)', value: '073' },
]);



// 生命周期
onMounted(() => {
  language.value = languageStore()?.getLanguageKey().language.value;
  pageParams.value = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
  languageStore().setPageParams({});

  // 获取列表
  getParams();
});



const getParams = () => {
  // 取到路由带过来的参数
  OrderApi.postTakeoutOrderDetail(
    {
      member_sale_order_uuid: route.query.member_sale_order_uuid,
    },
    true
  )
    .then((res) => {
      loading.value = false;
      Object.assign(detail, res.data);
      activities.value = [];
      res.data.operation_log.list.map((item) => {
        activities.value.push({
          refund_type: item.refund_type,
          pay_type: item.pay_type,
          content: item.description,
          timestamp:
            DTime(item.create_time) +
            ' ' +
            $t('操作人：') +
            (item.user_name ? item.user_name + (item.user_email ? `(${item.user_email})` : '') : item.user_email || '') +
            ' ' +
            item.source,
          color: '#0bbd87',
        });
      });
    })
    .catch((error) => {
      loading.value = false;
    });
};

const cancelFunc = () => {
  languageStore().setPageParams(pageParams.value);
  router.back(-1);
};

// const cancelClick = (item) => {
//   order_no.value = item.order_no;
//   sale_bill_uuid.value = item.sale_bill_uuid;
//   open_edit.value = true;
// };

// const refundClick = (item) => {
//   order_no.value = item.order_no;
//   sale_bill_uuid.value = item.sale_bill_uuid;
//   sale_order_uuid.value = item.sale_orders[0].sale_order_uuid;
//   pay_price.value = 0;
//   open_refund.value = true;
// };

// const closeDialogFunc = (e, f) => {
//   if (f == 'edit') {
//     open_edit.value = e.openDialog;
//     if (e.type == 'success') {
//       getParams();
//     }
//   }
// };

// const closerefundDialogFunc = (e, f) => {
//   if (f == 'edit') {
//     open_refund.value = e.openDialog;
//     if (e.type == 'success') {
//       getParams();
//     }
//   }
// };

// const closerefundAgainDialogFunc = (e) => {
//   open_refundAgain.value = e.openDialog;
//   if (e.type == 'success') {
//     getParams();
//   }
// };

// const delClick = (item) => {
//   ElMessageBox.confirm($t('删除后不可恢复，确认删除吗?'), $t('提示'), {
//     type: 'warning',
//   }).then(() => {
//     OrderApi.storedelete({
//       sale_bill_uuid: item.sale_bill_uuid,
//       sale_order_uuid: item.is_split === 'undefined' ? item.sale_order_uuid : 0,
//     }).then((data) => {
//       ElMessage({
//         message: $t('删除成功'),
//         type: 'success',
//       });
//       router.back(-1);
//     });
//   });
// };

const handleRetry = (item) => {
  refundOrder.value = item;
  if (item.value == '90333') {
    open_refundAgain.value = true;
  } else {
    ElMessageBox.confirm('确定重试退款操作?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
      .then(() => {
        loading.value = true;
        const params = {
          return_order_uuid: item.return_order_uuid,
          return_amount_uuid: item.return_amount_uuid,
        };
        OrderApi.orderRefundAgain(params, true)
          .then((data) => {
            loading.value = false;
            ElMessage({
              message: $t('操作成功'),
              type: 'success',
            });
            getParams();
          })
          .catch((error) => {
            loading.value = false;
          });
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: '已取消操作',
        });
      });
  }
};

const bankName = (value) => {
  let name = '';
  bankList.value.map((item) => {
    if (item.value == value) {
      name = item.name;
    }
  });
  return name;
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
<style scoped lang="scss">
  .common-button-wrapper {
    justify-content: flex-end;
  }
  .el-radio-button {
    margin-right: 0;
  }
  .text-line-through {
    text-decoration: line-through;
    color: var(--el-color-tips);
  }
  .text-line-through:nth-child(2) {
    margin-right: 16px;
  }
  .product-info {
    .info {
      .name {
        display: flex;
        align-items: center;
      }
    }
  }
  .mr-8 {
    margin-right: 8px;
  }
  .timeline-box {
    padding-left: 1px;

    .timeline-main {
      font-size: 14px;
    }
  }
  :deep(.el-timeline-item__content) {
    display: flex;
    align-items: center;
  }
  .flex {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
  }
  .cupon {
    cursor: pointer;
    margin-left: 4px;
  }
  .icon {
    cursor: pointer;
    margin-left: 4px;
    font-size: 16px;
  }

  .sub-order {
    display: flex;
    gap: 24px;
    padding: 12px 0 8px 0;
  }
</style>
