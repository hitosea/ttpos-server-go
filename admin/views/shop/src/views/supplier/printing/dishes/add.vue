<template>
  <div>
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      :selectorType="form.product_method == 1 ? 'category' : form.product_method == 2 ? 'label' : 'all'"
      :selectedProductIds="
        form.product_method == 1
          ? printProductsDataByCategory.map((item) => item.product_id)
          : form.product_method == 2
          ? printProductsDataByPrintTag.map((item) => item.product_id)
          : []
      "
    >
    </ProductSelector>
    <el-dialog class="product-add" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('添加商品打印')">
      <!--form表单-->
      <el-form size="small" ref="formRef" :model="form" label-position="top">
        <!--添加门店-->
        <el-form-item
          for="no_click"
          :label="$t('名称')"
          prop="name"
          :rules="[
            { required: true, message: $t('请输入名称') },
            { validator: uniqueNameValidator('supplier_printing', undefined, 'SINGLE'), trigger: 'blur' },
          ]"
          ><el-input v-model="form.name" :placeholder="$t('请输入名称')" :maxlength="50"></el-input
        ></el-form-item>
        <el-form-item for="no_click" :label="$t('是否开启')" prop="is_open" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio v-model="form.is_open" :label="1">{{ $t('开启') }}</el-radio>
            <el-radio v-model="form.is_open" :label="0">{{ $t('关闭') }}</el-radio>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印模式')" prop="print_type" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio v-model="form.print_type" :label="10">{{ $t('付款打印') }}</el-radio>
            <el-radio v-model="form.print_type" :label="30">{{ $t('送厨打印') }}</el-radio>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('按区域打印')">
          <el-select v-model="form.area_id" multiple :placeholder="$t('全部区域')">
            <el-option v-for="(item, index) in areaData" :key="index" :label="item.area_name" :value="item.area_id"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印方式')" prop="print_method" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio-group
              v-model="form.print_method"
              @change="
                () => {
                  form.is_open_one_food = 0;
                  form.print_select = 1;
                }
              "
            >
              <el-radio :label="10">{{ $t('整单打印') }}</el-radio>
              <el-radio :label="40">{{ $t('按一菜一单打印') }}</el-radio>
            </el-radio-group>
          </div>
        </el-form-item>

        <el-form-item for="no_click" :label="$t('打印商品选择')" prop="product_ids" :rules="[{ required: true, message: $t('请选择商品') }]">
          <div class="print_products_wrap">
            <div>
              <el-radio-group v-model="form.product_method">
                <el-radio :label="1">{{ $t('按商品分类') }}</el-radio>
                <el-radio :label="2">{{ $t('按打印标签') }}</el-radio>
              </el-radio-group>
            </div>
            <div class="print_products_content">
              <!-- <div class="print_products_content_trigger" @click="() => (openProductSelector = true)"></div> -->
              <el-select v-model="form.product_ids" :placeholder="$t('请选择商品')" multiple value-key="product_id" :style="{ display: 'none' }">
                <el-option v-for="item in form.product_ids" :key="item.product_id" :label="item.product_name_text" :value="item"></el-option>
              </el-select>

              <template v-if="form.product_method == 1">
                <div class="print-products-selector" @click="handleOpenProductSelector">
                  <div class="print-products-selector-content">
                    <el-tag v-for="tag in printProductsDisplayByCategory" :key="tag.category_id" closable type="info" @close="($event) => handleCloseTag($event, tag)">
                      {{ tag.name_text }} ({{ tag.count }})
                    </el-tag>
                  </div>
                  <div class="el-select__suffix">
                    <el-icon>
                      <ArrowDown />
                    </el-icon>
                  </div>
                </div>
              </template>

              <template v-if="form.product_method == 2">
                <div class="print-products-selector" @click="handleOpenProductSelector">
                  <div class="print-products-selector-content">
                    <el-tag v-for="tag in printProductsDisplayByPrintTag" :key="tag.label_id" closable type="info" @close="($event) => handleCloseTag($event, tag)">
                      {{ tag.label_name_text }} ({{ tag.count }})
                    </el-tag>
                  </div>
                  <div class="el-select__suffix">
                    <el-icon>
                      <ArrowDown />
                    </el-icon>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.type == 10" :label="$t('打印机')" prop="printer_id" :rules="[{ required: true, message: $t('请选择打印机') }]">
          <el-select v-model="form.printer_id" :placeholder="$t('请选择')" multiple>
            <el-option v-for="(item, index) in type" :key="index" :label="item.printer_name + (item.is_usb == 1 ? ' (USB)' : '')" :value="item.printer_id">
              {{ item.printer_name }}
              <el-tag v-if="item.is_usb == 1" size="small" type="warning">USB</el-tag>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.type == 20" :label="$t('打印机')" prop="printer_id" :rules="[{ required: true, message: $t('请选择打印机') }]">
          <el-select v-model="form.printer_id" :placeholder="$t('请选择')" multiple>
            <el-option v-for="(item, index) in typeTag" :key="index" :label="item.printer_name + (item.is_usb == 1 ? ' (USB)' : '')" :value="item.printer_id">
              {{ item.printer_name }}
              <el-tag v-if="item.is_usb == 1" size="small" type="warning">USB</el-tag>
            </el-option>
          </el-select>
        </el-form-item>

        <!-- <el-form-item for="no_click" v-if="form.product_type == 0 && form.print_method == 20" :label="$t('商品分类')"
          prop="category_id" :rules="[{ required: true, message: '请选择商品分类' }]">
          <el-select v-model="form.category_id" multiple :placeholder="$t('请选择')">
            <el-option v-for="item in storeList" :key="item.category_id" :label="item.name_text"
              :value="item.category_id + ''"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.product_type == 1 && form.print_method == 20" :label="$t('商品分类')"
          prop="category_id" :rules="[{ required: true, message: $t('请选择商品分类') }]">
          <el-select v-model="form.category_id" multiple :placeholder="$t('请选择')">
            <el-option v-for="item in storeList" :key="item.category_id" :label="item.name_text"
              :value="item.category_id + ''"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.print_method == 30" :label="$t('打印标签')" prop="label_id">
          <el-select v-model="form.label_id" multiple :placeholder="$t('请选择')">
            <el-option v-for="item in labelList" :key="item.label_id" :label="item.label_name_text"
              :value="item.label_id + ''"></el-option>
          </el-select>
          <div class="tips">{{ $t('不选择打印全部') }}</div>
        </el-form-item>

        <el-form-item for="no_click" v-if="form.print_method == 20 || form.print_method == 30" :label="$t('按一菜一单打印')"
          prop="is_open_one_food" :rules="[{ required: true, message: '' }]">
          <div>
            <el-radio v-model="form.is_open_one_food" :label="0">{{ $t('关闭') }}</el-radio>
            <el-radio v-model="form.is_open_one_food" :label="1">{{ $t('开启') }}</el-radio>
          </div>
        </el-form-item> -->

        <el-form-item
          for="no_click"
          v-if="form.print_method == 40 || form.is_open_one_food == 1"
          :label="$t('打印')"
          prop="print_select"
          :rules="[{ required: true, message: '' }]"
        >
          <div>
            <el-radio v-model="form.print_select" :label="1">{{ $t('合并') }}</el-radio>
            <el-radio v-model="form.print_select" :label="2">{{ $t('分开') }}</el-radio>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleClose">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import SupplierApi from '@/api/supplier.js';
  import StoreApi from '@/api/store.js';
  import ProductSelector from '@/components/product/Selector.vue';
  import { uniqueNameValidator } from '@/utils/form.js';

  export default {
    name: 'SupplierPrintingDishesAdd',
    components: {
      ProductSelector,
    },
    data() {
      return {
        /*切换菜单*/
        // activeIndex: '1',
        /*form表单数据*/
        form: {
          name: '',
          is_open: 1,
          printer_id: [],
          product_type: 0,
          print_type: 10,
          category_id: [],
          area_id: '',
          type: 10,
          print_method: 10,
          label_id: [],
          is_open_one_food: 0,
          print_select: 1,
          product_method: 1,
          product_ids: [],
        },
        loading: false,
        dialogVisible: false,
        type: [],
        typeTag: [],
        storeList: [],
        takeList: [],
        labelList: [],
        //
        options: [],
        categoryIds: [],
        areaData: [],
        openProductSelector: false,
        productSelectorList: [],

        printProductsDataByCategory: [],
        printProductsDataByPrintTag: [],
      };
    },
    props: ['open_add'],
    created() {
      this.dialogVisible = this.open_add;
      this.getAreaData();
    },
    // watch: {
    //     'categoryIds': {
    //         handler(val) {
    //             this.form.category_id = [];
    //             this.categoryIds.map(h=>{
    //                 if (h[1]) {
    //                     this.form.category_id.push(h[1])
    //                 }
    //             })
    //             //
    //             this.$refs?.form?.validate(_=>{})
    //         },
    //         deep: true,
    //         immediate: true,
    //     }
    // },
    computed: {
      printProductsDisplayByCategory() {
        const selectedCategoryIds = this.printProductsDataByCategory.map((item) => item.category_id);
        const list = [];

        for (const item of this.storeList) {
          let count = 0;
          if (selectedCategoryIds.includes(item.category_id)) {
            count = this.printProductsDataByCategory.filter((val) => val.category_id == item.category_id).length;
          }
          for (const child of item.child) {
            if (selectedCategoryIds.includes(child.category_id)) {
              count += this.printProductsDataByCategory.filter((val) => val.category_id == child.category_id).length;
            }
          }
          if (count > 0) {
            list.push({
              ...item,
              count,
            });
          }
        }

        return list;
      },
      printProductsDisplayByPrintTag() {
        const selectedPrintTagIds = this.printProductsDataByPrintTag.map((item) => item.label_id);
        return this.labelList
          .filter((item) => selectedPrintTagIds.includes(item.label_id))
          .map((item) => ({ ...item, count: this.printProductsDataByPrintTag.filter((val) => val.label_id == item.label_id).length }));
      },
    },
    methods: {
      getData() {
        SupplierApi.getPrinting({}, true)
          .then((data) => {
            this.storeList = data.data.storeList;
            this.takeList = data.data.takeList;
            this.type = data.data.printerList;
            this.typeTag = data.data.printerTagList;
            this.labelList = data.data.labelList;
            //
            this.options = [];
            this.storeList?.map((item) => {
              if (item.parent_id == 0) {
                let children = [];
                this.storeList?.map((val) => {
                  if (item.category_id == val.parent_id) {
                    children.push({
                      value: val.category_id,
                      label: val.name_text,
                      children: [],
                    });
                  }
                });
                this.options.push({
                  value: item.category_id,
                  label: item.name_text,
                  children: children,
                });
              }
            });
          })
          .catch(() => {});
      },

      /*获取列表*/
      getAreaData() {
        let self = this;
        self.loading = true;
        StoreApi.arealist({}, true)
          .then((data) => {
            self.loading = false;
            self.areaData = data.data.list.data.map((item) => {
              return {
                area_id: item.area_id.toString(),
                area_name: item.area_name,
              };
            });
            self.areaData.unshift({
              area_id: '0',
              area_name: this.$t('无区域 (点餐无桌台)'),
            });
            this.getData();
          })
          .catch(() => {
            self.loading = false;
          });
      },

      //提交表单
      onSubmit() {
        let self = this;
        let form = JSON.parse(JSON.stringify(self.form));
        //
        if (!form.print_method == 20) {
          form.category_id = [];
        }
        form.area_id = (form.area_id || []).filter((id) => id);
        form.area_id = (form.area_id || []).length > 0 ? form.area_id : '';
        form.product_ids =
          this.form.product_method == 1
            ? (this.printProductsDataByCategory || []).map((item) => item.product_id)
            : this.form.product_method == 2
            ? (this.printProductsDataByPrintTag || []).map((item) => item.product_id)
            : [];
        self.$refs.formRef.validate((valid) => {
          if (valid) {
            self.loading = true;
            SupplierApi.addPrinting(form, true)
              .then(() => {
                self.loading = false;
                self.$ElMessage({
                  message: self.$t('添加成功'),
                  type: 'success',
                });
                self.$emit('close', 1);
              })
              .catch(() => {
                self.loading = false;
              });
          }
        });
      },
      handleClose() {
        this.$emit('close');
      },
      handleProductSelectorClose(list) {
        if (Array.isArray(list)) {
          this.form.product_ids = list;
          if (this.form.product_method == 1) {
            this.printProductsDataByCategory = list;
          }
          if (this.form.product_method == 2) {
            this.printProductsDataByPrintTag = list;
          }
        }
        this.openProductSelector = false;
      },

      uniqueNameValidator: uniqueNameValidator,
      handleOpenProductSelector() {
        if (this.loading) return;
        this.openProductSelector = true;
      },
      handleCloseTag(event, tag) {
        event.stopPropagation();
        if (this.form.product_method == 1) {
          this.form.product_ids = this.form.product_ids.filter(
            (item) => item.category_id !== tag.category_id && item.parent_category_id !== tag.category_id && item.parent_id !== tag.category_id
          );
          this.printProductsDataByCategory = this.form.product_ids;
        }

        if (this.form.product_method == 2) {
          this.form.product_ids = this.form.product_ids.filter((item) => item.label_id !== tag.label_id);
          this.printProductsDataByPrintTag = this.form.product_ids;
        }
      },
    },
  };
</script>
<style scoped>
  :deep(.el-select--small .el-select__wrapper) {
    min-height: 32px !important;
    height: auto;
    padding: 4px 8px !important;
  }

  .print_products_wrap {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: stretch;
  }

  .print_products_content {
    position: relative;

    .print_products_content_trigger {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      z-index: 10;
    }

    .print-products-selector {
      display: flex;
      align-items: center;
      position: relative;
      box-sizing: border-box;
      cursor: pointer;
      text-align: left;
      gap: 4px;

      border-radius: var(--el-border-radius-base);
      padding: 4px 8px !important;
      background-color: var(--el-fill-color-blank);
      transition: var(--el-transition-duration);
      box-shadow: 0 0 0 1px var(--el-border-color) inset;

      height: auto;
      min-height: 32px !important;
      font-size: 14px;

      .print-products-selector-content {
        margin-left: -6px;
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
        align-items: center;
        position: relative;
        flex: 1;
        min-width: 0;
      }

      .print-products-selector-suffix {
        gap: 4px;
        display: flex;
        align-items: center;
        flex-shrink: 0;
        color: var(--el-input-icon-color, var(--el-text-color-placeholder));
      }
    }
  }
</style>
