<template>
  <div class="import-product">
    <div class="common-return">
      <el-icon class="return-icon" @click="dialogFormVisible"> <ArrowLeftBold /> </el-icon>{{ $t('导入商品') }}
    </div>

    <div v-loading="loading" class="dialog-content">
      <div v-if="step == 1">
        <div class="common-form">
          {{ $t('第一步，添加商品分类、税率管理-税类、规格库、单位库') }}
        </div>
        <div class="common-form">
          {{ $t('第二步，下载 Excel 模板，安装模板填写商品信息') }}
        </div>
        <el-button type="primary" @click="handleDownload" :loading="loading">
          {{ $t('下载') }}
        </el-button>
        <div class="common-form mt-24">
          {{ $t('第三步，上传文件并确认商品后提交') }}
        </div>

        <el-upload class="upload-demo" drag accept=".xls,.xlsx" :before-upload="onBeforeUpload">
          <SvgIcon class="data-box-icon" name="excel"></SvgIcon>
          <div class="el-upload__text">
            {{ $t('将文件拖拽到此处，或') }} <em>{{ $t('点击上传') }}</em>
          </div>
          <div class="el-upload__text">
            {{ $t('每次导入商品数量应小于1000条') }}
          </div>
        </el-upload>
      </div>
      <div v-if="step == 2">
        <el-form ref="tableFormRef" :model="nowTableData">
          <el-table class="ti-table-import" :data="nowTableData" virtual-scroll style="width: 100%" max-height="calc(100vh - 320px)">
            <el-table-column prop="row" width="80" :label="$t('序号')">
              <template #default="scope">
                <span>{{ scope.row.row }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="product_name" minWidth="300" :label="`*${$t('商品名称')}`">
              <template #default="scope">
                <template v-for="(item, index) in languageList" :key="index">
                  <el-form-item
                    v-if="hasObject(scope.row.product_name, item.key)"
                    :prop="`${scope.$index}.product_name.${item.key}`"
                    :rules="[
                      {
                        validator: (rule, value, callback) => {
                          if (!scope.row.product_name[item.key]) {
                            callback(new Error($t('请输入商品名称')));
                          } else if (scope.row.product_name_is_exist[item.key]) {
                            callback(new Error($t('此名称已存在')));
                          } else {
                            callback();
                          }
                        },
                      },
                    ]"
                    :label="item.value"
                  >
                    <mInput
                      v-model:valueData="scope.row.product_name[item.key]"
                      :value="scope.row.product_name[item.key]"
                      :maxlength="50"
                      :placeholder="$t('请输入商品名称')"
                      :langKey="item.name"
                      width="w-full"
                      @translate="(e) => translate(e, scope.row.row - 1)"
                      @change="handleClearNameExist(scope.row.row - 1, item.key)"
                    ></mInput>
                  </el-form-item>
                </template>
              </template>
            </el-table-column>
            <el-table-column prop="category_name" minWidth="200" :label="`*${$t('所属分类')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.category_id`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.category_id,
                      message: $t('请选择分类'),
                    },
                  ]"
                >
                  <el-cascader
                    :options="categoryList"
                    v-model="scope.row.category_id"
                    clearable
                    :props="{ value: 'category_id', label: 'name_text', children: 'child' }"
                    style="width: 100%"
                    :placeholder="$t('请选择分类')"
                  ></el-cascader>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="deduct_stock_type" width="120" :label="`*${$t('库存计算方式')}`">
              <template #default="scope">
                <el-radio-group v-model="scope.row.deduct_stock_type">
                  <el-radio :value="10">{{ $t('下单减库存') }}</el-radio>
                  <el-radio :value="20">{{ $t('付款减库存') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
            <el-table-column prop="product_unit" width="120" :label="`*${$t('商品单位')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.unit_id`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.unit_id,
                      message: $t('请选择商品单位'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.unit_id" clearable :placeholder="$t('请选择商品单位')">
                    <el-option v-for="item in unitList" :key="item.unit_id" :label="item.unit_name_text" :value="item.unit_id"></el-option>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="spec_name" width="120" :label="`*${$t('规格名称')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.spec_id`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.spec_id,
                      message: $t('请选择规格名称'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.spec_id" clearable :placeholder="$t('请选择规格名称')">
                    <el-option v-for="item in specList" :key="item.spec_id" :label="item.spec_name_text" :value="item.spec_id"></el-option>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="img_name" width="120" :label="$t('图片名称')">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.img_name`"
                  :rules="[
                    {
                      trigger: ['blur', 'change'],
                      validator: (rule, value, callback) => {
                        if (scope.row.img_name_is_exist) {
                          callback(new Error($t('图片名称已存在')));
                        } else {
                          callback();
                        }
                      },
                    },
                  ]"
                >
                  <el-input v-model="scope.row.img_name" clearable :placeholder="$t('图片名称')" @input="handleClearImgNameExist(scope.row.row - 1)" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="product_stock" width="160" :label="`*${$t('库存数量')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.product_stock`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => isNumber(scope.row.product_stock) && scope.row.product_stock >= 0,
                      message: $t('请输入库存数量'),
                    },
                  ]"
                >
                  <el-input-number
                    v-model.number="scope.row.product_stock"
                    :controls="false"
                    :min="0"
                    :max="1000000"
                    :precision="0"
                    :placeholder="$t('请输入库存数量')"
                  ></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="barcode" width="120" :label="$t('商品条码')">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.barcode`"
                  :rules="[
                    {
                      trigger: ['blur', 'change'],
                      validator: (rule, value, callback) => {
                        if (scope.row.barcode_is_exist) {
                          callback(new Error($t('商品条码已存在')));
                        } else {
                          callback();
                        }
                      },
                    },
                  ]"
                >
                  <el-input v-model="scope.row.barcode" clearable :placeholder="$t('商品条码')" @input="handleClearCodeNameExist(scope.row.row - 1)" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="product_price" width="130" :label="`*${$t('商品价格')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.product_price`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => isNumber(scope.row.product_price) && scope.row.product_price >= 0,
                      message: $t('请输入价格'),
                    },
                  ]"
                >
                  <el-input-number
                    v-model.number="scope.row.product_price"
                    :controls="false"
                    :min="0"
                    :max="100000000"
                    :precision="2"
                    :placeholder="$t('请输入价格')"
                  ></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="product_status" width="120" :label="`*${$t('商品状态')}`">
              <template #default="scope">
                <el-radio-group v-model="scope.row.product_status">
                  <el-radio :value="10">{{ $t('开启') }}</el-radio>
                  <el-radio :value="20">{{ $t('关闭') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
            <el-table-column prop="product_ratin_tax_type" width="180" :label="`*${$t('堂食税类')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.ratin_tax_id`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.ratin_tax_id,
                      message: $t('请选择堂食税类'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.ratin_tax_id" clearable :placeholder="$t('请选择堂食税类')">
                    <el-option v-for="item in taxList" :key="item.id" :label="item.name_text" :value="item.id"></el-option>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="product_takeout_tax_type" width="180" :label="`*${$t('外带税类')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.takeout_tax_id`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.takeout_tax_id,
                      message: $t('请选择外带税类'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.takeout_tax_id" clearable :placeholder="$t('请选择外带税类')">
                    <el-option v-for="item in taxList" :key="item.id" :label="item.name_text" :value="item.id"></el-option>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="num_type" width="120" :label="`*${$t('计价方式')}`">
              <template #default="scope">
                <el-radio-group v-model="scope.row.num_type">
                  <el-radio value="1">{{ $t('整数') }}</el-radio>
                  <el-radio value="2">{{ $t('小数') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
            <el-table-column prop="shows" width="120" :label="`*${$t('显示')}`">
              <template #default="scope">
                <el-form-item
                  :prop="`${scope.$index}.is_show_cashier`"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: (rule, value, callback) => {
                        if (
                          scope.row.is_show_cashier == 1 ||
                          scope.row.is_show_tablet == 1 ||
                          scope.row.is_show_kitchen == 1 ||
                          scope.row.is_show_assistant == 1 ||
                          scope.row.is_show_h5 == 1 ||
                          scope.row.is_show_delivery == 1
                        ) {
                          return true;
                        } else {
                          callback(new Error($t('请选择显示')));
                        }
                      },
                    },
                  ]"
                >
                  <el-checkbox v-model="scope.row.is_show_cashier" :true-value="1" :false-value="2" :label="$t('收银机')" />
                  <el-checkbox v-model="scope.row.is_show_tablet" :true-value="1" :false-value="2" :label="$t('平板')" :disabled="scope.row.num_type === 1" />
                  <el-checkbox v-model="scope.row.is_show_kitchen" :true-value="1" :false-value="2" :label="$t('厨显')" />
                  <el-checkbox v-model="scope.row.is_show_assistant" :true-value="1" :false-value="2" :label="$t('点餐助手')" :disabled="scope.row.num_type === 1" />
                  <el-checkbox v-model="scope.row.is_show_h5" :true-value="1" :false-value="2" :label="$t('扫码点餐')" :disabled="scope.row.num_type === 1" />
                  <el-checkbox v-if="showDelivery" v-model="scope.row.is_show_delivery" :true-value="1" :false-value="2" :label="$t('外送')" :disabled="scope.row.num_type === 1" />
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="product_sort" width="120" :label="$t('商品排序')">
              <template #default="scope">
                <el-form-item>
                  <el-input-number v-model.number="scope.row.product_sort" :controls="false" :min="0" :max="1000000" :precision="0" :placeholder="$t('商品排序')"></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="limit_num" width="120" :label="$t('限购数量')">
              <template #default="scope">
                <el-form-item prop="" scope.row.limit_num>
                  <el-input-number v-model.number="scope.row.limit_num" :controls="false" :min="0" :max="1000000" :precision="0" :placeholder="$t('0为不限购')"></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column width="120" :label="`*${$t('会员折扣')}`">
              <template #default="scope">
                <el-radio-group v-model="scope.row.is_enable_grade">
                  <el-radio value="1">{{ $t('开启') }}</el-radio>
                  <el-radio value="0">{{ $t('关闭') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
            <el-table-column width="120" :label="`*${$t('整单折扣')}`">
              <template #default="scope">
                <el-radio-group v-model="scope.row.open_overall_discount">
                  <el-radio value="1">{{ $t('开启') }}</el-radio>
                  <el-radio value="0">{{ $t('关闭') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
            <el-table-column prop="" width="80" fixed="right" align="center" :label="$t('操作')">
              <template #default="scope">
                <el-button type="primary" link @click="handleDelete(scope.row)">{{ $t('删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-form>
        <div class="pagination">
          <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            background
            :current-page="curPage"
            :page-size="pageSize"
            layout="total, prev, pager, next, jumper"
            :total="totalDataNumber"
          ></el-pagination>
        </div>
      </div>
    </div>

    <div class="common-button-wrapper" v-if="step == 2">
      <el-button class="dialog-add" type="primary" size="small" icon="Plus" @click="handleAdd">{{ $t('新增一行') }}</el-button>
      <div>
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="handleClick" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </div>
  </div>
</template>

<script>
  import * as XLSX from 'xlsx';
  import { languageStore } from '@/store/model/language';
  import ProductImportsApi from '@/api/productImports';
  import { cloneDeep, isNumber, has, isArray } from 'lodash-es';
  import mInput from '@/components/m-input/index.vue';
  //
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  //
  const languageList = languageStore().getLanguageList().languageList.value;
  const language = languageStore().getLanguageKey().language.value;

  import { useUserStore } from '@/store/index';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const showDelivery = (supplier.value?.delivery_status || 0) == 1;
  //
  export default {
    components: { SvgIcon, mInput },
    props: ['open'],
    created() {
      this.dialogVisible = this.open;
    },
    data() {
      return {
        dialogVisible: false,
        loading: false,
        form: {
          search: '',
        },
        step: 1,
        tableData: [], // excel数据数组列表
        categoryList: [], // 下拉分类列表
        specList: [], // 下拉规格列表
        taxList: [], // 下拉税类列表
        unitList: [], // 下拉单位列表
        languageList: languageList,
        curPage: 1,
        pageSize: 10,
        totalDataNumber: 0,
        nowTableData: [],
        showDelivery: showDelivery,
      };
    },
    watch: {
      tableData: {
        handler(newVal) {
          this.totalDataNumber = newVal.length || 0;
          newVal.forEach((row) => {
            if (row.num_type === '2') {
              row.is_show_tablet = 2;
              row.is_show_assistant = 2;
              row.is_show_h5 = 2;
              row.is_show_delivery = 2;
            }
          });
        },
        deep: true,
        immediate: true,
      },
    },
    computed: {},

    methods: {
      isNumber,
      dialogFormVisible() {
        this.$router.go(-1);
      },

      tablePageData() {
        const start = (this.curPage - 1) * this.pageSize;
        const end = this.curPage * this.pageSize;
        this.nowTableData = this.tableData.slice(start, end);
        this.$refs.tableFormRef?.validate();
      },

      hasObject(obj, key) {
        return has(obj, key);
      },

      handleClearNameExist(index, key) {
        if (this.tableData && this.tableData[index] && this.tableData[index].product_name_is_exist) this.tableData[index].product_name_is_exist[key] = false;
      },

      handleClearImgNameExist(index) {
        if (this.tableData && this.tableData[index] && this.tableData[index].img_name_is_exist) this.tableData[index].img_name_is_exist = false;
      },

      handleClearCodeNameExist(index) {
        if (this.tableData && this.tableData[index] && this.tableData[index].barcode_is_exist) this.tableData[index].barcode_is_exist = false;
      },

      handleClick() {
        // 去掉名称校验
        this.tableData.map((item) => {
          for (const key in item.product_name_is_exist) {
            if (Object.prototype.hasOwnProperty.call(item.product_name_is_exist, key)) {
              item.product_name_is_exist[key] = false;
            }
          }
          return item;
        });
        // 开始校验
        let self = this;
        this.$refs.tableFormRef?.validate(async (valid) => {
          // 校验名称是否重复
          if (!valid) {
            let have_name = false;
            self.tableData?.map((item) => {
              self.languageList?.map((lang) => {
                if (item.product_name_is_exist[lang.key]) {
                  have_name = true;
                }
              });
            });
            if (have_name) {
              ElMessageBox.confirm($t('商品上传失败，请稍后重试。失败原因：商品名称重名'), $t('导入失败'), {
                confirmButtonText: $t('确定'),
                cancelButtonText: $t('取消'),
                type: 'warning',
              })
                .then(() => {})
                .catch(() => {});
            }
            this.moveToError();
            return;
          }
          const list =
            cloneDeep(self.tableData)?.map((item) => {
              item.product_name = JSON.stringify(item.product_name);
              // 分类赋值
              if (typeof item.category_id == 'object' && item.category_id) {
                item.category_id = item.category_id[item.category_id.length - 1];
              }

              return item;
            }) || [];

          if (!list || list.length == 0) return;

          // 校验名字重复多规格
          let product_names = list.map((item) => item.product_name);
          let uniqueArr = [...new Set(product_names)];
          let hasDuplicates = product_names.length !== uniqueArr.length;

          if (hasDuplicates) {
            ElMessageBox.confirm($t('您批量录入了多规格商品，单位、分类、商品状态等通用属性都以第一个规格为准，是否继续？'), $t('多规格提示'), {
              confirmButtonText: $t('确定'),
              cancelButtonText: $t('取消'),
              type: 'warning',
            })
              .then(() => {
                self.handleUp(list);
              })
              .catch(() => {});
          } else {
            self.handleUp(list);
          }
        });
      },

      async handleUp(list) {
        let self = this;
        try {
          self.loading = true;
          const res = await ProductImportsApi.fetchProductImportsList({ list }, true);
          // 提示
          self.$ElMessage({ message: $t('导入成功'), type: 'success' });
          // 关闭
          self.step = 1;
          self.dialogFormVisible();
          // 重新刷新列表
        } catch (error) {
          // 提示错误
          const arr = isArray(error.data) ? error.data : [error.data];
          arr?.map((item) => {
            if (item.row > 0 && self.tableData && self.tableData[item.row - 1]) self.tableData[item.row - 1].product_name_is_exist = item.product_name_is_exist;
            if (item.row > 0 && self.tableData && self.tableData[item.row - 1]) self.tableData[item.row - 1].img_name_is_exist = item.img_name_is_exist;
            if (item.row > 0 && self.tableData && self.tableData[item.row - 1]) self.tableData[item.row - 1].barcode_is_exist = item.barcode_is_exist;
          });
          self.$refs.tableFormRef?.validate();
          this.moveToError();
        } finally {
          self.loading = false;
        }
      },

      handleDownload() {
        const downloadLink = document.createElement('a');
        downloadLink.href = `${import.meta.env.VITE_BASIC_URL || ''}/imports/product/${language}.xlsx`; // 文件地址
        downloadLink.download = `${$t('商品导入模板')}.xlsx`; // 保存的文件名
        document.body.appendChild(downloadLink);
        downloadLink.click();
        document.body.removeChild(downloadLink);
      },

      onBeforeUpload(file) {
        const reader = new FileReader();
        reader.onload = (e) => {
          const data = new Uint8Array(e.target.result);
          const workbook = XLSX.read(data, { type: 'array' });
          const firstSheetName = workbook.SheetNames[0];
          const worksheet = workbook.Sheets[firstSheetName];
          const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 });
          // 获取xlsx数据
          const uploadArr = [];
          jsonData.map((item, index) => {
            if (index > 0 && item[0]) {
              uploadArr.push({
                product_name: item[0] || '', // 商品名称
                category_name: item[1] || '', // 所属分类
                deduct_stock_type: item[2] || '0', // 库存计算方式
                product_unit: item[3] || '', // 商品单位
                spec_name: item[4] || '', // 规格名称
                img_name: item[5] || '', // 图片名称
                product_stock: item[6] || '', // 库存数量
                barcode: item[7] || '', // 商品条码
                product_price: item[8] || '', // 商品价格
                product_status: item[9] , // 商品状态
                product_ratin_tax_type: item[10] || '', // 堂食税类
                product_takeout_tax_type: item[11] || '', // 外带税类
                num_type: item[12] || '', // 计价方式
                shows: item[13] || '', // 显示
                product_sort: item[14] || '0', // 商品排序
                limit_num: item[15] || '', // 限购数量
                is_enable_grade: item[16] , // 会员折扣
                open_overall_discount: item[17] , // 整单折扣
                row: index,
              });
            }
          });
          // 上传文件，获取预览数据
          this.handleUploadXlsx(uploadArr);
        };
        reader.readAsArrayBuffer(file);

        return false; // 阻止默认的上传行为
      },

      // 上传xlsx文件，获取数据预览
      async handleUploadXlsx(list = []) {
        if (!list || list.length == 0) return;
        try {
          this.loading = true;
          const res = await ProductImportsApi.getProductImportsList({ list }, true);
          //
          const arr = res?.data?.list?.map((item) => {
            if (!item.category_id) item.category_id = '';
            if (!item.unit_id) item.unit_id = '';
            if (!item.spec_id) item.spec_id = '';
            if (!item.ratin_tax_id) item.ratin_tax_id = '';
            if (!item.takeout_tax_id) item.takeout_tax_id = '';
            if (item.product_stock) item.product_stock = Number(item.product_stock || 0);
            if (item.product_price) item.product_price = Number(item.product_price || 0);
            if (item.limit_num == '') item.limit_num = null;
            //
            try {
              item.product_name = JSON.parse(item.product_name);
            } catch (error) {
              item.product_name = {};
            }
            //
            this.languageList?.map(({ key }) => {
              if (!item.product_name[key]) item.product_name[key] = '';
            });
            return item;
          });
          this.tableData = arr || [];
          this.tablePageData();

          this.categoryList = res?.data?.category_list || [];
          this.specList = res?.data?.spec_list || [];
          this.taxList = res?.data?.tax_list || [];
          this.unitList = res?.data?.unit_list || [];
          //
          this.step = 2;
          setTimeout(() => {
            this.$refs.tableFormRef?.validate();
          }, 200);
        } catch (error) {
          //
        } finally {
          this.loading = false;
        }
      },

      handleAdd() {
        const row = this.tableData.length + 1;
        // 设置默认语言
        const product_name = {};
        const product_name_is_exist = {};
        this.languageList?.map(({ key }) => {
          product_name[key] = '';
          product_name_is_exist[key] = false;
        });
        this.tableData.push({
          product_name, // 商品名称
          product_name_is_exist,
          category_id: [this.categoryList[0]?.category_id] || '', // 所属分类
          deduct_stock_type: 10, // 库存计算方式
          num_type: 0, // 数量计算方法, 0-整数 1-小数
          unit_id: this.unitList[0]?.unit_id || '', // 商品单位
          spec_id: this.specList[0]?.spec_id || '', // 规格名称
          img_name: '', // 图片名称
          product_stock: 9999, // 库存数量
          barcode: '', // 商品条码
          product_price: 999, // 商品价格
          product_status: 10, // 商品状态
          ratin_tax_id: this.taxList[0]?.id || '', // 堂食税类
          takeout_tax_id: this.taxList[0]?.id || '', // 外带税类
          // shows: '', // 显示
          is_show_cashier: 1, // 是否显示在收银端 1-显示 2-不显示
          is_show_tablet: 1, // 是否显示在平板端 1-显示 2-不显示
          is_show_kitchen: 1, // 是否显示在送厨端 1-显示 2-不显示
          is_show_assistant: 1, // 是否显示在点餐助手 1-显示 2-不显示
          is_show_h5: 1, // 是否显示在h5 1-显示 2-不显示
          is_show_delivery: 1, // 是否显示在外送 1-显示 2-不显示
          product_sort: '0', // 商品排序
          limit_num: 0, // 限购数量
          is_enable_grade: '1', // 会员折扣
          row: this.tableData.length + 1,
        });
        this.tablePageData();
      },

      handleDelete(row) {
        this.tableData = this.tableData.filter((item) => item !== row);
        this.tableData.map((item, index) => {
          item.row = index + 1;
        });
        this.tablePageData();
      },

      /*选择第几页*/
      handleCurrentChange(val) {
        this.curPage = val;
        this.tablePageData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.pageSize = val;
      },

      moveToError() {
        // 设置一个定时器，在200毫秒后，获取所有带有el-form-item__error类的元素，并将第一个元素滚动到视图中
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
        }, 200);
      },

      /*翻译*/
      translate(lang, index) {
        this.languageList.map((item) => {
          lang.map((items) => {
            let key = item.name;
            if (key == 'zhtw') {
              key = 'zh-TW';
            }
            if (items[key]) {
              this.tableData[index].product_name[item.key] = items[key];
            }
            if (this.tableData && this.tableData[index] && this.tableData[index].product_name_is_exist) this.tableData[index].product_name_is_exist[item.key] = false;
          });
        });
      },
    },
  };
</script>
<style scoped lang="scss">
  .dialog-import-Qrcode {
    display: flex;
    flex-direction: column;
  }
  .dialog-import-Qrcode .el-dialog__body {
    flex: 1 1 auto;
    overflow: auto;
  }
</style>
<style lang="scss" scoped>
  .ti-table-import {
    :deep(.cell) {
      display: block;
    }
  }
  .dialog-footer {
    flex: 0 0 auto;
    flex-shrink: 0;
    .dialog-add {
      float: left;
    }
  }
  .mt-24 {
    margin-top: 24px;
  }
  .data-box-icon {
    width: 108px;
    height: 108px;
  }
  .import-product {
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
  .dialog-content {
    flex: 1 1 auto;
    overflow: auto;
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
    justify-content: space-between;
  }
</style>
