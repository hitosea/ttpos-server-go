<template>
  <el-dialog
    width="900"
    :title="hasEdit ? $t('编辑统一账号') : $t('添加统一账号')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="emits('update:show', false)"
  >
    <el-form ref="formElement" label-position="top" label-width="auto" :model="formData" :rules="formRules">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('邮箱：')" prop="email">
            <el-input v-model="formData.email" type="text" maxlength="255" clearable :placeholder="$t('请输入邮箱')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('手机号：')" prop="phone">
            <el-input v-model="formData.phone" type="text" maxlength="20" clearable :placeholder="$t('请输入手机号（可选）')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('姓名：')" prop="real_name">
            <el-input v-model="formData.real_name" type="text" maxlength="255" clearable :placeholder="$t('请输入姓名')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12" v-if="!hasEdit">
          <el-form-item :label="$t('关联门店：')" prop="company_uuid">
            <el-select v-model="formData.company_uuid" :placeholder="$t('请选择门店')" clearable>
              <el-option v-for="item in companyList" :value="item.app_id" :key="item.app_id" :label="item.shop_supplier_name || item.id" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20" v-if="!hasEdit">
        <el-col :span="24">
          <el-form-item :label="$t('角色：')" prop="role_uuids">
            <el-select v-model="formData.role_uuids" :placeholder="$t('请选择角色')" clearable multiple>
              <el-option v-for="item in roleList" :value="item.id" :key="item.id" :label="item.role_name" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20" v-if="hasEdit">
        <el-col :span="24">
          <el-form-item :label="$t('关联门店和角色：')" prop="company_list">
            <div v-for="(item, index) in formData.company_list" :key="index" class="mb-4 p-4 border rounded">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-select v-model="item.company_uuid" :placeholder="$t('请选择门店')" clearable @change="handleCompanyChange(index)">
                    <el-option v-for="company in companyList" :value="company.app_id" :key="company.app_id" :label="company.shop_supplier_name || company.id" />
                  </el-select>
                </el-col>
                <el-col :span="14">
                  <el-select v-model="item.role_uuids" :placeholder="$t('请选择角色')" clearable multiple>
                    <el-option v-for="role in roleList" :value="role.id" :key="role.id" :label="role.role_name" />
                  </el-select>
                </el-col>
                <el-col :span="2">
                  <el-button type="danger" icon="Delete" circle @click="handleRemoveCompany(index)" v-if="formData.company_list.length > 1"></el-button>
                </el-col>
              </el-row>
            </div>
            <el-button type="primary" icon="Plus" @click="handleAddCompany">{{ $t('添加门店') }}</el-button>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="$t('登录密码')" prop="password">
            <el-input type="password" v-model="formData.password" autocomplete="off" maxlength="50" show-password :placeholder="hasEdit ? $t('不修改请留空') : $t('请输入登录密码')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="$t('确认密码')" prop="confirm_password">
            <el-input type="password" v-model="formData.confirm_password" autocomplete="off" maxlength="50" show-password :placeholder="hasEdit ? $t('不修改请留空') : $t('请确认密码')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="emits('update:show', false)">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, computed } from 'vue';
  import { StaffAddType, StaffEditType, fetchStaffAdd, fetchStaffEdit } from '@/api/user/staff';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'change', value: any): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: StaffAddType & { uuid?: number; company_list?: any[] };
      companyList?: any[];
      roleList?: { id: number; role_name: string }[];
    }>(),
    {
      show: false,
      detail: () => ({}),
      companyList: () => [],
      roleList: () => [],
    },
  );
  const hasEdit = computed(() => !!props.detail?.uuid);
  const formData = ref<StaffAddType & StaffEditType>({
    email: '',
    phone: '',
    password: '',
    confirm_password: '',
    real_name: '',
    company_uuid: undefined,
    role_uuids: [],
    company_list: [],
  });
  const formRules = reactive({
    email: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          const re = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
          if (!value) {
            callback(new Error($t('请输入邮箱')));
          } else if (!re.test(value)) {
            callback(new Error($t('请确认邮箱格式')));
          }
          callback();
        },
      },
    ],
    phone: [
      {
        required: false,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          // 手机号允许空字符串，如果填写则需要验证格式
          if (value && value.trim() !== '' && !/^1[3-9]\d{9}$/.test(value)) {
            callback(new Error($t('请输入正确的手机号格式')));
          }
          callback();
        },
      },
    ],
    real_name: [{ required: false, message: $t('请输入姓名'), trigger: 'blur' }],
    company_uuid: [{ required: true, message: $t('请选择门店'), trigger: ['change', 'blur'] }],
    role_uuids: [{ required: false, message: $t('请选择角色'), trigger: ['change', 'blur'] }],
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
    company_list: [
      {
        required: false,
        trigger: ['change', 'blur'],
        validator: (_: object, value: any[], callback: any) => {
          if (hasEdit.value && (!value || value.length === 0)) {
            callback(new Error($t('请至少添加一个门店')));
          }
          callback();
        },
      },
    ],
  });
  const formLoading = ref(false);
  const formElement = ref();

  const handleAddCompany = () => {
    formData.value.company_list = formData.value.company_list || [];
    formData.value.company_list.push({
      company_uuid: undefined,
      role_uuids: [],
    });
  };

  const handleRemoveCompany = (index: number) => {
    formData.value.company_list?.splice(index, 1);
  };

  const handleCompanyChange = (index: number) => {
    // 当门店改变时，可以动态加载该门店的角色列表
    // 这里暂时使用全局角色列表
  };

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        const data: any = { ...formData.value };
        formLoading.value = true;
        let res = null;
        
        if (hasEdit.value) {
          // 编辑时，构建 company_list
          if (data.company_list && data.company_list.length > 0) {
            data.company_list = data.company_list.map((item: any) => ({
              company_uuid: item.company_uuid,
              role_uuids: item.role_uuids || [],
            }));
          }
          res = await fetchStaffEdit({ ...data, uuid: props.detail?.uuid });
        } else {
          // 新增时，不需要 company_list
          delete data.company_list;
          res = await fetchStaffAdd(data);
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
      formRules.company_uuid[0].required = !hasEdit.value;
      
      if (hasEdit.value) {
        // 编辑模式
        const detail = props.detail;
        const companyList = detail?.company_list || [];
        formData.value = {
          email: detail?.email || '',
          phone: detail?.phone || '',
          password: '',
          confirm_password: '',
          real_name: detail?.real_name || '',
          company_list: companyList.length > 0 
            ? companyList.map((item: any) => ({
                company_uuid: item.company_uuid,
                role_uuids: item.roles?.map((r: any) => r.role_uuid) || [],
              }))
            : [{ company_uuid: undefined, role_uuids: [] }],
        };
      } else {
        // 新增模式
        formData.value = {
          email: '',
          phone: '',
          password: '',
          confirm_password: '',
          real_name: '',
          company_uuid: undefined,
          role_uuids: [],
          company_list: [],
        };
      }
      
      // 清空校验
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>

<style lang="scss" scoped></style>
