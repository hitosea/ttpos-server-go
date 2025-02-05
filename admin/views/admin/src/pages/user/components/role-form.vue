<template>
  <el-dialog
    width="640"
    :title="props.detail?.id ? $t('编辑角色') : $t('添加角色')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="emits('update:show', false)"
  >
    <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="auto">
      <el-form-item :label="$t('角色名称：')" prop="role_name">
        <el-input v-model="formData.role_name" type="text" maxlength="50" clearable :placeholder="$t('请输入角色名称')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('权限：')" prop="access_id">
        <div v-if="props.show" class="rounded py-2 px-3 bg-[#f6f8fb] w-full max-h-[500px] overflow-auto">
          <el-tree
            :data="props.roleList"
            show-checkbox
            node-key="id"
            :props="{ children: 'children', label: 'name' }"
            :default-expand-all="true"
            :default-checked-keys="selectMenu"
            @check="handleCheckChange"
          ></el-tree>
        </div>
      </el-form-item>
      <!-- <el-form-item :label="$t('排序')" prop="sort" :rules="[{ required: true, message: $t('请输入排序') }]">
        <el-input-number v-model.number="formData.sort" :controls="false" :min="0" :max="999" style="width: 100%" :placeholder="$t('接近0，排序等级越高')"></el-input-number>
      </el-form-item> -->
    </el-form>
    <template #footer>
      <el-button @click="emits('update:show', false)">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, reactive, watch } from 'vue';
  import type { FormRules } from 'element-plus';
  import { RoleListData, RoleEditType, fetchRoleAdd, fetchRoleEdit } from '@/api/user/role';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'change', value: any): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: RoleListData;
      roleList?: { id: number; role_name: string }[];
      roleChecked?: number[];
    }>(),
    {
      show: false,
      detail: () => ({}),
      roleList: () => [],
      roleChecked: () => [],
    },
  );
  const formData = ref<RoleEditType>({
    role_name: '',
    access_id: [],
    sort: 0,
  });
  const formRules = reactive<FormRules>({
    role_name: [{ required: true, message: $t('请输入角色名称'), trigger: 'blur' }],
    access_id: [{ required: true, message: $t('请选择权限'), trigger: 'blur' }],
  });
  const formLoading = ref(false);
  const formElement = ref();
  const selectMenu = ref<number[]>();

  const handleSubmit = () => {
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) return;
      try {
        formLoading.value = true;
        let res = null;
        if (props.detail?.id) {
          res = await fetchRoleEdit({ ...formData.value, id: props.detail?.id });
        } else {
          res = await fetchRoleAdd(formData.value);
        }
        message.success(res?.msg);
        emits('update:show', false);
        emits('change', res);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  const handleCheckChange = (_: any, checked: any) => {
    // formData.value.access_id = checked.checkedKeys.concat(checked.halfCheckedKeys);
    formData.value.access_id = checked.checkedKeys;
    formElement.value?.validateField('access_id').catch(() => {});
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) return;
      formElement.value?.resetFields();
      selectMenu.value = props.roleChecked || [];
      //
      formData.value = {
        role_name: props.detail.role_name || '',
        access_id: props.roleChecked || [],
        sort: props.detail.sort || 0,
      };
      // 清空
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>

<style lang="scss" scoped>
  :deep(.el-tree-node__label) {
    text-transform: capitalize;
  }
</style>
