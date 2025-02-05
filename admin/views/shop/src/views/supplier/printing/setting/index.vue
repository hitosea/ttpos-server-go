<template>
  <!--

      时间：2019-10-26
      描述：设置-小票打印设置
    -->
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" label-position="top" label-width="200px">
      <el-form-item for="no_click" :label="$t('打印语言（送厨）')" class="cashier-item" prop="kitchen_language" :rules="[{ required: true, message: ' ' }]">
        <el-select class="max-w460" v-model="form.kitchen_language" :placeholder="$t('请选择')">
          <el-option v-for="(item, index) in langList" :key="index" :label="item.value" :value="item.key"> </el-option>
        </el-select>
      </el-form-item>
      <div class="cashier-desc">{{ $t('送厨小票显示的语言将根据选择的语言进行打印') }}</div>

      <el-form-item for="no_click" :label="$t('打印方式（送厨）')" class="cashier-item" prop="kitchen_print_method" :rules="[{ required: true, message: ' ' }]">
        <el-select class="max-w460" v-model="form.kitchen_print_method" :placeholder="$t('请选择')">
          <el-option v-for="(item, index) in printMethodList" :key="index" :label="item.name" :value="item.key"> </el-option>
        </el-select>
      </el-form-item>
      <div class="cashier-desc">{{ $t('图片打印可打印图片（如LOGO）或打印机不支持的语言') }}</div>

      <el-form-item for="no_click" :label="$t('打印语言（收银）')" class="cashier-item" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.language_method">
          <el-radio label="1">{{ $t('单语言') }}</el-radio>
          <el-radio label="2">{{ $t('多语言') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="cashier-desc"></div>

      <el-form-item for="no_click" :label="$t('打印语言（收银）')" class="cashier-item" prop="default_language" :rules="[{ required: true, message: ' ' }]">
        <el-select class="max-w460" v-model="form.default_language" :placeholder="$t('请选择')">
          <el-option v-for="(item, index) in langList" :key="index" :label="item.value" :value="item.key"> </el-option>
        </el-select>
      </el-form-item>
      <div class="cashier-desc">{{
        $t('交班单、营业数据、预结账单、结账单、发票、充值单小票显示的语言将根据选择的语言进行打印（开启多语言时，预结账单/结账单/发票/充值单支持多语言选择）')
      }}</div>

      <el-form-item for="no_click" :label="$t('打印方式（收银）')" class="cashier-item" prop="print_method" :rules="[{ required: true, message: ' ' }]">
        <el-select class="max-w460" v-model="form.print_method" :placeholder="$t('请选择')">
          <el-option v-for="(item, index) in printMethodList" :key="index" :label="item.name" :value="item.key"> </el-option>
        </el-select>
      </el-form-item>
      <div class="cashier-desc">{{ $t('图片打印可打印图片（如LOGO）或打印机不支持的语言') }}</div>

      <el-form-item for="no_click" :label="$t('消费税')" class="cashier-item" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.consumption_tax">
          <el-radio label="1">{{ $t('显示全部类型') }}</el-radio>
          <el-radio label="2">{{ $t('仅显示商品已含税') }}</el-radio>
          <el-radio label="3">{{ $t('仅显示商品未含税') }}</el-radio>
          <el-radio label="4">{{ $t('全部不显示') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="cashier-desc">{{ $t('仅针对发票/预结账单/结账单，如设置不显示，当消费税计算类型为“商品未含税”时，小票将缺少消费税的体现（小票上的计算将会有误差）') }}</div>
      <el-form-item for="no_click" :label="$t('自助餐标识')" class="cashier-item" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.buffet_sign_open">
          <el-radio label="1">{{ $t('开') }}</el-radio>
          <el-radio label="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="cashier-desc">{{ $t('开启后将在送厨小票中对自助餐商品增加标识') }}</div>
      <el-form-item for="no_click" :label="$t('货币单位')" class="cashier-item" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.monetary_unit_open">
          <el-radio label="1">{{ $t('开') }}</el-radio>
          <el-radio label="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <div class="cashier-desc">{{ $t('开启后将在小票中显示对应的货币单位') }}</div>

      <el-form-item for="no_click" :label="$t('日历')" class="cashier-item" :rules="[{ required: true, message: '' }]">
        <el-select class="max-w460" v-model="form.default_calendar" :placeholder="$t('请选择')">
          <el-option v-for="(item, index) in calendarList" :key="index" :label="item.name" :value="item.key"> </el-option>
        </el-select>
      </el-form-item>
      <div class="cashier-desc">{{ $t('开启后将在小票中对应日历') }}</div>

      <!--提交-->
      <div class="common-button-wrapper">
        <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script>
  import SettingApi from '@/api/setting.js';

  export default {
    data() {
      return {
        /*切换菜单*/
        // activeIndex: '1',
        /*form表单数据*/
        form: {
          default_language: 'en',
          kitchen_language: 'en',
          buffet_sign_open: '1',
          monetary_unit_open: '1',
          language_method: '1',
          consumption_tax: '1',
          default_calendar: 1,
          print_method: '1',
          kitchen_print_method: '1',
        },
        checked: false,
        langList: [],
        calendarList: [],
        loading: false,
        printMethodList: [
          { key: '1', name: $t('文本打印') },
          { key: '2', name: $t('图片打印') },
        ],
      };
    },
    created() {
      this.getData();
    },

    methods: {
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.printingDetail({}, true)
          .then((data) => {
            self.loading = false;
            self.form.default_language = '' + data.data.values.default_language;
            self.form.kitchen_language = '' + data.data.values.kitchen_language;
            self.form.buffet_sign_open = '' + data.data.values.buffet_sign_open;
            self.form.monetary_unit_open = '' + data.data.values.monetary_unit_open;
            self.form.consumption_tax = '' + data.data.values.consumption_tax;
            self.form.default_calendar = '' + data.data.values.default_calendar;

            self.form.print_method = '' + data.data.values.print_method;
            self.form.kitchen_print_method = '' + data.data.values.kitchen_print_method;

            self.form.language_method = '' + data.data.values.language_method;

            self.langList = data.data.values.language_list;
            self.calendarList = data.data.values.calendar_list;
            self.printMethodList = data.data.values.print_list;
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      //提交表单
      onSubmit() {
        let self = this;
        let params = this.form;
        self.loading = true;
        SettingApi.editPrinting(params, true)
          .then((data) => {
            self.loading = false;
            this.$ElMessage({
              message: $t('保存成功'),
              type: 'success',
            });
            // self.$router.push('/setting/Printing');
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      //监听复选框选中
      handleCheckedCitiesChange(e) {
        let self = this;
        if (e) {
          self.form.order_status[0] = 20;
        } else {
          self.form.order_status = [];
        }
      },
    },
  };
</script>

<style scoped>
  .tips {
    color: #ccc;
  }

  .cashier-item {
    margin-bottom: 0 !important;
  }

  .cashier-desc {
    font-size: 14px;
    color: #ccc;
    margin-bottom: 20px;
    margin-top: 4px;
  }
</style>
