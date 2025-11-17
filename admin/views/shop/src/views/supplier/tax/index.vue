<template>
  <div v-loading="winLoading || loading">
    <div class="supplier">
      <el-form size="small" ref="form" :model="form" label-position="top">
        <el-form-item :label="$t('消费税')" prop="is_open" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio v-model="form.is_open" :label="'1'">{{ $t('开启') }}</el-radio>
            <el-radio v-model="form.is_open" :label="'0'">{{ $t('关闭') }}</el-radio>
          </div>
          <div class="tips">{{ $t('注：开启之前没有过税类的商品、自助餐默认选择第一个税类') }}</div>
        </el-form-item>

        <el-form-item v-if="form.is_open == '1'" :label="$t('计算类型')" prop="calc_type" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio v-model="form.calc_type" :label="'1'">{{ $t('商品已含税') }}</el-radio>
            <el-radio v-model="form.calc_type" :label="'2'">{{ $t('商品未含税') }}</el-radio>
          </div>
        </el-form-item>

        <el-form-item
          v-if="form.is_open == '1'"
          :label="$t('税类')"
          prop="add_tax_category"
          :rules="[
            {
              validator: () => {
                return form.add_tax_category.length > 0 ? true : false;
              },
              message: $t('请添加税率'),
            },
          ]"
        >
          <template #label>
            <p class="label-p"><span>*</span>{{ $t('税类') }}</p>
          </template>
          <el-button type="primary" @click="add">{{ $t('添加') }}</el-button>
          <div class="tips">{{ $t('注：输入百分比，开始时计算所增加的消费税') }}</div>
        </el-form-item>
        <template v-if="form.is_open == '1'" v-for="(item, index) in form.add_tax_category">
          <div class="flex-box" v-if="item.action != 'delete'">
            <el-form-item
              class="max-w460"
              label=""
              :prop="`add_tax_category[${index}].name`"
              :rules="[
                {
                  validator: () => {
                    return item.name ? true : false;
                  },
                  message: $t('请输入名称'),
                },

                {
                  validator: (rule, value, callback) => {
                    if (!value) {
                      callback(new Error($t('请输入名称')));
                    } else {
                      callback();
                    }
                  },
                  trigger: 'change',
                },
              ]"
            >
              <el-input :placeholder="$t('请输入税率名称')" :maxlength="50" v-model="item.name"></el-input>
            </el-form-item>
            <el-form-item
              class="max-w460"
              label=""
              :prop="`add_tax_category[${index}].tax_rate`"
              :rules="[
                {
                  validator: () => {
                    return item.tax_rate != null ? true : false;
                  },
                  message: $t('请输入0-100的税率'),
                },
              ]"
            >
              <numInput :min="0" :max="100" :precision="2" v-model="item.tax_rate" :placeholder="$t('请输入税率')"></numInput>
            </el-form-item>
            <span class="span-p">%</span>
            <el-icon class="delete-icon" :class="unDelete ? 'delete-icon-none' : ''" @click="handleDelete(index)">
              <Delete />
            </el-icon>
          </div>
        </template>
      </el-form>
      <!--提交-->
      <div class="common-button-wrapper">
        <el-button @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
      </div>
    </div>
  </div>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import numInput from '@/components/num-input/index.vue';
  import { EEUIRELOAD } from '@/utils/platform.js';

  export default {
    components: {
      numInput,
    },
    data() {
      return {
        loading: false,
        unDelete: false,
        form: {
          is_open: '1',
          calc_type: '1',
          add_tax_category: [],
        },
        winLoading: false,
      };
    },
    created() {
      this.getData();
    },
    watch: {
      'form.add_tax_category': {
        handler(val) {
          let result = 0;
          val.map((item) => {
            if (item.action != 'delete') {
              result++;
            }
          });
          if (result == 1) {
            this.unDelete = true;
          } else {
            this.unDelete = false;
          }
        },
        deep: true,
        immediate: true,
      },
    },
    methods: {
      /*获取列表*/
      getData() {
        let self = this;
        self.loading = true;
        SettingApi.getTaxRate({}, true)
          .then((data) => {
            self.loading = false;
            self.form.is_open = data.data.vars.values.is_open.toString();
            self.form.calc_type = data.data.vars.values.calc_type.toString();
            self.form.add_tax_category = [];
            data.data.vars.values.add_tax_category.map((item) => {
              self.form.add_tax_category.push({
                id: item.id,
                name: item.name,
                tax_rate: item.tax_rate,
                action: 'edit',
              });
            });
            self.$refs.form.validate();
          })
          .catch((error) => {});
      },
      onSubmit() {
        let self = this;
        let params = JSON.parse(JSON.stringify(self.form));
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            SettingApi.setTaxRate(params, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });
                self.winLoading = true;
                setTimeout(() => {
                  EEUIRELOAD();
                }, 1000);
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },
      //添加一个税率
      add() {
        this.form.add_tax_category.push({
          id: 0,
          name: '',
          tax_rate: null,
          action: 'add',
        });
        this.$refs.form.validateField('add_tax_category');
      },

      handleDelete(index) {
        let result = 0;
        this.form.add_tax_category.map((item) => {
          if (item.action != 'delete') {
            result++;
          }
        });
        if (result == 1) return;
        if (this.form.add_tax_category[index].id == 0) {
          this.form.add_tax_category.splice(index, 1);
        } else {
          this.form.add_tax_category[index].action = 'delete';
        }
      },
    },
  };
</script>
<style scoped lang="scss">
  .flex-box {
    display: flex;
    align-items: center;
    width: 100%;
    gap: 12px;

    .span-p {
      margin-bottom: 16px;
    }

    .el-form-item {
      width: 100%;
    }

    .delete-icon {
      font-size: 24px;
      cursor: pointer;
      margin-bottom: 16px;
    }

    .delete-icon-none {
      cursor: not-allowed;
      color: var(--el-color-tips);
    }
  }

  .label-p span {
    font-size: var(--el-form-label-font-size);
    color: var(--el-color-danger);
    margin-right: 4px;
  }
</style>
