<template>
  <div class="supplier" v-loading="loading">
    <el-form class="form-box" size="small" ref="form" :model="form" label-position="top" :rules="formRules">
      <el-form-item for="no_click" :label="$t('优惠折扣自动抹零方式')" prop="zeroing_method">
        <el-select v-model="form.zeroing_method">
          <template v-for="(item, index) in zeroingMethods" :key="index">
            <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
          </template>
        </el-select>
        <el-tooltip effect="dark" placement="bottom">
          <template #content>
            <p>{{ $t('抹分：例268.25→268.2') }}</p>
            <p>{{ $t('抹角：例268.25→268') }}</p>
            <p>{{ $t('四舍五入保留一位小数：例268.25→268.3，268.54→268.5') }}</p>
            <p>{{ $t('四舍五入到整数：例268.25→268，268.54→269') }}</p>
          </template>
          <SvgIcon class="tip-icon" name="icon6"></SvgIcon>
        </el-tooltip>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('结账自动抹零方式')" prop="checkout_zeroing_method">
        <el-select v-model="form.checkout_zeroing_method">
          <template v-for="(item, index) in checkoutZeroingMethod" :key="index">
            <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
          </template>
        </el-select>
        <el-tooltip effect="dark" placement="bottom">
          <template #content>
            <p>{{ $t('抹分：例268.25→268.2') }}</p>
            <p>{{ $t('抹角：例268.25→268') }}</p>
            <p>{{ $t('抹元：例268.25→260') }}</p>
          </template>
          <SvgIcon class="tip-icon" name="icon6"></SvgIcon>
        </el-tooltip>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('赠菜计算方式')" prop="gift_method">
        <div>
          <el-select v-model="form.gift_method">
            <template v-for="(item, index) in giftMethods" :key="index">
              <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
            </template>
          </el-select>
        </div>
        <el-tooltip effect="dark" placement="bottom">
          <template #content>
            <p>{{ $t('计入总销售额、优惠折扣：赠菜的总金额将纳入总销售额、优惠折扣') }}</p>
            <p>{{ $t('不计入总销售额、优惠折扣：赠菜的总金额将不纳入总销售额、优惠折扣') }}</p>
          </template>
          <SvgIcon class="tip-icon" name="icon6"></SvgIcon>
        </el-tooltip>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('免单计算方式')" prop="free_method">
        <div>
          <el-select v-model="form.free_method">
            <template v-for="(item, index) in freeMethods" :key="index">
              <el-option :value="item.key" :label="item.label">{{ item.label }}</el-option>
            </template>
          </el-select>
        </div>
        <el-tooltip effect="dark" placement="bottom">
          <template #content>
            <p>{{ $t('计入总销售额、优惠折扣：免单总金额将纳入总销售额、优惠折扣、服务费、税费') }}</p>
            <p>{{ $t('不计入总销售额、优惠折扣：免单的总金额将不纳入总销售额、优惠折扣、服务费、税费 ') }}</p>
          </template>
          <SvgIcon class="tip-icon" name="icon6"></SvgIcon>
        </el-tooltip>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('开票信息')" prop="is_invoice" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_invoice">
          <el-radio :label="0">{{ $t('不需要填写') }}</el-radio>
          <el-radio :label="1">{{ $t('需要填写') }}</el-radio>
        </el-radio-group>
        <div class="tips">
          {{ $t('如需要填写，则在收银机打印发票时需要填写对应的开票信息') }}
        </div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('折扣计算方式')" prop="discount_method" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.discount_method">
          <el-radio :label="10">{{ $t('百分比打折') }}</el-radio>
          <el-radio :label="20">{{ $t('直接减免') }}</el-radio>
        </el-radio-group>
        <div class="tips line-24">
          {{ $t('百分比打折：如果订单原价为100，打8折，即表示消费者需要支付商品价格的80%，计算方式为：100 × 80% = 80') }}<br />
          {{ $t('直接减免：如果订单原价为100，20% OFF，表示从原价中减去20%的价格。计算方式为：100 - (100 × 20%) = 80') }}
        </div>
      </el-form-item>

      <el-form-item :label="$t('电子菜单')">
        <el-button @click="downloadFile('menu')" type="primary">{{ $t('二维码') }}</el-button>
        <div class="tips">
          {{ $t('仅用于显示商品，无法点餐下单') }}
        </div>
      </el-form-item>

      <el-form-item :label="$t('会员端')" v-auth="'/card/user/index'">
        <el-button @click="downloadFile('member')" type="primary">{{ $t('二维码') }}</el-button>
        <p class="copy-link" @click="handleCopyLink">{{ $t('复制链接') }}</p>
      </el-form-item>

      <el-form-item :label="$t('外送商品价格')" v-if="showDelivery" v-auth="'/card/user/index'">
        <el-input-number
          class="max-w320"
          v-model="form.delivery_price_ratio"
          :controls="false"
          :min="0"
          :max="300"
          :precision="0"
          :placeholder="$t('请输入外送商品价格')"
        ></el-input-number>
        <p class="ml8"> % {{ $t('商品原价') }} </p>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('结账后不清台')" prop="no_clear_table" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.no_clear_table">
          <el-radio :label="0">{{ $t('清台') }}</el-radio>
          <el-radio :label="1">{{ $t('不清台') }}</el-radio>
        </el-radio-group>
        <div class="tips">
          {{ $t('注：开启后结账不自动清台，结账后平板/H5再无法下单（对收银机/点餐助手）') }}
        </div>
      </el-form-item>
      <el-form-item :label="$t('免单/赠菜原因')">
        <el-button @click="setFreeReason()" type="primary">{{ $t('管理') }} ({{ freeTagCount }})</el-button>
      </el-form-item>
      <el-form-item :label="$t('退菜原因')">
        <el-button @click="setRefundReason()" type="primary">{{ $t('管理') }} ({{ returnReasonCount }})</el-button>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('整单备注')">
        <el-button @click="setOrderRemark()" type="primary">{{ $t('管理') }} ({{ orderRemarkCount }})</el-button>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('取消订单/退菜')" prop="is_need_password" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_need_password">
          <el-radio :label="1">{{ $t('需要密码') }}</el-radio>
          <el-radio :label="0">{{ $t('无需密码') }}</el-radio>
        </el-radio-group>
        <div class="tips">
          {{ $t('涉及到收银机、点餐助手，密码则为对应端的高级密码') }}
        </div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('菜品卡片样式')" prop="dish_card_style" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.dish_card_style">
          <el-radio :label="0">{{ $t('无图模式') }}</el-radio>
          <el-radio :label="1">{{ $t('图片模式') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('营业时间')" prop="opening_hours" :rules="[{ required: true, message: $t('请选择营业时间') }]">
        <TimePicker v-model="form.opening_hours" @update:modelValue="updateOpeningHours" />
      </el-form-item>
      <el-form-item for="no_click" :label="$t('起始流水号')" prop="start_serial_no" :rules="[{ required: true, message: '' }]" class="start-serial-no">
        <el-input v-model="form.start_serial_no" :placeholder="$t('请输入起始流水号')" class="max-w400" @input="form.start_serial_no = form.start_serial_no.replace(/[^0-9]/g, '')" />
      </el-form-item>
    </el-form>
    <!--提交-->
    <div class="common-button-wrapper">
      <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
      <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
  </div>
  <ManageFreeReason
    v-if="openFreeReasonDialog"
    :open="openFreeReasonDialog"
    @close="
      (refresh) => {
        openFreeReasonDialog = false;
        if (refresh) {
          this.getData();
        }
      }
    "
  >
  </ManageFreeReason>

  <ManageRefundReason
    v-if="openRefundReasonDialog"
    :open="openRefundReasonDialog"
    @close="
      (refresh) => {
        openRefundReasonDialog = false;
        if (refresh) {
          this.getData();
        }
      }
    "
  >
  </ManageRefundReason>

  <ManageOrderRemark
    v-if="openOrderRemarkDialog"
    :open="openOrderRemarkDialog"
    @close="
      (refresh) => {
        openOrderRemarkDialog = false;
        if (refresh) {
          this.getData();
        }
      }
    "
  >
  </ManageOrderRemark>

  <Qrcode :open="isQrcode" @close="closeQrcode" :type="qrcodeType"></Qrcode>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import { useUserStore } from '@/store';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import ManageFreeReason from './ManageFreeReason.vue';
  import ManageRefundReason from './ManageRefundReason.vue';
  import ManageOrderRemark from './ManageOrderRemark.vue';
  import Qrcode from './Qrcode.vue';
  import TimePicker from '@/components/time-picker/index.vue';

  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const showDelivery = (supplier.value?.delivery_status || 0) == 1;
  const { currency } = useUserStore();
  export default {
    components: {
      SvgIcon,
      ManageFreeReason,
      ManageRefundReason,
      ManageOrderRemark,
      Qrcode,
      TimePicker,
    },
    data() {
      return {
        showDelivery: showDelivery,
        currency: currency,
        loading: false,
        openFreeReasonDialog: false,
        openRefundReasonDialog: false,
        openOrderRemarkDialog: false,
        isQrcode: false,
        qrcodeType: '',
        form: {
          zeroing_method: 0,
          checkout_zeroing_method: 0,
          gift_method: 10,
          free_method: 10,
          is_invoice: 0,
          discount_method: 10,
          no_clear_table: 1,
          is_need_password: 1,
          dish_card_style: 1,
          opening_hours: '',
          delivery_price_ratio: 100,
          start_serial_no: '0001',
        },
        company_link: '',
        formRules: {
          zeroing_method: [
            {
              required: true,
              message: $t('请选择'),
              trigger: 'blur',
            },
          ],
          checkout_zeroing_method: [
            {
              required: true,
              message: $t('请选择'),
              trigger: 'blur',
            },
          ],

          gift_method: [
            {
              required: true,
              message: $t('请选择'),
              trigger: 'blur',
            },
          ],

          free_method: [
            {
              required: true,
              message: $t('请选择'),
              trigger: 'blur',
            },
          ],
        },
        checkoutZeroingMethod: [
          {
            key: 0,
            label: $t('实款实收'),
          },
          {
            key: 1,
            label: $t('抹分'),
          },
          {
            key: 2,
            label: $t('抹角'),
          },
          {
            key: 5,
            label: $t('抹元'),
          },
        ],
        zeroingMethods: [
          {
            key: 0,
            label: $t('实款实收'),
          },
          {
            key: 1,
            label: $t('抹分'),
          },
          {
            key: 2,
            label: $t('抹角'),
          },
          {
            key: 3,
            label: $t('四舍五入保留一位小数'),
          },
          {
            key: 4,
            label: $t('四舍五入到整数'),
          },
        ],
        giftMethods: [
          {
            key: 10,
            label: $t('计入总销售额、优惠折扣'),
          },
          {
            key: 20,
            label: $t('不计入总销售额、优惠折扣'),
          },
        ],
        freeMethods: [
          {
            key: 10,
            label: $t('计入总销售额、优惠折扣、服务费、税费'),
          },
          {
            key: 20,
            label: $t('不计入总销售额、优惠折扣、服务费、税费'),
          },
        ],
        freeTagCount: 0,
        returnReasonCount: 0,
        orderRemarkCount: 0,
      };
    },
    created() {
      this.getData();
    },
    methods: {
      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.getBusiness({}, true)
          .then((data) => {
            self.loading = false;

            self.form.free_method = Number(data.data.vars.values.free_method) || 10;
            if (Array.isArray(data.data.vars.values.free_method_list)) {
              self.freeMethods = data.data.vars.values.free_method_list.map((item) => ({ key: Number(item.key), label: item.name }));
            }

            self.form.gift_method = Number(data.data.vars.values.gift_method) || 10;
            if (Array.isArray(data.data.vars.values.gift_method_list)) {
              self.giftMethods = data.data.vars.values.gift_method_list.map((item) => ({ key: Number(item.key), label: item.name }));
            }

            self.form.zeroing_method = Number(data.data.vars.values.zeroing_method) || 0;
            if (Array.isArray(data.data.vars.values.zeroing_method_list)) {
              self.zeroingMethods = data.data.vars.values.zeroing_method_list.map((item) => ({ key: Number(item.key), label: item.name }));
            }

            self.form.checkout_zeroing_method = Number(data.data.vars.values.checkout_zeroing_method) || 0;
            if (Array.isArray(data.data.vars.values.checkout_zeroing_method_list)) {
              self.checkoutZeroingMethod = data.data.vars.values.checkout_zeroing_method_list.map((item) => ({ key: Number(item.key), label: item.name }));
            }

            self.form.is_invoice = Number(data.data.vars.values.is_invoice) || 0;

            self.form.discount_method = Number(data.data.vars.values.discount_method) || 0;

            self.form.no_clear_table = Number(data.data.vars.values.no_clear_table) || 0;

            self.form.is_need_password = Number(data.data.vars.values.is_need_password) || 0;

            self.form.dish_card_style = Number(data.data.vars.values.dish_card_style) || 0;
            // 营业时间 格式：00:00-23:59 转为[00:00, 23:59]
            self.form.opening_hours = data.data.vars.values.opening_hours
              ? [data.data.vars.values.opening_hours.split('-')[0], data.data.vars.values.opening_hours.split('-')[1]]
              : [];
            self.form.start_serial_no = data.data.vars.values.start_serial_no || '0001';
            self.form.delivery_price_ratio = Number(data.data.vars.values.delivery_price_ratio) || 100;

            self.freeTagCount = Number(data.data.free_tag_count) || 0;
            self.returnReasonCount = Number(data.data.return_reason_count) || 0;
            self.orderRemarkCount = Number(data.data.order_remark_count) || 0;
            self.company_link = data.data.company_link;

            self.$nextTick(() => {
              self.$refs.form.validate();
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      onSubmit() {
        let self = this;
        let params = JSON.parse(JSON.stringify(self.form));
        params.opening_hours = params.opening_hours.join('-');
        params.start_serial_no = params.start_serial_no.trim();
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            SettingApi.setBusiness(params, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });
                self.getData();
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },
      setFreeReason() {
        this.openFreeReasonDialog = true;
      },
      setRefundReason() {
        this.openRefundReasonDialog = true;
      },

      setOrderRemark() {
        this.openOrderRemarkDialog = true;
      },

      downloadFile(type) {
        let self = this;
        self.isQrcode = true;
        self.qrcodeType = type;
      },

      handleCopyLink() {
        this.$copyText(this.company_link).then(
          () => {
            this.$ElMessage({
              message: this.$t('复制成功'),
              type: 'success',
            });
          },
          () => {
            this.$ElMessage({
              message: this.$t('复制失败'),
              type: 'error',
            });
          }
        );
      },

      closeQrcode() {
        let self = this;
        self.isQrcode = false;
      },

      updateOpeningHours(value) {
        this.form.opening_hours = value;
      },
    },
  };
</script>
<style lang="scss" scoped>
  .supplier {
    display: flex;
    flex-direction: column;
    height: 100%;
    .form-box {
      flex: 1 1 auto;
      overflow: auto;
    }
    .common-button-wrapper {
      flex: 0 0 auto;
      flex-shrink: 1;
    }
  }
  .tip-icon {
    margin-left: 8px;
    width: 24px;
    height: 24px;
  }

  .el-form-item__content {
    .el-select--small {
      width: 320px;
    }
  }
  .line-24 {
    line-height: 24px;
    margin-top: 6px;
  }
  .copy-link {
    color: #409eff;
    cursor: pointer;
    margin-left: 16px;
  }
  .max-w320 {
    max-width: 320px;
  }
  .max-w400 {
    max-width: 400px;
  }
</style>
