<template>
  <el-dialog width="820" :title="props?.title" :modelValue="props.show" :close-on-click-modal="false" :close-on-press-escape="false" align-center @close="handleClose()">
    <div class="max-h-[75vh] overflow-auto pr-4">
      <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="auto">
        <el-form-item :label="$t('品牌')" prop="brand">
          <el-radio-group v-model="formData.brand" :disabled="hasDetail">
            <el-radio :value="1">TTPOS</el-radio>
            <el-radio :value="2">JBCレジ</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('版本号')" prop="version_name">
          <el-input v-model="formData.version_name" :disabled="true" type="text" maxlength="20" clearable :placeholder="$t('请输入版本号')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('强制更新')" prop="forced_update">
          <el-radio-group v-model="formData.forced_update" :disabled="props.detail.is_publish == '1'">
            <el-radio :value="0">{{ $t('否') }}</el-radio>
            <el-radio :value="1">{{ $t('是') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('安装包')" prop="package_url" v-if="!hasDetail">
          <file-upload ref="fileUpLoadRef" v-model="formData.package_url" @change="upDate"></file-upload>
        </el-form-item>
        <UniqueNameForm
          v-if="UniqueNameFormShow"
          ref="uniqueNameFormRef"
          :disabled="props.detail.is_publish == '1'"
          :inputType="'textarea'"
          :maxlength="500"
          :labelPrefix="$t('更新日志')"
          :placeholder="$t('请输入更新日志')"
          :overrideLanguages="formData.update_log ? formData.update_log : undefined"
          apiSource="product_barcode"
        />
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleClose()">{{ hasDetail ? $t('关闭') : $t('取消') }}</el-button>
      <el-button v-if="props.detail.is_publish == '0'" :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
  import { type addDataType, type detailData, publishClient } from '@/api/client';
  import { ref, reactive, watch } from 'vue';
  import { $t } from '@/i18n';
  import fileUpload from './file-upload/index.vue';
  import { message } from '@/utils/feedback';
  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'change', value: any): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      title?: string;
      type?: string;
      detail?: detailData;
    }>(),
    {
      show: false,
      title: '',
      type: '1',
      detail: () => ({}),
    },
  );
  const formLoading = ref(false);
  const UniqueNameFormShow = ref(false);
  const hasDetail = ref(false);
  const formElement = ref();
  const fileUpLoadRef = ref();
  const uniqueNameFormRef = ref();
  const formData = ref<addDataType>({
    brand: 1, //品牌
    type: '1', // 类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
    version_name: '', // 版本名称
    version_number: '', // 版本号
    forced_update: 0, // 强制更新 0否 1是
    package_url: '', // 包地址
    update_log: '', // 更新日志
  });
  const formRules = reactive({
    brand: [{ required: true, message: $t('请选择品牌'), trigger: 'blur' }],
    version_name: [{ required: true, message: $t('请输入版本名称'), trigger: 'blur' }],
    version_number: [{ required: true, message: $t('请输入版本号'), trigger: 'blur' }],
    forced_update: [{ required: true, message: $t('请选择更新类型'), trigger: 'blur' }],
    package_url: [{ required: true, message: $t('请上传安装包'), trigger: 'blur' }],
    update_log: [{ required: true, message: $t('请输入更新日志'), trigger: 'blur' }],
  });

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      const validUniqueName = await uniqueNameFormRef.value.validate();
      if (!validUniqueName) return;
      try {
        formLoading.value = true;
        let res = null;
        const upData = {
          id: '',
          update_log: '',
          forced_update: 0,
        };
        upData.id = props.detail?.id || '';
        upData.forced_update = formData.value.forced_update || 0;
        upData.update_log = JSON.stringify(uniqueNameFormRef.value.data) || '';
        res = await publishClient(upData);
        message.success($t('发布成功'));
        emits('update:show', false);
        emits('change', res);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  const upDate = (data: any) => {
    formData.value.version_number = data.version_number;
    formData.value.version_name = data.version_name;
  };

  const handleClose = () => {
    setTimeout(() => {
      formData.value = {
        brand: 1, //品牌
        type: '', // 类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
        version_name: '', // 版本名称
        version_number: '', // 版本号
        forced_update: 0, // 强制更新 0否 1是
        package_url: '', // 包地址
        update_log: '', // 更新日志
      };
      if (fileUpLoadRef.value) {
        fileUpLoadRef.value.clearName();
      }
      hasDetail.value = false;
    }, 300);
    emits('update:show', false);
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) {
        setTimeout(() => {
          UniqueNameFormShow.value = false;
        }, 300);
        return;
      }
      if (props.detail?.id) {
        hasDetail.value = true;
        UniqueNameFormShow.value = true;
        formElement.value?.resetFields();
        //
        // 默认值
        formData.value = {
          brand: props.detail?.brand || 1,
          type: props.detail?.type || '',
          version_name: props.detail?.version_name || '',
          version_number: props.detail?.version_number || '',
          forced_update: props.detail?.forced_update || 0,
          package_url: props.detail?.package_url || '',
          update_log: JSON.parse(props.detail?.update_log || '{}'),
        };
      }
      // 清空
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>
<style lang=""></style>
