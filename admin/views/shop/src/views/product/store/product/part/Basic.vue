<template>
  <div class="basic-setting-content">
    <!--基本信息-->
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-form-item :label="$t('类型：')" v-if="baseSale == '1'">
      <el-radio-group v-model="form.model.type" :disabled="disableChangeType" @change="changeType">
        <el-radio :value="10">{{ $t('成品') }}</el-radio>
        <el-radio :value="30">{{ $t('套餐') }}</el-radio>
        <el-radio :value="20">{{ $t('材料') }}</el-radio>
      </el-radio-group>
    </el-form-item>

    <UniqueNameForm
      ref="uniqueNameFormRef"
      :labelPrefix="form.model.type == 30 ? $t('套餐名称') : $t('商品名称')"
      apiSource="product"
      width="460px"
      :maxlength="150"
      :apiId="form.model.product_id ? form.model.product_id : undefined"
      :overrideLanguages="form.model.product_id ? form.model.product_name : undefined"
    />

    <el-form-item for="no_click" :label="$t('所属分类：')" :rules="[{ required: true, message: $t('请选择所属分类') }]" prop="model.category_id">
      <el-cascader class="max-w460 mr8" :options="options" v-model="form.model.category_id" clearable style="width: 100%" :placeholder="$t('请选择分类')"></el-cascader>
      <el-button size="small" type="primary" class="el-icon-circle-plus" @click="addCategory">{{ $t('添加分类') }}+</el-button>
    </el-form-item>
    <el-form-item for="no_click" :label="$t('供应商：')" v-if="baseSale == '1' && form.model.type == 20">
      <el-select v-model="form.model.erp_supplier_id" filterable clearable class="max-w460" size="default" :placeholder="$t('请选择供应商')">
        <template v-for="item in supplierList" :key="item.id">
          <el-option :value="item.id" :label="item.name"></el-option>
        </template>
      </el-select>
    </el-form-item>

    <el-form-item for="no_click" :label="$t('商品图片：')" prop="model.image">
      <div class="draggable-list">
        <transition-group>
          <div class="item" v-for="(item, index) in form.model.image" :key="item.file_path">
            <img v-img-url="item.file_path" />
            <a href="javascript:void(0);" class="delete-btn" @click.stop="deleteImg(index)"
              ><el-icon> <Close /> </el-icon
            ></a>
          </div>
        </transition-group>

        <div v-if="form.model.image && form.model.image.length == 0" class="item img-select" @click="openProductUpload">
          <el-icon>
            <Plus />
          </el-icon>
        </div>
      </div>
      <div class="gray9">
        {{ $t('支持JPG、JPEG、PNG、WEBP格式，小于15MB，尺寸：160*120px') }}
      </div>
    </el-form-item>
    <el-form-item
      for="no_click"
      :label="$t('图片名称：')"
      prop="model.img_name"
      :rules="[{ validator: uniqueNameValidator('product_img', form.model.product_id, 'SINGLE', undefined, false), trigger: 'blur' }]"
    >
      <el-input type="text" :placeholder="$t('请输入图片名称')" v-model="form.model.img_name" :maxlength="50" class="max-w460"></el-input>
    </el-form-item>

    <el-form-item for="no_click" :label="$t('套餐价格：')" prop="model.package_price" v-if="form.model.type == 30" :rules="[{ required: true, message: $t('请输入套餐价格') }]">
      <numInput type="text" :placeholder="$t('请输入套餐价格')" v-model="form.model.package_price" :maxlength="50" class="max-w460"></numInput>
    </el-form-item>

    <!--商品图片组件-->
    <Upload v-if="isProductUpload" :config="{ total: 1 }" :isupload="isProductUpload" :aspectRatio="1.333" @returnImgs="returnProductImgsFunc">{{ $t('上传图片') }}</Upload>

    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
  </div>
</template>

<script setup>
  import { ref, inject, watch, onMounted, nextTick } from 'vue';
  import Upload from '@/components/file/Upload.vue';
  import Add from '../../category/Add.vue';
  import PurchaseApi from '@/api/purchase.js';
  import { useUserStore } from '@/store';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
  import { uniqueNameValidator } from '@/utils/form.js';

  // 获取用户信息
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;

  // 定义props
  const props = defineProps({
    disableChangeType: {
      type: Boolean,
      default: false,
    },
  });

  // 定义emit
  const emit = defineEmits(['validateField']);

  // 注入form
  const form = inject('form', {});

  // 响应式数据
  const isProductUpload = ref(false);
  const open_add = ref(false);
  const options = ref([]);
  const supplierList = ref([]);

  // 模板引用
  const uniqueNameFormRef = ref(null);

  // 组件挂载时初始化
  onMounted(() => {
    getData();
  });

  // 监听form变化
  watch(
    () => form,
    (val) => {
      options.value = [];
      val.category.map((item, index) => {
        options.value.push({
          value: item.category_id,
          label: item.name_text,
          children: [],
        });
        item.child.map((items) => {
          options.value[index].children.push({
            value: items.category_id,
            label: items.name_text,
          });
        });
      });
    },
    { deep: true, immediate: true }
  );

  // 方法定义
  //获取供应商
  const getData = async () => {
    try {
      let Params = {};
      Params.list_rows = 1000;
      const data = await PurchaseApi.supplierList(Params, true);
      supplierList.value = data.data.list.data;
      nextTick(() => {
        let arr = [];
        supplierList.value.map((item) => {
          arr.push(item.id);
        });
        if (!arr.includes(form.model.erp_supplier_id)) {
          form.model.erp_supplier_id = '';
        }
      });
    } catch (error) {
      // 错误处理
    }
  };

  const addCategory = () => {
    open_add.value = true;
  };

  /*关闭弹窗*/
  const closeDialogFunc = (e, f) => {
    if (f == 'add') {
      open_add.value = e.openDialog;
      if (e.type == 'success' && e.data) {
        e.data.parent_id = Number(e.data.parent_id);
        if (e.data.parent_id) {
          form.category.map((item) => {
            if (item.category_id == e.data.parent_id) {
              item.child.push(e.data);
            }
          });
        } else {
          e.data.child = [];
          form.category.unshift(e.data);
        }
      }
    }
  };

  const changeType = () => {
    form.model.sku[0].material = [];
    form.model.sku = form.model.sku.slice(0, 1);
    form.single_select_list = [];
    form.many_select_list = [[]];
    form.ing_select_list = [[]];
    form.model.type == 10 ? (form.model.spec_type = 20) : (form.model.spec_type = 10);
  };

  /*打开上传图片*/
  const openProductUpload = () => {
    isProductUpload.value = true;
  };

  /*上传商品图片*/
  const returnProductImgsFunc = (e) => {
    if (e != null) {
      let imgs = form.model.image.concat(e);
      form.model['image'] = imgs;
    }
    isProductUpload.value = false;
    emit('validateField', 'model.image');
  };

  /*删除商品图片*/
  const deleteImg = (index) => {
    form.model.image.splice(index, 1);
  };
</script>

<style lang="scss" scoped></style>
