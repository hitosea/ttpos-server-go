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
            <el-radio value="2">{{ $t('部分退款') }}</el-radio>
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
              <numInput
                :min="0"
                :max="Number(scope.row.total_num)"
                :precision="scope.row.num_type == 1 ? 2 : 0"
                :placeholder="$t('请输入')"
                v-model="scope.row.refund_num_updata"
                :controls="true"
              ></numInput>
            </template>
          </el-table-column>
          <el-table-column :label="$t('退款金额')" align="right">
            <template #default="scope">
              <div class="flex">
                <p>
                  <span v-if="currency.unit_position == '0'">
                    {{ currency.unit }}
                  </span>
                  {{ this.$formatPrice(Number(scope.row.refund_num_updata) * Number(scope.row.price)) }}
                  <span v-if="currency.unit_position == '1'">
                    {{ currency.unit }}
                  </span>
                </p>
                <p class="tips">
                  {{ $t('可退款金额：') }}
                  <span v-if="currency.unit_position == '0'">
                    {{ currency.unit }}
                  </span>
                  {{ this.$formatPrice(Number(scope.row.total_price)) }}
                  <span v-if="currency.unit_position == '1'">
                    {{ currency.unit }}
                  </span>
                </p>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-form-item v-if="manual_return_points" for="no_click" :label="$t('扣除积分')" :label-width="formLabelWidth">
          <div class="flex-row">
            <el-input-number
              class="flex-1"
              :min="0"
              :disabled="deductible_points == 0"
              :max="deductible_points"
              :precision="2"
              :controls="false"
              v-model="form.points"
              :placeholder="$t('请输入扣除积分')"
            />
            <span>{{ $t('可退积分') + ' ' + deductible_points }}</span>
          </div>
        </el-form-item>
      </el-form>

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

