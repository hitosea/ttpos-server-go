<template>
  <div class="supplier" v-loading="loading">
    <el-form size="small" ref="form" :model="form" label-position="top" :rules="formRules">
      <el-form-item :label="$t('服务费')" prop="is_open">
        <div>
          <el-radio v-model="form.is_open" :label="'1'">{{ $t('开启') }}</el-radio>
          <el-radio v-model="form.is_open" :label="'0'">{{ $t('关闭') }}</el-radio>
        </div>
      </el-form-item>
      <template v-if="form.is_open == '1'">
        <el-form-item :label="$t('计算方式')" prop="charge_type">
          <div>
            <el-radio v-model="form.charge_type" :label="'1'">{{ $t('按固定金额') }}</el-radio>
            <el-radio v-model="form.charge_type" :label="'2'">{{ $t('按比例') }}</el-radio>
          </div>
        </el-form-item>

        <template v-if="form.charge_type == '1'">
          <el-form-item :label="$t('金额')" prop="service_charge">
            <span v-if="currency.unit_position == '0'">{{ currency.unit }}</span>
            <numInput class="max-w460" :min="0" :precision="2" v-model:valueData="form.service_charge" :value="form.service_charge" :placeholder="$t('请输入')"></numInput>
            <span v-if="currency.unit_position == '1'">{{ currency.unit }}</span>
            <div class="tips">{{ $t('收银/桌台订单所需要增加的服务费') }}</div>
          </el-form-item>
        </template>
        <template v-if="form.charge_type == '2'">
          <el-form-item prop="service_charge_rate">
            <numInput
              class="max-w460"
              :min="0"
              :max="100"
              :precision="2"
              v-model:valueData="form.service_charge_rate"
              :value="form.service_charge_rate"
              :placeholder="$t('请输入')"
            ></numInput>
            <span>%</span>
            <div class="tips">{{ $t('收银/桌台订单所需要增加的服务费') }}</div>
          </el-form-item>
          <el-form-item :label="$t('税费')" prop="is_open_tax" :rules="[{ required: true, message: $t('请选择是否收取税费') }]">
            <div>
              <el-radio v-model="form.is_open_tax" :label="'1'">{{ $t('收取税费') }}</el-radio>
              <el-radio v-model="form.is_open_tax" :label="'0'">{{ $t('不收取税费') }}</el-radio>
            </div>
            <div class="tips">{{ $t('如选择收取税费，需要开启VAT方可收取税费') }}</div>
          </el-form-item>
        </template>
      </template>
    </el-form>
    <!--提交-->
    <div class="common-button-wrapper">
      <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
      <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
  </div>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import { useUserStore } from '@/store';
  const { currency } = useUserStore();
  export default {
    data() {
      return {
        currency: currency,
        loading: false,
        form: {
          is_open: '1',
          charge_type: '1',
          service_charge: 0,
          service_charge_rate: 0,
          is_open_tax: '0',
        },
        formRules: {
          is_open: [
            {
              required: true,
              message: $t('请选择是否开启'),
              trigger: 'blur',
            },
          ],

          charge_type: [
            {
              required: true,
              message: $t('请输入计算方式'),
              trigger: 'blur',
            },
          ],

          service_charge: [
            {
              required: true,
              message: $t('请输入数字'),
              trigger: 'blur',
            },
          ],
        },
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
        SettingApi.getServiceCharge({}, true)
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.form.charge_type = (data.data.vars.values.charge_type ?? 1).toString();
            self.form.service_charge = Number(self.form.service_charge) || 0;
            self.form.service_charge_rate = Number(self.form.service_charge_rate) || 0;
            self.form.is_open = (data.data.vars.values.is_open ?? 0).toString();
            self.form.is_open_tax = (data.data.vars.values.is_open_tax ?? 0).toString();
            self.$refs.form.validate();
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      onSubmit() {
        let self = this;
        let params = JSON.parse(JSON.stringify(self.form));
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            SettingApi.setServiceCharge(params, true)
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
      numChange() {
        this.$nextTick(() => {
          this.form.service_charge = Number(this.$priceTwo(this.form.service_charge));
        });
      },
    },
  };
</script>
<style lang=""></style>
