<template>
  <div class="supplier" v-loading="loading">
    <el-form size="small" ref="form" :model="form" label-position="top" :rules="formRules">
      <el-form-item :label="$t('主货币单位(系统显示)')" prop="unit">
        <el-input class="max-w460" v-model="form.unit" :maxlength="50" :placeholder="$t('请输入')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('主货币单位(打印显示)')" prop="print_unit">
        <el-input class="max-w460" v-model="form.print_unit" :maxlength="50" :placeholder="$t('请输入')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('主货币显示位置')" prop="unit_position">
        <div>
          <el-radio v-model="form.unit_position" label="0">{{ $t('金额前') }}</el-radio>
          <el-radio v-model="form.unit_position" label="1">{{ $t('金额后') }}</el-radio>
        </div>
      </el-form-item>
      <el-form-item :label="$t('副货币单位')" prop="is_open">
        <div>
          <el-radio v-model="form.is_open" label="1">{{ $t('开启') }}</el-radio>
          <el-radio v-model="form.is_open" label="0">{{ $t('关闭') }}</el-radio>
        </div>
      </el-form-item>
      <el-form-item v-if="form.is_open == '1'" :label="$t('副货币单位')" prop="vice_unit">
        <el-input class="max-w460" v-model="form.vice_unit" :placeholder="$t('请输入')" :maxlength="50"></el-input>
      </el-form-item>
      <el-form-item v-if="form.is_open == '1'" :label="$t('副货币汇率')" prop="unit_rate">
        <el-input-number class="max-w460" :controls="false"  :min="0" :placeholder="$t('请输入')" v-model.number="form.unit_rate"></el-input-number>
      </el-form-item>
      <el-form-item v-if="form.is_open == '1'" :label="$t('副货币显示位置')" prop="vice_unit_position">
        <div>
          <el-radio v-model="form.vice_unit_position" label="0">{{ $t('金额前') }}</el-radio>
          <el-radio v-model="form.vice_unit_position" label="1">{{ $t('金额后') }}</el-radio>
        </div>
      </el-form-item>
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

  import { nextTick } from 'vue';
  import { useUserStore } from '@/store';
  import { setStorage } from '@/utils/storageData';

  const { currency } = useUserStore();
  export default {
    data() {
      return {
        currency: currency,
        loading: false,
        form: {
          unit: '',
          print_unit: '',
          unit_position: '0',
          is_open: '1',
          vice_unit: '',
          unit_rate: null,
          vice_unit_position: '0',
        },
        formRules: {
          unit: [
            {
              required: true,
              message: $t('请输入主货币单位(系统显示)'),
              trigger: 'blur',
            },
          ],
          print_unit: [
            {
              required: true,
              message: $t('请输入主货币单位(打印显示)'),
              trigger: 'blur',
            },
          ],
          unit_position: [
            {
              required: true,
              message: $t('主货币显示位置'),
              trigger: 'blur',
            },
          ],
          is_open: [
            {
              required: true,
              message: $t('请输入主货币单位'),
              trigger: 'blur',
            },
          ],
          vice_unit: [
            {
              required: true,
              message: $t('请输入副货币单位'),
              trigger: 'blur',
            },
          ],
          unit_rate: [
            {
              required: true,
              message: $t('请输入货币汇率'),
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
        SettingApi.getCurrencyUnit({}, true)
          .then((data) => {
            self.loading = false;
            self.form = data.data.vars.values;
            self.form.unit_rate = Number(self.form.unit_rate);
            self.form.is_open = data.data.vars.values.is_open.toString();
            nextTick(() => {
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
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            SettingApi.setCurrencyUnit(params, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });

                self.currency.is_open = params.is_open;
                self.currency.unit = params.unit;
                self.currency.unit_position = params.unit_position;

                self.currency.vices.unit_rate = params.unit_rate;
                self.currency.vices.vice_unit = params.vice_unit;
                self.currency.vices.vice_unit_position = params.vice_unit_position;

                setStorage(JSON.stringify(self.currency), 'currency');
                self.dialogFormVisible(true);
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },
    },
  };
</script>
