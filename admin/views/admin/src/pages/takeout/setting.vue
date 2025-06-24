<template>
  <div class="p-4 bg-white rounded min-h-full">
    <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="124px">
      <div class="text-lg font-bold mb-4">{{ $t('选择外送渠道') }}</div>
      <el-form-item :label="$t('Token')" prop="takeout">
        <el-select class="max-width" v-model="formData.takeout" :placeholder="$t('选择外送渠道')">
          <el-option value="1" :label="'SKootar' + $t('（默认选中）')" />
          <el-option value="2" :label="'Grab'" />
        </el-select>
      </el-form-item>
      <div class="text-lg font-bold mb-4 mt-12">{{ $t('基础设置') }}</div>
      <el-row :gutter="24">
        <el-col :span="8">
          <el-form-item :label="$t('外送基础服务费')" prop="takeout">
            <el-input-number
              class="!w-full"
              v-model="formData.base_service_fee"
              :controls="false"
              :precision="2"
              :min="0"
              :max="999999"
              :placeholder="$t('请输入外送基础服务费')"
            ></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="$t('起步配送费')" prop="takeout">
            <el-input-number
              class="!w-full"
              v-model="formData.starting_fee"
              :controls="false"
              :precision="2"
              :min="0"
              :max="999999"
              :placeholder="$t('请输入起步配送费')"
            ></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="$t('骑手未接单取消时间')" prop="takeout">
            <el-input-number
              class="!w-full"
              v-model="formData.time"
              :controls="false"
              :precision="0"
              :min="0"
              :max="999999"
              :placeholder="$t('请输入骑手未接单取消时间')"
            ></el-input-number>
            <div class="text-[#ccc] w-full">{{ $t('骑手未接单多少分钟后自动取消订单（1-60分钟）') }}</div>
          </el-form-item>
        </el-col>
      </el-row>
      <div class="text-lg font-bold mb-4">{{ $t('距离范围设置') }}</div>
      <!-- 距离范围设置 -->
      <div class="flex flex-col gap-4">
        <el-card class="w-full" shadow="hover" v-for="(item, index) in formData.distance_range_setting" :key="index">
          <el-row justify="space-between" align="middle" class="mb-4 mt-2">
            <h3 class="text-base font-bold">{{ $t('距离范围设置') }}</h3>
            <el-button class="!mr-0" type="danger" @click="removeDistanceRange(index)">{{ $t('删除') }}</el-button>
          </el-row>
          <el-row :gutter="24">
            <el-col :span="8">
              <el-form-item :label="$t('起始距离 (公里)')" prop="takeout">
                <el-input-number
                  class="!w-full"
                  v-model="item.starting_distance"
                  :controls="false"
                  :precision="1"
                  :min="0"
                  :max="999999"
                  :placeholder="$t('请输入起始距离 (公里)')"
                ></el-input-number>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('结束距离 (公里)')" prop="takeout">
                <div class="flex items-center gap-2 !w-full">
                  <el-input-number
                    class="flex-1"
                    v-model="item.ending_distance"
                    :controls="false"
                    :precision="1"
                    :min="0"
                    :max="999999"
                    :placeholder="$t('请输入结束距离 (公里)')"
                  ></el-input-number>
                  <el-checkbox v-model="checked1" :label="$t('最大')" />
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item :label="$t('单价 (元/公里)')" prop="takeout">
                <el-input-number
                  class="!w-full"
                  v-model="item.unit_price"
                  :controls="false"
                  :precision="2"
                  :min="0"
                  :max="999999"
                  :placeholder="$t('请输入单价 (元/公里)')"
                ></el-input-number>
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </div>
      <el-button class="mt-4" @click="addDistanceRange" icon="plus">{{ $t('添加范围距离') }}</el-button>
    </el-form>
    <div class="border-t border-[#eee] flex items-center justify-center p-4 mt-4">
      <el-button @click="handleReset">{{ $t('重置') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit">{{ $t('保存') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { $t } from '@/i18n';
  const formData = ref({
    takeout: '1',
    base_service_fee: 0,
    starting_fee: 0,
    time: 0,
    distance_range_setting: [
      {
        starting_distance: 0,
        ending_distance: 0,
        unit_price: 0,
      },
    ],
  });
  const formLoading = ref(false);

  const checked1 = ref(false);

  const formRules = ref({});

  const formElement = ref();

  const handleSubmit = () => {
    formElement.value.validate((valid: boolean) => {
      if (valid) {
        console.log(formData.value);
      }
    });
  };

  const handleReset = () => {
    formElement.value.resetFields();
  };

  const addDistanceRange = () => {
    formData.value.distance_range_setting.push({
      starting_distance: 0,
      ending_distance: 0,
      unit_price: 0,
    });
  };

  const removeDistanceRange = (index: number) => {
    formData.value.distance_range_setting.splice(index, 1);
  };
</script>
