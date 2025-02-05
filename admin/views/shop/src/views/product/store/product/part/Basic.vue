<template>
  <!--

          时间：2019-10-26
          描述：商品管理-商品编辑-基础信息
      -->
  <div class="basic-setting-content">
    <!--基本信息-->
    <div class="common-form">{{ $t('基本信息') }}</div>
    <el-form-item :label="$t('类型：')" v-if="baseSale == '1'">
      <el-radio-group v-model="form.model.type" :disabled="disableChangeType" @change="changeType">
        <el-radio :label="10">{{ $t('成品') }}</el-radio>
        <el-radio :label="20">{{ $t('材料') }}</el-radio>
      </el-radio-group>
    </el-form-item>

    <UniqueNameForm
      ref="uniqueNameFormRef"
      :labelPrefix="$t('商品名称')"
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
        <draggable class="wrapper" v-model="form.model.image">
          <transition-group>
            <div class="item" v-for="(item, index) in form.model.image" :key="item.file_path">
              <img v-img-url="item.file_path" />
              <a href="javascript:void(0);" class="delete-btn" @click.stop="deleteImg(index)"
                ><el-icon> <Close /> </el-icon
              ></a>
            </div>
          </transition-group>
        </draggable>
        <div v-if="form.model.image.length == 0" class="item img-select" @click="openProductUpload">
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
      :rules="[{ validator: uniqueNameValidator('product_img', formData.model.product_id, 'SINGLE', undefined, false), trigger: 'blur' }]"
    >
      <el-input type="text" :placeholder="$t('请输入图片名称')" v-model="form.model.img_name" :maxlength="50" class="max-w460"></el-input>
    </el-form-item>

    <!--商品图片组件-->
    <Upload v-if="isProductUpload" :config="{ total: 1 }" :isupload="isProductUpload" :aspectRatio="1.333" @returnImgs="returnProductImgsFunc">{{ $t('上传图片') }}</Upload>

    <!--添加-->
    <Add v-if="open_add" :open_add="open_add" @closeDialog="closeDialogFunc($event, 'add')"> </Add>
  </div>
</template>

<script>
  import Upload from '@/components/file/Upload.vue';
  import mInput from '@/components/m-input/index.vue';
  import Add from '../../category/Add.vue';
  import PurchaseApi from '@/api/purchase.js';
  import { useUserStore } from '@/store';
  import UniqueNameForm from '@/components/product/UniqueNameForm.vue';
  import { uniqueNameValidator } from '@/utils/form.js';

  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const baseSale = supplier.value?.sale_stock || 0;

  export default {
    name: 'ProductStoreProductPartBasic',
    components: {
      Upload,
      mInput,
      UniqueNameForm,
      Add,
    },
    data() {
      return {
        formData: {},
        isProductUpload: false,
        open_add: false,
        options: [],
        supplierList: [],
        baseSale: baseSale,
      };
    },
    props: {
      disableChangeType: {
        type: Boolean,
        default: false,
      },
    },
    inject: ['form'],
    watch: {
      form: {
        handler(val) {
          this.options = [];
          val.category.map((item, index) => {
            this.options.push({
              value: item.category_id,
              label: item.name_text,
              children: [],
            });
            item.child.map((items) => {
              this.options[index].children.push({
                value: items.category_id,
                label: items.name_text,
              });
            });
          });
        },
        deep: true,
        immediate: true,
      },
    },
    created() {
      this.formData = this.form;
      // this['formData'] = this.form;
      this.getData();
    },
    methods: {
      //获取供应商
      getData() {
        let self = this;
        let Params = {};
        Params.list_rows = 1000;
        PurchaseApi.supplierList(Params, true)
          .then((data) => {
            self.loading = false;
            self.supplierList = data.data.list.data;
            self.$nextTick(() => {
              let arr = [];
              self.supplierList.map((item) => {
                arr.push(item.id);
              });
              if (!arr.includes(self.form.model.erp_supplier_id)) {
                self.form.model.erp_supplier_id = '';
              }
            });
          })
          .catch(() => {});
      },

      addCategory() {
        this.open_add = true;
      },

      /*关闭弹窗*/
      closeDialogFunc(e, f) {
        if (f == 'add') {
          this.open_add = e.openDialog;
          if (e.type == 'success' && e.data) {
            e.data.parent_id = Number(e.data.parent_id);
            if (e.data.parent_id) {
              this.form.category.map((item) => {
                if (item.category_id == e.data.parent_id) {
                  item.child.push(e.data);
                }
              });
            } else {
              e.data.child = [];
              this.form.category.unshift(e.data);
            }
          }
        }
      },

      changeType() {
        this.form.model.sku[0].material = [];
        this.form.model.sku = this.form.model.sku.slice(0, 1);
        this.form.single_select_list = [];
        this.form.many_select_list = [[]];
        this.form.ing_select_list = [[]];
        this.form.model.type == 10 ? (this.form.model.spec_type = 20) : (this.form.model.spec_type = 10);
      },
      /*打开上传图片*/
      openProductUpload: function () {
        this.isProductUpload = true;
      },

      /*上传商品图片*/
      returnProductImgsFunc(e) {
        if (e != null) {
          let imgs = this.form.model.image.concat(e);
          this.form.model['image'] = imgs;
        }
        this.isProductUpload = false;
        this.$emit('validateField', 'model.image');
      },

      /*删除商品图片*/
      deleteImg(index) {
        this.form.model.image.splice(index, 1);
      },
      uniqueNameValidator: uniqueNameValidator,
    },
  };
</script>
