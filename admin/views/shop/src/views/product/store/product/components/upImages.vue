<template>
  <el-form-item :label="$t('上传图片')" for="no_click">
    <div class="upload-btn">
      <el-button type="primary" :loading="loading">{{ $t('上传图片') }}</el-button>
      <input class="file-upload-input" type="file" id="input-id-files" @change="handleFolderSelect" webkitdirectory directory multiple />
    </div>
  </el-form-item>
  <div class="product-list" v-if="folderName.length > 0">
    <template v-for="(item, index) in folderName" :key="index">
      <div class="product-item">
        {{ item }}
        <i class="el-icon el-tag__close" @click="deleteFolder(index)">
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
  <el-table size="small" :data="tableData" border style="width: 100%" v-loading="loading">
    <el-table-column :label="$t('序号')" width="80">
      <template #default="scope">
        {{ scope.$index + 1 }}
      </template>
    </el-table-column>
    <el-table-column prop="product_name_text" :label="$t('商品名称')" minWidth="300"> </el-table-column>
    <el-table-column prop="img_name" :label="$t('图片名称')" minWidth="300">
      <template #default="scope">
        {{ scope.row.img_name || '-' }}
      </template>
    </el-table-column>
    <el-table-column prop="path_name_text" :label="$t('所属分类')" width="180">
      <template #default="scope">
        {{ scope.row.path_name_text || '-' }}
      </template>
    </el-table-column>
    <el-table-column prop="type" width="200" align="center">
      <template #header="scope">
        <el-tooltip effect="dark" placement="top" :content="$t('图片') + $t('支持JPG、JPEG、PNG、WEBP格式，小於15MB，尺寸：160*120px')">
          <div>
            {{ $t('图片') }}<span class="tips">{{ $t('支持JPG、JPEG、PNG、WEBP格式，小於15MB，尺寸：160*120px') }}</span>
          </div>
        </el-tooltip>
      </template>
      <template #default="scope">
        <div class="draggable-lists">
          <draggable class="wrapper" v-model="imgList" v-if="imgList[scope.$index]?.url">
            <transition-group>
              <div class="item">
                <img v-img-url="imgList[scope.$index]?.url" />
                <a href="javascript:void(0);" class="delete-btn" @click.stop="deleteImg(scope.$index)"
                  ><el-icon> <Close /> </el-icon
                ></a>
              </div>
            </transition-group>
          </draggable>
          <el-upload
            class="avatar-uploader"
            accept="image/jpeg,image/png,image/jpg,image/webp"
            :show-file-list="false"
            :before-upload="(e) => beforeAvatarUpload(e, scope.$index)"
            v-if="!imgList[scope.$index]?.url"
          >
            <div class="item img-select" @click="openProductUpload">
              <el-icon>
                <Plus />
              </el-icon>
            </div>
          </el-upload>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog :title="$t('图片名称已存在')" v-model="dialogVisible" append-to-body @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
    <div>
      <p class="mb16">{{ $t('以下图片名称已存在，是否需要覆盖为最新的图片') }}</p>
      <el-table size="small" :data="repeat_list" border style="width: 100%" v-loading="loading">
        <el-table-column :label="$t('序号')" width="80">
          <template #default="scope">
            {{ scope.$index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="product_name_text" :label="$t('商品名称')" minWidth="300"> </el-table-column>
        <el-table-column prop="path_name_text" :label="$t('所属分类')" width="180"></el-table-column>
        <el-table-column prop="file_name" :label="$t('图片名称')" width="180"></el-table-column>
      </el-table>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="packaging" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
  import JSZip from 'jszip';
  import ProductApi from '@/api/product.js';
  export default {
    props: ['product_list'],
    data() {
      return {
        tableData: [],
        files: [],
        filesList: [],
        imgList: [],
        folderName: [],
        nameList: [],
        repeat_list: [],
        request_cache: '',
        dialogVisible: false,
        loading: false,
      };
    },
    watch: {
      product_list: {
        async handler(val) {
          // 新的加料id
          const newIds = [];
          (val || []).map((item) => {
            newIds.push(item.product_id);
          });

          // 需要新增的id
          const _pushIds = await newIds.filter((item) => {
            return this.tableData.findIndex((items) => items.product_id == item) == -1;
          });

          //需要删除的id
          const _needToDeleteIds = await this.tableData.filter((item) => {
            return newIds.findIndex((items) => items == item.product_id) == -1;
          });

          // 遍历需要删除的id数组
          await _needToDeleteIds.map((item) => {
            this.tableData.map((items, indexs) => {
              if (items.product_id == item.product_id) {
                this.tableData.splice(indexs, 1);
                this.imgList.splice(indexs, 1);
              }
            });
          });

          // 遍历val数组
          await (val || []).map((item) => {
            if (_pushIds.indexOf(item.product_id) != -1) {
              this.tableData.push(item);
              this.imgList.push({
                product_name_text: item.product_name_text,
                product_id: item.product_id,
                url: '',
                file: '',
              });
            }
          });
        },
        immediate: true,
        deep: true,
      },
    },

    methods: {
      //上传文件夹
      async handleFolderSelect(event) {
        if (event.target.files.length == 0) return;

        this.files = [];
        this.filesList = [];
        this.folderName = [];
        this.nameList = [];
        const selectedFiles = event.target.files;
        this.filesList.push([]);
        Array.from(selectedFiles).map((file) => {
          this.filesList[0].push(file);
        });

        // 获取文件夹名称
        this.folderName.push(this.filesList[0][0].webkitRelativePath.split('/')[0]);

        for (const item of this.filesList) {
          for (const file of item) {
            if ((file.type.includes('jpg') || file.type.includes('png') || file.type.includes('jpeg') || file.type.includes('webp')) && file.size < 15 * 1024 * 1024) {
              const isType = await this.checkImageType(file);
              if (isType) {
                this.files.push(file);
              } else {
                const changeFile = await this.convertWebp(file);
                this.files.push(changeFile);
              }
            }
          }
        }

        this.filterList();
      },

      //删除文件夹
      deleteFolder(index) {
        this.filesList.splice(index, 1);
        this.folderName.splice(index, 1);
        this.files = [];
        this.imgList.map((item) => {
          item.img_name = '';
          item.url = '';
        });

        document.getElementById('input-id-files').value = '';
        this.filterList(true);
      },

      //过滤列表
      filterList(e) {
        //过滤只剩图片文件
        const filteredFiles = this.files.filter((file) => {
          const fileType = file.type.toLowerCase();
          return (fileType.includes('jpg') || fileType.includes('png') || fileType.includes('jpeg') || fileType.includes('webp')) && file.size < 15 * 1024 * 1024;
        });
        //过滤重复名字的图片
        const uniqueArray = filteredFiles.filter((obj, index, self) => index === self.findIndex((t) => t.name === obj.name));
        this.files = uniqueArray;

        if (this.files.length > 0) {
          // 显示图片
          // 用于存储文件名的数组
          this.nameList = [];
          this.files.map((file) => {
            this.nameList.push(file.name.replace(/\.(png|jpg|jpeg|webp)/gi, ''));
          });

          //图片数组
          this.tableData.map((item, index) => {
            this.imgList[index].file = this.files[this.nameList.indexOf(item.img_name)];
            if (this.nameList.includes(item.img_name)) {
              const reader = new FileReader();
              reader.onload = (e) => {
                const base64String = e.target.result;
                this.imgList[index].url = base64String;
              };
              reader.readAsDataURL(this.files[this.nameList.indexOf(item.img_name)]);
            }
          });
        } else if (this.files.length == 0 && e) {
          this.$ElMessage({
            type: 'warning',
            message: this.$t('没有相匹配的图片'),
          });
        }
      },

      //提交数据
      async packaging() {
        const zip = new JSZip();
        for (const imgData of this.imgList) {
          if (imgData.file) {
            zip.file(imgData.file.name, imgData.file);
          }
        }

        const list = [];
        this.imgList.map((item) => {
          if (item.file) {
            list.push({
              product_id: item.product_id,
              file_name: item.file.name,
              img_name: item.file.name.replace(/\.(png|jpg|jpeg|webp)/gi, ''),
              product_name_text: item.product_name_text,
            });
          }
        });

        if (list.length === 0) return;
        const content = await zip.generateAsync({ type: 'blob' });
        const formData = new FormData();
        formData.append('iFile', content);
        formData.append('file_type', content.type);
        formData.append('size', content.size);
        formData.append('name', 'filename.zip');
        formData.append('list', JSON.stringify(list));
        formData.append('is_overlay', this.request_cache ? 1 : 0);

        this.loading = true;
        this.$emit('loading', true);
        ProductApi.batchReplaceProductImage(formData, true)
          .then((data) => {
            this.dialogVisible = false;
            this.loading = false;
            this.$emit('loading', false);
            this.$ElMessage({
              type: 'success',
              message: this.$t('上传成功'),
            });
            this.$emit('close');
          })
          .catch((error) => {
            this.loading = false;
            this.$emit('loading', false);
            if (error.data?.repeat_list?.length > 0) {
              this.repeat_list = error.data.repeat_list;
              this.request_cache = error.data.request_cache;
              if (this.repeat_list.length > 0) {
                (this.repeat_list || []).map((item) => {
                  this.tableData.map((item2) => {
                    if (item2.product_id === item.product_id) {
                      item.path_name_text = item2.path_name_text;
                    }
                  });
                });
              }
              this.dialogVisible = true;
            }
          });
      },

      //删除图片
      deleteImg(index) {
        this.imgList[index].url = '';
        this.imgList[index].file = '';
      },

      //上传图片
      async beforeAvatarUpload(file, index) {
        const fileType = file.type.toLowerCase();
        if ((fileType.includes('jpg') || fileType.includes('png') || fileType.includes('jpeg') || fileType.includes('webp')) && file.size < 15 * 1024 * 1024) {
          const isType = await this.checkImageType(file);
          if (!isType) {
            const changeFile = await this.convertWebp(file);
            this.imgList[index].file = changeFile;
          } else {
            this.imgList[index].file = file;
          }

          const reader = new FileReader();
          reader.onload = (e) => {
            const base64String = e.target.result;
            this.imgList[index].url = base64String;
          };
          reader.readAsDataURL(file);
        } else {
          this.$ElMessage({
            type: 'warning',
            message: this.$t('请上传大小在15M以内的jpg、png、webp图片'),
          });
        }

        return false;
      },

      async checkImageType(file) {
        return await new Promise((resolve, reject) => {
          const reader = new FileReader();
          reader.onloadend = (e) => {
            const arr = new Uint8Array(e.target?.result).subarray(0, 12);
            let header = '';
            for (let i = 0; i < arr.length; i++) {
              header += arr[i].toString(16).padStart(2, '0');
            }

            // WebP 的魔数是 "52494646"(RIFF) 后跟文件大小，然后是 "57454250"(WEBP)
            if (header.startsWith('52494646') && header.includes('57454250')) {
              resolve(false);
            } else if (header.startsWith('89504e47')) {
              resolve(true);
            } else if (header.startsWith('ffd8ff')) {
              resolve(true);
            } else {
              resolve(false);
            }
          };
          reader.readAsArrayBuffer(file);
        });
      },

      async convertWebp(file) {
        let self = this;
        return await new Promise((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = function (event) {
            const img = new Image();
            img.onload = function () {
              const canvas = document.createElement('canvas');
              canvas.width = img.width;
              canvas.height = img.height;
              const ctx = canvas.getContext('2d');
              ctx.drawImage(img, 0, 0, img.width, img.height);
              const pngUrl = canvas.toDataURL('image/jpg');
              resolve(self.base64ToFile(pngUrl, file.name.replace(/\.(png|jpg|jpeg|webp)/gi, '') + '.png'));
            };
            img.src = event.target.result;
          };
          reader.readAsDataURL(file);
        });
      },

      base64ToFile(base64, filename) {
        // 分割 MIME 类型和 Base64 数据部分
        const [metadata, base64Data] = base64.split(',');
        const mimeString = metadata.match(/:(.*?);/)[1]; // 提取 MIME 类型
        const byteString = atob(base64Data); // 解码 Base64
        const arrayBuffer = new Uint8Array(byteString.length);

        // 将解码后的数据填充到 Uint8Array
        for (let i = 0; i < byteString.length; i++) {
          arrayBuffer[i] = byteString.charCodeAt(i);
        }

        // 创建 Blob 对象
        const blob = new Blob([arrayBuffer], { type: mimeString });

        // 创建 File 对象
        const file = new File([blob], filename, { type: mimeString });

        return file;
      },

      dialogFormVisible() {
        this.dialogVisible = false;
        this.request_cache = '';
      },
    },
  };
</script>
<style scoped>
  .upload-btn {
    position: relative;
    overflow: hidden;
    display: inline-block;
    cursor: pointer;
  }
  .file-upload-input {
    cursor: pointer;
    font-size: 100px;
    position: absolute;
    left: 0;
    top: 0;
    right: 0;
    bottom: 0;
    opacity: 0;
  }
  .img-table {
    margin: auto;
  }

  .draggable-lists {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
  }

  .draggable-lists .wrapper > span {
    display: flex;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .draggable-lists .item {
    position: relative;
    width: 110px;
    height: 110px;
    margin-top: 0px;
    margin-right: 0px;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid #dddddd;
  }

  .draggable-lists .delete-btn {
    position: absolute;
    top: 0;
    right: 0;
    width: 16px;
    height: 16px;
    background: red;
    line-height: 16px;
    font-size: 16px;
    color: #ffffff;
    display: none;
  }

  .draggable-lists .item:hover .delete-btn {
    display: block;
  }

  .draggable-lists .item img {
    position: absolute;
    top: 50%;
    left: 50%;
    -webkit-transform: translate(-50%, -50%);
    transform: translate(-50%, -50%);
    max-height: 100%;
    max-width: 100%;
  }

  .draggable-lists .img-select {
    display: flex;
    justify-content: center;
    align-items: center;
    border: 1px dashed #dddddd;
    font-size: 30px;
    cursor: pointer;
  }

  .draggable-lists .img-select i {
    color: #409eff;
  }
</style>
<style scoped lang="scss">
  .product-list {
    display: flex;
    box-shadow: 0 0 0 1px var(--el-input-border-color, var(--el-border-color)) inset;
    border-radius: 4px;
    padding: 6px 11px;
    margin: 16px 0;
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
</style>
