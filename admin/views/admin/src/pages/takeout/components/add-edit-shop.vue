<template>
  <el-dialog
    width="960"
    :title="hasEdit ? $t('编辑商家外送渠道设置') : $t('新增商家外送渠道设置')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="handleClose()"
  >
    <div class="max-h-[75vh] overflow-auto pr-4">
      <el-form :model="formData" ref="formElement" label-position="top" label-width="auto">
        <el-form-item :label="$t('选择商家')" prop="uuid" :rules="[{ required: true, message: $t('请选择商家'), trigger: 'blur' }]">
          <el-select v-model="formData.uuid" type="text" filterable remote :remote-method="remoteMethod" :disabled="hasEdit" :placeholder="$t('请选择商家')">
            <el-option class="!h-auto" :disabled="item.channel_names.length > 0" v-for="item in companyList" :key="item.uuid" :value="item.uuid" :label="item.name">
              <div>
                <p class="">{{ item.name }}</p>
                <p class="text-xs text-[#999]" :class="item.channel_names.length == 0 ? 'mb-2' : ''">ID: {{ item.uuid }}</p>
                <p class="text-xs text-[#999] mb-2" v-if="item.channel_names.length > 0">{{ $t('已添加，渠道：') }} {{ item.channel_names.join(',') }}</p>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('外送渠道')" prop="takeout" :rules="[{ required: true, message: $t('请选择外送渠道'), trigger: 'blur' }]">
          <el-checkbox-group v-model="formData.takeout" @change="handleTakeoutChange">
            <el-checkbox v-for="item in channels" :key="item.value" :value="item.value" :label="$t(item.name)" />
          </el-checkbox-group>
        </el-form-item>
        <div class="flex flex-col gap-4">
          <div class="rounded p-4 bg-[#fff] border border-[#dcdfe6]" v-for="(item, index) in formData.channels" :key="item.channel">
            <h3 class="text-base font-bold mb-4">{{ item.channel }} {{ $t('渠道参数设置') }}</h3>
            <el-form-item :label="$t('参数同步方式')" prop="takeout">
              <el-radio-group v-model="item.config_type" @change="handleConfigTypeChange(index, item.channel)">
                <el-radio value="auto_sync" :label="$t('自动同步默认')" />
                <el-radio value="manual" :label="$t('手动设置')" />
              </el-radio-group>
            </el-form-item>
            <el-row :gutter="24">
              <el-col :span="8">
                <el-form-item
                  :label="$t('外送基础服务费')"
                  :prop="'channels.' + index + '.basic_fee'"
                  :rules="[{ required: true, message: $t('请输入外送基础服务费'), trigger: 'blur' }]"
                >
                  <el-input-number
                    class="!w-full"
                    v-model="item.basic_fee"
                    :controls="false"
                    :disabled="item.config_type === 'auto_sync'"
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
                  :prop="'channels.' + index + '.base_delivery_fee'"
                  :rules="[{ required: true, message: $t('请输入起步配送费'), trigger: 'blur' }]"
                >
                  <el-input-number
                    class="!w-full"
                    v-model="item.base_delivery_fee"
                    :controls="false"
                    :disabled="item.config_type === 'auto_sync'"
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
                  :prop="'channels.' + index + '.rider_acceptance_timeout'"
                  :rules="[{ required: true, message: $t('请输入骑手未接单取消时间'), trigger: 'blur' }]"
                >
                  <el-input-number
                    class="!w-full"
                    v-model="item.rider_acceptance_timeout"
                    :controls="false"
                    :disabled="item.config_type === 'auto_sync'"
                    :precision="0"
                    :min="0"
                    :max="999999"
                    :placeholder="$t('请输入骑手未接单取消时间')"
                  ></el-input-number>
                  <div class="text-[#ccc] w-full text-xs mt-1">{{ $t('骑手未接单多少分钟后自动取消订单（1-60分钟）') }}</div>
                </el-form-item>
              </el-col>
            </el-row>

            <h3 class="text-base font-bold mb-4">{{ $t('距离范围设置') }}</h3>
            <!-- 距离范围设置 -->
            <div class="flex flex-col">
              <el-form-item
                :prop="'channels.' + index + '.distance_range'"
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
                <div class="flex flex-col gap-3 w-full">
                  <div class="w-full rounded p-4 bg-[#f9f9f9]" v-for="(distanceRange, distanceRangeIndex) in item.distance_range" :key="distanceRangeIndex">
                    <el-row justify="space-between" align="middle" class="mb-4 mt-2">
                      <h3 class="text-base font-bold">{{ $t('距离范围') }} {{ distanceRangeIndex + 1 }}</h3>
                      <el-icon v-if="item.config_type === 'manual'" class="!mr-0 cursor-pointer text-gray-500 text-lg" @click="removeDistanceRange(index, distanceRangeIndex)">
                        <Delete />
                      </el-icon>
                    </el-row>
                    <el-row :gutter="24">
                      <el-col :span="12">
                        <el-form-item
                          :label="$t('结束距离 (公里)')"
                          :prop="'channels.' + index + '.distance_range.' + distanceRangeIndex + '.end'"
                          :rules="[
                            {
                              required: true,
                              trigger: 'blur',
                              validator: (_rule: any, value: any, callback: any) => {
                                if (!value) {
                                  callback(new Error($t('请输入结束距离')));
                                }
                                if (distanceRangeIndex > 0 && value <= item.distance_range[distanceRangeIndex - 1].end && !distanceRange.is_unlimited) {
                                  callback(new Error($t('结束距离必须大于前一个范围的结束距离')));
                                } else {
                                  callback();
                                }
                              },
                            },
                          ]"
                        >
                          <div class="flex items-center gap-2 !w-full">
                            <el-input-number
                              class="flex-1"
                              v-model="distanceRange.end"
                              :controls="false"
                              :disabled="distanceRange.is_unlimited || item.config_type === 'auto_sync'"
                              :precision="1"
                              :min="distanceRangeIndex === 0 ? 0 : item.distance_range[distanceRangeIndex - 1].end || 0"
                              :max="999999"
                              :placeholder="$t('请输入结束距离 (公里)')"
                            ></el-input-number>
                            <el-checkbox
                              v-if="distanceRangeIndex === item.distance_range.length - 1"
                              v-model="distanceRange.is_unlimited"
                              :disabled="item.config_type === 'auto_sync'"
                              :label="$t('最大')"
                              @change="handleIsUnlimitedChange(index, distanceRangeIndex)"
                            />
                          </div>
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item
                          :label="$t('单价')"
                          :prop="'channels.' + index + '.distance_range.' + distanceRangeIndex + '.price_per_km'"
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
                            :disabled="item.config_type === 'auto_sync'"
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
                v-if="item.config_type === 'manual'"
                :disabled="item.distance_range?.map((distanceRange) => distanceRange.is_unlimited).includes(true)"
                type="primary"
                plain
                class="!mr-0"
                @click="addDistanceRange(index)"
                icon="plus"
                >{{ $t('添加范围距离') }}</el-button
              >
            </div>
          </div>
        </div>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleClose()">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleSubmit()" :loading="formLoading">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick } from 'vue';
  import { $t } from '@/i18n';
  import { addEditCompany, getCompanySelect, type TakeoutShopSelectData, getChannels, type TakeoutShopData, type TakeoutShopAddEditDataItem } from '@/api/takeout/shop';
  import { getTakeoutSetting, type TakeoutSettingData } from '@/api/takeout/setting';
  import { message } from '@/utils/feedback';
  const emits = defineEmits(['update:show', 'getList']);

  const props = defineProps({
    show: {
      type: Boolean,
      default: false,
    },
    hasEdit: {
      type: Boolean,
      default: false,
    },
    editData: {
      type: Object,
      default: () => ({}),
    },
  });

  const companyList = ref<TakeoutShopSelectData[]>([]);
  const formData = ref({
    uuid: '',
    takeout: [],
    channels: [] as TakeoutShopAddEditDataItem[],
  });
  const formElement = ref();
  const channels = ref<TakeoutShopData[]>([]);
  const takeoutSettingData = ref<TakeoutSettingData[]>([]);
  const formLoading = ref(false);

  watch(
    () => props.show,
    () => {
      if (props.show) {
        // getCompanyListData('');
        getChannelsData();
        getTakeoutSettingData();
        if (props.hasEdit) {
          formData.value = {
            uuid: props.editData.uuid,
            takeout: props.editData.channel_names,
            channels: props.editData.delivery_config,
          };
        }
      }
    },
  );

  const handleClose = () => {
    clearForm();
    emits('update:show', false);
  };

  // 清空表单
  const clearForm = () => {
    setTimeout(() => {
      formElement.value.resetFields();
      formData.value = {
        uuid: '',
        takeout: [],
        channels: [] as TakeoutShopAddEditDataItem[],
      };
    }, 300);
  };

  // 提交表单
  const handleSubmit = () => {
    formElement.value.validate(async (valid: boolean) => {
      if (valid) {
        try {
          formLoading.value = true;
          await addEditCompany({
            uuid: Number(formData.value.uuid),
            channels: formData.value.channels,
          });
          handleClose();
          emits('getList');
          message.success($t('保存成功'));
        } catch (error) {
          console.error('添加失败:', error);
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

  // 删除距离范围
  const removeDistanceRange = (index: number, distanceRangeIndex: number) => {
    formData.value.channels[index].distance_range.splice(distanceRangeIndex, 1);
    formElement.value.validateField('channels.' + index + '.distance_range');
  };

  // 添加距离范围
  const addDistanceRange = (index: number) => {
    formData.value.channels[index].distance_range.push({
      end: 0,
      price_per_km: 0,
      is_unlimited: false,
    });
    // 所有的is_unlimited设置为false
    formData.value.channels[index].distance_range?.forEach((item) => {
      item.is_unlimited = false;
    });
    formElement.value.validateField('channels.' + index + '.distance_range');
  };

  // 远程搜索
  const remoteMethod = (keyword: string) => {
    setTimeout(() => {
      getCompanyListData(keyword);
    }, 200);
  };

  // 获取商家列表
  const getCompanyListData = async (keyword: string) => {
    try {
      const res = await getCompanySelect({ keyword: keyword, list_rows: 50 });
      companyList.value = res.data?.list?.data || [];
    } catch (error) {
      console.error('获取渠道失败:', error);
    }
  };

  // 获取渠道列表
  const getChannelsData = async () => {
    try {
      const res = await getChannels({ configured: 1 });
      channels.value = res.data || [];
    } catch (error) {
      console.error('获取渠道失败:', error);
    }
  };

  // 获取设置列表
  const getTakeoutSettingData = async () => {
    try {
      const res = await getTakeoutSetting();
      takeoutSettingData.value = res?.data || [];
    } catch (error) {
      console.error('获取设置失败:', error);
    }
  };

  // 外送渠道变化
  const handleTakeoutChange = (val: string[]) => {
    // 以val为基准，对比formData.value.channels，如果存在，就从formData.value.channels中删除
    formData.value.channels = formData.value.channels.filter((item) => val.includes(item.channel));
    // formData.value.channels 如果不存在，从takeoutSettingData.value中获取添加
    val.forEach((item) => {
      if (!formData.value.channels.some((channel) => channel.channel === item)) {
        const settingData = JSON.parse(JSON.stringify(takeoutSettingData.value.find((setting) => setting.channel === item)));
        formData.value.channels.push({
          channel: item,
          config_type: 'auto_sync',
          basic_fee: settingData?.basic_fee || 0,
          base_delivery_fee: settingData?.base_delivery_fee || 0,
          rider_acceptance_timeout: settingData?.rider_acceptance_timeout || 0,
          distance_range: settingData?.distance_range || [],
        });
      }
    });
  };

  // 参数设置变化
  const handleConfigTypeChange = (index: number, channel: string) => {
    // 如果config_type为auto_sync，distance_range设置为takeoutSettingData.value中对应channel的distance_range
    if (formData.value.channels[index].config_type === 'auto_sync') {
      console.log(takeoutSettingData.value.find((setting) => setting.channel === channel)?.distance_range);
      formData.value.channels[index] = JSON.parse(JSON.stringify(takeoutSettingData.value.find((setting) => setting.channel === channel)));
      formData.value.channels[index].config_type = 'auto_sync';
    }
  };

  // 处理is_unlimited变化
  const handleIsUnlimitedChange = (index: number, distanceRangeIndex: number) => {
    if (!formData.value.channels[index].distance_range || !formData.value.channels[index].distance_range[distanceRangeIndex]) return;
    const distanceRange = formData.value.channels[index].distance_range[distanceRangeIndex];
    distanceRange.end = distanceRange.is_unlimited ? 999999 : 0;
  };
</script>
