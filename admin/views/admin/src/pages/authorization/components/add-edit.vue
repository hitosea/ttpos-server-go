<template>
  <el-dialog
    width="960"
    :title="hasEdit ? $t('編輯商家') : $t('选择商家')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="handleClose()"
  >
    <el-form :model="formData" :rules="formRules" ref="formRef" label-position="top" label-width="auto">
      <el-form-item :label="$t('选择商家')" prop="uuid">
        <el-select v-model="formData.uuid" :placeholder="$t('请选择商家')" clearable filterable :disabled="props.hasEdit">
          <el-option v-for="item in companyList" :key="item.uuid" :value="item.uuid" :label="item.name" :disabled="item.erpnext_site_code !== ''" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('所属erpnext的site')" prop="erpnext_site_code">
        <el-radio-group v-model="formData.erpnext_site_code" @change="getErpnextCompanyList(formData.erpnext_site_code)" :disabled="props.hasEdit">
          <el-radio v-for="item in erpnextSiteList" :key="item.code" :value="item.code" size="large">{{ item.name }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item :label="$t('所属erpnext公司')" prop="erpnext_company_abbr">
        <el-select v-model="formData.erpnext_company_abbr" :placeholder="$t('请选择所属erpnext公司')" filterable clearable :disabled="props.hasEdit">
          <el-option v-for="item in erpnextCompanyList" :key="item.company_abbr" :value="item.company_abbr" :label="item.company_name" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="isLoading">{{ $t('确定') }}</el-button>
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
        getErpnextCompanyList(res.data.list[0].code);
      } else if (props.hasEdit) {
        formData.value.erpnext_site_code = props.editRow.erpnext_site_code;
        getErpnextCompanyList(props.editRow.erpnext_site_code);
      }
    } catch (error) {
      console.log(error);
    }
  };

  const getErpnextCompanyList = async (site_code: string) => {
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

  const formData = ref({
    uuid: '',
    erpnext_site_code: '',
    erpnext_company_abbr: '',
  });

  const formRules = ref({
    uuid: [{ required: true, message: $t('请选择商家') }],
    erpnext_site_code: [{ required: true, message: $t('请选择所属erpnext的site') }],
    erpnext_company_abbr: [{ required: true, message: $t('请选择所属erpnext公司') }],
  });

  const handleClose = () => {
    emit('update:show', false);
    handleReset();
  };

  const handleReset = () => {
    formRef.value?.resetFields();
  };

  const handleSubmit = async () => {
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
  };

  watch(
    () => props.show,
    (newVal: boolean) => {
      if (newVal) {
        if (props.hasEdit) {
          formData.value = {
            uuid: props.editRow.uuid.toString(),
            erpnext_site_code: props.editRow.erpnext_site_code,
            erpnext_company_abbr: props.editRow.erpnext_company_abbr,
          };
        }
        getCompanyList();
        getErpnextSiteCodeList();
      }
    },
    { immediate: true, deep: true },
  );
</script>
<style lang=""></style>
