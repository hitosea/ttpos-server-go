<template>
  <!--
      
      时间：2019-10-26
      描述：设置-小票打印设置
    -->
  <div class="product-add" v-loading="loading">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" label-position="top" label-width="200px">
      <el-form-item for="no_click" :label="$t('收银打印')" :rules="[{ required: true }]">
        <div>
          <el-radio v-model="form.cashier_open" :label="'1'">{{ $t('开启') }}</el-radio>
          <el-radio v-model="form.cashier_open" :label="'0'">{{ $t('关闭') }}</el-radio>
        </div>
      </el-form-item>

      <template v-if="form.cashier_open == 1 && cashierList.length > 0" v-for="(items, indexs) in cashierList">
        <el-form-item for="no_click" :label="items.cashier_name" class="cashier-item">
          <el-select class="max-w460" v-model="items.printer_id" :placeholder="$t('请选择')">
            <el-option v-for="(item, index) in printerList" :key="index" :label="item.printer_name" :value="item.printer_id + ''"> </el-option>
          </el-select>
          <div class="max-w460 flex-center">
            <span >{{ $t('收银机SN') }}</span>
            <el-input  v-model="items.sn" :placeholder="$t('请输入收银机SN')"></el-input>
          </div>
        </el-form-item>
        <div class="cashier-desc">
          {{ $t('交班单、营业数据、预结账单、结账单、发票、充值单') }}
        </div>
      </template>
      <template v-if="form.cashier_open == 1 && cashierList.length == 0">
        {{ $t('暂无绑定收银机') }}
      </template>
      <!--提交-->
      <div class="common-button-wrapper">
        <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
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
          cashier_open: '0',
          cashier_printer: [],
        },
        checked: false,
        printerList: [],
        cashierList: [],
        loading: false,
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
            let vars = data.data.values;
            self.form.cashier_open = vars.cashier_open;
            self.form.cashier_printer = vars.cashier_printer;
            self.printerList = data.data.printerList;
            self.cashierList = data.data.cashierList;

            if (vars.order_status != null && vars.order_status[0] == 20) {
              self.checked = true;
            }

            let printerKeyArr = [];
            let printerIdArr = [];
            let printerSnArr = [];
            (self.form.cashier_printer || []).map((item) => {
              printerKeyArr.push(item.key);
              printerIdArr.push(item.printer_id);
              printerSnArr.push(item.sn);
            });

            (self.cashierList || []).map((item) => {
              if (printerKeyArr.indexOf(item.cashier_key) != -1) {
                item.printer_id = printerIdArr[printerKeyArr.indexOf(item.cashier_key)];
                item.sn = printerSnArr[printerKeyArr.indexOf(item.cashier_key)];
              } else {
                item.printer_id = '';
              }
            });

            let arr = [];
            self.printerList.map((item) => {
              arr.push(item.printer_id);
            });
            (self.form.cashier_printer || []).map((item) => {
              if (!arr.includes(item.printer_id)) {
                item.printer_id = '';
              }
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      //提交表单
      onSubmit() {
        let self = this;
        let params = {};
        params.cashier_open = self.form.cashier_open;
        params.cashier_printer = [];
        self.cashierList.map((item) => {
          params.cashier_printer.push({
            key: item.cashier_key,
            printer_id: item.printer_id,
            sn: item.sn,
          });
        });
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

<style scoped lang="scss">
  .tips {
    color: #ccc;
  }

  .cashier-item {
    margin-bottom: 0 !important;
    :deep(.el-form-item__content) {
      gap: 10px;
    }
  }
  .flex-center {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 4px;
    span {
      flex-shrink: 0;
    }
  }
  .cashier-desc {
    font-size: 14px;
    color: #ccc;
    margin-bottom: 20px;
    margin-top: 4px;
  }
</style>
