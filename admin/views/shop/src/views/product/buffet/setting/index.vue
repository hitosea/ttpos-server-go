<template>
  <div class="setting">
    <el-form class="form-setting" size="small" ref="form" :model="form" label-position="top" v-loading="loading">
      <el-form-item :label="$t('自助餐')" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_open">
          <el-radio value="1">{{ $t('开') }}</el-radio>
          <el-radio value="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item
        v-if="this.is_open_tablet || this.is_open_scan"
        :label="nameReturn()"
        prop="tablet_end_time"
        :rules="[{ required: true, message: $t('请输入平板/扫码H5结束时间提醒') }]"
      >
        <div class="max-w460 color-box">
          <el-input-number
            :controls="false"
            :min="0"
            :max="999"
            style="width: 200px !important"
            :placeholder="$t('请输入平板/扫码H5结束时间提醒')"
            v-model.number="form.tablet_end_time"
          ></el-input-number>
          {{ $t('分') }}
        </div>
      </el-form-item>

      <el-form-item :label="$t('是否显示非自助餐商品')" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_show_non_buffet_product">
          <el-radio value="1">{{ $t('开') }}</el-radio>
          <el-radio value="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="form.is_show_non_buffet_product == '1'" :label="$t('用餐时间到后可继续选购非自助餐商品')" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_buy_continue">
          <el-radio value="1">{{ $t('开') }}</el-radio>
          <el-radio value="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item
        class="customer-type"
        :label="$t('顾客类型')"
        prop="customer_type"
        :rules="[
          {
            required: true,
            validator: () => {
              return form.customer_type.length > 0 ? true : false;
            },
            message: $t('请添加顾客类型'),
          },
        ]"
      >
        <el-button type="primary" @click="addCustomerType" :disabled="checkLength >= 5">{{ $t('添加') }}</el-button>
        <template v-for="(item, index) in form.customer_type">
          <div class="customer-type-box" v-if="item.action != 'delete'">
            <el-form-item
              label=""
              :prop="`customer_type[${index}].name`"
              style="width: 100%"
              :rules="[
                {
                  required: true,
                  validator: () => {
                    return item.name ? true : false;
                  },
                  message: $t('请输入名称'),
                },
              ]"
            >
              <el-input type="text" :maxlength="50" style="margin-top: 16px; width: 100%" v-model="item.name" :placeholder="$t('请输入名称')"></el-input>
            </el-form-item>
            <el-icon class="delete-icon" :class="unCusDelete ? 'delete-icon-none' : ''" @click="handleDeleteCustomer(index)">
              <Delete />
            </el-icon>
          </div>
        </template>
      </el-form-item>

      <el-form-item :label="$t('加钟')" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.is_add_clock" @change="handleChange">
          <el-radio value="1">{{ $t('开') }}</el-radio>
          <el-radio value="0">{{ $t('关') }}</el-radio>
        </el-radio-group>
        <div class="limit-list" v-if="form.is_add_clock == 1">
          <el-button type="primary" @click="add">{{ $t('添加') }}</el-button>

          <div class="limit-product-list">
            <template v-for="(item, index) in form.add_clock" :key="index">
              <div v-if="item.action != 'delete'" class="limit-product-box">
                <el-form-item
                  label=""
                  :prop="`add_clock[${index}].name`"
                  :rules="[
                    {
                      required: true,
                      validator: () => {
                        return item.name ? true : false;
                      },
                      message: $t('请输入名称'),
                    },
                  ]"
                >
                  <el-input type="text" style="margin-top: 16px; min-width: 160px" v-model="item.name" :placeholder="$t('名称')"></el-input>
                </el-form-item>
                <el-form-item
                  label=""
                  :prop="`add_clock[${index}].delay_time`"
                  :rules="[
                    {
                      required: true,
                      validator: () => {
                        return item.delay_time ? true : false;
                      },
                      message: $t('请输入时间'),
                    },
                  ]"
                >
                  <el-input-number
                    :controls="false"
                    :min="0"
                    :max="999"
                    style="width: 160px !important; margin-top: 16px"
                    :placeholder="$t('请输入时间')"
                    v-model.number="item.delay_time"
                  ></el-input-number>
                </el-form-item>
                <p class="limit-product-p">{{ $t('分') }}</p>
                <el-form-item
                  label=""
                  :prop="`add_clock[${index}].price`"
                  :rules="[
                    {
                      required: true,
                      validator: () => {
                        return typeof item.price == 'object' ? false : true;
                      },
                      message: $t('请输入价格'),
                    },
                  ]"
                >
                  <el-input-number
                    :controls="false"
                    :min="0"
                    :max="100000000"
                    style="width: 160px !important; margin-top: 16px"
                    :placeholder="$t('请输入价格')"
                    v-model.number="item.price"
                  ></el-input-number>
                </el-form-item>
                <p class="limit-product-p">{{ currency.unit }}</p>
                <el-icon class="delete-icon" :class="unDelete ? 'delete-icon-none' : ''" @click="handleDelete(index)">
                  <Delete />
                </el-icon>
              </div>
            </template>
          </div>
        </div>
      </el-form-item>
    </el-form>
    <div class="common-button-wrapper">
      <el-button size="small" @click="getData">{{ $t('重置') }}</el-button>
      <el-button size="small" type="primary" @click="onSubmit" :loading="loading">{{ $t('保存') }}</el-button>
    </div>
    <productList v-if="open_product" :open_product="open_product" :openIndex="openIndex" :multiple_selection="multiple_selection" @closeDialogFunc="closeDialogFunc($event)">
    </productList>
  </div>
