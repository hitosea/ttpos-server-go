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
        <!-- <div class="lh18 mt10 gray9">
          <p>注：修改积分名称后，在买家端的所有页面里，看到的都是自定义的名称</p>
          <p>例：商家使用自定义的积分名称来做品牌运营。如京东把积分称为“京豆”，淘宝把积分称为“淘金币”</p>
        </div> -->
      </el-form-item>
      <!-- <el-form-item :label="$t('积分说明')" :rules="[{required: true,message: ' '}]">
        <el-input type="textarea" rows="5" v-model="form.describe" autocomplete="off"></el-input>
      </el-form-item> -->
      <!-- <div class="common-form">{{ $t('积分赠送') }}</div> -->
      <el-form-item :label="$t('购物送积分')">
        <el-radio-group v-model="form.is_shopping_gift">
          <el-radio :label="1">{{ $t('开启') }}</el-radio>
          <el-radio :label="0">{{ $t('关闭') }}</el-radio>
        </el-radio-group>
        <div class="lh18 mt10 gray9">
          <p>{{ $t('打开后订单完成后赠送用户积分') }}</p>
          <p>{{ $t('注：退款后已赠送的订单积分对应扣除') }}</p>
        </div>
      </el-form-item>
      <el-form-item v-if="form.is_shopping_gift == 1" :label="$t('积分赠送比例')" prop="gift_ratio" :rules="[{ required: true, message: ' ' }]">
        <el-input-number :controls="false" class="max-w460" :min="0" :max="100" :placeholder="$t('请输入内容')" v-model.number="form.gift_ratio"></el-input-number>
        <span>%</span>
        <div class="lh18 mt10 gray9">
          <p> {{ $t('注：请填写数字0~100；') }}</p>
          <p> {{ $t('例：订单付款金额(100) * 积分赠送比例(100%) = 实际赠送的积分(100积分)') }}</p>
        </div>
      </el-form-item>
      <!-- <div class="common-form">积分抵扣</div>
      <el-form-item label=" 是否允许下单使用积分抵扣 ">
        <el-radio-group v-model="form.is_shopping_discount" class="max-w460">
          <el-radio :label="1">允许</el-radio>
          <el-radio :label="0">不允许</el-radio>
        </el-radio-group>
        <div class="lh18 mt10 gray9">
          <p>注：如开启则用户下单时可选择使用积分抵扣订单金额</p>
        </div>
      </el-form-item>
      <el-form-item label=" 积分抵扣比例">
        <el-input placeholder="请输入内容" v-model="form.discount.discount_ratio" class="max-w460">
          <template #prepend>1个积分可抵扣</template>
<template #append>元</template>
</el-input>
<div class="lh18 mt10 gray9">
  <p>例如：1积分可抵扣0.01元，100积分则可抵扣1元，1000积分则可抵扣10元</p>
</div>
</el-form-item>
<el-form-item label=" 抵扣条件">
  <el-input placeholder="请输入内容" v-model="form.discount.full_order_price" class="max-w460">
    <template #prepend>订单满</template>
    <template #append>元</template>
  </el-input>
</el-form-item>
<el-form-item label=" ">
  <el-input placeholder="请输入内容" v-model="form.discount.max_money_ratio" class="max-w460">
    <template #prepend>最高可抵扣金额</template>
    <template #append>%</template>
  </el-input>
  <div class="lh18 mt10 gray9">
    <p>温馨提示：例如订单金额100元，最高可抵扣10%，此时用户可抵扣10元</p>
  </div>
</el-form-item> -->
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
  export default {
    data() {
      return {
        form: {
          deduct_order: 1,
          deduct_ratio_main: 100,
          deduct_ratio_gift: 0,
          is_shopping_gift: 1,
          gift_ratio: 10,
          is_shopping_discount: 1,
          discount: {
            discount_ratio: 0,
            full_order_price: 0,
            max_money_ratio: 0,
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
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*保存*/
      onSubmit() {
        let self = this;
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
