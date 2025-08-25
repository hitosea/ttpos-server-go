<!-- 描述：商品-加料库-添加加料 -->
<template>
  <el-dialog :title="$t('添加加料')" v-model="dialogVisible" @close="handleClose" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" ref="formRef">
      <UniqueNameForm ref="uniqueNameFormRef" :labelPrefix="$t('加料名称')" apiSource="feed" />

      <el-form-item
        for="no_click"
        :label="$t('价格')"
        prop="price"
        :rules="[
          { required: true, message: $t('请输入价格') },
          { type: 'number', message: $t('请输入数字') },
        ]"
      >
        <el-input-number :controls="false" :precision="2" :min="0" :max="100000000" :placeholder="$t('请输入价格')" v-model.number="form.price" autocomplete="off"></el-input-number>
      </el-form-item>

      <template v-if="baseSale == '1'">
        <el-form-item for="no_click" :label="$t('材料：')">
          <el-button type="primary" @click="addMaterials">{{ $t('添加材料') }}</el-button>
        </el-form-item>
        <div class="materials-one" label="" v-for="(item, index) in ing_select_list" :key="item.product_id">
          <el-form-item for="no_click" label="" class="max-w230">
            <el-input v-model="item.product_name_text" disabled></el-input>
          </el-form-item>
          <el-form-item
            for="no_click"
            label=""
            class="max-w230 flex"
            :prop="`material.${index}.material_num`"
            :rules="[
              {
                required: true,
                message: $t('请输入数量'),
              },
              {
                type: 'number',
                message: $t('请输入数字'),
              },
            ]"
          >
            <el-input-number
              :controls="false"
              :min="0"
              :placeholder="$t('请输入数量')"
              v-model="form.material[index].material_num"
              @change="(e) => handleInput(e, index)"
            ></el-input-number>
            <span class="unit">{{ item.product_unit_text }}</span>
          </el-form-item>
          <el-icon class="delete-icon" @click="handleDeleteOne(index)">
            <Delete />
          </el-icon>
        </div>
      </template>
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
  <ProductList v-if="open_product" :open_product="open_product" :multiple_selection="multiple_selection" :material_type="20" @closeDialogFunc="closeDialogFunc($event)">
  </ProductList>
</template>

<script setup>
import { ref, reactive, onMounted, getCurrentInstance, nextTick } from 'vue';
import ProductApi from '@/api/product.js';
import ProductList from '@/components/productList/productList.vue';
import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
import { DecimalPointFour } from '@/utils/formatPrice.js';
import { useUserStore } from '@/store';

// 获取组件实例
const { proxy } = getCurrentInstance();

// 获取用户store
const { computedSupplier } = useUserStore();
const supplier = computedSupplier().supplier;
const baseSale = supplier.value?.sale_stock || 0;

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
const open_product = ref(false);
const ing_select_list = ref([]);
const multiple_selection = ref([]);

const form = reactive({
  feed_name: {},
  price: NaN,
  material: [],
});

// 初始化数据
onMounted(() => {
  dialogVisible.value = props.open_add;
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
    params.feed_name = JSON.stringify(_name);

    const res = await ProductApi.addFeed(params, true);
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

// 添加材料
const addMaterials = () => {
  multiple_selection.value = ing_select_list.value;
  open_product.value = true;
};

// 关闭对话框
const closeDialogFunc = (e) => {
  open_product.value = e.openDialog;
  if (e.type == 'select') {
    let map = new Map();
    [ing_select_list.value, e.data].flat().forEach((obj) => map.set(obj.product_id, obj));
    ing_select_list.value = Array.from(map.values());

    let arr = [];
    if (form.material.length > 0) {
      form.material.map((item) => {
        arr.push(item.product_id);
      });
    }

    ing_select_list.value.map((item) => {
      if (!arr.includes(item.product_id)) {
        form.material.push({
          product_id: item.product_id,
          material_num: null,
        });
      }
    });
  }
};

// 删除一个材料
const handleDeleteOne = (index) => {
  ing_select_list.value.splice(index, 1);
  form.material.splice(index, 1);
};

// 处理输入
const handleInput = (val, index) => {
  // material_num 只能是数字
  nextTick(() => {
    form.material[index].material_num = DecimalPointFour(val);
  });
};
</script>

<style scoped lang="scss">
  .img {
    margin-top: 10px;
  }

  .materials-one {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 8px;

    .delete-icon {
      cursor: pointer;
      font-size: 24px;
      margin-top: -16px;
      margin-right: 0;
    }
  }

  .max-w230 {
    width: 100%;
  }
  .flex {
    :deep(.el-form-item__content) {
      display: flex;
      align-items: center;
      flex-wrap: nowrap;
      gap: 8px;
      .unit {
        flex-shrink: 1;
      }
    }
  }
</style>
