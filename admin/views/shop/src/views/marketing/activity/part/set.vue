<template>
  <div class="basic-setting-content pl16 pr16">
    <div class="common-form">{{ $t('规则设置') }}</div>
    <el-form-item for="no_click" :label="$t('奖励条件')" :rules="[{ required: true, message: $t('请输入有效期') }]" prop="model.expire">
      <el-radio-group v-model="form.model.status">
        <el-radio :label="0"
          >{{ $t('推荐会员入会并消费') }}
          <numInput class="w-267" width="m-full" :min="0" :precision="2" v-model:valueData="form.service_charge" :value="form.service_charge" :placeholder="$t('请输入')"></numInput
        ></el-radio>
      </el-radio-group>

      <div class="gray9">{{ $t('注：老会员推荐好友入会并消费，好友再邀请的新会员消费，也能为老会员累计奖励。') }}</div>
    </el-form-item>
    <el-form-item for="no_click" :label="$t('奖励限制')">
      <el-checkbox-group v-model="form.model.limit_type">
        <el-checkbox :label="$t('限定每个会员获得奖励次数')" :value="1">
          {{ $t('限定每个会员获得奖励次数') }}
        </el-checkbox>
      </el-checkbox-group>
      <div class="gray9">{{ $t('为了避免会员滥用活动获取不正当利益，建议勾选限定次数以规避潜在风险。') }}</div>
    </el-form-item>

    <el-form-item for="no_click" :label="$t('次数限制')" :rules="[{ required: true, message: $t('请输入次数限制') }]" prop="model.content">
      <el-radio-group v-model="form.model.status" class="flex-box">
        <el-radio :label="0"
          >{{ $t('限制活动有效期内每个会员最多获取奖品') }}
          <el-input-number :controls="false" :min="0" :max="999" class="max-w140" :placeholder="$t('请输入次数限制')" v-model.number="form.model.content"></el-input-number>
          <div class="text-gray9">{{ $t('次') }}</div>
        </el-radio>
      </el-radio-group>
    </el-form-item>
  </div>
</template>

<script>
  export default {
    data() {
      return {};
    },
    created() {},
    inject: ['form'],
    methods: {},
  };
</script>

<style lang="scss" scoped>
  .w-267 {
    width: 267px;
    margin-left: 40px;
  }
  .max-w140 {
    max-width: 140px;
    margin-left: 40px;
  }
  .text-gray9 {
    display: inline-block;
    color: #999;
    font-size: 12px;
    margin-left: 8px;
  }
  .flex-box {
    display: flex;
    flex-direction: column;
    align-items: start;
    gap: 8px;
    :deep(.el-radio.el-radio--small) {
      height: 30px;
    }
  }
</style>
