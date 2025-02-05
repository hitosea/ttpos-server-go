<template>
  <el-upload
    class="upload-demo"
    accept="image/jpeg,image/png,image/jpg,image/webp,video/mp4"
    action=""
    multiple
    :show-file-list="false"
    :before-upload="onBeforeUploadImage"
    :on-change="fileChange"
    :http-request="UploadImage"
  >
    <el-button type="primary">{{ $t('上传') }}</el-button>
    <template #tip>
      <div class="el-upload__tip">
        <p v-for="item in tips">{{ item }}</p>
      </div>
    </template>
  </el-upload>
</template>
<script>
  import FileApi from '@/api/file.js';

  export default {
    props: {
      tips: {
        type: Array,
        default: [$t('图片：支持JPG、JPEG、PNG、WEBP格式，小于2MB，尺寸：1024*600px'), $t('视频：支持AVI、MPEG、MOV、MP4格式，小于10MB，尺寸：1024*600px')],
      },
      imgSize: {
        type: [String, Number],
        default: 15,
      },
      videoSize: {
        type: [String, Number],
        default: 30,
      },
      source: {
        type: [String, Number],
        default: '',
      },
    },
    methods: {
      /*选择上传图片*/
      fileChange(e) {},

      /*选择图片之前*/
      onBeforeUploadImage(file) {
        let str = file.type.toLowerCase();
        if (str.includes('png') || str.includes('jpg') || str.includes('jpeg') || str.includes('webp') || str.includes('mp4')) {
          if (file.size > this.imgSize * 1024 * 1024 && file.type.includes('image')) {
            this.$ElMessage({
              message: $t('图片大小超出限制！'),
              type: 'warning',
            });
            return false;
          }
          if (file.size > this.videoSize * 1024 * 1024 && file.type.includes('video')) {
            this.$ElMessage({
              message: $t('视频大小超出限制！'),
              type: 'warning',
            });
            return false;
          }
          return true;
        } else {
          this.$ElMessage({
            message: $t('请上传符合要求的图片/视频文件'),
            type: 'warning',
          });
          return false;
        }
      },

      /*上传图片*/
      async UploadImage(param) {
        if (param.file.type.includes('png') || param.file.type.includes('jpg') || param.file.type.includes('jpeg') || param.file.type.includes('webp')) {
          const isType = await this.checkImageType(param.file);
          if (!isType) {
            const changeFile = await this.convertWebp(param.file);
            param.file = changeFile;
          }
        }

        let self = this;
        const loading = ElLoading.service({
          lock: true,
          text: $t('上传中，请等待'),
          background: 'rgba(0, 0, 0, 0.7)',
        });
        let type = '';
        let size = 30;
        if (param.file.type.includes('video')) {
          type = 'video';
          size = this.videoSize;
        }
        if (param.file.type.includes('image')) {
          type = 'image';
          size = this.imgSize;
        }
        const formData = new FormData();
        formData.append('iFile', param.file);
        formData.append('file_type', type);
        formData.append('size', size);
        formData.append('source', self.source);
        FileApi.uploadFile(formData)
          .then((response) => {
            loading.close();
            this.$ElMessage({
              message: $t('本次上传成功'),
              type: 'success',
            });
            self.$emit('upLoad', response.data);
          })
          .catch((response) => {
            loading.close();
            this.$ElMessage({
              message: $t('本次上传失败'),
              type: 'warning',
            });
          });
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
              const pngUrl = canvas.toDataURL('image/png');
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
    },
  };
</script>
<style scoped>
  .el-upload__tip {
    line-height: 1.7;
    color: var(--el-color-tips);
  }
</style>