</template>
<script>
  import PorductApi from '@/api/product.js';
  import autoTips from '../list/autoTips.vue';
  import productList from './productList.vue';
  import { useUserStore } from '@/store';
  const { currency, computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_tablet = supplier.value?.is_open_tablet || 0;
  const is_open_scan = supplier.value?.is_open_scan || 0;
  export default {
    components: { productList, autoTips },
    data() {
      return {
        open_product: false,

        currency: currency,
        loading: false,
        form: {
          is_open: '1',
          tablet_end_time: null,
          is_buy_continue: '1',
          is_buffet_discount: '1',
          add_buffet_discount: [],
          customer_type: [],
          is_add_clock: '1',
          add_clock: [],
          fav: [],
          is_show_non_buffet_product: '1',
          // is_remain_continue: '0',
          // remain_continue_notice_time: '',
          // remain_continue_time: '',
        },
        unDelete: false,
        unFavDelete: false,
        unCusDelete: false,
        select_list: [],
        multiple_selection: [],
        openIndex: 0,
        is_open_tablet: is_open_tablet,
        is_open_scan: is_open_scan,
      };
    },
    mounted() {
      this.getData();
    },
    watch: {
      'form.add_clock': {
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
      'form.add_buffet_discount': {
        handler(val) {
          let result = 0;
          val.map((item) => {
            if (item.action != 'delete') {
              result++;
            }
          });
          if (result == 1) {
            this.unFavDelete = true;
          } else {
            this.unFavDelete = false;
          }
        },
        deep: true,
        immediate: true,
      },
      'form.customer_type': {
        handler(val) {
          let result = 0;
          val.map((item) => {
            if (item.action != 'delete') {
              result++;
            }
          });
          if (result == 1) {
            this.unCusDelete = true;
          } else {
            this.unCusDelete = false;
          }
        },
        deep: true,
        immediate: true,
      },
    },
    computed: {
      checkLength() {
        let result = 0;
        this.form.customer_type.map((item) => {
          if (item.action == 'add' || item.action == 'edit') {
            result++;
          }
        });
        return result;
      },
    },
    methods: {
      getData() {
        let self = this;
        self.loading = true;
        PorductApi.getSettingBuffet()
          .then((data) => {
            self.loading = false;
            this.form = data.data.vars.values;
            this.form.add_clock.map((item, index) => {
              this.form.add_clock[index].action = 'edit';
              this.form.add_clock[index].price = Number(item.price);
            });
            this.form.tablet_end_time = Number(this.form.tablet_end_time);
            this.select_list = [];

            this.form.add_buffet_discount.map((item, index) => {
              this.select_list.push(item.buffets);
              this.form.add_buffet_discount[index].buffet_ids = [];
              this.form.add_buffet_discount[index].action = 'edit';
            });

            this.form.customer_type.map((item, index) => {
              this.form.customer_type[index].action = 'edit';
            });

            this.from.remain_continue_notice_time = Number(this.from.remain_continue_notice_time);
            this.from.remain_continue_time = Number(this.from.remain_continue_time);

            // 暂时改为默认
            this.form.is_buffet_discount = '0';
            this.form.add_buffet_discount = [];
            // 暂时改为默认

            this.select_list.map((item, index) => {
              item.map((items) => {
                this.form.add_buffet_discount[index].buffet_ids.push(items.id);
              });
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },
      onSubmit() {
        // 定义一个变量self，指向当前组件实例
        let self = this;
        // 将当前组件实例的form属性转换为JSON字符串，再解析为对象，赋值给params
        let params = JSON.parse(JSON.stringify(self.form));
        // 将当前组件实例的form属性转换为JSON字符串，再解析为对象，赋值给copyData
        let copyData = JSON.parse(JSON.stringify(self.form));

        // 将copyData的add_clock属性中的每个对象转换为指定格式，并赋值给params的add_clock属性
        params.add_clock = copyData.add_clock.map(({ id, name, delay_time, price, action }) => ({ id, name, delay_time, price, action }));

        // 将copyData的customer_type属性中的每个对象转换为指定格式，并赋值给params的customer_type属性
        params.customer_type = copyData.customer_type.map(({ id, name, action }) => ({ id, name, action }));

        // 调用当前组件实例的$refs属性中的form属性的validate方法，验证表单是否合法
        self.$refs.form.validate((valid) => {
          // 如果表单验证通过
          if (valid) {
            // 将当前组件实例的loading属性设置为true，表示正在加载
            self.loading = true;
            // 调用PorductApi的setSettingBuffet方法，传入params和true，发送请求
            PorductApi.setSettingBuffet(params, true)
              // 请求成功后
              .then((data) => {
                // 将当前组件实例的loading属性设置为false，表示加载完成
                self.loading = false;
                // 调用当前组件实例的$ElMessage方法，显示保存成功的提示信息
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });
                // 调用当前组件实例的getData方法，刷新数据
                this.getData();
              })
              // 请求失败后
              .catch((error) => {
                // 将当前组件实例的loading属性设置为false，表示加载完成
                self.loading = false;
              });
          }
        });
      },

      //名字返回
      nameReturn() {
        let result = '';
        if (this.is_open_tablet && this.is_open_scan) {
          result = $t('平板/扫码H5结束时间提醒');
        }
        if (this.is_open_tablet && !this.is_open_scan) {
          result = $t('平板结束时间提醒');
        }
        if (!this.is_open_tablet && this.is_open_scan) {
          result = $t('扫码H5结束时间提醒');
        }
        return result;
      },

      add() {
        this.form.add_clock.push({
          id: 0,
          name: '',
          delay_time: null,
          price: null,
          action: 'add',
        });
      },

      handleChange() {
        // 如果add_clock的长度为0，则调用add方法
        if (this.form.add_clock.length == 0) {
          this.add();
        }
      },

      handleChangeDiscount() {
        // 如果add_buffet_discount的长度为0，则调用addFavorable方法
        if (this.form.add_buffet_discount.length == 0) {
          this.addFavorable();
        }
      },

      handleDelete(index) {
        // 定义一个变量result，初始值为0
        let result = 0;
        // 遍历add_clock数组，如果action不等于delete，则result加1
        this.form.add_clock.map((item) => {
          if (item.action != 'delete') {
            result++;
          }
        });
        // 如果result等于1，则返回
        if (result == 1) return;
        // 如果add_clock数组的id等于0，则删除该元素
        if (this.form.add_clock[index].id == 0) {
          this.form.add_clock.splice(index, 1);
        } else {
          // 否则，将action设置为delete
          this.form.add_clock[index].action = 'delete';
        }
      },

      selectFavorable(item, index) {
        // 将openIndex设置为index
        this.openIndex = index;
        // 将multiple_selection设置为item.buffet_ids
        this.multiple_selection = item.buffet_ids;
        // 将open_product设置为true
        this.open_product = true;
      },

      handleChangeDiscountType(item, index) {
        // 如果discount_type等于1，则调用validateField方法，验证add_buffet_discount数组中index位置的discount_ratio
        item.discount_type == '1'
          ? this.$refs.form.validateField(`add_buffet_discount[${index}].discount_ratio`)
          : // 否则，调用validateField方法，验证add_buffet_discount数组中index位置的discount_price
            this.$refs.form.validateField(`add_buffet_discount[${index}].discount_price`);
      },

      addFavorable() {
        // 将select_list数组的第一个元素设置为空数组
        this.select_list.unshift([]);
        // 将add_buffet_discount数组的第一个元素设置为对象
        this.form.add_buffet_discount.unshift({
          id: 0,
          name: '',
          discount_type: 1,
          discount_ratio: null,
          discount_price: null,
          buffet_ids: [],
          action: 'add',
        });
      },

      addCustomerType() {
        // 将customer_type数组的第一个元素设置为对象
        this.form.customer_type.unshift({
          id: 0,
          name: '',
          action: 'add',
        });
      },

      handleDeleteCustomer(index) {
        // 定义一个变量result，初始值为0
        let result = 0;
        // 遍历customer_type数组，如果action不等于delete，则result加1
        this.form.customer_type.map((item) => {
          if (item.action != 'delete') {
            result++;
          }
        });
        // 如果result等于1，则返回
        if (result == 1) return;
        // 如果customer_type数组的id等于0，则删除该元素
        if (this.form.customer_type[index].id == 0) {
          this.form.customer_type.splice(index, 1);
        } else {
          // 否则，将action设置为delete
          this.form.customer_type[index].action = 'delete';
        }
      },

      //判断重名
      checkName(name, index) {
        let result = true;
        this.form.customer_type.map((item, e) => {
          if (item.name == name && e != index && item.action != 'delete') {
            result = false;
          }
        });
        return result;
      },

      handleDeleteFav(index) {
        let result = 0;
        this.form.add_buffet_discount.map((item) => {
          if (item.action != 'delete') {
            result++;
          }
        });
        if (result == 1) return;
        if (this.form.add_buffet_discount[index].id == 0) {
          this.form.add_buffet_discount.splice(index, 1);
        } else {
          this.form.add_buffet_discount[index].action = 'delete';
        }
      },

      deleteOne(index, indexs) {
        this.select_list[index].splice(indexs, 1);
        this.form.add_buffet_discount[index].buffet_ids = [];
        this.select_list[index].map((item) => {
          this.form.add_buffet_discount[index].buffet_ids.push(item.id);
        });
        this.$refs.form.validateField(`add_buffet_discount[${index}].buffet_ids`);
      },

      closeDialogFunc(e) {
        this.open_product = e.openDialog;
        if (e.type == 'submit') {
          let map = new Map();
          [this.select_list[e.index], e.data].flat().forEach((obj) => map.set(obj.id, obj));
          this.select_list[e.index] = Array.from(map.values());

          this.form.add_buffet_discount[e.index].buffet_ids = [];
          this.select_list[e.index].map((item) => {
            this.form.add_buffet_discount[e.index].buffet_ids.push(item.id);
          });
          this.$refs.form.validateField(`add_buffet_discount[${e.index}].buffet_ids`);
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .setting {
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    overflow-x: hidden;
    height: 100%;
  }

  .form-setting {
    flex: 1 1 auto;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .common-button-wrapper {
    flex: 0 0 auto;
    flex-shrink: 0;
  }

  .color-box {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .p-tips {
    color: var(--el-color-tips);
  }

  .limit-list {
    width: 100%;
    margin-top: 12px;

    .limit-product-list {
      width: 100%;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      margin-top: 16px;

      .limit-product-box {
        display: flex;
        gap: 12px;
        align-items: center;

        .limit-product-p {
          flex-shrink: 0;
        }

        .delete-icon {
          font-size: 24px;
          cursor: pointer;
        }

        .delete-icon-none {
          cursor: not-allowed;
          color: var(--el-color-tips);
        }
      }
    }
  }

  .favorable-list {
    width: 100%;
    margin-top: 12px;

    .favorable-list-box {
      width: 100%;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      margin-top: 16px;
      gap: 12px;

      .favorable-box {
        display: flex;
        gap: 12px;
        max-width: 620px;
        overflow: hidden;

        .favorable-box-l {
          width: 100%;
          max-width: 584px;
          border: 1px solid var(--el-color-border);
          border-radius: 4px;
          padding: 0 16px;

          .favorable-box-item {
            :deep(.el-form-item__content) {
              display: flex;
              align-items: center;
              flex-wrap: nowrap;
              gap: 12px;

              .p-unit {
              }
            }
          }
        }

        .delete-icon {
          font-size: 24px;
          cursor: pointer;
          margin-top: 18px;
          flex-shrink: 0;
        }

        .delete-icon-none {
          cursor: not-allowed;
          color: var(--el-color-tips);
        }
      }
    }
  }

  .select-list {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding: 7px 7px 0 0;
    max-height: 400px;
    overflow: auto;
  }

  .select-button {
    flex: 1;
    max-width: calc(33% - 6px);
    min-width: 30%;
    border: solid 1px var(--el-color-tips);
    color: var(--el-color-tips);
    border-radius: 4px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .price-p {
      margin-right: 6px;
      flex-shrink: 0;
      color: #c80000;
    }

    .select-p {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      position: relative;
    }

    .select-icon {
      position: absolute;
      right: -7px;
      top: -7px;
      cursor: pointer;
      color: #c80000;
    }
  }

  .customer-type {
    :deep(.el-form-item__content) {
      flex-direction: column;
      align-items: start;
    }

    .customer-type-box {
      width: 100%;
      max-width: 609px;
      display: flex;
      align-items: center;
      gap: 8px;

      .delete-icon {
        font-size: 24px;
        cursor: pointer;
        flex-shrink: 0;
      }

      .delete-icon-none {
        cursor: not-allowed;
        color: var(--el-color-tips);
      }
    }
  }
</style>
