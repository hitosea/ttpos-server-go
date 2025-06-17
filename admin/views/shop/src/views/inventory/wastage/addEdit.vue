<template>
  <el-dialog :title="title" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <el-form size="small" :model="form" label-position="top" :rules="formRules" ref="form">
      <el-form-item :label="$t('类型')" prop="type" :rules="[{ required: true, message: '' }]">
        <el-radio-group v-model="form.type">
          <el-radio :label="1">{{ $t('丢失') }}</el-radio>
          <el-radio :label="2">{{ $t('损耗') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item
        label=""
        prop="product_id"
        :rules="[
          {
            required: true,
            validator: () => {
              return form.product_id ? true : false;
            },
            message: $t('请选中商品'),
          },
        ]"
      >
        <el-button type="primary" @click="selectList()">{{ $t('选中商品') }}</el-button>
      </el-form-item>
      <div class="product-content" v-if="tableData.length > 0">
        <div class="table-wrap">
          <el-table size="small" ref="multipleTable" :data="tableData" border style="width: 100%" v-loading="loading">
            <el-table-column :label="$t('类型')">
              <template #default="scope">
                {{ scope.row.product.type == 10 ? $t('成品') : $t('材料') }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('商品名称')" width="300px">
              <template #default="scope">
                <div class="product-info">
                  <div class="pic"><img v-img-url="scope.row.product.image[0]?.file_path" alt="" /></div>
                  <div class="info">
                    <div class="name">{{ scope.row.product.product_name_text }}</div>
                    <div class="price">{{ $t('销售价：') }}{{ scope.row.product_price }}</div>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="category.path_name_text" width="200" :label="$t('分类名称')">
              <template #default="scope">
                {{ scope.row.product.category.path_name_text || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="spec_name_text" :label="$t('规格')">
              <template #default="scope">
                {{ scope.row.spec_name_text || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="product_sales" :label="$t('实际销量')">
              <template #default="scope">
                {{ scope.row.product_sales }}
              </template>
            </el-table-column>
            <el-table-column prop="stock_num" :label="$t('库存')">
              <template #default="scope">
                {{ scope.row.product.type == 10 ? scope.row.stock_num : scope.row.material_stock }}
              </template>
            </el-table-column>
            <el-table-column prop="create_time" :label="$t('添加时间')" width="180">
              <template #default="scope">
                <p class="create-time">{{ scope.row.create_time.split(' ')[0] || '-' }}</p>
                <p class="create-time">{{ scope.row.create_time.split(' ')[1] || '' }}</p>
              </template>
            </el-table-column>
            <el-table-column prop="create_time" fixed="right" :label="$t('报损数量')" width="180">
              <template #default="scope">
                <el-form-item label="" prop="num" style="margin-top: 16px" :rules="[{ required: true, message: $t('请输入') }]">
                  <el-input-number :controls="false" :min="0" :placeholder="$t('请输入')" v-model.number="form.num"></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <el-form-item :label="$t('备注')" prop="remark" :rules="[{ required: true, message: $t('请输入备注') }]">
        <el-input type="text" v-model="form.remark" :placeholder="$t('请输入备注')" :maxlength="100"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible()">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="submit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
  <skuProductList v-if="open_product" :open_product="open_product" :multiple_choice="1" @closeDialogFunc="closeDialogFunc($event)"> </skuProductList>
</template>
<script>
  import skuProductList from '@/components/skuProductList/skuProductList.vue';
  import InventoryApi from '@/api/inventory.js';
  export default {
    components: { skuProductList },
    props: {
      open_dialog: {
        type: Boolean,
        default: false,
      },
      title: {
        default: '',
      },
      editData: {
        default: '',
      },
      formRules: [],
    },
    data() {
      return {
        dialogVisible: false,
        loading: false,

        form: {
          type: 1,
          product_id: '',
          num: null,
          remark: null,
          product_sku_id: '',
        },

        //材料
        open_product: false,
        select_list: [],
        tableData: [],
      };
    },
    watch: {
      'form.num': {
        handler(val) {
          if (val) {
            //成品整数
            if (this.tableData[0].product.type == 10) {
              this.$nextTick(() => {
                this.form.num = Number(String(val).replace(/(\.\d{0,0})\d*/, '$1'));
              });
            }

            // 材料库存4位小数
            if (this.tableData[0].product.type == 20) {
              this.$nextTick(() => {
                this.form.num = Number(String(val).replace(/(\.\d{1,4})\d*/, '$1'));
              });
            }
          }
        },
        deep: true,
        immediate: true,
      },
    },
    created() {
      this.dialogVisible = this.open_dialog;
      if (this.editData) {
        this.form.type = this.editData.type;
        this.form.num = this.editData.num;
        this.form.remark = this.editData.remark;
        this.form.product_id = this.editData.product_id;
        this.form.product_sku_id = this.editData.product_sku_id;
        this.tableData.push(this.editData.sku);
      }
    },
    methods: {
      /*提交*/
      submit() {
        let self = this;
        self.$refs.form.validate((valid) => {
          if (valid) {
            let params = JSON.parse(JSON.stringify(self.form));
            if (params.num == 0) {
              this.$ElMessage({
                message: $t('报损数量不能为0！'),
                type: 'warning',
              });
              return;
            }
            self.loading = true;
            if (this.editData) {
              params.id = this.editData.id;
              InventoryApi.erpDamagedProductRecordUpdate(params, true)
                .then((data) => {
                  self.loading = false;
                  this.$ElMessage({
                    message: $t('操作成功'),
                    type: 'success',
                  });
                  self.dialogFormVisible(true);
                })
                .catch((error) => {
                  self.loading = false;
                });
            } else {
              InventoryApi.erpDamagedProductRecordAdd(params, true)
                .then((data) => {
                  self.loading = false;
                  this.$ElMessage({
                    message: $t('操作成功'),
                    type: 'success',
                  });
                  self.dialogFormVisible(true);
                })
                .catch((error) => {
                  self.loading = false;
                });
            }
          }
        });
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        if (e) {
          this.$emit('closeDialog', {
            type: 'success',
            openDialog: false,
          });
        } else {
          this.$emit('closeDialog', {
            type: 'error',
            openDialog: false,
          });
        }
      },

      selectList() {
        this.open_product = true;
      },

      closeDialogFunc(e) {
        this.open_product = e.openDialog;
        if (e.type == 'select') {
          this.tableData = e.data;
          this.form.product_id = e.data[0].product_id;
          this.form.product_sku_id = e.data[0].product_sku_id;
        }
      },
    },
  };
</script>
<style lang="scss" scoped>
  .select-list {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding: 7px 7px 0 0;
    max-height: 400px;
    overflow: auto;
  }

  .customer-type {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding: 7px 7px 0 0;
    max-height: 400px;
    overflow: auto;

    .customer-button {
      width: 100%;
      display: flex;
      border: solid 1px var(--el-color-tips);
      border-radius: 4px;
      padding: 6px 12px;
      gap: 12px;
      align-items: center;
      position: relative;

      .delete-icon {
        position: absolute;
        right: -7px;
        top: -7px;
        cursor: pointer;
        color: #c80000;
      }
    }
  }
</style>
