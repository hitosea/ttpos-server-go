<template>
  <div class="p-4 bg-white rounded min-h-full">
    <el-form :model="formData" ref="formElement" label-position="top" label-width="124px">
      <div class="bg-[#fff] rounded p-4 mb-6 border border-[#dcdfe6]" v-for="(item, index) in formData" :key="index">
        <div class="text-lg font-bold mb-4">{{ item.channel }} {{ $t('基础设置') }}</div>
        <el-row :gutter="24">
          <el-col :span="8">
            <el-form-item
              :label="$t('外送基础服务费')"
              :prop="`${index}.basic_fee`"
              :rules="[
                {
                  required: true,
                  trigger: 'blur',
                  validator: (_rule: any, value: any, callback: any) => {
                    if (!value) {
                      callback(new Error($t('请输入外送基础服务费')));
                    } else {
                      callback();
                    }
                  },
                },
              ]"
            >
              <el-input-number
                class="!w-full"
                v-model="item.basic_fee"
                :controls="false"
                :precision="2"
                :min="0"
                :max="999999"
                :placeholder="$t('请输入外送基础服务费')"
              ></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item
              :label="$t('起步配送费')"
              :prop="`${index}.base_delivery_fee`"
              :rules="[
                {
                  required: true,
                  trigger: 'blur',
                  validator: (_rule: any, value: any, callback: any) => {
                    if (!value) {
                      callback(new Error($t('请输入起步配送费')));
                    } else {
                      callback();
                    }
                  },
                },
              ]"
            >
              <el-input-number
                class="!w-full"
                v-model="item.base_delivery_fee"
                :controls="false"
                :precision="2"
                :min="0"
                :max="999999"
                :placeholder="$t('请输入起步配送费')"
              ></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item
              :label="$t('骑手未接单取消时间')"
              :prop="`${index}.rider_acceptance_timeout`"
              :rules="[
                {
                  required: true,
                  trigger: 'blur',
                  validator: (_rule: any, value: any, callback: any) => {
                    if (!value) {
                      callback(new Error($t('请输入骑手未接单取消时间')));
                    } else {
                      callback();
                    }
                  },
                },
              ]"
            >
              <el-input-number
                class="!w-full"
                v-model="item.rider_acceptance_timeout"
                :controls="false"
                :precision="0"
                :min="1"
                :max="60"
                :placeholder="$t('请输入骑手未接单取消时间')"
              ></el-input-number>
              <div class="text-[#ccc] w-full">{{ $t('骑手未接单多少分钟后自动取消订单（1-60分钟）') }}</div>
            </el-form-item>
          </el-col>
        </el-row>
        <div class="text-lg font-bold mb-4">{{ $t('距离范围设置') }}</div>
        <!-- 距离范围设置 -->
        <el-form-item
          :prop="`${index}.distance_range`"
          :rules="[
            {
              required: true,
              trigger: 'blur',
              validator: (_rule: any, value: any, callback: any) => {
                if (value.length === 0) {
                  callback(new Error($t('请至少添加一个距离范围')));
                } else {
                  callback();
                }
              },
            },
          ]"
        >
          <div class="flex flex-col gap-4 w-full">
            <div class="w-full rounded p-4 bg-[#f9f9f9]" v-for="(distanceRange, distanceRangeIndex) in item.distance_range" :key="distanceRangeIndex">
              <el-row justify="space-between" align="middle" class="mb-4 mt-2">
                <h3 class="text-base font-bold">{{ $t('距离范围') }} {{ distanceRangeIndex + 1 }}</h3>
                <el-icon class="!mr-0 cursor-pointer text-gray-500 text-lg" @click="removeDistanceRange(index, distanceRangeIndex)"><Delete /></el-icon>
              </el-row>
              <el-row :gutter="24">
                <el-col :span="12">
                  <el-form-item
                    :label="$t('结束距离 (公里)')"
                    :prop="`${index}.distance_range.${distanceRangeIndex}.end`"
                    :rules="[
                      {
                        required: true,
                        trigger: 'blur',
                        validator: (_rule: any, value: any, callback: any) => {
                          // 检查是否为空
                          if (!value) {
                            callback(new Error($t('请输入结束距离')));
                            return;
                          }

                          // 如果不是第一个距离范围，检查是否大于前一个结束距离
                          if (distanceRangeIndex > 0) {
                            const prevEnd = formData[index]?.distance_range?.[distanceRangeIndex - 1]?.end;
                            if (prevEnd !== undefined && Number(value) <= Number(prevEnd) && !distanceRange.is_unlimited) {
                              callback(new Error($t('结束距离必须大于上一个结束距离')));
                              return;
                            }
                          }
                          // 如果是最后一个距离范围，则提示必须勾选最大
                          if (distanceRangeIndex === (formData[index].distance_range && formData[index].distance_range?.length - 1)) {
                            callback(new Error($t('最后一个距离范围必须勾选最大')));
                            return;
                          }
                          callback();
                        },
                      },
                    ]"
                  >
                    <div class="flex items-center gap-2 !w-full">
                      <el-input-number
                        class="flex-1"
                        v-model="distanceRange.end"
                        :controls="false"
                        :precision="1"
                        :min="distanceRangeIndex === 0 ? 0 : (formData[index].distance_range && formData[index].distance_range[distanceRangeIndex - 1].end) || 0"
                        :max="999999"
                        :disabled="distanceRange.is_unlimited"
                        :placeholder="$t('请输入结束距离')"
                      ></el-input-number>
                      <el-checkbox
                        v-model="distanceRange.is_unlimited"
                        :label="$t('最大')"
                        @change="handleIsUnlimitedChange(index, distanceRangeIndex)"
                        v-if="distanceRangeIndex === (formData[index].distance_range && formData[index].distance_range?.length - 1)"
                      />
                    </div>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item
                    :label="$t('单价')"
                    :prop="`${index}.distance_range.${distanceRangeIndex}.price_per_km`"
                    :rules="[
                      {
                        required: true,
                        trigger: 'blur',
                        validator: (_rule: any, value: any, callback: any) => {
                          if (!value) {
                            callback(new Error($t('请输入单价')));
                          } else {
                            callback();
                          }
                        },
                      },
                    ]"
                  >
                    <el-input-number
                      class="!w-full"
                      v-model="distanceRange.price_per_km"
                      :controls="false"
                      :precision="2"
                      :min="0"
                      :max="999999"
                      :placeholder="$t('请输入单价')"
                    ></el-input-number>
                  </el-form-item>
                </el-col>
              </el-row>
            </div>
          </div>
        </el-form-item>
        <el-button
          class="w-full"
          type="primary"
          :disabled="item.distance_range?.map((distanceRange) => distanceRange.is_unlimited).includes(true)"
          plain
          @click="addDistanceRange(index)"
          icon="plus"
          >{{ $t('添加范围距离') }}</el-button
        >
      </div>
    </el-form>
    <div class="border-t border-[#eee] flex items-center justify-center p-4 mt-4">
      <el-button @click="handleReset">{{ $t('重置') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit">{{ $t('保存') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, nextTick } from 'vue';
  import { $t } from '@/i18n';
  import { getTakeoutSetting, fetchTakeoutSetting, type TakeoutSettingData } from '@/api/takeout/setting';
  import { message } from '@/utils/feedback';
  const formData = ref<TakeoutSettingData[]>([{}]);
  const formLoading = ref(false);

  const formElement = ref();

  const handleSubmit = () => {
    formElement.value.validate(async (valid: boolean) => {
      if (valid) {
        try {
          formLoading.value = true;
          const data = {
            channels: <TakeoutSettingData[]>[],
          };
          formData.value.forEach((item) => {
            data.channels.push({
              channel: item.channel,
              basic_fee: item.basic_fee,
              base_delivery_fee: item.base_delivery_fee,
              rider_acceptance_timeout: item.rider_acceptance_timeout,
              distance_range: item.distance_range,
            });
          });
          const res = await fetchTakeoutSetting(data);
          if (res.code === 1) {
            message.success($t('保存成功'));
          }
        } catch (error) {
          console.error('保存设置失败:', error);
        } finally {
          formLoading.value = false;
        }
      } else {
        // 验证失败页面滑动到错误位置
        nextTick(() => {
          const errorField = document.querySelector('.is-error');
          if (errorField) {
            errorField.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        });
      }
    });
  };

  const getTakeoutSettingData = async () => {
    try {
      formLoading.value = true;
      const res = await getTakeoutSetting();
      formData.value = res?.data || [];
    } catch (error) {
      console.error('获取设置失败:', error);
    } finally {
      formLoading.value = false;
    }
  };

  const handleReset = () => {
    formElement.value.resetFields();
    getTakeoutSettingData();
  };

  // 添加距离范围
  const addDistanceRange = (index: number) => {
    formData.value[index].distance_range?.push({
      end: 0,
      price_per_km: 0,
      is_unlimited: false,
    });
    // 所有的is_unlimited设置为false
    formData.value[index].distance_range?.forEach((item) => {
      item.is_unlimited = false;
    });
    // 校验距离范围表单
    formElement.value.validateField(`${index}.distance_range`);
  };

  // 删除距离范围
  const removeDistanceRange = (index: number, distanceRangeIndex: number) => {
    formData.value[index].distance_range?.splice(distanceRangeIndex, 1);
    // 校验距离范围表单
    formElement.value.validateField(`${index}.distance_range`);
  };

  // 处理is_unlimited变化
  const handleIsUnlimitedChange = (index: number, distanceRangeIndex: number) => {
    if (!formData.value[index].distance_range || !formData.value[index].distance_range[distanceRangeIndex]) return;
    const distanceRange = formData.value[index].distance_range[distanceRangeIndex];
    distanceRange.end = distanceRange.is_unlimited ? 999999 : 0;
  };
  onMounted(() => {
    getTakeoutSettingData();
  });
</script>
