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
      <!--分页-->
      <div class="pagination">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          background
          :current-page="page"
          :page-size="page_size"
          layout="total, prev, pager, next, jumper"
          :total="totalDataNumber"
        ></el-pagination>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
        <el-button type="primary" @click="packaging" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, watch, getCurrentInstance } from 'vue';
  import { ElMessage } from 'element-plus';
  import JSZip from 'jszip';
  import ProductApi from '@/api/product.js';

  // 获取组件实例
  const { proxy } = getCurrentInstance();

  // 定义props
  const props = defineProps({
    product_list: {
      type: Array,
      default: () => [],
    },
  });

  // 定义emits
  const emit = defineEmits(['loading', 'close']);

  // 响应式数据
  const tableData = ref([]);
  const files = ref([]);
  const filesList = ref([]);
  const imgList = ref([]);
  const folderName = ref([]);
  const nameList = ref([]);
  const repeat_list = ref([]);
  const request_cache = ref(false);
  const dialogVisible = ref(false);
  const loading = ref(false);
  const page = ref(1);
  const page_size = ref(10);
  const totalDataNumber = ref(0);

  // 监听product_list变化
  watch(
    () => props.product_list,
    async (val) => {
      // 新的加料id
      const newIds = await Promise.resolve().then(() => {
        const ids = [];
        (val || []).map((item) => {
          ids.push(item.product_id);
        });
        return ids;
      });

      // 需要新增的id
      const _pushIds = await Promise.resolve().then(() => {
        return newIds.filter((item) => tableData.value.findIndex((items) => items.product_id == item) == -1);
      });

      // 需要删除的id
      const _needToDeleteIds = await Promise.resolve().then(() => {
        return tableData.value.filter((item) => newIds.findIndex((items) => items == item.product_id) == -1);
      });

      // 第一步：删除不需要的id
      await Promise.resolve().then(() => {
        _needToDeleteIds.map((item) => {
          tableData.value.map((items, indexs) => {
            if (items.product_id == item.product_id) {
              tableData.value.splice(indexs, 1);
              imgList.value.splice(indexs, 1);
            }
          });
        });
      });

      // 第二步：插入需要新增的id（
      await Promise.resolve().then(() => {
        (val || []).map((item) => {
          if (_pushIds.indexOf(item.product_id) != -1) {
            tableData.value.push(item);
            imgList.value.push({
              product_name_text: item.product_name_text,
              product_id: item.product_id,
              url: '',
              file: '',
            });
          }
        });
      });
    },
    { immediate: true, deep: true }
  );

  // 方法定义
  //上传文件夹
  const handleFolderSelect = async (event) => {
    if (event.target.files.length == 0) return;

    files.value = [];
    filesList.value = [];
    folderName.value = [];
    nameList.value = [];
    const selectedFiles = event.target.files;
    filesList.value.push([]);
    Array.from(selectedFiles).map((file) => {
      filesList.value[0].push(file);
    });

    // 获取文件夹名称
    folderName.value.push(filesList.value[0][0].webkitRelativePath.split('/')[0]);

    for (const item of filesList.value) {
      for (const file of item) {
        if ((file.type.includes('jpg') || file.type.includes('png') || file.type.includes('jpeg') || file.type.includes('webp')) && file.size < 15 * 1024 * 1024) {
          const isType = await checkImageType(file);
          if (isType) {
            files.value.push(file);
          } else {
            const changeFile = await convertWebp(file);
            files.value.push(changeFile);
          }
        }
      }
    }

    filterList();
  };

  //删除文件夹
  const deleteFolder = (index) => {
    filesList.value.splice(index, 1);
    folderName.value.splice(index, 1);
    files.value = [];
    imgList.value.map((item) => {
      item.img_name = '';
      item.url = '';
    });

    document.getElementById('input-id-files').value = '';
    filterList(true);
  };

  //过滤列表
  const filterList = (e) => {
    //过滤只剩图片文件
    const filteredFiles = files.value.filter((file) => {
      const fileType = file.type.toLowerCase();
      return (fileType.includes('jpg') || fileType.includes('png') || fileType.includes('jpeg') || fileType.includes('webp')) && file.size < 15 * 1024 * 1024;
    });
    //过滤重复名字的图片
    const uniqueArray = filteredFiles.filter((obj, index, self) => index === self.findIndex((t) => t.name === obj.name));
    files.value = uniqueArray;

    if (files.value.length > 0) {
      // 显示图片
      // 用于存储文件名的数组
      nameList.value = [];
      files.value.map((file) => {
        nameList.value.push(file.name.replace(/\.(png|jpg|jpeg|webp)/gi, ''));
      });

      //图片数组
      tableData.value.map((item, index) => {
        imgList.value[index].file = files.value[nameList.value.indexOf(item.img_name)];
        if (nameList.value.includes(item.img_name)) {
          const reader = new FileReader();
          reader.onload = (e) => {
            const base64String = e.target.result;
            imgList.value[index].url = base64String;
          };
          reader.readAsDataURL(files.value[nameList.value.indexOf(item.img_name)]);
        }
      });
    } else if (files.value.length == 0 && e) {
      ElMessage({
        type: 'warning',
        message: $t('没有相匹配的图片'),
      });
    }
  };

  /*选择第几页*/
  const handleCurrentChange = (val) => {
    loading.value = true;
    page.value = val;
    repeatList();
  };

  /*每页多少条*/
  const handleSizeChange = (val) => {
    page_size.value = val;
    repeatList();
  };

  // 批量修改分类验证（方法二）
  const repeatList = async () => {
    // 第一步：构建列表（方法二）
    const list = await Promise.resolve().then(() => {
      const temp = [];
      imgList.value.map((item) => {
        if (item.file) {
          temp.push({
            product_id: item.product_id,
            file_name: item.file.name,
            img_name: item.file.name.replace(/\.(png|jpg|jpeg|webp)/gi, ''),
            product_name_text: item.product_name_text,
          });
        }
      });
      return temp;
    });

    if (list.length === 0) {
      ElMessage({
        type: 'warning',
        message: $t('没有相匹配的图片'),
      });
      return;
    }

    try {
      loading.value = true;
      // 第二步：请求接口（真实异步）
      const res = await ProductApi.repeatList(
        {
          list,
          page: page.value,
          page_size: page_size.value,
        },
        true
      );
      loading.value = false;

      if (res.data?.data?.length > 0) {
        // 第三步：写入响应数据（方法二）
        await Promise.resolve().then(() => {
          repeat_list.value = res.data?.data;
          request_cache.value = true;
          totalDataNumber.value = res.data?.total;
        });

        // 第四步：补充 path_name_text（方法二）
        await Promise.resolve().then(() => {
          if (repeat_list.value.length > 0) {
            (repeat_list.value || []).map((item) => {
              tableData.value.map((item2) => {
                if (item2.product_id === item.product_id) {
                  item.path_name_text = item2.path_name_text;
                }
              });
            });
          }
        });

        dialogVisible.value = true;
      } else {
        // 第五步：无重复则直接打包上传（真实异步）
        await packaging();
      }
    } catch (error) {
      loading.value = false;
    }
  };

  //提交数据
  const packaging = async () => {
    const zip = new JSZip();
    for (const imgData of imgList.value) {
      if (imgData.file) {
        zip.file(imgData.file.name, imgData.file);
      }
    }

    const list = [];
    imgList.value.map((item) => {
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
    formData.append('is_overlay', request_cache.value ? 1 : 0);

    loading.value = true;
    emit('loading', true);

    try {
      await ProductApi.batchReplaceProductImage(formData, true);
      dialogVisible.value = false;
      loading.value = false;
      emit('loading', false);
      ElMessage({
        type: 'success',
        message: $t('上传成功'),
      });
      emit('close');
    } catch (error) {
      loading.value = false;
      emit('loading', false);
    }
  };

  //删除图片
  const deleteImg = (index) => {
    imgList.value[index].url = '';
    imgList.value[index].file = '';
  };

  //上传图片
  const beforeAvatarUpload = async (file, index) => {
    const fileType = file.type.toLowerCase();
    if ((fileType.includes('jpg') || fileType.includes('png') || fileType.includes('jpeg') || fileType.includes('webp')) && file.size < 15 * 1024 * 1024) {
      const isType = await checkImageType(file);
      if (!isType) {
        const changeFile = await convertWebp(file);
        imgList.value[index].file = changeFile;
      } else {
        imgList.value[index].file = file;
      }

      const reader = new FileReader();
      reader.onload = (e) => {
        const base64String = e.target.result;
        imgList.value[index].url = base64String;
      };
      reader.readAsDataURL(file);
    } else {
      ElMessage({
        type: 'warning',
        message: $t('请上传大小在15M以内的jpg、png、webp图片'),
      });
    }

    return false;
  };

  const checkImageType = async (file) => {
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
  };

  const convertWebp = async (file) => {
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
          resolve(base64ToFile(pngUrl, file.name.replace(/\.(png|jpg|jpeg|webp)/gi, '') + '.png'));
        };
        img.src = event.target.result;
      };
      reader.readAsDataURL(file);
    });
  };

  const base64ToFile = (base64, filename) => {
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
  };

  const dialogFormVisible = () => {
    dialogVisible.value = false;
    request_cache.value = '';
  };

  defineExpose({
    repeatList,
  });
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
