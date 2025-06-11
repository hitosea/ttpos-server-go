<template>
  <div class="pb50" v-loading="loading">
    <el-form ref="form" size="small" :model="form" label-position="top" label-width="200px">
      <div class="common-form">{{ $t('扣款设置') }}</div>
      <el-form-item>
        <template #label>
          <div class="label-wrap">
            <span>{{ $t('扣款顺序') }}</span>
            <el-tooltip effect="dark" placement="bottom">
              <template #content>
                <p>{{
                  $t('先主账户后赠送账户：按照先扣主账户金额，再扣除赠送账户的金额。例：充值1000赠送100，每次消费先扣主账户（1000）里的余额，扣完后扣赠送账户（100）里的余额')
                }}</p>
                <p>{{
                  $t('先赠送账户后主账户：按照先扣赠送账户金额，再扣除主账户的金额。例：充值1000赠送100，每次消费先扣赠送账户（100）里的余额，扣完后扣主账户（1000）里的余额')
                }}</p>
                <p
                  >{{
                    $t(
                      '按比例：按照主账户和赠送账户比例进行扣款。例如，如果主账户比例为60%，赠送账户比例为40%，则根据比例从各账户中扣款。如果赠送账户余额不足，则从主账户中补足扣款。反之亦然。'
                    )
                  }}
                </p>
              </template>
              <SvgIcon class="tip-icon" name="icon6"></SvgIcon>
            </el-tooltip>
          </div>
        </template>
        <el-radio-group v-model="form.deduct_order">
          <el-radio :label="1">{{ $t('先主账户后赠送账户') }}</el-radio>
          <el-radio :label="2">{{ $t('先赠送账户后主账户') }}</el-radio>
          <el-radio :label="3">{{ $t('按比例') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <template v-if="form.deduct_order == 3">
        <el-form-item :label="$t('主账户')" prop="deduct_ratio_main" :rules="[{ required: true, message: $t('请输入主账户比例') }]">
          <el-input-number
            :controls="false"
            class="max-w460"
            :min="0"
            :max="100"
            :precision="2"
            :placeholder="$t('请输入主账户比例')"
            v-model.number="form.deduct_ratio_main"
            @input="handleDeductRatioMainInput"
            @change="handleDeductRatioChange"
          ></el-input-number>
          <span>%</span>
        </el-form-item>
        <el-form-item :label="$t('赠送账户')" prop="deduct_ratio_gift" :rules="[{ required: true, message: $t('请输入赠送账户比例') }]">
          <el-input-number
            :controls="false"
            class="max-w460"
            :min="0"
            :max="100"
            :precision="2"
            :placeholder="$t('请输入赠送账户比例')"
            v-model.number="form.deduct_ratio_gift"
            @input="handleDeductRatioGiftInput"
            @change="handleDeductRatioChange"
          ></el-input-number>
          <span>%</span>
        </el-form-item>
      </template>
      <div class="common-form">{{ $t('积分设置') }}</div>
      <el-form-item :label="$t('积分名称')" prop="points_name" :rules="[{ required: true, message: ' ' }]">
        <el-input v-model="form.points_name" :placeholder="$t('自定义您店铺的积分名称')" autocomplete="off" class="max-w460"></el-input>
      </el-form-item>

      <div class="common-form">{{ $t('购物赠送积分规则') }}</div>
      <el-form-item :label="$t('按付款金额比例赠送')">
        <el-radio-group v-model="form.shopping_gift_rules[0].is_open">
          <el-radio label="1">{{ $t('开启') }}</el-radio>
          <el-radio label="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
        <div class="lh18 mt10 gray9">
          <p>{{ $t('例：订单付款金额（100.00元）*积分赠送比例（100%）=实际赠送的积分（100积分）') }}</p>
        </div>
      </el-form-item>

      <template v-if="form.shopping_gift_rules[0].is_open == 1">
        <el-form-item :label="$t('是否按会员等级赠送')" prop="shopping_gift_rules.0.is_member_level_related" :rules="[{ required: true, message: $t('请选择是否按会员等级赠送') }]">
          <el-radio-group v-model="form.shopping_gift_rules[0].is_member_level_related">
            <el-radio label="0">{{ $t('所有会员等级相同') }}</el-radio>
            <el-radio label="1">{{ $t('按等级区分') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.shopping_gift_rules[0].is_member_level_related == 0">
          <el-form-item :label="$t('赠送积分比例')" prop="shopping_gift_rules.0.value" :rules="[{ required: true, message: $t('请输入赠送积分比例') }]">
            <el-input class="max-w460" @input="(e) => handleValueInput(e, 0)" :placeholder="$t('请输入赠送积分比例')" v-model="form.shopping_gift_rules[0].value"></el-input>
            <span>%</span>
          </el-form-item>
        </template>

        <template v-if="form.shopping_gift_rules[0].is_member_level_related == 1">
          <el-form-item
            v-for="(item, levelIndex) in form.shopping_gift_rules[0].member_levels"
            :key="levelIndex"
            :label="item.name"
            :prop="`shopping_gift_rules.0.member_levels.${levelIndex}.value`"
            :rules="[{ required: true, message: $t('请输入积分赠送') }]"
          >
            <el-input class="max-w460" @input="(e) => handleMemberLevelsInput(e, 0, levelIndex)" :placeholder="$t('请输入积分赠送')" v-model="item.value"></el-input>
            <span> %</span>
            <div class="lh18 mt10 gray9">
              <p> {{ $t('注：请输入大于0的数字') }}</p>
            </div>
          </el-form-item>
        </template>
        <el-form-item
          :label="$t('赠送积分所需付款金额')"
          prop="shopping_gift_rules.0.payment_amount_requirement"
          :rules="[{ required: true, message: $t('请输入赠送积分所需付款金额') }]"
        >
          <el-input
            class="max-w460"
            @input="(e) => handleMoneyInput(e, 0)"
            :placeholder="$t('请输入金额')"
            v-model="form.shopping_gift_rules[0].payment_amount_requirement"
          ></el-input>
          <div class="lh18 mt10 gray9">
            <p>{{ $t('注：请输入大于0的数字') }}</p>
          </div>
        </el-form-item>
        <el-form-item :label="$t('适用就餐类型')" prop="shopping_gift_rules.0.meal_type" :rules="[{ required: true, message: $t('请选择适用就餐类型') }]">
          <el-checkbox-group v-model="form.shopping_gift_rules[0].meal_type">
            <el-checkbox label="non-buffet">{{ $t('非自助餐') }}</el-checkbox>
            <el-checkbox v-if="is_open_buffet == 1" label="buffet">{{ $t('自助餐') }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item :label="$t('会员余额支付是否赠送')" prop="shopping_gift_rules.0.balance_payment_get_points" :rules="[{ required: true, message: ' ' }]">
          <el-radio-group v-model="form.shopping_gift_rules[0].balance_payment_get_points">
            <el-radio label="1">{{ $t('是') }}</el-radio>
            <el-radio label="0">{{ $t('否') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('退款自动扣积分')" prop="shopping_gift_rules.0.refund_return_points" :rules="[{ required: true, message: ' ' }]">
          {{ $t('开启') }}
          <div class="lh18 mt10 gray9">
            <p> {{ $t('注：该规则下默认开启退款自动退积分，若开启了积分自动抵扣消费金额则需要订单退款时手动扣积分') }}</p>
          </div>
        </el-form-item>
      </template>

      <el-divider />

      <el-form-item :label="$t('按桌台人数赠送')" v-if="is_open_buffet == 1">
        <el-radio-group v-model="form.shopping_gift_rules[1].is_open">
          <el-radio label="1">{{ $t('开启') }}</el-radio>
          <el-radio label="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <template v-if="form.shopping_gift_rules[1].is_open == 1 && is_open_buffet == 1">
        <el-form-item :label="$t('是否按会员等级赠送')" prop="shopping_gift_rules.1.is_member_level_related" :rules="[{ required: true, message: ' ' }]">
          <el-radio-group v-model="form.shopping_gift_rules[1].is_member_level_related">
            <el-radio label="0">{{ $t('所有会员等级相同') }}</el-radio>
            <el-radio label="1">{{ $t('按等级区分') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.shopping_gift_rules[1].is_member_level_related == 0">
          <el-form-item :label="$t('赠送积分')" prop="shopping_gift_rules.1.value" :rules="[{ required: true, message: $t('请输入赠送积分') }]">
            <el-input class="max-w460" @input="(e) => handleValuesInput(e, 1)" :placeholder="$t('请输入积分赠送')" v-model="form.shopping_gift_rules[1].value"></el-input>
            <span> {{ $t('积分/人') }}</span>
          </el-form-item>
        </template>

        <template v-if="form.shopping_gift_rules[1].is_member_level_related == 1">
          <el-form-item
            v-for="(item, levelIndex) in form.shopping_gift_rules[1].member_levels"
            :key="levelIndex"
            :label="item.name"
            :prop="`shopping_gift_rules.1.member_levels.${levelIndex}.value`"
            :rules="[{ required: true, message: $t('请输入积分赠送') }]"
          >
            <el-input class="max-w460" @input="(e) => handleMemberLevelInput(e, 1, levelIndex)" :placeholder="$t('请输入积分赠送')" v-model="item.value"></el-input>
            <span> {{ $t('积分/人') }}</span>
            <div class="lh18 mt10 gray9">
              <p> {{ $t('注：请输入大于0的数字') }}</p>
            </div>
          </el-form-item>
        </template>

        <el-form-item :label="$t('适用就餐类型')" :rules="[{ required: true, message: ' ' }]">
          <!-- <el-checkbox-group v-model="form.shopping_gift_rules[1].meal_type">
            <el-checkbox label="non-buffet">{{ $t('非自助餐') }}</el-checkbox>
            <el-checkbox  disabled label="buffet">{{ $t('自助餐') }}</el-checkbox>
          </el-checkbox-group> -->
          <div class="lh18 mt10 gray9">
            <p> {{ $t('该规则仅支持自助餐') }}</p>
          </div>
        </el-form-item>

        <el-form-item :label="$t('会员余额支付是否赠送')" prop="shopping_gift_rules.1.balance_payment_get_points" :rules="[{ required: true, message: ' ' }]">
          <el-radio-group v-model="form.shopping_gift_rules[1].balance_payment_get_points">
            <el-radio label="1">{{ $t('是') }}</el-radio>
            <el-radio label="0">{{ $t('否') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('退款自动扣积分')" prop="shopping_gift_rules.1.refund_return_points" :rules="[{ required: true, message: ' ' }]">
          {{ $t('关闭') }}
          <div class="lh18 mt10 gray9">
            <p> {{ $t('注：该规则仅支持退款时手动输入扣减积分') }}</p>
          </div>
        </el-form-item>
      </template>
      <div class="common-form">{{ $t('积分抵扣') }}</div>
      <el-form-item :label="$t('会员积分抵扣订单金额')">
        <el-radio-group v-model="form.exchange.open_points_exchange">
          <el-radio label="1">{{ $t('开启') }}</el-radio>
          <el-radio label="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
        <div class="lh18 mt10 gray9">
          <p> {{ $t('注：不开启不可设置抵扣比例') }}</p>
        </div>
      </el-form-item>

      <template v-if="form.exchange.open_points_exchange == 1">
        <el-form-item :label="$t('每积分抵扣应付金额')">
          <el-input class="max-w460" @input="(e) => handlePointsExchangeRateInput(e)" :placeholder="$t('请输入')" v-model="form.exchange.points_exchange_rate"></el-input>
        </el-form-item>

        <el-form-item :label="$t('是否自动抵扣')">
          <el-radio-group v-model="form.exchange.auto_points_exchange">
            <el-radio label="1">{{ $t('开启') }}</el-radio>
            <el-radio label="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </template>
      <!--提交-->
      <div class="common-button-wrapper">
        <el-button size="small" @click="getData" :loading="loading">{{ $t('重置') }}</el-button>
        <el-button type="primary" size="small" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </el-form>
  </div>
</template>
<script>
  import PointsApi from '@/api/points.js';
  import { useUserStore } from '@/store/index';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const is_open_buffet = supplier.value?.is_open_buffet || 0;
  export default {
    data() {
      return {
        is_open_buffet: is_open_buffet,
        form: {
          deduct_order: 1,
          deduct_ratio_main: 100,
          deduct_ratio_gift: 0,
          points_name: '',
          is_shopping_gift: 1,
          gift_ratio: 10,
          is_member_level_related: 0,
          meal_type: [],
          is_member_balance_payment: '1',
          refund_return_points: '0',
          is_shopping_discount: 1,
          discount: {
            discount_ratio: 0,
            full_order_price: 0,
            max_money_ratio: 0,
          },
          shopping_gift_rules: [
            {
              // 按付款金额比例赠送
              is_open: '0',
              is_member_level_related: '0',
              value: '',
              payment_amount_requirement: '',
              meal_type: [],
              balance_payment_get_points: '1',
              refund_return_points: '0',
              member_levels: [],
            },
            {
              // 按桌台人数赠送
              is_open: '0',
              is_member_level_related: '0',
              value: '',
              payment_amount_requirement: '',
              meal_type: [],
              balance_payment_get_points: '1',
              refund_return_points: '0',
              member_levels: [],
            },
          ],
          exchange: {
            open_points_exchange: '0',
            points_exchange_rate: '0',
            auto_points_exchange: '0',
          },
        },
        loading: false,
      };
    },
    created() {
      /*获取数据*/
      this.getData();
    },
    methods: {
      /*获取数据*/
      getData() {
        let self = this;
        let Params = {};
        self.loading = true;
        PointsApi.getPoints(Params, true)
          .then((data) => {
            self.loading = false;
            self.form = data.data.values;
            self.form.deduct_order = parseInt(data.data.values.deduct_order ?? 1);
            self.form.deduct_ratio_main = Number(data.data.values.deduct_ratio_main ?? 100);
            self.form.deduct_ratio_gift = Number(data.data.values.deduct_ratio_gift ?? 0);
            self.form.is_shopping_gift = parseInt(data.data.values.is_shopping_gift);
            self.form.is_shopping_discount = parseInt(data.data.values.is_shopping_discount);
            self.form.shopping_gift_rules = data.data.values.shopping_gift_rules ?? [
              {
                is_open: '1',
                is_member_level_related: '0',
                value: '',
                payment_amount_requirement: '',
                meal_type: [],
                balance_payment_get_points: '1',
                refund_return_points: '0',
                member_levels: [],
              },
              {
                is_open: '0',
                is_member_level_related: '0',
                value: '',
                payment_amount_requirement: '',
                meal_type: [],
                balance_payment_get_points: '1',
                refund_return_points: '0',
                member_levels: [],
              },
            ];
            self.form.exchange = data.data.values.exchange ?? {};

            // 初始化会员等级数据（如果后端有提供会员等级列表）
            if (data.data.member_levels) {
              data.data.member_levels.forEach((level) => {
                // 为每个规则初始化会员等级数据
                self.form.shopping_gift_rules.forEach((rule) => {
                  if (!rule.member_levels) rule.member_levels = [];
                  const existingLevel = rule.member_levels.find((ml) => ml.id === level.id);
                  if (!existingLevel) {
                    rule.member_levels.push({
                      id: level.id,
                      name: level.name,
                      value: '',
                    });
                  }
                });
              });
            }
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*保存*/
      onSubmit() {
        let self = this;
        // 判断是否开启自助餐,只能有一个自助餐
        if (self.form.shopping_gift_rules[1].is_open == 1 && self.form.shopping_gift_rules[0].meal_type.includes('buffet')) {
          this.$ElMessage({
            message: this.$t('自助餐只可适用于一个规则'),
            type: 'error',
          });
          return;
        }
        let form = self.form;

        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            PointsApi.setPoints(form, true)
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('保存成功'),
                    type: 'success',
                  });
                } else {
                  self.loading = false;
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },
      handleDeductRatioMainInput(value) {
        const main = value > 100 ? 100 : value < 0 ? 0 : value;
        this.form.deduct_ratio_main = main;
        this.form.deduct_ratio_gift = 100 - main;
      },
      handleDeductRatioGiftInput(value) {
        const gift = value > 100 ? 100 : value < 0 ? 0 : value;
        this.form.deduct_ratio_gift = gift;
        this.form.deduct_ratio_main = 100 - gift;
      },
      handleDeductRatioChange() {
        const main = this.form.deduct_ratio_main > 100 ? 100 : this.form.deduct_ratio_main < 0 ? 0 : this.form.deduct_ratio_main;
        const gift = 100 - main;
        if (this.form.deduct_ratio_main !== main) {
          this.form.deduct_ratio_main = main;
        }
        if (this.form.deduct_ratio_gift !== gift) {
          this.form.deduct_ratio_gift = gift;
        }
      },

      handleValueInput(e, index) {
        let value = e;
        // 只允许输入数字和小数点
        value = value.replace(/[^0-9.]/g, '');

        // 确保只有一个小数点
        const parts = value.split('.');
        if (parts.length > 2) {
          value = parts[0] + '.' + parts.slice(1).join('');
        }

        // 限制最多两位小数
        if (parts[1] && parts[1].length > 2) {
          value = parts[0] + '.' + parts[1].substring(0, 2);
        }

        // 转换为数字进行范围检查
        const numValue = parseFloat(value);
        if (numValue > 100) {
          value = '100';
        }

        // 更新表单值
        this.form.shopping_gift_rules[index].value = value;
      },

      handleValuesInput(e, index) {
        let value = e;
        //只允许数字，1-9999999
        value = value.replace(/[^0-9.]/g, '');
        // 转换为数字进行范围检查
        const numValue = parseFloat(value);
        if (numValue > 9999999) {
          value = '9999999';
        }
        if (numValue < 1) {
          value = '1';
        }
        this.form.shopping_gift_rules[index].value = value;
      },

      handlePointsExchangeRateInput(e) {
        let value = e;

        // 只允许输入数字和小数点
        value = value.replace(/[^0-9.]/g, '');

        // 确保只有一个小数点
        const parts = value.split('.');
        if (parts.length > 2) {
          value = parts[0] + '.' + parts.slice(1).join('');
        }

        // 限制最多两位小数
        if (parts[1] && parts[1].length > 2) {
          value = parts[0] + '.' + parts[1].substring(0, 2);
        }

        // 转换为数字进行范围检查
        const numValue = parseFloat(value);
        if (numValue > 9999999) {
          value = '9999999';
        }
        if (numValue < 0) {
          value = '0';
        }
        this.form.exchange.points_exchange_rate = value;
      },
      handleMoneyInput(e, index) {
        let value = e;

        // 只允许输入数字和小数点
        value = value.replace(/[^0-9.]/g, '');

        // 确保只有一个小数点
        const parts = value.split('.');
        if (parts.length > 2) {
          value = parts[0] + '.' + parts.slice(1).join('');
        }

        // 限制最多两位小数
        if (parts[1] && parts[1].length > 2) {
          value = parts[0] + '.' + parts[1].substring(0, 2);
        }

        // 转换为数字进行范围检查
        const numValue = parseFloat(value);
        if (numValue > 9999999) {
          value = '9999999';
        }

        // 不允许输入0或负数
        if (numValue <= 0 && value !== '' && value !== '0.') {
          value = '';
        }

        // 更新表单值
        this.form.shopping_gift_rules[index].payment_amount_requirement = value;
      },

      // 处理会员等级积分输入
      handleMemberLevelInput(value, ruleIndex, levelIndex) {
        // 只允许输入数字
        let newValue = value.replace(/[^0-9]/g, '');

        // 限制最小值为1
        if (newValue <= 0 && value !== '' && value !== '0.') {
          newValue = '';
        }

        // 限制最大值
        const numValue = parseInt(newValue);
        if (numValue > 9999999) {
          newValue = '9999999';
        }

        // 更新对应的会员等级积分值
        this.form.shopping_gift_rules[ruleIndex].member_levels[levelIndex].value = newValue;
      },

      handleMemberLevelsInput(e, index, levelIndex) {
        let value = e;
        // 只允许输入数字和小数点
        value = value.replace(/[^0-9.]/g, '');

        // 确保只有一个小数点
        const parts = value.split('.');
        if (parts.length > 2) {
          value = parts[0] + '.' + parts.slice(1).join('');
        }

        // 限制最多两位小数
        if (parts[1] && parts[1].length > 2) {
          value = parts[0] + '.' + parts[1].substring(0, 2);
        }

        // 转换为数字进行范围检查
        const numValue = parseFloat(value);
        if (numValue > 100) {
          value = '100';
        }
        if (numValue < 0) {
          value = '0';
        }
        this.form.shopping_gift_rules[index].member_levels[levelIndex].value = value;
      },
    },
  };
</script>

<style lang="scss" scoped>
  .label-wrap {
    display: flex;
    align-items: center;
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
</style>
