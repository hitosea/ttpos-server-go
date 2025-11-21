<template>
  <el-dialog :title="$t('添加管理员')" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" label-position="top" :rules="formRules" :label-width="formLabelWidth">
      <el-form-item for="no_click" :label="$t('邮箱')" prop="user_name">
        <el-input v-model="form.user_name" maxlength="64" :placeholder="$t('请输入邮箱')"> </el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('手机号')" prop="phone">
        <el-input v-model="form.phone" maxlength="20" :placeholder="$t('请输入手机号')"> </el-input>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('角色')" prop="role_id">
        <el-select v-model="form.role_id" :multiple="true" @change="selectChange" :placeholder="$t('请选择角色')">
          <el-option v-for="item in roleList" :value="item.role_id" :key="item.role_id" :label="item.role_name_h1"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('登录密码')" prop="password">
        <el-input v-model="form.password" :maxlength="16" :placeholder="$t('请输入登录密码')" type="password" @input="changePassword"></el-input>
        <div class="tips">
          {{ $t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种') }}
        </div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('确认密码')" :maxlength="16" prop="confirm_password"
        ><el-input v-model="form.confirm_password" :placeholder="$t('请输入确认密码')" type="password"></el-input
      ></el-form-item>
      <el-form-item for="no_click" :label="$t('权限密码')" prop="permission_password">
        <el-input v-model="form.permission_password" :maxlength="8" :placeholder="$t('请输入权限密码（必填）')" type="password"></el-input>
        <div class="tips">
          {{ $t('密码必须为 4 - 8 位数字') }}
        </div>
      </el-form-item>
      <el-form-item for="no_click" :label="$t('姓名')" prop="real_name">
        <el-input v-model="form.real_name" :maxlength="50" :placeholder="$t('请输入姓名')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import AuthApi from '@/api/auth.js';

  export default {
    data() {
      let validatePass1 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入登录密码')));
        } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
          callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
        } else {
          callback();
        }
      };

      let user_name = (rule, value, callback) => {
        const re = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        if (!value) {
          callback(new Error($t('请输入邮箱')));
        } else if (!re.test(value)) {
          callback(new Error($t('请确认邮箱格式')));
        }
        callback();
      };

      let validatePass2 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入确认密码')));
        } else if (value !== this.form.password) {
          callback(new Error($t('两次密码不一致！')));
        } else {
          callback();
        }
      };

      let validatePermissionPassword = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('权限密码不能为空')));
        } else if (!/^\d{4,8}$/.test(value)) {
          callback(new Error($t('密码必须为 4 - 8 位数字')));
        } else {
          callback();
        }
      };
      return {
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        loading: false,
        /*是否显示*/
        dialogVisible: false,
        /*form表单对象*/
        form: {
          user_name: '',
          phone: '',
          access_id: [],
          permission_password: '',
        },
        /*form验证*/
        formRules: {
          user_name: [
            {
              required: true,
              validator: user_name,
              trigger: ['blur', 'change'],
            },
          ],
          phone: [
            {
              required: true,
              message: $t('请输入手机号'),
              trigger: ['blur', 'change'],
            },
          ],
          role_id: [
            {
              required: true,
              message: $t('请选择角色'),
              trigger: 'blur',
            },
          ],
          password: [
            {
              required: true,
              validator: validatePass1,
              trigger: ['blur', 'change'],
            },
          ],
          confirm_password: [
            {
              required: true,
              validator: validatePass2,
              trigger: ['blur', 'change'],
            },
          ],
          real_name: [
            {
              required: true,
              message: $t('请输入姓名'),
              trigger: ['blur', 'change'],
            },
          ],
          permission_password: [
            {
              required: true,
              validator: validatePermissionPassword,
              trigger: ['blur', 'change'],
            },
          ],
        },
      };
    },
    props: ['open', 'roleList'],

    created() {
      this.dialogVisible = this.open;
    },
    methods: {
      /*添加*/
      onSubmit() {
        let self = this;
        self.$refs.form.validate((valid) => {
          if (valid) {
            self.loading = true;
            let params = self.form;
            AuthApi.userAdd(params, true)
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('添加成功'),
                  type: 'success',
                });
                self.dialogFormVisible(true);
              })
              .catch((error) => {
                self.loading = false;
              });
          }
        });
      },

      changePassword() {
        if (this.form.confirm_password) {
          this.$refs.form.validateField('confirm_password');
        }
      },

      selectChange() {
        this.$refs.form.validateField('role_id');
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        this.form = {
          user_name: '',
          access_id: [],
          permission_password: '',
        };
        if (e) {
          this.$emit('close', {
            type: 'success',
            openDialog: false,
          });
        } else {
          this.$emit('close', {
            type: 'error',
            openDialog: false,
          });
        }
      },
    },
  };
</script>
<style scoped lang="scss">
  :deep(.el-select-dropdown__item) {
    max-width: 999px !important;
  }
  :deep(.el-select--small .el-select__wrapper) {
    min-height: 32px !important;
    height: auto;
    padding: 4px 8px !important;
  }
</style>
