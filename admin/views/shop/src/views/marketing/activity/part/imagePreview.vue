<template>
  <div>
    <div class="image-preview-wrapper">
      <div class="image-preview-item">
        <div class="image-preview-bg"></div>
        <div class="image-preview-content">
          <div class="image-preview-content-item">
            <div class="logo-wrapper">
              <img :src="userInfo.logoUrl" />
            </div>
            <h3 class="item-title"> {{ userInfo.shopName || '点餐系统连锁总店' }} </h3>
            <h4 class="item-subtitle">
              {{ $t('邀请您参加') }} <span>{{ imgName }}</span>
            </h4>
            <p class="item-desc"> {{ imgDescription }} </p>
            <img :src="qrcode" alt="" class="item-qrcode" />
            <p class="item-qrcode-desc"> {{ $t('到店消费时出示此二维码') }} </p>
            <img src="@/assets/img/activeImgTop.svg" alt="" class="item-img" />
          </div>
        </div>
      </div>

      <div class="image-preview-item-hide" id="image-preview-item">
        <div class="image-preview-bg"></div>
        <div class="image-preview-content">
          <div class="image-preview-content-item">
            <div class="logo-wrapper">
              <img :src="userInfo.logoUrl" />
            </div>
            <h3 class="item-title"> {{ userInfo.shopName || '点餐系统连锁总店' }} </h3>
            <h4 class="item-subtitle">
              {{ $t('邀请您参加') }} <span>{{ imgName }}</span>
            </h4>
            <p class="item-desc"> {{ imgDescription }} </p>
            <img :src="qrcode" alt="" class="item-qrcode" />
            <p class="item-qrcode-desc"> {{ $t('到店消费时出示此二维码') }} </p>
            <img src="@/assets/img/activeImgTop.svg" alt="" class="item-img" />
          </div>
        </div>
      </div>
    </div>
    <div class="image-preview-btn-wrapper">
      <el-button class="image-preview-btn" @click="previewImage" :loading="loading"> {{ $t('预览') }} </el-button>
      <el-button class="image-preview-btn" type="primary" @click="downloadImage" :loading="loading"> {{ $t('保存图片') }} </el-button>
    </div>
    <el-dialog v-model="previewImageVisible" :title="$t('预览')" width="560">
      <img :src="previewImageUrl" alt="" class="preview-image" />
    </el-dialog>
  </div>
</template>
<script setup>
  import { useUserStore } from '@/store';
  import html2canvas from 'html2canvas';
  import { ref, defineProps } from 'vue';
  const props = defineProps({
    form: {
      type: Object,
      default: () => {},
    },
    qrcode: {
      type: String,
      default: '',
    },
    imgName: {
      type: String,
      default: '',
    },
    imgDescription: {
      type: String,
      default: '',
    },
  });

  const { userInfo } = useUserStore();
  const previewImageVisible = ref(false);
  const previewImageUrl = ref('');
  const loading = ref(false);
  //div 转换为base64
  const convertToBase64 = async () => {
    const imagePreviewItem = document.getElementById('image-preview-item');
    try {
      const canvas = await html2canvas(imagePreviewItem, {
        useCORS: true,
        allowTaint: true,
        scale: 1,
      });
      return canvas.toDataURL('image/png');
    } catch (err) {
      console.log(err);
      throw err;
    }
  };

  const downloadImage = async () => {
    loading.value = true;
    const imgUrl = await convertToBase64();
    const a = document.createElement('a');
    a.href = imgUrl;
    a.download = 'image.png';
    a.click();
    loading.value = false;
  };

  const previewImage = async () => {
    loading.value = true;
    const imgUrl = await convertToBase64();
    previewImageUrl.value = imgUrl;
    previewImageVisible.value = true;
    loading.value = false;
  };

  defineExpose({
    convertToBase64,
  });
