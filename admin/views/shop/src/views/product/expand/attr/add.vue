<!-- 描述：商品-属性库-添加属性 -->
<template>
  <el-dialog :title="$t('添加属性')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <el-form-item>
        <el-radio-group v-model="type">
          <el-radio :value="1">{{ $t('属性组') }}</el-radio>
          <el-radio :value="2">{{ $t('属性值') }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <template v-if="type === 2">
        <el-form-item
          for="no_click"
          :label="$t('属性组')"
          prop="parent_id"
          :rules="[
            { required: true, message: $t('请选择属性组') },
            {
              validator: (rule, value, callback) => {
                return typeof value === 'number' && value > 0 ? callback() : callback($t('请选择属性组'));
              },
            },
          ]"
        >
          <el-select v-model="form.parent_id" :placeholder="$t('请选择属性组')">
            <template v-for="item in groupData" :key="item.attribute_id">
              <el-option :value="item.attribute_id" :label="item.attribute_name_text">{{ item.attribute_name_text }}</el-option>
            </template>
          </el-select>
        </el-form-item>
      </template>

      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('属性名称')" apiSource="attribute" :parent_id="type == '2' ? form.parent_id : 0" />
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ $t('取消') }}</el-button>
        <el-tooltip effect="dark" placement="top" :content="$t('如长时间无响应，请刷新后重试。')">
          <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
        </el-tooltip>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, reactive, onMounted, getCurrentInstance } from 'vue';
  import ProductApi from '@/api/product.js';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';

  // 获取组件实例
  const { proxy } = getCurrentInstance();

  // 定义props
  const props = defineProps({
    open_add: {
      type: Boolean,
      default: false,
    },
    addform: {
      type: Object,
      default: () => ({}),
    },
  });

  // 定义emits
  const emit = defineEmits(['closeDialog']);

  // 响应式数据
  const formRef = ref(null);
  const uniqueNameFormRef = ref(null);
  const dialogVisible = ref(false);
  const loading = ref(false);
  const type = ref(2); // 1: 属性组，2：属性值
  const groupData = ref([]);

  const form = reactive({
    attribute_name: {},
    parent_id: null,
  });

  // 初始化数据
  onMounted(() => {
    dialogVisible.value = props.open_add;
    getGroupData();
  });

  // 提交方法
  const submit = async () => {
    loading.value = true;
    try {
      const validForm = await formRef.value.validate();
      if (!validForm) return;

      const validUniqueName = await uniqueNameFormRef.value.validate();
      if (!validUniqueName) return;

      const _name = uniqueNameFormRef.value.data;
      const params = JSON.parse(JSON.stringify(form));
      params.attribute_name = JSON.stringify(_name);

      if (type.value === 1) {
        params.parent_id = '';
      }

      const res = await ProductApi.addAttribute(params, true);
      proxy.$ElMessage({
        message: $t('保存成功'),
        type: 'success',
      });

      handleClose(true, res.data);
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  // 关闭弹窗
  const handleClose = (isSuccess = false, data) => {
    emit('closeDialog', {
      type: isSuccess ? 'success' : 'error',
      openDialog: false,
      data: data,
    });
  };

  // 获取属性组数据
  const getGroupData = async () => {
    try {
      const { data } = await ProductApi.AttributeList(
        {
          page: 1,
          list_rows: 10000,
          type: 1,
        },
        true
      );
      groupData.value = data.list.data;
    } catch (err) {
      // 错误处理
    }
  };
</script>
