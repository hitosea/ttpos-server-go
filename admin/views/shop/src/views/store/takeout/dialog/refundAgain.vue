<template>
  <div>
    <el-dialog :title="$t('提示')" v-model="dialogVisible" @close="dialogFormVisible(false)" :close-on-click-modal="false" :close-on-press-escape="false">
      <div>
        <p class="tips-QR" v-if="dataChange">{{ $t('上一次') }}<span> QR PromptPay </span>{{ $t('退款信息如下') }}</p>
        <p class="tips-QR" v-else>{{ $t('此前已退款，已修改退款信息（以下为最新信息）') }}</p>
        <el-form size="small" ref="bankFormRef" :model="bankForm" label-position="top">
          <!-- 退款类型 -->
          <el-form-item for="no_click" :label="$t('选择银行')" prop="bank_code" :rules="[{ required: true, message: $t('请选择银行') }]">
            <el-select v-model="bankForm.bank_code" clearable size="default" :placeholder="$t('请选择银行')" :disabled="edit">
              <template v-for="item in bankList">
                <el-option :value="item.value" :label="item.name"></el-option>
              </template>
            </el-select>
          </el-form-item>
          <!-- 账户名  -->
          <el-form-item for="no_click" :label="$t('账户名')" prop="account_name" :rules="[{ required: true, message: $t('请输入账户名') }]">
            <el-input v-model="bankForm.account_name" :placeholder="$t('请输入账户名')" clearable :disabled="edit"></el-input>
          </el-form-item>
          <!-- 账号 -->
          <el-form-item for="no_click" :label="$t('账号')" prop="account_no" :rules="[{ required: true, message: $t('请输入账号') }]">
            <el-input v-model="bankForm.account_no" :placeholder="$t('请输入账号')" clearable :disabled="edit"></el-input>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <template v-if="edit">
            <el-button type="primary" @click="handleEdit" :loading="loading">{{ $t('修改') }}</el-button>
            <el-button @click="dialogFormVisible(false)" :loading="loading">{{ $t('取消') }}</el-button>
            <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('重试退款') }}</el-button>
          </template>
          <template v-else>
            <el-button @click="dialogFormVisible(false)" :loading="loading">{{ $t('取消') }}</el-button>
            <el-button type="primary" @click="handleEdit" :loading="loading">{{ $t('确定') }}</el-button>
          </template>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script setup>
  import { ref, watch } from 'vue';
  import OrderApi from '@/api/order.js';
  import { message } from '@/utils/message.js';

  const props = defineProps({
    open_edit: Boolean,
    refundOrder: Object,
  });

  const dialogVisible = ref(props.open_edit);

  const edit = ref(true);
  const loading = ref(false);
  const bankForm = ref({
    bank_code: '',
    account_no: '',
    account_name: '',
  });
  const bankFormRef = ref(null);
  const dataChange = ref(true);

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

  watch(
    () => props.open_edit,
    (val) => {
      if (val) {
        bankForm.value.bank_code = props.refundOrder.bank_code;
        bankForm.value.account_name = props.refundOrder.account_name;
        bankForm.value.account_no = props.refundOrder.account_no;
      }
    },
    {
      deep: true,
      immediate: true,
    }
  );

  const emits = defineEmits(['closeDialog']);

  /*关闭弹窗*/
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

  /*判断是否是旧数据*/
  const oldData = () => {
    return (
      bankForm.value.bank_code == props.refundOrder.bank_code &&
      bankForm.value.account_name == props.refundOrder.account_name &&
      bankForm.value.account_no == props.refundOrder.account_no
    );
  };

  const handleClick = async () => {
    bankFormRef.value.validate(async (valid) => {
      if (valid) {
        try {
          loading.value = true;
          const params = {
            return_order_uuid: props.refundOrder.return_order_uuid,
            return_amount_uuid: props.refundOrder.return_amount_uuid,
            bank_code: bankForm.value.bank_code,
            account_no: bankForm.value.account_no,
            account_name: bankForm.value.account_name,
          };
          const res = await OrderApi.orderRefundAgain(params, true);
          loading.value = false;
          message({
            message: res.msg,
            type: 'success',
          });
          dialogFormVisible(true);
        } catch (error) {
          loading.value = false;
        }
      }
    });
  };

  const handleEdit = () => {
    if (!oldData()) {
      dataChange.value = false;
    }
    edit.value = !edit.value;
  };
</script>
<style lang="scss" scoped>
  .tips-QR {
    font-size: 15px;
    font-weight: 500;
    margin-bottom: 12px;
    span {
      color: var(--el-color-primary);
      font-weight: 700;
    }
  }
</style>
