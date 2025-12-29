<template>
  <div class="pb50" v-loading="loading">
    <div class="product-content">
      <div class="common-form">{{ $t('基本信息') }}</div>
      <div class="table-wrap">
        <el-row>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单类型：') }}</span>
              {{ $t('充值订单') }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单号：') }}</span>
              {{ detail.order_no }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.member">
            <div class="pb16">
              <span class="gray9">{{ $t('会员：') }}</span>
              <span>{{ detail.member?.nickname + '(' + detail.member?.phone + ')'}}</span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('订单金额：') }}</span>
              <main-currency>
                {{ proxy.$formatPrice(detail.recharge_amount) }}
              </main-currency>
              <span style="padding-left: 8px">
                <sub-currency>
                  {{ proxy.$formatPrice(Number(detail.recharge_amount) * Number(currency.vices?.unit_rate)) }}
                </sub-currency>
              </span>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 1">
            <div class="pb16">
              <span class="gray9">{{ $t('实付金额：') }}</span>
              <main-currency>
                {{ proxy.$formatPrice(detail.amount || 0) }}
              </main-currency>
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.status == 1" v-for="item in detail.payment_methods">
            <div class="pb16">
              <span class="gray9">{{ $t('支付方式：') }}</span>
              <span>
                {{ item.name }} {{ item.source_text ? `(${item.source_text})` : '' }}
              </span>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('交易状态：') }}</span>
              {{ detail.status == 0 ? $t('进行中') : detail.status == 2 ? $t('已取消') : $t('已完成') }}
            </div>
          </el-col>
          <el-col :span="6" v-if="detail.cashier">
            <div class="pb16">
              <span class="gray9">{{ $t('收银员：') }}</span>
              {{ detail.cashier?.real_name || '-' }}
            </div>
          </el-col>
          <el-col :span="6">
            <div class="pb16">
              <span class="gray9">{{ $t('支付时间：') }}</span>
              {{ detail.payment_time || '-' }}
            </div>
          </el-col>
        </el-row>
      </div>

      <div class="common-form mt16">
        {{ $t('订单信息') }}
      </div>
      <div class="table-wrap">
        <el-table size="small" :data="tableData" border style="width: 100%">
          <el-table-column prop="recharge_amount" :label="$t('充值金额')"></el-table-column>
          <el-table-column prop="gift_amount" :label="$t('赠送金额')"></el-table-column>
          <el-table-column prop="gift_point" :label="$t('赠送积分')"></el-table-column>
        </el-table>
      </div>

      <div class="common-form mt16">{{ $t('操作记录') }}</div>

      <div class="timeline-box">
        <el-timeline class="timeline-main">
          <el-timeline-item v-for="(activity, index) in activities" placement="top" :key="index" :color="activity.color" :hollow="activity.hollow" :timestamp="activity.timestamp">
            <div class="flex">
              {{ activity.content }}
              <template v-if="(activity.pay_type || []).length > 0">&nbsp;
                (<span class="flex" v-for="(item, itemIndex) in activity.pay_type" :key="itemIndex">
                  <span>{{ item.name }}:{{ item.unit + proxy.$formatPrice(item.refund_money) }}</span>
                  <el-tag v-if="item.refund_status == '0'" class="cupon" type="danger" @click="handleRetry(item)">{{ $t('退款失败，重试') }}</el-tag>
                  <el-tooltip v-if="item.value == '90333'" placement="bottom" trigger="click">
                    <el-icon class="icon"><WarningFilled /></el-icon>
                    <template #content>
                      <div> {{ $t('银行') }}{{ $t('：') }}{{ bankName(item.bank_code) }}</div>
                      <div> {{ $t('账户名') }}{{ $t('：') }}{{ item.account_no }} </div>
                      <div> {{ $t('账号') }}{{ $t('：') }}{{ item.account_name }} </div>
                    </template>
                  </el-tooltip>
                  <span v-if="itemIndex < (activity.pay_type || []).length - 1">，</span>
                </span>)
              </template> 
            </div>
          </el-timeline-item>
        </el-timeline>
      </div>
    </div>
    <div class="common-button-wrapper">
      <el-button size="small" @click="cancelFunc">{{ $t('返回') }}</el-button>
      <el-button
        v-if="detail.extra?.is_cell_refund"
        @click="refundClick(detail)"
        type="danger"
        size="small"
        v-auth="'/store/recharge/refund'"
        >{{ $t('退款') }}</el-button
      >
      <el-button v-if="detail.extra?.is_cell_cancel" @click="cancelClick(detail)" type="danger" size="small" v-auth="'/store/recharge/cancel'">{{ $t('取消') }} </el-button>
    </div>

    <!--处理-->
    <Cancel v-if="open_edit" :open_edit="open_edit" :order_no="order_no" :id="id" @closeDialog="closeDialogFunc($event, 'edit')"> </Cancel>
    <!--处理-->
    <refund v-if="open_refund" :open_edit="open_refund" :id="id" @closeDialog="closerefundDialogFunc($event, 'edit')"> </refund>
    <!--处理-->
    <refundAgain v-if="open_refundAgain" :open_edit="open_refundAgain" :refundOrder="refundOrder" @closeDialog="closerefundAgainDialogFunc($event)"> </refundAgain>
  </div>