<script>
  import { useUserStore } from '@/store';
  const { currency } = useUserStore();
  import { languageStore } from '@/store/model/language';
  import OrderApi from '@/api/order.js';
  export default {
    components: {},
    data() {
      return {
        loading: false, //加载状态
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        dialogVisible: false,
        dialogVisible2: false,
        form: {
          refund_type: '1', //退款类型
          order_id: '', //订单id
          sub_order_id: '', // 子单id
          pay_price: '', //支付金额
          points: null,
        },
        deductible_points: 0, //扣除积分
        manual_return_points: false,
        tableData: [], //表格数据
        pay_list: [], //支付记录（原路退款会按照以下顺序退回）
        currency: currency, //货币
        language: '',
        bankForm: {
          bank_code: '',
          account_no: '',
          account_name: '',
        },
        bankList: [
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
        ],
      };
    },
    props: ['open_edit', 'order_id', 'sub_order_id', 'pay_price'], //父组件传递的参数
    created() {
      this.dialogVisible = this.open_edit;
      this.form.order_id = this.order_id;
      this.form.sub_order_id = this.sub_order_id;
      this.form.pay_price = this.$priceTwo(this.pay_price);
      this.language = languageStore()?.getLanguageKey().language.value;
      // this.getData();
      this.getStoreRefundInfo();
    },
    methods: {
      /*处理*/
      submit() {
        let self = this;
        let form = {
          sale_bill_uuid: this.form.order_id,
          sale_order_uuid: this.form.sub_order_id,
          refund_type: this.form.refund_type,
          points: this.form.points,
          refund_product: [],
          refund_buffet: [],
          refund_delay: [],
        };
        this.tableData.map((item) => {
          if (item.refund_num_updata > 0 && item.type == 1) {
            form.refund_buffet.push({
              id: item.id,
              refund_num: item.refund_num_updata,
            });
          }
          if (item.refund_num_updata > 0 && item.type == 2) {
            form.refund_delay.push({
              id: item.id,
              refund_num: item.refund_num_updata,
            });
          }
          if (item.refund_num_updata > 0 && item.type == 3) {
            form.refund_product.push({
              order_product_id: item.id,
              refund_num: item.refund_num_updata,
            });
          }
        });
        if (this.form.refund_type == '2' && form.refund_product.length == 0 && form.refund_buffet.length == 0 && form.refund_delay.length == 0) {
          this.$ElMessage({
            type: 'warning',
            message: $t('请选择退款商品'),
          });
          return;
        }
        if (this.dialogVisible2) {
          form.bank_code = this.bankForm.bank_code;
          form.account_no = this.bankForm.account_no;
          form.account_name = this.bankForm.account_name;
        }

        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            OrderApi.storeRefund(form, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: data.msg,
                  type: 'success',
                });
                self.dialogFormVisible(true);
                self.dialogFormVisible2(false);
              })
              .catch((error) => {
                self.loading = false;
                if (error?.data?.code == -901) {
                  self.dialogVisible2 = true;
                }
              });
          }
        });
      },
      getData() {
        let self = this;
        self.loading = true;
        OrderApi.orderProductList({ order_id: this.form.order_id })
          .then((data) => {
            self.loading = false;
            this.tableData = [];

            (data.data.buffetList || []).map((item) => {
              this.tableData.push({
                type: 1,
                product_name_text: item.buffet_name_text + `(${item.customer_type_name_text})`, //名称
                product_attr: '', //规格
                total_num: item.num, //总数量
                total_price: item.consumption_tax_pay_price * Number(item.num), //总价钱
                refund_num: item.refund_num, //已退数量
                refund_money: item.consumption_tax_pay_price * Number(item.refund_num), //已退价钱
                id: item.id, //id
                refund_num_updata: 0, //提交退款的数量
              });
            });

            (data.data.delayList || []).map((item) => {
              this.tableData.push({
                type: 2,
                product_name_text: item.name_text, //名称
                product_attr: '', //规格
                total_num: item.num, //总数量
                total_price: item.price * Number(item.num), //总价钱
                refund_num: item.refund_num, //已退数量
                refund_money: item.refund_money, //已退价钱
                id: item.id, //id
                refund_num_updata: 0, //提交退款的数量
              });
            });

            (data.data.productList || []).map((item) => {
              this.tableData.push({
                type: 3,
                product_name_text: item.product_name_text, //名称
                product_attr: item.product_attr, //规格
                total_num: item.total_num, //总数量
                total_price: item.consumption_tax_pay_price * Number(item.total_num), //总价钱
                refund_num: item.refund_num, //已退数量
                refund_money: item.consumption_tax_pay_price * Number(item.refund_num), //已退价钱
                id: item.order_product_id, //id
                refund_num_updata: 0, //提交退款的数量
                num_type: item.num_type,
              });
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      getStoreRefundInfo() {
        let self = this;
        self.loading = true;
        OrderApi.getStoreRefund(
          {
            sale_bill_uuid: this.form.order_id,
            sale_order_uuid: this.sub_order_id,
          },
          true
        )
          .then((data) => {
            self.loading = false;
            self.form.pay_price = this.$priceTwo(data.data.can_return_amount);
            self.pay_list = data.data.payment_records;
            self.deductible_points = data.data.deductible_points;
            self.manual_return_points = data.data.manual_return_points;
            (data.data.products || []).map((item) => {
              this.tableData.push({
                type: 3,
                product_name_text: item.locale_name[self.language], //名称
                product_attr: item.locale_attribute_name[self.language], //规格
                total_num: item.num, // 可退数量
                total_price: item.can_return_amount, // 可退金额
                price: item.price, // 商品单价
                id: item.sale_order_product_uuid, //id
                refund_num_updata: 0, //提交退款的数量
                num_type: item.num_type, //数量类型
              });
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      /*关闭弹窗*/
      dialogFormVisible(e) {
        if (e) {
          this.$emit('closeDialog', {
            type: 'success',
            openDialog: false,
          });
        } else {
          this.$emit('closeDialog', {
            type: 'error',
            openDialog: false,
          });
        }
      },

      handleClick() {
        this.$refs.bankForm.validate((valid) => {
          if (valid) {
            this.submit();
          }
        });
      },

      dialogFormVisible2() {
        this.dialogVisible2 = false;
        this.bankForm = {
          bank_code: '',
          account_no: '',
          account_name: '',
        };
      },
    },
  };
</script>
<style scoped lang="scss">
  .flex {
    display: flex;
    flex-direction: column;
  }
  .flex-row {
    display: flex;
    flex-direction: row;
    align-items: center;
    width: 100%;
    gap: 8px;
    .flex-1 {
      flex: 1;
      width: 100%;
    }
    span {
      flex-shrink: 0;
    }
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
