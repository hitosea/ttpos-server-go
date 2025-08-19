<template>
  <div>
    <!-- 退款弹窗 -->
    <el-dialog :title="$t('退款')" v-model="dialogVisible" @close="dialogFormVisible(false)" :close-on-click-modal="false" :close-on-press-escape="false">
      <!-- 表单 -->
      <el-form size="small" ref="form" :model="form" label-position="top">
        <!-- 退款类型 -->
        <el-form-item for="no_click" :rules="[{ required: true, message: '' }]">
          <el-radio-group v-model="form.refund_type">
            <el-radio value="1">{{ $t('整单退款') }}</el-radio>
            <!-- <el-radio label="2">{{ $t('部分退款') }}</el-radio> -->
          </el-radio-group>
        </el-form-item>
        <!-- 整单退款 -->
        <el-form-item
          v-if="form.refund_type == '1'"
          for="no_click"
          :label="$t('可退款金额')"
          :label-width="formLabelWidth"
          prop="pay_price"
          :rules="[{ required: true, message: ' ' }]"
        >
          <el-input v-model="form.pay_price" disabled :placeholder="$t('可退款金额')" :readonly="true">
            <template #prepend v-if="currency.unit_position == '0'">
              {{ currency.unit }}
            </template>
            <template #append v-if="currency.unit_position == '1'">
              {{ currency.unit }}
            </template>
          </el-input>
          <div class="tips">{{ $t('注：确定后则为整单操作退款') }}</div>
        </el-form-item>
      </el-form>
      <!-- 部分退款 -->
      <el-table v-if="form.refund_type == '2'" size="small" :data="tableData" border style="width: 100%" class="mb16" v-loading="loading">
        <el-table-column prop="product_name_text" :label="$t('商品名称')">
          <template #default="scope">
            {{ scope.row.product_name_text }}
            <span class="tips" v-if="scope.row.product_attr"> ({{ scope.row.product_attr }}) </span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('可退数量')">
          <template #default="scope">
            {{ Number(scope.row.total_num) }}
          </template>
        </el-table-column>
        <el-table-column prop="refund_num_updata" :label="$t('退菜数量')">
          <template #default="scope">
            <el-input-number :min="0" :max="Number(scope.row.total_num)" :placeholder="$t('请输入')" v-model.number="scope.row.refund_num_updata"></el-input-number>
          </template>
        </el-table-column>
        <el-table-column :label="$t('退款金额')" align="right">
          <template #default="scope">
            <div class="flex">
              <p>
                <main-currency>
                  {{ proxy.$formatPrice(Number(scope.row.refund_num_updata) * Number(scope.row.price)) }}
                </main-currency>
              </p>
              <p class="tips">
                {{ $t('可退款金额：') }}
                <main-currency>
                  {{ proxy.$formatPrice(Number(scope.row.total_price)) }}
                </main-currency>
              </p>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="refund-total">
        <p class="refund-total-title">{{ $t('支付记录（原路退款会按照以下顺序退回）') }}</p>
        <p class="refund-total-text" v-for="(item, index) in pay_list" :key="index">
          <span>{{ item.payment_method_name }}</span>
          <b>
            <span>{{ item.payment_amount }}</span>
            <span>
              (
              {{ $t('剩余可退') }}
              <span class="span">{{ item.can_return_amount }}</span>
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
        <el-form size="small" ref="bankForm" :model="bankForm" label-position="top">
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
  import { ref, reactive, onMounted, getCurrentInstance, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language';
  import OrderOldApi from '@/api/orderOld.js';
  import draggable from 'vuedraggable';

  // 获取Vue实例
  const { proxy } = getCurrentInstance();

  // 获取store数据
  const { currency } = useUserStore();

  // 定义props
  const props = defineProps({
    open_edit: {
      type: Boolean,
      default: false,
    },
    order_id: {
      type: [String, Number],
      default: '',
    },
    sub_order_id: {
      type: [String, Number],
      default: '',
    },
    pay_price: {
      type: [String, Number],
      default: '',
    },
  });

  // 定义emits
  const emit = defineEmits(['closeDialog']);

  // 响应式变量
  const loading = ref(false); // 加载状态
  const formLabelWidth = ref('120px'); // 左边长度
  const dialogVisible = ref(false); // 是否显示
  const dialogVisible2 = ref(false);
  const language = ref('');

  // 表单数据
  const form = reactive({
    refund_type: '1', // 退款类型
    order_id: '', // 订单id
    sub_order_id: '', // 子单id
    pay_price: '', // 支付金额
  });

  const tableData = ref([]); // 表格数据
  const pay_list = ref([]); // 支付记录（原路退款会按照以下顺序退回）

  // 银行表单
  const bankForm = reactive({
    bank_code: '',
    account_no: '',
    account_name: '',
  });

  // 银行列表
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

  // 监听props变化
  watch(
    () => props.open_edit,
    (newVal) => {
      dialogVisible.value = newVal;
    }
  );

  // 方法定义

  // 处理提交
  const submit = async () => {
    let submitForm = {
      sale_bill_uuid: form.order_id,
      sale_order_uuid: form.sub_order_id,
      refund_type: form.refund_type,
      refund_product: [],
      refund_buffet: [],
      refund_delay: [],
    };

    tableData.value.map((item) => {
      if (item.refund_num_updata > 0 && item.type == 1) {
        submitForm.refund_buffet.push({
          id: item.id,
          refund_num: item.refund_num_updata,
        });
      }
      if (item.refund_num_updata > 0 && item.type == 2) {
        submitForm.refund_delay.push({
          id: item.id,
          refund_num: item.refund_num_updata,
        });
      }
      if (item.refund_num_updata > 0 && item.type == 3) {
        submitForm.refund_product.push({
          order_product_id: item.id,
          refund_num: item.refund_num_updata,
        });
      }
    });

    if (form.refund_type == '2' && submitForm.refund_product.length == 0 && submitForm.refund_buffet.length == 0 && submitForm.refund_delay.length == 0) {
      ElMessage({
        type: 'warning',
        message: proxy.$t('请选择退款商品'),
      });
      return;
    }

    if (dialogVisible2.value) {
      submitForm.bank_code = bankForm.bank_code;
      submitForm.account_no = bankForm.account_no;
      submitForm.account_name = bankForm.account_name;
    }

    try {
      const valid = await proxy.$refs.form.validate();
      if (valid) {
        loading.value = true;
        const data = await OrderOldApi.storeRefund(submitForm, true);
        loading.value = false;
        ElMessage({
          message: data.msg,
          type: 'success',
        });
        dialogFormVisible(true);
        dialogFormVisible2(false);
      }
    } catch (error) {
      loading.value = false;
      if (error?.data?.code == -901) {
        dialogVisible2.value = true;
      }
    }
  };
  // 2025年04月29日10:50:36 旧订单只有整单退款
  //   getData() {
  //     let self = this;
  //     self.loading = true;
  //     OrderOldApi.orderProductList({ order_id: this.form.order_id })
  //       .then((data) => {
  //         self.loading = false;
  //         this.tableData = [];

  //         (data.data.buffetList || []).map((item) => {
  //           this.tableData.push({
  //             type: 1,
  //             product_name_text: item.buffet_name_text + `(${item.customer_type_name_text})`, //名称
  //             product_attr: '', //规格
  //             total_num: item.num, //总数量
  //             total_price: item.consumption_tax_pay_price * Number(item.num), //总价钱
  //             refund_num: item.refund_num, //已退数量
  //             refund_money: item.consumption_tax_pay_price * Number(item.refund_num), //已退价钱
  //             id: item.id, //id
  //             refund_num_updata: 0, //提交退款的数量
  //           });
  //         });

  //         (data.data.delayList || []).map((item) => {
  //           this.tableData.push({
  //             type: 2,
  //             product_name_text: item.name_text, //名称
  //             product_attr: '', //规格
  //             total_num: item.num, //总数量
  //             total_price: item.price * Number(item.num), //总价钱
  //             refund_num: item.refund_num, //已退数量
  //             refund_money: item.refund_money, //已退价钱
  //             id: item.id, //id
  //             refund_num_updata: 0, //提交退款的数量
  //           });
  //         });

  //         (data.data.productList || []).map((item) => {
  //           this.tableData.push({
  //             type: 3,
  //             product_name_text: item.product_name_text, //名称
  //             product_attr: item.product_attr, //规格
  //             total_num: item.total_num, //总数量
  //             total_price: item.consumption_tax_pay_price * Number(item.total_num), //总价钱
  //             refund_num: item.refund_num, //已退数量
  //             refund_money: item.consumption_tax_pay_price * Number(item.refund_num), //已退价钱
  //             id: item.order_product_id, //id
  //             refund_num_updata: 0, //提交退款的数量
  //           });
  //         });
  //       })
  //       .catch((error) => {
  //         self.loading = false;
  //       });
  //   },

  // 获取退款信息
  const getStoreRefundInfo = async () => {
    loading.value = true;
    try {
      const data = await OrderOldApi.getStoreRefund(
        {
          sale_bill_uuid: form.order_id,
          sale_order_uuid: props.sub_order_id,
        },
        true
      );
      loading.value = false;
      form.pay_price = proxy.$priceTwo(data.data.can_return_amount);
      pay_list.value = data.data.payment_records;
      (data.data.products || []).map((item) => {
        tableData.value.push({
          type: 3,
          product_name_text: item.locale_name[language.value], // 名称
          product_attr: item.locale_attribute_name[language.value], // 规格
          total_num: item.num, // 可退数量
          total_price: item.can_return_amount, // 可退金额
          price: item.price, // 商品单价
          id: item.sale_order_product_uuid, // id
          refund_num_updata: 0, // 提交退款的数量
        });
      });
    } catch (error) {
      loading.value = false;
    }
  };

  // 关闭弹窗
  const dialogFormVisible = (e) => {
    if (e) {
      emit('closeDialog', {
        type: 'success',
        openDialog: false,
      });
    } else {
      emit('closeDialog', {
        type: 'error',
        openDialog: false,
      });
    }
  };

  const handleClick = async () => {
    try {
      const valid = await proxy.$refs.bankForm.validate();
      if (valid) {
        submit();
      }
    } catch (error) {
      // 处理验证失败
    }
  };

  const dialogFormVisible2 = () => {
    dialogVisible2.value = false;
    Object.assign(bankForm, {
      bank_code: '',
      account_no: '',
      account_name: '',
    });
  };

  // 生命周期 - 组件创建时
  onMounted(() => {
    dialogVisible.value = props.open_edit;
    form.order_id = props.order_id;
    form.sub_order_id = props.sub_order_id;
    form.pay_price = proxy.$priceTwo(props.pay_price);
    language.value = languageStore()?.getLanguageKey().language.value;
    getStoreRefundInfo();
  });
</script>
<style scoped lang="scss">
  .flex {
    display: flex;
    flex-direction: column;
  }
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
