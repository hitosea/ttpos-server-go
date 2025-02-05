<template>
  <el-dialog width="720" :title="$t('修改密码')" :modelValue="props.show" :close-on-click-modal="false" :close-on-press-escape="false" @close="emits('update:show', false)">
    <el-form ref="formElement" :model="formData" :rules="formRules" label-position="top" label-width="auto">
      <el-form-item :label="$t('原密码')" prop="oldPass">
        <el-input type="password" v-model="formData.oldPass" autocomplete="off" maxlength="50" show-password :placeholder="$t('请输入原密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('新密码')" prop="pass">
        <el-input type="password" v-model="formData.pass" autocomplete="off" maxlength="50" show-password :placeholder="$t('请输入新密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('确认新密码')" prop="checkPass">
        <el-input type="password" v-model="formData.checkPass" autocomplete="off" maxlength="50" show-password :placeholder="$t('请确认新密码')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emits('update:show', false)">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" :disabled="buttonDisabled" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, computed } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { message } from '@/utils/feedback';
  import type { FormRules } from 'element-plus';
  import { clearAuthInfo } from '@/utils/auth';
  import { editPassword } from '@/api/login';
  import { $t } from '@/i18n';

  const route = useRoute();
  const router = useRouter();
  //
  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
    }>(),
    {
      show: false,
    },
  );
  const formData = ref({
    oldPass: '',
    pass: '',
    checkPass: '',
  });
  const formElement = ref();
  const formLoading = ref(false);
  const formRules = reactive<FormRules>({
    oldPass: [
      {
        required: true,
        message: $t('请输入原密码'),
        trigger: 'change',
      },
    ],
    pass: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (!value) {
            callback(new Error($t('请输入新密码')));
          } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
            callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
          }
          callback();
        },
      },
    ],
    checkPass: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (!value) {
            callback(new Error($t('请确认新密码')));
          } else if (value !== formData.value.pass) {
            callback(new Error($t('两次密码不一致！')));
          }
          callback();
        },
      },
    ],
  });

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        formLoading.value = true;
        const res = await editPassword(formData.value);
        message.success(res.msg);
        emits('update:show', false);
        //
        clearAuthInfo();
        router.replace(`/login?redirect=${route.fullPath}`);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  const buttonDisabled = computed(() => {
    return !formData.value.oldPass || !formData.value.pass || !formData.value.checkPass;
  });

  watch(
    () => props.show,
    (val) => {
      if (!val) return;
      formElement.value?.resetFields();
      //
      formData.value = {
        oldPass: '',
        pass: '',
        checkPass: '',
      };
      // 清空
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>

<style lang="scss" scoped></style>
