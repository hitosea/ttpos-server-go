<template>
  <el-dialog
    width="960"
    :title="hasEdit ? $t('查看商家') : $t('选择商家')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="handleClose()"
  >
    <el-form :model="formData" :rules="formRules" ref="formRef" label-position="top" label-width="auto">
      <el-form-item :label="$t('选择商家')" prop="uuid">
        <el-select v-model="formData.uuid" :placeholder="$t('请选择商家')" clearable filterable :disabled="props.hasEdit || confirmPassword">
          <el-option v-for="item in companyList" :key="item.uuid" :value="item.uuid" :label="item.name" :disabled="item.erpnext_site_code !== ''" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('所属erpnext的site')" prop="erpnext_site_code">
        <el-radio-group v-model="formData.erpnext_site_code" @change="handleSiteChange(formData.erpnext_site_code)" :disabled="props.hasEdit || confirmPassword">
          <el-radio v-for="item in erpnextSiteList" :key="item.code" :value="item.code" size="large">{{ item.name }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('所属erpnext公司')" prop="erpnext_company_abbr">
        <el-cascader
          class="w-full"
          v-model="formData.erpnext_company_abbr"
          :options="erpnextCompanyList"
          :props="{
            label: 'company_name',
            value: 'company_abbr',
            disabled: 'is_used', // 添加禁用字段
            checkStrictly: true,
            expandTrigger: 'hover',
          }"
          clearable
          filterable
          :disabled="props.hasEdit || confirmPassword"
          @change="handleCompanyChange"
        >
          <template #default="{ data }">
            <span class="!w-full block" :class="{ 'text-gray-400 cursor-not-allowed': data.is_used, 'cursor-pointer': !data.is_used }" @click="handleCompanyClick(data)">
              {{ data.company_name }}
            </span>
          </template>
        </el-cascader>
      </el-form-item>
      <el-form-item v-if="confirmPassword && !props.hasEdit" :label="$t('密码验证')" prop="password">
        <el-input v-model="formData.password" type="password" :placeholder="$t('请输入密码')" clearable :disabled="props.hasEdit" />
        <div class="text-center text-sm text-red-500 mt-2">
          {{ $t('请确定授权信息，该操作不可逆，请仔细确认！') }}
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">{{ $t('取消') }}</el-button>
      <el-button v-if="!props.hasEdit" type="primary" @click="handleSubmit" :loading="isLoading">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
  import { ref, defineProps, defineEmits, watch, PropType } from 'vue';
  import { $t } from '@/i18n';
  import {
    authorizationListTypeItem,
    erpnextSiteCodeItem,
    erpnextSiteCompanyItem,
    getAuthorizationList,
    getErpnextSiteCode,
    getErpnextSiteCompany,
    erpnextAdd,
    erpnextAddParams,
  } from '@/api/authorization';
  import { ElMessage, FormInstance } from 'element-plus';
  const emit = defineEmits(['update:show', 'refresh']);
  const props = defineProps({
    show: {
      type: Boolean,
      default: false,
    },
    hasEdit: {
      type: Boolean,
      default: false,
    },
    editRow: {
      type: Object as PropType<authorizationListTypeItem>,
      default: () => ({}),
    },
  });

  const formRef = ref<FormInstance>();
  const companyList = ref<authorizationListTypeItem[]>([]);
  const erpnextSiteList = ref<erpnextSiteCodeItem[]>([]);
  const erpnextCompanyList = ref<erpnextSiteCompanyItem[]>([]);
  // 确认密码输入
  const confirmPassword = ref(false);
  const isLoading = ref(false);
  const getCompanyList = async () => {
    try {
      const res = await getAuthorizationList({
        keyword: '',
        page: 1,
        list_rows: 1000,
        configured: 0,
      });
      companyList.value = res.data.list.data;
    } catch (error) {
      console.log(error);
    }
  };

  const getErpnextSiteCodeList = async () => {
    try {
      const res = await getErpnextSiteCode();
      erpnextSiteList.value = res.data.list;
      if (res.data.list.length > 0 && !props.hasEdit) {
        formData.value.erpnext_site_code = res.data.list[0].code;
        formData.value.erpnext_default_company_abbr = res.data.list[0].default_company;
        getErpnextCompanyList(res.data.list[0].code);
      } else if (props.hasEdit) {
        // 保存编辑模式下的原始值
        const originalCompanyAbbr = props.editRow.erpnext_company_abbr || '';

        formData.value.erpnext_site_code = props.editRow.erpnext_site_code;
        formData.value.erpnext_default_company_abbr = props.editRow.erpnext_default_company_abbr || '';

        await getErpnextCompanyList(props.editRow.erpnext_site_code);

        // 恢复原始值
        formData.value.erpnext_company_abbr = originalCompanyAbbr;

        // 触发公司选择变化，更新pos profile列表
        if (originalCompanyAbbr) {
          handleCompanyChange(originalCompanyAbbr);
        }
      }
    } catch (error) {
      console.log(error);
    }
  };

  const handleSiteChange = (site_code: string) => {
    // 清空erpnext_company_abbr
    formData.value.erpnext_company_abbr = '';
    getErpnextCompanyList(site_code);
  };
  const getErpnextCompanyList = async (site_code: string) => {
    erpnextCompanyList.value = [];

    // 循环erpnextSiteList获取erpnext_default_company_abbr
    for (const item of erpnextSiteList.value) {
      if (item.code === site_code) {
        formData.value.erpnext_default_company_abbr = item.default_company;
        break;
      }
    }
    try {
      const res = await getErpnextSiteCompany({
        site_code: site_code,
        company_abbr: '',
      });
      erpnextCompanyList.value = res.data.list;
    } catch (error) {
      console.log(error);
    }
  };

  const handleCompanyChange = (value: any) => {
    // 确保转换为字符串
    if (Array.isArray(value)) {
      // 如果是数组，取最后一个值（通常是选中的叶子节点）
      formData.value.erpnext_company_abbr = value[value.length - 1] || '';
    } else {
      // 如果已经是字符串，直接使用
      formData.value.erpnext_company_abbr = value || '';
    }
  };

  // 处理公司点击事件，阻止已被使用的项目被选中
  const handleCompanyClick = (data: any) => {
    // 如果已被使用，阻止选择
    if (data.is_used) {
      return;
    }
    // 如果未被使用，允许选择
    handleCompanyChange(data.company_abbr);
  };

  const formData = ref<erpnextAddParams>({
    uuid: undefined,
    erpnext_site_code: '',
    erpnext_default_company_abbr: '',
    erpnext_company_abbr: '',
    password: '',
  });

  const formRules = ref({
    uuid: [{ required: true, message: $t('请选择商家') }],
    erpnext_site_code: [{ required: true, message: $t('请选择所属erpnext的site') }],
    erpnext_company_abbr: [{ required: true, message: $t('请选择所属erpnext公司') }],
    password: [{ required: true, message: $t('请输入密码') }],
  });

  const handleClose = () => {
    emit('update:show', false);
    handleReset();
  };

  const handleReset = () => {
    formData.value = {
      uuid: undefined,
      erpnext_site_code: '',
      erpnext_default_company_abbr: '',
      erpnext_company_abbr: '',
      password: '',
    };
    confirmPassword.value = false;
  };

  const handleSubmit = async () => {
    formRef.value?.validate(async (valid) => {
      if (valid) {
        if (!confirmPassword.value) {
          confirmPassword.value = true;
          return;
        }
        try {
          isLoading.value = true;
          const res = await erpnextAdd(formData.value);
          ElMessage.success(res.msg);
          handleClose();
          emit('refresh');
        } catch (error) {
          console.log(error);
        } finally {
          isLoading.value = false;
        }
      }
    });
  };

  watch(
    () => props.show,
    async (newVal: boolean) => {
      if (newVal) {
        if (props.hasEdit) {
          formData.value = {
            uuid: props.editRow.uuid,
            erpnext_site_code: props.editRow.erpnext_site_code,
            erpnext_default_company_abbr: props.editRow.erpnext_default_company_abbr || '',
            erpnext_company_abbr: props.editRow.erpnext_company_abbr || '',
            password: '',
          };
        }

        await getCompanyList();
        await getErpnextSiteCodeList();
        // 表单验证取消
        formRef.value?.clearValidate();
      }
    },
    { immediate: true, deep: true },
  );
</script>
<style lang=""></style>
