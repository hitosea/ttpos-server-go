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

        <!--服务费计算-->
        <el-form-item class="mt-24" :label="$t('服务费计算')" prop="apply_scope" :rules="[{ required: true, message: $t('请选择服务费计算') }]">
          <div>
            <el-radio v-model="form.service_fee_base" :label="'0'" true-label>{{ $t('商品惠后价') }}</el-radio>
            <el-radio v-model="form.service_fee_base" :label="'1'">{{ $t('商品价格合计') }}</el-radio>
          </div>
        </el-form-item>

        <!-- 应用范围 -->
        <el-form-item class="mt-24" :label="$t('应用范围')" prop="apply_scope" :rules="[{ required: true, message: $t('请选择应用范围') }]">
          <div>
            <el-radio v-model="form.apply_scope" :label="'1'" true-label>{{ $t('全部应用') }}</el-radio>
            <el-radio v-model="form.apply_scope" :label="'2'">{{ $t('部分应用') }}</el-radio>
          </div>
        </el-form-item>
        <template v-if="form.apply_scope == '2'">
          <el-form-item label="">
            <el-checkbox v-model="form.apply_scope_ordering" true-label="1" false-label="0">{{ $t('点餐方式') }}</el-checkbox>
            <el-checkbox v-model="form.apply_scope_table" true-label="1" false-label="0">{{ $t('桌台方式') }}</el-checkbox>
          </el-form-item>
          <div class="table-selector" @click="handleOpenTableSelector" v-if="form.apply_scope_table == '1'">
            <div class="table-selector-content">
              <el-tag v-for="tag in selectTableList" :key="tag.area_id" :disable-transitions="true" closable type="info" @close="($event) => handleCloseTag($event, tag)">
                {{ tag.area_name }} ({{ tag.count }})
              </el-tag>
              <p class="tips" v-if="selectTableList.length == 0">{{ $t('请选择桌台') }}</p>
            </div>
            <div class="el-select__suffix">
              <el-icon>
                <ArrowDown />
              </el-icon>
            </div>
          </div>
        </template>
      </template>
    </el-form>
    <!--提交-->
    <div class="common-button-wrapper">
      <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
      <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
    <DownloadQrcode
      v-if="is_open_batch_download_qrcode"
      :open="is_open_batch_download_qrcode"
      :include="form.apply_scope_table_list"
      :Dtype="Dtype"
      @close="closeDownloadQrcode"
      @selectTable="selectTable"
    >
    </DownloadQrcode>
  </div>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import StoreApi from '@/api/store.js';
  import { useUserStore } from '@/store';
  import DownloadQrcode from '@/views/supplier/table/table/batch/DownloadQrcode.vue';
  const { currency } = useUserStore();
  export default {
    components: {
      DownloadQrcode,
    },
    data() {
      return {
        currency: currency,
        loading: false,
        is_open_batch_download_qrcode: false,
        Dtype: 'service',
        tableList: [],
        allTableList: [],
        form: {
          is_open: '1',
          charge_type: '1',
          service_charge: 0,
          service_charge_rate: 0,
          is_open_tax: '0',
          apply_scope: '1',
          apply_scope_ordering: '0',
          apply_scope_table: '0',
          apply_scope_table_list: [],
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
      this.getTableList();
    },
    computed: {
      selectTableList() {
        //筛选出选中的桌台
        this.tableList = this.allTableList.filter((item) => this.form.apply_scope_table_list.includes(item.table_id));
        // 将tableList中的area_id和area_name组合成一个对象
        const areaListWithName = this.tableList.map((item) => ({
          area_id: item.area_id,
          area_name: item.area_name,
          count: 0,
        }));
        // 根据area_id去重
        const areaListWithNameUnique = areaListWithName.filter((item, index, self) => index === self.findIndex((t) => t.area_id === item.area_id));
        // 根据area_id将areaListWithNameUnique中的count加1
        areaListWithNameUnique.forEach((item) => {
          item.count = this.tableList.filter((t) => t.area_id === item.area_id).length;
        });
        return areaListWithNameUnique;
      },
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
      /*获取列表*/
      getTableList() {
        let params = {};
        params.page = 1;
        params.list_rows = 1000;
        StoreApi.tablelist(params, true)
          .then((data) => {
            this.allTableList = data.data.list.data;
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

      closeDownloadQrcode(e) {
        this.is_open_batch_download_qrcode = false;
      },

      selectTable(e) {
        this.form.apply_scope_table_list = e;
        this.is_open_batch_download_qrcode = false;
      },

      handleOpenTableSelector() {
        this.is_open_batch_download_qrcode = true;
      },

      handleCloseTag(event, tag) {
        event.stopPropagation();
        // 删除tableList中area_id为tag.area_id的元素
        this.tableList.map((item) => {
          if (item.area_id === tag.area_id) {
            this.form.apply_scope_table_list.splice(this.form.apply_scope_table_list.indexOf(item.table_id), 1);
          }
        });
      },
    },
  };
</script>
<style lang="scss" scoped>
  .mt-24 {
    margin-top: 24px;
  }
  .table-selector {
    display: flex;
    align-items: center;
    position: relative;
    box-sizing: border-box;
    cursor: pointer;
    text-align: left;
    gap: 4px;
    margin-bottom: 16px;

    border-radius: var(--el-border-radius-base);
    padding: 4px 8px !important;
    background-color: var(--el-fill-color-blank);
    transition: var(--el-transition-duration);
    box-shadow: 0 0 0 1px var(--el-border-color) inset;
    width: 480px;
    height: auto;
    min-height: 32px !important;
    font-size: 14px;

    .table-selector-content {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      align-items: center;
      position: relative;
      flex: 1;
      min-width: 0;
    }

    .table-selector-suffix {
      gap: 4px;
      display: flex;
      align-items: center;
      flex-shrink: 0;
      color: var(--el-input-icon-color, var(--el-text-color-placeholder));
    }
  }
</style>
