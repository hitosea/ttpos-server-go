<template>
  <el-dialog :title="$t('修改密码')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false" width="30%">
    <el-form size="small" :model="form" label-position="top" ref="form" :rules="rules">
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

<script>
  import UserApi from '@/api/user.js';
  import { useUserStore } from '@/store';
  import { languageStore } from '@/store/model/language.js';
  import { EEUIRELOAD } from '@/utils/platform.js';

  const isCloudDeploy = languageStore().getIsCloudDeploy().isCloudDeploy;
  const { afterLogout } = useUserStore();
  export default {
    data() {
      let validatePass1 = function () {};

      validatePass1 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入登录密码')));
        } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
          callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
        } else {
          callback();
        }
      };

      let validatePass2 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入确认新密码')));
        } else if (value !== this.form.password) {
          callback(new Error($t('两次密码不一致！')));
        } else {
          callback();
        }
      };
      return {
        isCloudDeploy: isCloudDeploy,
        loading: false,
        /*左边长度*/
        formLabelWidth: '100px',
        /*是否显示*/
        dialogVisible: true,
        /*表单*/
        form: {},
        rules: {
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
          confirmPass: [{ required: true, validator: validatePass2, trigger: ['blur', 'change'] }],
        },
      };
    },
    props: [],
    created() {},
    methods: {
      /*确认事件*/
      submitFunc(e) {
        let self = this;
        let form = self.form;
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            UserApi.EditPass(form, true)
              .then((data) => {
                self.loading = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: data.msg,
                    type: 'success',
                  });
                  self.dialogFormVisible(true);
                  setTimeout(() => {
                    this.logout();
                  }, 2000);
                } else {
                  self.loading = false;
                }
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },

      async logout() {
        afterLogout();
        this.$router.push('/login');
        EEUIRELOAD();
      },

      /*关闭弹窗*/
      dialogFormVisible() {
        this.$emit('close', false);
      },

      changePassword() {
        if (this.form.confirmPass) {
          this.$refs.form.validateField('confirmPass');
        }
      },
    },
  };
</script>
