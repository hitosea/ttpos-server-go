<template>
  <el-dialog
    width="820"
    :title="$t('支付')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="emits('update:show', false)"
  >
    <h2 class="border-l-4 border-[#ffbe00] pl-2 text-[16px] mb-4">{{ $t('申请信息') }}(LianLian)</h2>
    <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="auto">
      <el-form-item :label="$t('白名单IP')" prop="ll_white_ip">
        <el-input v-model="formData.ll_white_ip" type="text" maxlength="50" disabled clearable :placeholder="$t('请输入白名单IP')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('商户号')" prop="ll_merchant_id">
        <el-input v-model="formData.ll_merchant_id" type="text" maxlength="50" clearable :placeholder="$t('请输入商户号')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('LianLianpay公钥')" prop="ll_public_key">
        <el-input v-model="formData.ll_public_key" type="text" clearable :placeholder="$t('请输入LianLianpay公钥')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('商户私钥')" prop="ll_merchant_private_key">
        <el-input v-model="formData.ll_merchant_private_key" type="text" clearable :placeholder="$t('请输入商户私钥')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('Token')" prop="ll_token">
        <el-input v-model="formData.ll_token" type="text" maxlength="255" clearable :placeholder="$t('请输入Token')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('站点ID')" prop="ll_store_id">
        <el-input v-model="formData.ll_store_id" type="text" maxlength="255" clearable :placeholder="$t('请输入站点ID')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emits('update:show', false)">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
  import { $t } from '@/i18n';
  import { ref, watch, reactive } from 'vue';
  import { getPayment, setPayment } from '@/api/merchant';
  import { message } from '@/utils/feedback';
  //
  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'change', value: any): void;
  }>();

  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: any;
    }>(),
    {
      show: false,
      detail: () => ({}),
    },
  );
  const formLoading = ref(false);
  const formData = ref({
    ll_white_ip: '',
    ll_merchant_id: '',
    ll_public_key: '',
    ll_merchant_private_key: '',
    ll_token: '',
    ll_store_id: '',
  });
  const formRules = reactive({
    ll_white_ip: [{ required: true, message: $t('请输入白名单IP'), trigger: 'blur' }],
    ll_merchant_id: [{ required: true, message: $t('请输入商户号'), trigger: 'blur' }],
    ll_public_key: [{ required: true, message: $t('请输入LianLianpay公钥'), trigger: 'blur' }],
    ll_merchant_private_key: [{ required: true, message: $t('请输入商户私钥'), trigger: 'blur' }],
    ll_token: [{ required: true, message: $t('请输入商户Token'), trigger: 'blur' }],
    ll_store_id: [{ required: true, message: $t('请输入站点ID'), trigger: 'blur' }],
  });

  //
  const handleSubmit = async () => {
    try {
      let data: any = {};
      data = formData.value;
      data.app_id = props.detail.app_id;
      data.shop_supplier_id = props.detail.shop_supplier_id;
      formLoading.value = true;
      const res = await setPayment(data);
      message.success(res.msg);
      emits('update:show', false);
    } catch (error) {
      //
    } finally {
      //
      formLoading.value = false;
    }
  };

  const getPaymentData = async () => {
    formLoading.value = true;
    try {
      const res = await getPayment(props.detail.shop_supplier_id);
      formData.value.ll_white_ip = res.data?.ll_white_ip || '';
      formData.value.ll_merchant_id = res.data?.ll_merchant_id || '';
      formData.value.ll_public_key = res.data?.ll_public_key || '';
      formData.value.ll_merchant_private_key = res.data?.ll_merchant_private_key || '';
      formData.value.ll_token = res.data?.ll_token || '';
      formData.value.ll_store_id = res.data?.ll_store_id || '';
    } catch (error) {
      //
    } finally {
      //
      formLoading.value = false;
    }
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) return;
      getPaymentData();
    },
  );
</script>
<style lang="scss" scoped></style>
