<template>
  <el-dialog class="product-add" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('添加打印机')">
    <!--form表单-->
    <el-form size="small" ref="formRef" :model="form" label-position="top">
      <!--添加门店-->
      <el-form-item
        for="no_click"
        :label="$t('打印机名称')"
        prop="printer_name"
        :rules="[
          { required: true, message: $t('请输入打印机名称') },
          {
            validator: uniqueNameValidator('printer', undefined, 'SINGLE'),
            trigger: 'blur',
          },
        ]"
      >
        <el-input v-model="form.printer_name" :maxlength="50" :placeholder="$t('请输入打印机名称')"></el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('打印机类型')" prop="printer_type" :rules="[{ required: true, message: ' ' }]">
        <el-select v-model="form.printer_type" :placeholder="$t('请选择类型')" style="width: 100%">
          <el-option v-for="(item, index) in type" :key="index" :label="item" :value="index"> </el-option>
        </el-select>
      </el-form-item>

      <!-- 飞鹅打印机 -->
      <div v-if="form.printer_type == 'FEI_E_YUN'">
        <el-form-item for="no_click" label="USER" prop="FEI_E_YUN.USER" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN.USER"></el-input>
          <div class="tips">{{ $t('飞鹅云后台注册用户名') }}</div>
        </el-form-item>

        <el-form-item for="no_click" label="UKEY" prop="FEI_E_YUN.UKEY" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN.UKEY"></el-input>
          <div class="tips">{{ $t('飞鹅云后台登录生成的UKEY') }}</div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机编号')" prop="FEI_E_YUN.SN" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN.SN"></el-input>
          <div class="tips">{{ $t('打印机编号为9位数字，查看飞鹅打印机底部贴纸上面的编号') }}</div>
        </el-form-item>
      </div>

      <!-- 飞鹅打印机 -->
      <div v-if="form.printer_type == 'FEI_E_YUN_TAG'">
        <el-form-item for="no_click" label="USER" prop="FEI_E_YUN_TAG.USER" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN_TAG.USER"></el-input>
          <div class="tips">{{ $t('飞鹅云后台注册用户名') }}</div>
        </el-form-item>

        <el-form-item for="no_click" label="UKEY" prop="FEI_E_YUN_TAG.UKEY" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN_TAG.UKEY"></el-input>
          <div class="tips">{{ $t('飞鹅云后台登录生成的UKEY') }}</div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机编号')" prop="FEI_E_YUN_TAG.SN" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.FEI_E_YUN_TAG.SN"></el-input>
          <div class="tips">{{ $t('打印机编号为9位数字，查看飞鹅打印机底部贴纸上面的编号') }}</div>
        </el-form-item>
      </div>

      <!-- 365云打印 -->
      <div v-if="form.printer_type == 'PRINT_CENTER'">
        <el-form-item for="no_click" :label="$t('打印机编号')" prop="PRINT_CENTER.deviceNo" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.PRINT_CENTER.deviceNo"></el-input>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机秘钥')" prop="PRINT_CENTER.key" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.PRINT_CENTER.key"></el-input>
        </el-form-item>
      </div>

      <!-- 商米打印 -->
      <div v-if="form.printer_type == 'SUNMI_LAN'">
        <el-form-item for="no_click" :label="$t('打印机IP')" prop="SUNMI_LAN.IP" :rules="[{ required: true, message: ' ' }]">
          <el-autocomplete
            v-model="form.SUNMI_LAN.IP"
            :fetch-suggestions="getIPSuggestions"
            clearable
            :placeholder="$t('请选择或输入IP地址')"
            style="width: 100%">
            <template #default="{ item }">
              <div>{{ item.label }}</div>
            </template>
          </el-autocomplete>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机SN')" prop="SUNMI_LAN.SN" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.SUNMI_LAN.SN"></el-input>
        </el-form-item>
      </div>

      <!-- 商米云打印 -->
      <div v-if="form.printer_type == 'SUNMI_CLOUD'">
        <el-form-item for="no_click" :label="$t('打印机APPID')" prop="SUNMI_CLOUD.APP_ID" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.SUNMI_CLOUD.APP_ID"></el-input>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机APPKEY')" prop="SUNMI_CLOUD.APP_KEY" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.SUNMI_CLOUD.APP_KEY"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('打印机SN')" prop="SUNMI_CLOUD.SN" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.SUNMI_CLOUD.SN"></el-input>
        </el-form-item>
      </div>

      <!-- 佳博云打印 -->
      <div v-if="form.printer_type == 'GP_CLOUD'">
        <el-form-item for="no_click" :label="$t('打印机APPID')" prop="GP_CLOUD.APP_ID" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.GP_CLOUD.APP_ID"></el-input>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机APPKEY')" prop="GP_CLOUD.APP_KEY" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.GP_CLOUD.APP_KEY"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('打印机SN')" prop="GP_CLOUD.SN" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.GP_CLOUD.SN"></el-input>
        </el-form-item>
      </div>

      <!-- 芯烨打印 -->
      <div v-if="form.printer_type == 'XPRINTER_LAN' || form.printer_type == 'XPRINTER_WIFI'">
        <el-form-item for="no_click" :label="$t('打印机IP')" prop="XPRINTER_LAN.IP" :rules="[{ required: true, message: ' ' }]">
          <el-autocomplete
            v-model="form.XPRINTER_LAN.IP"
            :fetch-suggestions="getIPSuggestions"
            clearable
            :placeholder="$t('请选择或输入IP地址')"
            style="width: 100%">
            <template #default="{ item }">
              <div>{{ item.label }}</div>
            </template>
          </el-autocomplete>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机PORT')" prop="XPRINTER_LAN.PORT" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.XPRINTER_LAN.PORT"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('打印机SN')">
          <el-input v-model="form.sn"></el-input>
        </el-form-item>
      </div>

      <!-- CODESOFT打印 -->
      <div v-if="form.printer_type == 'CODESOFT_LAN' || form.printer_type == 'CODESOFT_WIFI'">
        <el-form-item for="no_click" :label="$t('打印机IP')" prop="CODESOFT_LAN.IP" :rules="[{ required: true, message: ' ' }]">
          <el-autocomplete
            v-model="form.CODESOFT_LAN.IP"
            :fetch-suggestions="getIPSuggestions"
            clearable
            :placeholder="$t('请选择或输入IP地址')"
            style="width: 100%">
            <template #default="{ item }">
              <div>{{ item.label }}</div>
            </template>
          </el-autocomplete>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印机PORT')" prop="CODESOFT_LAN.PORT" :rules="[{ required: true, message: ' ' }]">
          <el-input v-model="form.CODESOFT_LAN.PORT"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('打印机SN')">
          <el-input v-model="form.sn"></el-input>
        </el-form-item>
      </div>

      <!-- 排序、打印联数、纸张宽度、状态检查合并为一行 -->
      <el-row :gutter="20">
        <el-col :span="6">
          <el-form-item for="no_click" :label="$t('排序')" prop="sort" :rules="[{ required: true, message: $t('接近0，排序等级越高') }]">
            <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="form.sort" autocomplete="off" style="width: 100%"></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item for="no_click" :label="$t('打印联数')" prop="print_times" :rules="[{ required: true, message: $t('请输入打印联数') }]">
            <el-input-number
              :controls="false"
              :min="1"
              :max="10"
              :precision="0"
              :placeholder="$t('请输入打印联数')"
              v-model.number="form.print_times"
              autocomplete="off"
              style="width: 100%"
            ></el-input-number>
            <div class="tips">{{ $t('同一订单，打印的次数') }}</div>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item for="no_click" :label="$t('纸张宽度（mm）')" prop="width">
            <el-select v-model="form.width" :placeholder="$t('请选择纸张宽度')" style="width: 100%" clearable>
              <el-option :label="'58mm'" :value="58"></el-option>
              <el-option :label="'80mm'" :value="80"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item for="no_click" :label="$t('状态检查')" prop="enable_status_check">
            <div style="text-align: center;">
              <el-switch
                v-model="form.enable_status_check"
                :active-value="1"
                :inactive-value="0"
              />
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item for="no_click" :label="$t('打印关联收银机')" prop="is_main" v-if="is_usb != 1">
        <el-select v-model="form.source_device_sn" :placeholder="$t('选择打印机所关联的收银机')" style="width: 100%" clearable>
          <el-option v-for="(item, index) in cashierList" :key="index" :label="item.cashier_name" :value="item.cashier_key"> </el-option>
        </el-select>
      </el-form-item>

      <el-form-item for="no_click" :label="$t('打印方式')" prop="print_method">
        <el-select v-model="form.print_method" :placeholder="$t('选择打印方式')" style="width: 100%" clearable>
          <el-option :label="$t('文本打印')" :value="1"></el-option>
          <el-option :label="$t('图片打印')" :value="2"></el-option>
        </el-select>
      </el-form-item>
      <!--提交-->
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
  import SettingApi from '@/api/setting.js';
  import { uniqueNameValidator } from '@/utils/form.js';

  export default {
    name: 'SupplierPrintingPrinterAdd',
    props: ['open_add'],
    created() {
      this.dialogVisible = this.open_add;
      this.getData();
    },
    data() {
      return {
        /*切换菜单*/
        // activeIndex: '1',
        /*form表单数据*/
        form: {
          printer_name: '',
          printer_type: '',
          print_method: '',
          sort: null,
          print_times: 1,
          width: 80, // 纸张宽度，默认80mm
          enable_status_check: 1, // 是否启用状态检查，默认开启
          source_device_sn: '',
          sn: '',
          FEI_E_YUN: {
            USER: '',
            UKEY: '',
            SN: '',
          },
          FEI_E_YUN_TAG: {
            USER: '',
            UKEY: '',
            SN: '',
          },
          PRINT_CENTER: {
            deviceNo: '',
            key: '',
          },
          SUNMI_LAN: {
            IP: '',
            SN: '',
          },
          XPRINTER_LAN: {
            IP: '',
            PORT: 9100,
          },
          SUNMI_CLOUD: {
            APP_ID: '',
            APP_KEY: '',
            SN: '',
          },
          GP_CLOUD: {
            APP_ID: '',
            APP_KEY: '',
            SN: '',
          },
          CODESOFT_LAN: {
            IP: '',
            PORT: 9100,
          },
        },
        loading: false,

        type: [],
        cashierList: [],
        dialogVisible: false,
        ipOptions: [],
      };
    },

    methods: {
      getData() {
        SettingApi.printerType({}, true)
          .then((data) => {
            this.type = data.data.printerType;
            this.cashierList = data.data.cashierList;
            this.ipOptions = data.data.lanPrinter || [];
          })
          .catch(() => {});
      },
      //提交表单
      onSubmit() {
        const self = this;
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
            if (self.form.SUNMI_LAN.IP && !ipRegex.test(self.form.SUNMI_LAN.IP)) {
              self.$ElMessage({
                message: self.$t('请输入正确IP地址'),
                type: 'error',
              });
              return;
            }
            if (self.form.XPRINTER_LAN.IP && !ipRegex.test(self.form.XPRINTER_LAN.IP)) {
              self.$ElMessage({
                message: self.$t('请输入正确IP地址'),
                type: 'error',
              });
              return;
            }
            self.loading = true;
            self.form.source_device_sn = self.form.source_device_sn || '';
            self.form.print_method = self.form.print_method || 0;
            SettingApi.addPrinter(self.form, true)
              .then(() => {
                self.loading = false;
                self.$ElMessage({
                  message: self.$t('添加成功'),
                  type: 'success',
                });
                self.$emit('close', 1);
              })
              .catch(() => {
                self.loading = false;
              });
          }
        });
      },

      handleClose() {
        this.$emit('close');
      },

      uniqueNameValidator: uniqueNameValidator,

      getIPSuggestions(queryString, cb) {
        const suggestions = this.ipOptions;
        const results = queryString ? suggestions.filter(item => item.value.indexOf(queryString) === 0) : suggestions;
        cb(results);
      },
    },
  };
</script>

<style scoped>
  .tips {
    color: #ccc;
  }
</style>
