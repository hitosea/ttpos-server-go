<template>
  <div class="import-box">
    <div class="common-return">
      <el-icon class="return-icon" @click="dialogFormVisible"> <ArrowLeftBold /> </el-icon>{{ $t('导入桌位') }}
    </div>

    <div v-loading="loading" class="import-content">
      <div v-if="step == 1">
        <div class="common-form">
          {{ $t('第一步，下载 Excel 模板，安装模板填写桌位信息') }}
        </div>
        <el-button type="primary" @click="handleDownload" :loading="loading">{{ $t('下载') }}</el-button>
        <div class="common-form mt-24">
          {{ $t('第二步，上传文件并确认桌位后提交') }}
        </div>

        <el-upload class="upload-demo" drag accept=".xls,.xlsx" :before-upload="onBeforeUpload">
          <SvgIcon class="data-box-icon" name="excel"></SvgIcon>
          <div class="el-upload__text">
            {{ $t('将文件拖拽到此处，或') }} <em>{{ $t('点击上传') }}</em>
          </div>
          <div class="el-upload__text">
            {{ $t('每次导入桌位数量应小于100条') }}
          </div>
        </el-upload>
      </div>
      <div v-if="step == 2">
        <el-form ref="tableFormRef" :model="tableData">
          <el-table :data="tableData" style="width: 100%" max-height="calc(100vh - 270px)">
            <el-table-column prop="sort" :label="$t('序号')" width="80">
              <template #default="scope">
                <div class="mb-16">{{ scope.$index + 1 }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="table_no" :label="$t('桌位名称')" width="280">
              <template #default="scope">
                <el-form-item
                  prop="scope.row.product_price"
                  :rules="[
                    {
                      trigger: ['blur', 'change'],
                      validator: (rule, value, callback) => {
                        if (!scope.row.table_no) {
                          callback(new Error($t('请输入桌位名称')));
                        } else if (scope.row.table_no_is_exist) {
                          callback(new Error($t('桌位名称已存在')));
                        } else {
                          callback();
                        }
                      },
                    },
                  ]"
                >
                  <el-input @input="handleClearNameExist(scope.$index)" v-model="scope.row.table_no" :placeholder="$t('请输入桌位名称')" :maxlength="50"></el-input>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="type_id" :label="$t('所属类型')">
              <template #default="scope">
                <el-form-item
                  prop="scope.row.type_id"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.type_id,
                      message: $t('请选择所属类型'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.type_id" :placeholder="$t('请选择所属类型')">
                    <template v-for="item in typeList">
                      <el-option :label="item.type_name" :value="item.type_id"></el-option>
                    </template>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="area_id" :label="$t('所属区域')">
              <template #default="scope">
                <el-form-item
                  prop="scope.row.area_id"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => !!scope.row.area_id,
                      message: $t('请选择所属区域'),
                    },
                  ]"
                >
                  <el-select v-model="scope.row.area_id" :placeholder="$t('请选择所属区域')">
                    <template v-for="item in areaList">
                      <el-option :label="item.area_name" :value="item.area_id"></el-option>
                    </template>
                  </el-select>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column prop="sort" :label="$t('排序')">
              <template #default="scope">
                <el-form-item
                  prop="scope.row.sort"
                  :rules="[
                    {
                      required: true,
                      trigger: ['blur', 'change'],
                      validator: () => (scope.row.sort != null ? true : false),
                      message: $t('请输入排序'),
                    },
                  ]"
                >
                  <el-input-number :controls="false" :min="0" :max="999" :placeholder="$t('接近0，排序等级越高')" v-model.number="scope.row.sort"></el-input-number>
                </el-form-item>
              </template>
            </el-table-column>
            <el-table-column :label="$t('操作')" width="80" fixed="right" align="center">
              <template #default="scope">
                <el-button type="primary" class="mb-16" link @click="handleDelete(scope.row)">{{ $t('删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-form>
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
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import SupplierApi from '@/api/supplier.js';
  import { cloneDeep } from 'lodash-es';
  import { languageStore } from '@/store/model/language';

  const language = languageStore().getLanguageKey().language.value;
  export default {
    components: { SvgIcon },

    data() {
      return {
        loading: false,
        form: {
          search: '',
        },
        step: 1,
        tableData: [],
        areaList: [],
      };
    },
    methods: {
      dialogFormVisible() {
        this.$router.go(-1);
      },

      handleClick() {
        this.dialogFormVisible();
      },

      handleDownload() {
        const downloadLink = document.createElement('a');
        downloadLink.href = `${import.meta.env.VITE_BASIC_URL || ''}/imports/table/${language}.xlsx`; // 文件地址
        downloadLink.download = `${$t('桌台导入模版')}.xlsx`; // 保存的文件名
        downloadLink.click();
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
                table_no: item[0] || '', // 桌位名称
                type_name: item[1] || '', // 所属类型
                area_name: item[2] || '', // 所属区域
                sort: item[3] || 0, // 排序
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
        let params = {
          list,
          mode: 'get',
        };
        try {
          this.loading = true;
          const res = await SupplierApi.postImportsTableList(params, true);
          const arr = res?.data?.list?.map((item) => {
            if (!item.area_id) item.area_id = '';
            if (!item.type_id) item.type_id = '';
            //
          });
          //
          this.tableData = res?.data?.list || [];
          this.areaList = res?.data?.area_list || [];
          this.typeList = res?.data?.type_list || [];
          //
          this.step = 2;
        } catch (error) {
          //
        } finally {
          this.loading = false;
        }
      },

      handleClearNameExist(index) {
        if (this.tableData && this.tableData[index] && this.tableData[index].table_no_is_exist) this.tableData[index].table_no_is_exist = false;
      },

      handleClick() {
        let self = this;
        this.$refs.tableFormRef?.validate(async (valid) => {
          if (!valid) return;
          const list = cloneDeep(self.tableData) || [];
          if (!list || list.length == 0) return;
          let params = {
            list,
            mode: 'save',
          };
          try {
            self.loading = true;
            const res = await SupplierApi.postImportsTableList(params, true);
            // 提示
            self.$ElMessage({ message: $t('导入成功'), type: 'success' });
            // 关闭
            self.step = 1;
            self.dialogFormVisible();
            // 重新刷新列表
            self.$emit('success', res);
          } catch (error) {
            //
          } finally {
            self.loading = false;
          }
        });
      },

      handleAdd() {
        this.tableData.push({
          table_no: '', // 桌位名称
          type_id: '', // 所属类型
          area_id: '', // 所属区域
          sort: 0, // 排序
          row: this.tableData.length + 1,
        });
      },

      handleDelete(row) {
        this.tableData = this.tableData.filter((item) => item !== row);
        this.tableData.map((item, index) => {
          item.row = index + 1;
        });
      },
    },
  };
</script>
<style>
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
  .mt-24 {
    margin-top: 24px;
  }
  .mb-16 {
    margin-bottom: 16px;
  }
  .data-box-icon {
    width: 108px;
    height: 108px;
  }
  :deep(.el-table__row) {
    .cell {
      padding-top: 16px;
    }
  }
  .import-box {
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
  .import-content {
    flex: 1 1 auto;
    overflow: hidden;
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
