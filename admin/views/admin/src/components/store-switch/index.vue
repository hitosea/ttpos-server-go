<template>
  <div class="store-switch-component">
    <!-- 门店选择下拉框 -->
    <el-select v-model="selectedCompanyUuid" :placeholder="$t('请选择门店')" :loading="loading" :disabled="disabled" @change="handleCompanyChange" clearable>
      <el-option v-for="company in companyList" :key="company.company_uuid" :value="company.company_uuid" :label="company.company_name">
        <div class="flex items-center justify-between">
          <span>{{ company.company_name }}</span>
          <el-tag v-if="company.is_default" type="success" size="small">{{ $t('默认') }}</el-tag>
        </div>
      </el-option>
    </el-select>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, watch } from 'vue';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';
  import { $post } from '@/api/request';

  const props = withDefaults(
    defineProps<{
      modelValue?: number; // 当前选中的门店UUID
      disabled?: boolean; // 是否禁用
      autoLoad?: boolean; // 是否自动加载门店列表
    }>(),
    {
      modelValue: undefined,
      disabled: false,
      autoLoad: true,
    },
  );

  const emits = defineEmits<{
    (e: 'update:modelValue', value: number | undefined): void;
    (e: 'change', value: number | undefined): void;
  }>();

  const selectedCompanyUuid = ref<number | undefined>(props.modelValue);
  const companyList = ref<Array<{ company_uuid: number; company_name: string; is_default?: boolean }>>([]);
  const loading = ref(false);

  // 获取门店列表
  const fetchCompanyList = async () => {
    try {
      loading.value = true;
      // 调用获取门店列表 API（新管理端）
      const res = await $post<{ list: Array<{ company_uuid: number; company_name: string; is_default?: boolean }> }>(
        '/auth/company_list',
        {},
        {
          baseURL: '/api/v1/',
        },
      );
      companyList.value = res.data?.list || [];

      // 如果没有选中值，且列表中有默认门店，则选中默认门店
      if (!selectedCompanyUuid.value && companyList.value.length > 0) {
        const defaultCompany = companyList.value.find((item) => item.is_default);
        if (defaultCompany) {
          selectedCompanyUuid.value = defaultCompany.company_uuid;
          emits('update:modelValue', defaultCompany.company_uuid);
        } else {
          // 如果没有默认门店，选中第一个
          selectedCompanyUuid.value = companyList.value[0].company_uuid;
          emits('update:modelValue', companyList.value[0].company_uuid);
        }
      }
    } catch (error) {
      console.error('获取门店列表失败:', error);
    } finally {
      loading.value = false;
    }
  };

  // 门店切换
  const handleCompanyChange = async (companyUuid: number) => {
    if (!companyUuid) {
      emits('update:modelValue', undefined);
      emits('change', undefined);
      return;
    }

    try {
      loading.value = true;
      // 调用门店切换 API
      const res = await $post(
        '/auth/store_switch',
        { company_uuid: companyUuid },
        {
          baseURL: '/api/v1/',
        },
      );
      message.success(res.msg || $t('切换成功'));
      emits('update:modelValue', companyUuid);
      emits('change', companyUuid);

      // 触发页面刷新（如果需要）
      if (typeof window !== 'undefined') {
        window.location.reload();
      }
    } catch (error) {
      console.error('门店切换失败:', error);
      // 切换失败时恢复原值
      selectedCompanyUuid.value = props.modelValue;
    } finally {
      loading.value = false;
    }
  };

  // 监听外部值变化
  watch(
    () => props.modelValue,
    (newVal) => {
      selectedCompanyUuid.value = newVal;
    },
  );

  // 自动加载
  if (props.autoLoad) {
    onMounted(() => {
      fetchCompanyList();
    });
  }

  // 暴露方法供外部调用
  defineExpose({
    fetchCompanyList,
    companyList,
  });
</script>

<style lang="scss" scoped>
  .store-switch-component {
    :deep(.el-select) {
      width: 100%;
    }
  }
</style>