</template>
<script setup>
  import { ref, getCurrentInstance, onMounted } from 'vue';
  import { useUserStore } from '@/store';
  import OrderApi from '@/api/order.js';
  import { useRoute, useRouter } from 'vue-router';
  import { languageStore } from '@/store/model/language';
  import Cancel from './dialog/cancel.vue';
  import refund from './dialog/refund.vue';
  import refundAgain from './dialog/refundAgain.vue';
  import { message } from '@/utils/message.js';

  const route = useRoute();
  const router = useRouter();

  const open_edit = ref(false);
  const open_refund = ref(false);
  const open_refundAgain = ref(false);
  const refundOrder = ref({});
  const order_no = ref('');
  const id = ref('');
  const pay_price = ref(0);
  const pageParams = ref({});
  const { proxy } = getCurrentInstance();
  const { currency } = useUserStore();
  const loading = ref(false);
  const detail = ref({});
  const tableData = ref([]);
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
          OrderApi.postRechargeOrderRefundAgain(params, true)
            .then((data) => {
              loading.value = false;
              proxy.$ElMessage({
                message: $t('操作成功'),
                type: 'success',
              });
              getDetail();
            })
            .catch((error) => {
              loading.value = false;
            });
        })
        .catch(() => {
          proxy.$ElMessage({
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

  const getDetail = async (id) => {
    loading.value = true;
    try {
      const res = await OrderApi.getRechargeOrderDetail({
        id: route.query.id,
      });
      detail.value = res.data;
      tableData.value = [];
      tableData.value.push({
        recharge_amount: detail.value.recharge_amount,
        gift_amount: detail.value.gift_amount,
        gift_point: detail.value.gift_point,
      });

      activities.value = [];
      res.data.operation_log.list.map((item) => {
        activities.value.push({
          refund_type: item.refund_type,
          pay_type: item.refund_pay_types,
          content: item.description,
          timestamp:
            item.create_time +
            ' ' +
            $t('操作人：') +
            (item.real_name ? item.real_name + (item.username ? `(${item.username})` : '') : item.username || '') + ' ' + item.client,
          color: '#0bbd87',
        });
      });
      loading.value = false;
    } catch (error) {
      loading.value = false;
    }
  };

  const refundClick = (item) => {
    id.value = item.uuid;
    open_refund.value = true;
  };

  /*返回*/
  const cancelFunc = () => {
    languageStore().setPageParams(pageParams.value);
    router.back(-1);
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
        await getDetail();
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

  /*关闭弹窗*/
  const closeDialogFunc = (e, f) => {
    if (f == 'edit') {
      open_edit.value = e.openDialog;
      if (e.type == 'success') {
        getDetail();
      }
    }
  };
  /*关闭弹窗*/
  const closerefundDialogFunc = (e, f) => {
    if (f == 'edit') {
      open_refund.value = e.openDialog;
      if (e.type == 'success') {
        getDetail();
      }
    }
  };

  const closerefundAgainDialogFunc = (e) => {
    open_refundAgain.value = e.openDialog;
    if (e.type == 'success') {
      getDetail();
    }
  };

  onMounted(() => {
    pageParams.value = JSON.parse(JSON.stringify(languageStore().getPageParams().pageParams.value));
    languageStore().setPageParams({});
    getDetail();
  });
</script>
<style lang="scss" scoped>
  .common-button-wrapper {
    justify-content: flex-end;
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
</style>
