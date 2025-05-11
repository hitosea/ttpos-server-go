<template>
  <!--
            
            时间：2020-06-01
            描述：插件中心-分销-提现申请-弹窗
        -->
  <div>
    <el-dialog :title="$t('选择桌台')" width="820" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <div v-loading="loading">
        <el-form size="small" :inline="true" :model="form" class="demo-form-inline d-s-c">
          <el-form-item>
            <a-select size="small" v-model:value="form.area_id" :placeholder="$t('选择区域')" @change="onSearch">
              <el-option :label="$t('全部')" value="0"></el-option>
              <el-option v-for="(item, index) in area_list" :key="index" :label="item.area_name" :value="item.area_id"> </el-option>
            </a-select>
          </el-form-item>
          <el-form-item>
            <a-select size="small" v-model:value="form.type_id" :placeholder="$t('选择类型')" @change="onSearch">
              <el-option :label="$t('全部')" value="0"></el-option>
              <el-option v-for="(item, index) in type_list" :key="index" :label="item.type_name" :value="item.type_id"> </el-option>
            </a-select>
          </el-form-item>
          <el-form-item>
            <el-input v-model="form.search" autocomplete="off" :placeholder="$t('桌位名称')" @input="onSearch"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button size="small" type="primary" icon="Search" class="search-button" @click="onSearch">
              {{ $t('查询') }}
            </el-button>
          </el-form-item>
        </el-form>
        <div class="table-wrap">
          <el-table
            ref="multipleTableRef"
            id="multipleTableRef"
            size="small"
            :data="tableData"
            style="width: 100%"
            :max-height="400"
            v-loading="loading"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" :selectable="selectable" />
            <el-table-column prop="table_no" :label="$t('桌位名称')"></el-table-column>
          </el-table>
        </div>
      </div>
      <!-- <div class="pagination" v-if="Dtype == 'delete'">
          <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            background
            :current-page="curPage"
            :page-size="pageSize"
            layout="total, prev, pager, next, jumper"
            :total="totalDataNumber"
          ></el-pagination>
        </div> -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="handleClick" :loading="downloadLoading">{{ $t('确定') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import { useUserStore } from '@/store';
  import StoreApi from '@/api/store.js';
  import JSZip from 'jszip';

  const { token, supplier } = useUserStore();
  export default {
    data() {
      return {
        /*是否显示*/
        dialogVisible: false,
        loading: false,
        downloadLoading: false,
        token,
        /*列表数据*/
        tableData: [],
        form: {
          search: '',
          area_id: '',
          type_id: '',
        },
        type_list: [],
        area_list: [],
        /*一页多少条*/
        pageSize: 10,
        /*一共多少条数据*/
        totalDataNumber: 0,
        /*当前是第几页*/
        curPage: 1,
        last_page: 999,
        searchLoading: '',
        supplier,
        selectedTableIds: [],
      };
    },
    props: ['open', 'Dtype', 'include'],

    created() {
      this.dialogVisible = this.open;
      if (this.include && this.include.length) {
        this.selectedTableIds = [...this.include];
      }

      // this.handleScroll();
    },
    mounted() {
      this.getData('first');
    },
    methods: {
      /*选择第几页*/
      handleCurrentChange(val) {
        let self = this;
        self.curPage = val;
        self.getData();
      },

      /*每页多少条*/
      handleSizeChange(val) {
        this.curPage = 1;
        this.pageSize = val;
        this.getData();
      },

      /*搜索查询*/
      onSearch() {
        clearTimeout(this.searchLoading);
        this.tableData = [];
        this.searchLoading = setTimeout(() => {
          this.curPage = 1;
          this.getData();
        }, 200);
      },

      /*获取列表*/
      getData(e) {
        let self = this;

        self.loading = true;
        let params = self.form;
        params.page = self.curPage;
        params.list_rows = 1000;
        StoreApi.tablelist(params, true)
          .then((data) => {
            self.loading = false;
            self.type_list = data.data.type_list;
            self.area_list = data.data.area_list;
            if (e == 'first') {
              self.tableData = [];
            }
            if (this.Dtype == 'delete') {
              self.tableData = data.data.list.data;
              self.totalDataNumber = data.data.list.total;
            } else {
              data.data.list.data.map((item) => {
                self.tableData.push(item);
              });
              self.last_page = data.data.list.last_page;
              self.totalDataNumber = data.data.list.total;
            }
            //自动打钩
            this.$nextTick(() => {
              const okSelected = this.selectedTableIds;
              self.tableData.forEach((item) => {
                if (okSelected.includes(item.table_id)) {
                  self.$refs.multipleTableRef.toggleRowSelection(item, true);
                }
              });
            });
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      handleClick() {
        if (this.Dtype == 'delete') {
          this.deleteQrcode();
        } else if (this.Dtype == 'service') {
          // 使用selectedTableIds查找完整的表格行数据
          this.$emit('selectTable', this.selectedTableIds);
        } else {
          this.downloadQrcode();
        }
      },

      downloadQrcode() {
        let self = this;
        let ids = this.selectedTableIds.join(',');

        if (!ids.length) {
          self.$ElMessage({
            message: $t('请选择桌台'),
            type: 'error',
          });
          return;
        }
        self.downloadLoading = true;
        self
          .getQrcode(ids)
          .then((res) => {
            const zip = new JSZip();
            for (const imgData of res) {
              zip.file(imgData.name + '.png', imgData.url.split(',')[1], { base64: true });
            }
            zip.generateAsync({ type: 'blob' }).then((blob) => {
              const url = window.URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.style.display = 'none';
              a.href = url;
              a.download = `${supplier?.name || 'qrcode'}.zip`;
              document.body.appendChild(a);
              a.click();
              window.URL.revokeObjectURL(url);
              document.body.removeChild(a);

              self.$ElMessage({
                message: $t('开始下载'),
                type: 'success',
              });
            });
          })
          .catch((err) => {
            self.$ElMessage({
              message: $t('处理失败'),
              type: 'error',
            });
          })
          .finally(() => {
            self.downloadLoading = false;
          });
      },

      deleteQrcode() {
        let self = this;
        self.downloadLoading = true;

        let params = {
          ids: this.selectedTableIds.join(','),
        };
        StoreApi.batchTableDelete(params, true)
          .then((data) => {
            self.$ElMessage({
              message: $t('操作成功'),
              type: 'success',
            });
            self.dialogFormVisible(true);
          })
          .catch((err) => {
            self.$ElMessage({
              message: err.msg,
              type: 'error',
            });
          })
          .finally(() => {
            self.downloadLoading = false;
          });
      },

      getQrcode(ids) {
        let self = this;
        return StoreApi.batchQRcode(
          {
            ids,
            source: 'h5',
            action: 'get',
          },
          true
        ).then((data) => {
          return Promise.all(data.data.map((item) => self.modifyQRCode(item)));
        });
      },

      selectable(row) {
        if (row.status == 30 && this.Dtype == 'delete') {
          return false;
        } else {
          return true;
        }
      },

      modifyQRCode({ table_id, table_no, qrcode }) {
        return new Promise((resolve, reject) => {
          if (!qrcode || !table_id || !table_no) {
            reject(new Error('missing params'));
          }
          try {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            const img = new Image();
            img.src = qrcode;
            img.onload = () => {
              canvas.width = img.width + 160; // 增加宽度以容纳边框
              canvas.height = img.height + 160; // 增加高度以容纳边框和文字
              // 设置背景色为白色
              ctx.fillStyle = '#fff';
              ctx.fillRect(0, 0, canvas.width, canvas.height);

              // 绘制边框
              ctx.strokeStyle = 'black';
              ctx.lineWidth = 2;
              const radius = 20; // 圆角半径
              ctx.beginPath();
              ctx.moveTo(10 + radius, 10);
              ctx.arcTo(canvas.width - 10, 10, canvas.width - 10, canvas.height - 10, radius);
              ctx.arcTo(canvas.width - 10, canvas.height - 10, 10, canvas.height - 10, radius);
              ctx.arcTo(10, canvas.height - 10, 10, 10, radius);
              ctx.arcTo(10, 10, canvas.width - 10, 10, radius);
              ctx.closePath();
              ctx.stroke();

              // 绘制顶部文字
              ctx.fillStyle = 'black';
              ctx.font = '40px Arial';
              ctx.textAlign = 'center';
              ctx.fillText(self.$t('扫码点餐'), canvas.width / 2, 64);

              // 绘制二维码
              ctx.drawImage(img, 80, 80);

              // 绘制底部文字
              ctx.font = '54px Arial';

              // 计算最大允许宽度
              const maxWidth = canvas.width * 0.8;

              // 测量文本宽度
              const textWidth = ctx.measureText(table_no).width;

              let displayText = table_no;

              // 如果文本宽度超过最大允许宽度，则截断并添加省略号
              if (textWidth > maxWidth) {
                displayText = table_no.slice(0, 8) + '...';
              }

              // 绘制文本
              ctx.fillText(displayText, canvas.width / 2, canvas.height - 40);

              resolve({ id: table_id, name: table_no, url: canvas.toDataURL('image/png') });
            };
          } catch (error) {
            reject(error);
          }
        });
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        this.reset();
        this.$emit('close', e);
      },

      handleSelectionChange(val) {
        // 不直接覆盖，而是更新selectedTableIds
        const currentIds = val.map((item) => item.table_id);

        // 获取当前表格中显示的所有ID
        const visibleIds = this.tableData.map((item) => item.table_id);

        // 从selectedTableIds中移除当前页面中未被选中的ID
        this.selectedTableIds = this.selectedTableIds.filter((id) => !visibleIds.includes(id) || currentIds.includes(id));

        // 添加新选中的ID
        currentIds.forEach((id) => {
          if (!this.selectedTableIds.includes(id)) {
            this.selectedTableIds.push(id);
          }
        });
      },

      reset() {
        this.tableData = [];
        this.form.search = '';
        this.curPage = 1;
        this.pageSize = 10;
        this.totalDataNumber = 0;
        this.last_page = 999;
        this.selectedTableIds = [];
      },

      // handleScroll() {
      //   this.$nextTick(() => {
      //     const scrollableDiv = document.querySelector('#multipleTableRef .el-scrollbar__wrap');
      //     scrollableDiv.addEventListener('scroll', () => {
      //       const { scrollTop, scrollHeight, clientHeight } = scrollableDiv;
      //       // 判断是否滚动到底部
      //       if (scrollTop + clientHeight >= scrollHeight - 50 && !this.loading && this.curPage < this.last_page) {
      //         this.curPage++;
      //         this.getData();
      //       }
      //     });
      //   });
      // },
    },
  };
</script>
<style scoped>
  .click-button {
    cursor: pointer;
    text-align: center;
    font-weight: bold;
  }
</style>
