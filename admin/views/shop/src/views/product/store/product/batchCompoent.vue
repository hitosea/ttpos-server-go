<template>
  <div class="product-batch" v-loading="loading">
    <div class="common-return">
      <el-icon class="return-icon" @click="returnBack"> <ArrowLeftBold /> </el-icon>{{ title }}
    </div>
    <div class="product-body">
      <div class="common-form">
        {{ $t('选择商品') }}
      </div>
      <el-form size="small" ref="form" class="product-form" :model="form" label-position="top">
        <el-form-item
          :label="$t('商品')"
          for="no_click"
          prop="product_ids"
          :rules="[
            {
              validator: () => {
                return form.product_ids.length > 0 ? true : false;
              },
              required: true,
              message: $t('请选择商品'),
            },
          ]"
        >
          <div class="product-body">
            <div>
              <el-button size="small" type="primary" @click="addProduct" :loading="loading">
                {{ $t('选择商品') }}（{{ $t('已选商品') }}{{ product_list.length }}{{ $t('个') }}）
              </el-button>
            </div>
            <div class="product-list" v-if="product_list.length > 0">
              <template v-for="(item, index) in product_list" :key="index">
                <div class="product-item">
                  {{ item.product_name_text }}
                  <i class="el-icon el-tag__close" @click="removeProduct(index)">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
                      <path
                        fill="currentColor"
                        d="M764.288 214.592 512 466.88 259.712 214.592a31.936 31.936 0 0 0-45.12 45.12L466.752 512 214.528 764.224a31.936 31.936 0 1 0 45.12 45.184L512 557.184l252.288 252.288a31.936 31.936 0 0 0 45.12-45.12L557.12 512.064l252.288-252.352a31.936 31.936 0 1 0-45.12-45.184z"
                      ></path>
                    </svg>
                  </i>
                </div>
              </template>
            </div>
          </div>
        </el-form-item>
        <!-- 上传图片 -->
        <upImages
          v-if="type == 1 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="upImages"
          :product_list="product_list || []"
        ></upImages>
        <!-- 修改分类 -->
        <typeChange
          v-if="type == 2 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="typeChange"
          :product_list="product_list || []"
        ></typeChange>
        <!-- 修改税类 -->
        <taxChange
          v-if="type == 3 && product_list.length > 0"
          @close="dialogFormVisible"
          @loading="
            (e) => {
              loading = e;
            }
          "
          ref="taxChange"
          :product_list="product_list || []"
        ></taxChange>
      </el-form>
    </div>

    <div class="common-button-wrapper" v-if="product_list.length > 0">
      <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
      <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
    </div>

    <!-- 商品选择器 -->
    <ProductSelector
      v-if="openProductSelector"
      :open="openProductSelector"
      @close="handleProductSelectorClose"
      selectorType="all"
      type="all"
      :selectedProductIds="model?.product_ids?.map((item) => item.product_id) ?? []"
    >
    </ProductSelector>
  </div>
</template>
<script>
  import ProductSelector from '@/components/product/Selector.vue';
  import upImages from './components/upImages.vue';
  import typeChange from './components/typeChange.vue';
  import taxChange from './components/taxChange.vue';

  export default {
    components: { ProductSelector, upImages, typeChange, taxChange },

    data() {
      return {
        loading: false,
        openProductSelector: false,
        /*当前编辑的对象*/
        model: {},
        form: {
          product_ids: [],
          category_id: '',
          productTaxes: [
            {
              product_tax_type: '1',
              tax_category_id: '',
            },
            {
              product_tax_type: '2',
              tax_category_id: '',
            },
          ],
        },
        product_list: [],
        type: 1,
        title: '',
      };
    },
    provide: function () {
      return {
        form: this.form,
      };
    },
    watch: {
      product_list: {
        handler(val) {
          this.form.product_ids = val.map((item) => {
            return item.product_id;
          });

          if (this.$refs.form) {
            this.$refs.form.validateField('product_ids');
          }
        },
        deep: true,
        immediate: true,
      },
    },
    mounted() {
      this.type = this.$route.query.type;
      switch (this.type) {
        case '1':
          this.title = $t('修改图片');
          break;
        case '2':
          this.title = $t('修改分类');
          break;
        case '3':
          this.title = $t('修改税类');
          break;
        case '5':
          this.title = $t('商品批量导入');
          break;
      }
    },
    methods: {
      dialogFormVisible() {
        this.$router.go(-1);
      },
      handleClick() {
        this.$refs.form.validate((valid) => {
          if (valid) {
            if (this.type == 1) {
              this.$refs.upImages.repeatList();
            }
            if (this.type == 2) {
              this.$refs.typeChange.submit();
            }
            if (this.type == 3) {
              this.$refs.taxChange.submit();
            }
          }
        });
      },
      addProduct() {
        this.openProductSelector = true;
        this.model.product_ids = [];
        if (this.product_list.length > 0) {
          this.product_list.map((item) => {
            this.model.product_ids.push({
              product_id: item.product_id,
            });
          });
        }
      },
      handleProductSelectorClose(list, categories) {
        if (Array.isArray(list)) {
          this.product_list = list;
        }
        if (Array.isArray(categories)) {
          this.product_list.map((item) => {
            item.path_name_text = '';
            categories.map((item2) => {
              if (item.category_id == item2.category_id) {
                item.path_name_text = item2.path_name_text;
              }
              if (item2.child.length > 0) {
                item2.child.map((item3) => {
                  if (item.category_id == item3.category_id) {
                    item.path_name_text = item3.path_name_text;
                  }
                });
              }
            });
          });
        }

        this.openProductSelector = false;
      },
      returnBack() {
        this.$router.go(-1);
      },

      removeProduct(index) {
        this.product_list.splice(index, 1);
      },
    },
  };
</script>
<style lang="scss" scoped>
  .product-body {
    display: flex;
    flex-direction: column;
    width: 100%;
    overflow: auto;
    .product-list {
      display: flex;
      box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
      border-radius: 4px;
      padding: 6px 11px;
      margin-top: 16px;
      gap: 8px;
      flex-wrap: wrap;
      .product-item {
        color: var(--el-tag-text-color);
        display: inline-flex;
        justify-content: center;
        align-items: center;
        vertical-align: middle;
        font-size: var(--el-tag-font-size);
        line-height: 1;
        border-width: 1px;
        border-style: solid;
        box-sizing: border-box;
        white-space: nowrap;
        padding: 0 7px;
        height: 20px;
        font-size: 12px;
        border-radius: 4px;
        background-color: var(--el-color-info-light-9);
        border: solid 1px var(--el-color-info-light-8);
      }
      .el-tag__close {
        font-size: 12px;
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }
  .product-batch {
    display: flex;
    flex-direction: column;
    padding: 16px;
    background-color: #fff;
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    overflow: hidden;
  }

  .product-body {
    flex: 1 1 auto;
  }

  .common-return {
    font-size: 20px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 700;

    border-bottom: solid 1px var(--el-border-color);
    .return-icon {
      cursor: pointer;
    }
  }
  .common-button-wrapper {
    flex: 0 0 auto;
    flex-shrink: 0;
  }
</style>
