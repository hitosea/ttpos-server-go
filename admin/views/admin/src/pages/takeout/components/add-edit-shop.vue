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
      <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="auto">
        <el-form-item :label="$t('选择商家')" prop="name">
          <el-select v-model="formData.name" type="text" filterable maxlength="100" clearable :placeholder="$t('请选择商家')">
            <el-option value="1" label="商家1" />
            <el-option value="2" label="商家2" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('外送渠道')" prop="takeout">
          <el-checkbox-group v-model="formData.takeout">
            <el-checkbox value="1" :label="$t('SKootar')" />
            <el-checkbox value="2" :label="$t('Grab')" />
          </el-checkbox-group>
        </el-form-item>
        <div class="rounded p-4 border border-[#dcdfe6]">
          <h3 class="text-base font-bold mb-4">{{ $t('SKootar 渠道参数设置') }}</h3>
          <el-form-item :label="$t('参数同步方式')" prop="takeout">
            <el-radio-group v-model="formData.takeout">
              <el-radio value="1" :label="$t('自动同步默认')" />
              <el-radio value="2" :label="$t('手动设置')" />
            </el-radio-group>
          </el-form-item>
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
                <div class="text-[#ccc] w-full text-xs mt-1">{{ $t('骑手未接单多少分钟后自动取消订单（1-60分钟）') }}</div>
              </el-form-item>
            </el-col>
          </el-row>

          <h3 class="text-base font-bold mb-4">{{ $t('距离范围设置') }}</h3>
          <!-- 距离范围设置 -->
          <div class="flex flex-col gap-4">
            <el-card class="w-full" shadow="hover" v-for="(item, index) in formData.distance_range_setting" :key="index">
              <el-row justify="space-between" align="middle" class="mb-4 mt-2">
                <h3 class="text-base font-bold">{{ $t('距离范围1') }}</h3>
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
            <el-button class="!mr-0" @click="addDistanceRange" icon="plus">{{ $t('添加范围距离') }}</el-button>
          </div>
        </div>

        <div class="rounded p-4 border border-[#dcdfe6] mt-4">
          <h3 class="text-base font-bold mb-4">{{ $t('Grab 渠道参数设置') }}</h3>
          <el-form-item :label="$t('参数同步方式')" prop="takeout">
            <el-radio-group v-model="formData.takeout">
              <el-radio value="1" :label="$t('自动同步默认')" />
              <el-radio value="2" :label="$t('手动设置')" />
            </el-radio-group>
          </el-form-item>
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
                <div class="text-[#ccc] w-full text-xs mt-1">{{ $t('骑手未接单多少分钟后自动取消订单（1-60分钟）') }}</div>
              </el-form-item>
            </el-col>
          </el-row>

          <h3 class="text-base font-bold mb-4">{{ $t('距离范围设置') }}</h3>
          <!-- 距离范围设置 -->
          <div class="flex flex-col gap-4">
            <el-card class="w-full" shadow="hover" v-for="(item, index) in formData.distance_range_setting" :key="index">
              <el-row justify="space-between" align="middle" class="mb-4 mt-2">
                <h3 class="text-base font-bold">{{ $t('距离范围1') }}</h3>
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
            <el-button class="!mr-0" @click="addDistanceRange" icon="plus">{{ $t('添加范围距离') }}</el-button>
          </div>
        </div>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleClose()">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { $t } from '@/i18n';
  const formData = ref({
    name: '',
    takeout: [],
    distance_range_setting: [
      {
        starting_distance: 0,
        ending_distance: 0,
        unit_price: 0,
      },
    ],
  });
  const formRules = ref({
    name: [{ required: true, message: $t('请选择商家'), trigger: 'blur' }],
  });
  const formElement = ref();
  const props = defineProps({
    show: {
      type: Boolean,
      default: false,
    },
    hasEdit: {
      type: Boolean,
      default: false,
    },
  });

  const handleClose = () => {
    emits('update:show', false);
  };

  const handleSubmit = () => {
    formElement.value.validate((valid: boolean) => {
      if (valid) {
        console.log(formData.value);
      }
    });
  };

  const removeDistanceRange = (index: number) => {
    formData.value.distance_range_setting.splice(index, 1);
  };

  const addDistanceRange = () => {
    formData.value.distance_range_setting.push({
      starting_distance: 0,
      ending_distance: 0,
      unit_price: 0,
    });
  };

  const emits = defineEmits(['update:show']);
</script>