</script>
<style lang="scss" scoped>
  .image-preview-wrapper {
    width: 100%;
    padding: 36px;
    border-radius: 16px;
    background-color: #f3f5f6;
    position: relative;
    overflow: hidden;
    .image-preview-item {
      padding: 63px 0 45px 0;
      border-radius: 16px;
      background: #fff;
      position: relative;
      z-index: 2;
      overflow: visible;
      .image-preview-bg {
        width: 100%;
        height: 249px;
        background: linear-gradient(180deg, rgba(255, 190, 0, 0.25) 0%, rgba(255, 190, 0, 0) 100%);
        position: absolute;
        top: 0;
        left: 0;
        border-radius: 16px 16px 0 0;
      }
      .image-preview-content {
        background-image: url('@/assets/img/activeImgBottom.svg');
        background-repeat: no-repeat;
        background-position: bottom center;
        background-size: 100%;
        width: 222px;
        margin: 0 auto;
        position: relative;
        z-index: 1;
        padding-bottom: 12px;

        .image-preview-content-item {
          position: relative;
          z-index: 2;
          width: 197px;
          min-height: 290px;
          border-radius: 12px 12px 0 0;
          background: #fff;
          margin: 0 auto;
          padding-top: 40px;
          padding-bottom: 130px;
          .logo-wrapper {
            display: flex;
            justify-content: center;
            align-items: center;
            background-color: #fff;
            width: 65px;
            height: 65px;
            border-radius: 50%;
            overflow: hidden;
            position: absolute;
            margin: 0 auto;
            top: -33px;
            left: 0;
            right: 0;
            img {
              width: 57px;
              height: 57px;
              border-radius: 50%;
            }
          }
          .item-title {
            color: #100a05;
            text-align: center;
            font-size: 12px;
            font-style: normal;
            font-weight: 600;
            margin-bottom: 4px;
            padding: 0 18px;
          }
          .item-subtitle {
            color: rgba(36, 22, 11, 0.65);
            text-align: center;
            font-size: 8px;
            font-style: normal;
            font-weight: 400;
            margin-bottom: 16px;
            padding: 0 18px;
            span {
              color: #ffbe00;
              font-size: 9px;
              font-style: normal;
              font-weight: 600;
            }
          }
          .item-desc {
            color: #100a05;
            text-align: center;
            font-size: 10px;
            font-style: normal;
            font-weight: 400;
            margin-bottom: 12px;
            padding: 0 18px;
          }
          .item-qrcode {
            width: 96px;
            height: 96px;
            margin: 0 auto 4px;
            display: block;
          }
          .item-qrcode-desc {
            color: #999;
            text-align: center;
            font-size: 8px;
            font-style: normal;
            font-weight: 400;
            text-transform: capitalize;
            margin-top: 4px;
            margin-bottom: -24px;
          }
          .item-img {
            width: 207px;
            position: absolute;
            bottom: -5px;
            left: -5px;
            z-index: 3;
          }
        }
      }
    }

    .image-preview-item-hide {
      position: absolute;
      top: 0;
      left: 0;
      z-index: -2;
      padding: 105px 0 65px 0;
      background: #fff;
      overflow: visible;
      width: 390px;
      .image-preview-bg {
        width: 100%;
        height: 400px;
        background: linear-gradient(180deg, rgba(255, 190, 0, 0.25) 0%, rgba(255, 190, 0, 0) 100%);
        position: absolute;
        top: 0;
        left: 0;
      }
      .image-preview-content {
        background-image: url('@/assets/img/activeImgBottom.svg');
        background-repeat: no-repeat;
        background-position: bottom center;
        background-size: 100%;
        width: 357px;
        margin: 0 auto;
        position: relative;
        z-index: 1;
        padding-bottom: 12px;

        .image-preview-content-item {
          position: relative;
          z-index: 2;
          width: 320px;
          min-height: 460px;
          border-radius: 12px 12px 0 0;
          background: #fff;
          margin: 0 auto;
          padding-top: 55px;
          padding-bottom: 210px;
          .logo-wrapper {
            display: flex;
            justify-content: center;
            align-items: center;
            background-color: #fff;
            width: 96px;
            height: 96px;
            border-radius: 50%;
            overflow: hidden;
            position: absolute;
            margin: 0 auto;
            top: -45px;
            left: 0;
            right: 0;
            img {
              width: 85px;
              height: 85px;
              border-radius: 50%;
            }
          }
          .item-title {
            color: #100a05;
            text-align: center;
            font-size: 18px;
            font-style: normal;
            font-weight: 600;
            margin-bottom: 8px;
            padding: 0 30px;
          }
          .item-subtitle {
            color: rgba(36, 22, 11, 0.65);
            text-align: center;
            font-size: 13px;
            font-style: normal;
            font-weight: 400;
            margin-bottom: 36px;
            padding: 0 18px;
            span {
              color: #ffbe00;
              font-size: 14px;
              font-style: normal;
              font-weight: 600;
            }
          }
          .item-desc {
            color: #100a05;
            text-align: center;
            font-size: 15px;
            font-style: normal;
            font-weight: 400;
            margin-bottom: 16px;
            padding: 0 30px;
          }
          .item-qrcode {
            width: 140px;
            height: 140px;
            margin: 0 auto;
            display: block;
          }
          .item-qrcode-desc {
            color: #999;
            text-align: center;
            font-size: 14px;
            font-style: normal;
            font-weight: 400;
            text-transform: capitalize;
            margin-top: 4px;
            margin-bottom: -24px;
          }
          .item-img {
            width: 333px;
            position: absolute;
            bottom: -2px;
            left: -4px;
            z-index: 3;
          }
        }
      }
    }
  }
  .image-preview-btn-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: 20px;
  }
  .preview-image {
    display: block;
    margin: 0 auto;
  }
</style>
