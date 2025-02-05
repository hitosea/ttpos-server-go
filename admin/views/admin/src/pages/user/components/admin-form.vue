<template>
  <el-dialog
    width="820"
    :title="hasEdit ? $t('编辑管理員') : $t('添加管理員')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="emits('update:show', false)"
  >
    <el-form ref="formElement" label-position="top" label-width="auto" :model="formData" :rules="formRules">
      <el-form-item :label="$t('邮箱：')" prop="user_name">
        <el-input v-model="formData.user_name" type="text" maxlength="50" clearable :placeholder="$t('请输入邮箱')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('手机号：')" prop="phone">
        <el-input v-model="formData.phone" type="text" maxlength="50" clearable :placeholder="$t('请输入手机号')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('角色：')" prop="role_id">
        <el-select v-model="formData.role_id" :placeholder="$t('请选择角色')" clearable multiple>
          <el-option v-for="item in roleList" :value="item.id" :key="item.id" :label="item.role_name" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('登录密码')" prop="password">
        <el-input type="password" v-model="formData.password" autocomplete="off" maxlength="50" show-password :placeholder="$t('请输入登录密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('确认密码')" prop="confirm_password">
        <el-input type="password" v-model="formData.confirm_password" autocomplete="off" maxlength="50" show-password :placeholder="$t('请确认密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('姓名：')" prop="real_name">
        <el-input v-model="formData.real_name" type="text" maxlength="50" clearable :placeholder="$t('请输入姓名')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emits('update:show', false)">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, computed } from 'vue';
  import { AdminAddType, fetchAdminAdd, fetchAdminEdit } from '@/api/user/admin';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'change', value: any): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: AdminAddType & { admin_user_id?: number; userRole?: any };
      roleList?: { id: number; role_name: string }[];
    }>(),
    {
      show: false,
      detail: () => ({}),
      roleList: () => [],
    },
  );
  const hasEdit = computed(() => !!props.detail?.admin_user_id);
  const formData = ref<AdminAddType>({
    user_name: '',
    phone: '',
    password: '',
    confirm_password: '',
    real_name: '',
    role_id: [],
  });
  const formRules = reactive({
    user_name: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          const re = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
          if (!value) {
            callback(new Error($t('请输入用户名')));
          } else if (!re.test(value)) {
            callback(new Error($t('请确认邮箱格式')));
          }
          callback();
        },
      },
    ],
    phone: [{ required: true, message: $t('请输入手机号'), trigger: ['change', 'blur'] }],
    role_id: [{ required: true, message: $t('请选择角色'), trigger: ['change', 'blur'] }],
    real_name: [{ required: true, message: $t('请输入姓名'), trigger: 'blur' }],
    password: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (!value && hasEdit.value) callback();
          if (!value) {
            callback(new Error($t('请输入登录密码')));
          } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
            callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
          }
          callback();
        },
      },
    ],
    confirm_password: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (!value && hasEdit.value) callback();
          if (!value) {
            callback(new Error($t('请确认密码')));
          } else if (value !== formData.value.password) {
            callback(new Error($t('两次密码不一致！')));
          }
          callback();
        },
      },
    ],
  });
  const formLoading = ref(false);
  const formElement = ref();

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        const data = { ...formData.value };
        formLoading.value = true;
        let res = null;
        if (hasEdit.value) {
          res = await fetchAdminEdit({ ...data, admin_user_id: props.detail?.admin_user_id });
        } else {
          res = await fetchAdminAdd(data);
        }
        message.success(res.msg);
        emits('update:show', false);
        emits('change', res);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) return;
      formElement.value?.resetFields();
      // 重置校验
      formRules.password[0].required = !hasEdit.value;
      formRules.confirm_password[0].required = !hasEdit.value;
      //
      const data: any = {
        user_name: props.detail?.user_name || '',
        phone: props.detail?.phone || '',
        password: props.detail?.password || '',
        confirm_password: props.detail?.confirm_password || '',
        real_name: props.detail?.real_name || '',
        role_id: props.detail?.userRole?.map((item: any) => item.role_id) || [],
      };
      formData.value = data;
      // 清空
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>

<style lang="scss" scoped></style>
