<template>
  <div>
    <!-- 退款弹窗 -->
    <el-dialog :title="$t('退款')" v-model="dialogVisible" @close="dialogFormVisible(false)" :close-on-click-modal="false" :close-on-press-escape="false">
      <!-- 表单 -->
      <el-form size="small" v-loading="loading" ref="formRef" :model="form" label-position="top">
        <!-- 退款类型 -->
        <el-form-item for="no_click" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.refund_type">
            <el-radio value="1">{{ $t('整单退款') }}</el-radio>
            <el-radio value="2">{{ $t('部分退款') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- 整单退款 -->
        <el-form-item for="no_click" :label="$t('可退款金额')" prop="refundable_amount">
          <el-input v-model="info.refundable_amount" disabled :placeholder="$t('可退款金额')" :readonly="true">
            <template #prepend v-if="currency.unit_position == '0'">
              {{ currency.unit }}
            </template>
            <template #append v-if="currency.unit_position == '1'">
              {{ currency.unit }}
            </template>
          </el-input>
          <div class="tips" v-if="form.refund_type == '1'">{{ $t('注：确定后则为整单操作退款') }}</div>
        </el-form-item>
        <!-- 部分退款 -->
        <el-form-item v-if="form.refund_type == '2'" for="no_click" :label="$t('退款金额')" prop="refund_money" :rules="[{ required: true, message: $t('请输入退款金额') }]">
          <el-input-number
            :controls="false"
            :min="0"
            :max="info.cell_refund_money"
            @change="handleInput"
            :placeholder="$t('请输入本次退款金额')"
            v-model.number="form.refund_money"
          ></el-input-number>
        </el-form-item>
      </el-form>

      <div class="common-form mt24">{{ $t('订单充值信息') }}</div>
      <div class="detail-wrap">
        <div class="detail-item">
          <span class="span">{{ $t('充值金额：') }}</span>
          <span class="span">{{ proxy.$formatPrice(info?.recharge_amount || 0) }}</span>
        </div>
        <div class="detail-item">
          <span class="span">{{ $t('赠送金额：') }}</span>
          <span class="span">{{ proxy.$formatPrice(info?.gift_amount || 0) }}</span>
        </div>
        <div class="detail-item">
          <span class="span">{{ $t('赠送积分：') }}</span>
          <span class="span">{{ info?.gift_point || 0 }}</span>
        </div>
      </div>

      <div class="common-form mt16">{{ $t('当前会员剩余') }}</div>
      <div class="detail-wrap">
        <div class="detail-item">
          <span class="span">{{ $t('主账户：') }}</span>
          <span class="span">{{ proxy.$formatPrice(info?.recharge_member_info?.balance || 0) }}</span>
        </div>
        <div class="detail-item">
          <span class="span">{{ $t('赠送账户：') }}</span>
          <span class="span">{{ proxy.$formatPrice(info?.recharge_member_info?.gift_balance || 0) }}</span>
        </div>
        <div class="detail-item">
          <span class="span">{{ $t('积分：') }}</span>
          <span class="span">{{ info?.recharge_member_info?.points || 0 }}</span>
        </div>
      </div>
      <p class="tips mt4">{{ $t('退款后不变更赠送账户及积分，如需变更，请前去商家后台操作') }}</p>

      <div class="refund-total mt24">
        <p class="refund-total-title">{{ $t('支付记录（原路退款会按照以下顺序退回）') }}</p>
        <p class="refund-total-text" v-for="(item, index) in pay_list" :key="index">
          <span>{{ item.payment_name }}</span>
          <b>
            <span>{{ item.payment_amount }}</span>
            <span>
              (
              {{ $t('剩余可退') }}
              <span class="span">{{ item.refundable_amount }}</span>
              )
            </span>
          </b>
        </p>
      </div>
      <!-- 底部按钮 -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible(false)">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="submit()" :loading="loading">{{ $t('原路退回') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog :title="$t('提示')" v-model="dialogVisible2" @close="dialogFormVisible2(false)" :close-on-click-modal="false" :close-on-press-escape="false">
      <div>
        <p class="tips-QR">{{ $t('本次退款涉及') }}<span> QR PromptPay </span>{{ $t('退款，请填写退款所需的银行卡信息') }}</p>
        <el-form size="small" ref="bankFormRef" :model="bankForm" label-position="top">
          <!-- 退款类型 -->
          <el-form-item for="no_click" :label="$t('选择银行')" prop="bank_code" :rules="[{ required: true, message: $t('请选择银行') }]">
            <el-select v-model="bankForm.bank_code" clearable size="default" :placeholder="$t('请选择银行')">
              <template v-for="item in bankList">
                <el-option :value="item.value" :label="item.name"></el-option>
              </template>
            </el-select>
          </el-form-item>
          <!-- 账户名  -->
          <el-form-item for="no_click" :label="$t('账户名')" prop="account_name" :rules="[{ required: true, message: $t('请输入账户名') }]">
            <el-input v-model="bankForm.account_name" :placeholder="$t('请输入账户名')" clearable></el-input>
          </el-form-item>
          <!-- 账号 -->
          <el-form-item for="no_click" :label="$t('账号')" prop="account_no" :rules="[{ required: true, message: $t('请输入账号') }]">
            <el-input v-model="bankForm.account_no" :placeholder="$t('请输入账号')" clearable></el-input>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible2(false)">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script setup>
  import { onMounted, reactive, ref, getCurrentInstance, nextTick } from 'vue';
  import { useUserStore } from '@/store';
  import OrderApi from '@/api/order.js';
  import { message } from '@/utils/message.js';

  const props = defineProps({
    open_edit: Boolean,
    id: Number,
  });

  const emits = defineEmits(['closeDialog']);

  const { proxy } = getCurrentInstance();
  const { currency } = useUserStore();
  const dialogVisible = ref(props.open_edit);
  const loading = ref(false);
  const form = reactive({
    refund_type: '1',
    refund_money: null,
  });
  const formRef = ref(null);
  const pay_list = ref([]);
  const info = ref({});

  const dialogVisible2 = ref(false);
  const bankForm = reactive({
    bank_code: '',
    account_no: '',
    account_name: '',
  });
  const bankFormRef = ref(null);

  const bankList = [
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
  ];

  const getStoreRefundInfo = async () => {
    try {
      loading.value = true;
      const res = await OrderApi.getRechargeOrderRefund({ id: props.id }, true);
      loading.value = false;
      pay_list.value = res.data.pay_list;
      info.value = res.data.info;
    } catch (error) {
      loading.value = false;
    }
  };

  const handleClick = () => {
    bankFormRef.value.validate((valid) => {
      if (valid) {
        submit();
      }
    });
  };

  const submit = () => {
    formRef.value.validate(async (valid) => {
      if (valid) {
        loading.value = true;
        try {
          const params = {
            id: props.id,
            refund_type: form.refund_type,
            refund_money: form.refund_money,
          };
          if (dialogVisible2.value) {
            params.bank_code = bankForm.bank_code;
            params.account_no = bankForm.account_no;
            params.account_name = bankForm.account_name;
          }
          const res = await OrderApi.postRechargeOrderRefund(params, true);
          loading.value = false;
          message({
            message: res.msg,
            type: 'success',
          });
          console.log(res, 123123213);
          dialogFormVisible(true);
        } catch (error) {
          console.log(error, 456456);
          loading.value = false;
          if (error?.data?.code == -901) {
            dialogVisible2.value = true;
          }
        }
      }
    });
  };

  const dialogFormVisible = (e) => {
    if (e) {
      emits('closeDialog', {
        type: 'success',
        openDialog: false,
      });
    } else {
      emits('closeDialog', {
        type: 'error',
        openDialog: false,
      });
    }
  };

  const handleInput = (val) => {
    nextTick(() => {
      if (typeof val == 'number') {
        form.refund_money = proxy.$priceTwo(val || 0);
      }
    });
  };

  const dialogFormVisible2 = () => {
    dialogVisible2.value = false;
    bankForm.value = {
      bank_code: '',
      account_no: '',
      account_name: '',
    };
  };

  onMounted(() => {
    getStoreRefundInfo();
  });
</script>
<style lang="scss" scoped>
  .refund-total {
    padding: 16px;
    background: var(--el-color-primary-light-9);
    border-radius: 4px;
    display: flex;
    flex-direction: column;
    .refund-total-title {
      font-size: 14px;
      font-weight: 700;
      color: var(--el-color-black);
      margin-bottom: 4px;
    }
    .refund-total-text {
      font-size: 14px;
      font-weight: 400;
      color: var(--el-color-black);
      margin-top: 8px;
      display: flex;
      justify-content: space-between;
      .span {
        color: var(--el-color-danger);
      }
    }
  }
  .detail-wrap {
    padding: 16px;
    background: var(--el-color-primary-light-9);
    border-radius: 4px;
    display: flex;
    gap: 16px;
    .detail-item {
      flex: 1;
    }
  }
</style>
