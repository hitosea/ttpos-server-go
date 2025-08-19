<template>
  <el-dialog :title="$t('修改密码')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false" width="30%">
    <el-form size="small" :model="form" label-position="top" ref="formRef" :rules="rules">
      <el-form-item :label="$t('原始密码')" :label-width="formLabelWidth" prop="oldpass">
        <el-input type="password" v-model="form.oldpass" autocomplete="off" :placeholder="$t('请输入登录密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('新密码')" :label-width="formLabelWidth" prop="password">
        <el-input type="password" v-model="form.password" :maxlength="16" autocomplete="off" @input="changePassword" :placeholder="$t('请输入确认新密码')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('确认新密码')" :label-width="formLabelWidth" prop="confirmPass">
        <el-input type="password" v-model="form.confirmPass" :maxlength="16" autocomplete="off" :placeholder="$t('请输入确认新密码')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="submitFunc(form.user_id)" :loading="loading">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, reactive } from 'vue';
  import { useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import UserApi from '@/api/user.js';
  import { useUserStore } from '@/store';
  import { EEUIRELOAD } from '@/utils/platform.js';

  const router = useRouter();
  const { afterLogout } = useUserStore();

  const emit = defineEmits(['close']);

  // 响应式数据
  const loading = ref(false);
  const formLabelWidth = ref('100px');
  const dialogVisible = ref(true);
  const formRef = ref(null);

  const form = reactive({
    oldpass: '',
    password: '',
    confirmPass: '',
  });

  // 密码验证函数
  const validatePass1 = (rule, value, callback) => {
    if (!value) {
      callback(new Error($t('请输入登录密码')));
    } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
      callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
    } else {
      callback();
    }
  };

  const validatePass2 = (rule, value, callback) => {
    if (!value) {
      callback(new Error($t('请输入确认新密码')));
    } else if (value !== form.password) {
      callback(new Error($t('两次密码不一致！')));
    } else {
      callback();
    }
  };

  // 表单验证规则
  const rules = reactive({
    oldpass: [
      {
        required: true,
        message: $t('请输入原始密码'),
        trigger: ['blur', 'change'],
      },
    ],
    password: [
      {
        required: true,
        validator: validatePass1,
        trigger: ['blur', 'change'],
      },
    ],
    confirmPass: [
      {
        required: true,
        validator: validatePass2,
        trigger: ['blur', 'change'],
      },
    ],
  });

  // 确认修改密码
  const submitFunc = (e) => {
    formRef.value.validate((valid) => {
      if (valid) {
        loading.value = true;
        UserApi.EditPass(form, true)
          .then((data) => {
            loading.value = false;
            if (data.code == 1) {
              ElMessage({
                message: data.msg,
                type: 'success',
              });
              dialogFormVisible();
              setTimeout(() => {
                logout();
              }, 2000);
            } else {
              loading.value = false;
            }
          })
          .catch((error) => {
            loading.value = false;
          });
      }
    });
  };

  // 登出
  const logout = async () => {
    await afterLogout();
    router.push('/login');
    EEUIRELOAD();
  };

  // 关闭弹窗
  const dialogFormVisible = () => {
    emit('close', false);
  };

  // 密码变化时验证确认密码
  const changePassword = () => {
    if (form.confirmPass) {
      formRef.value.validateField('confirmPass');
    }
  };
</script>
