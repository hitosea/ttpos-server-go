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
            <h3 class="item-title">
              {{ $t('麥樂寶  (泰國分店)') }}
            </h3>
            <p class="item-desc"> 您的好友 张总，邀请你成为商家会员，共享商家多样优惠您的好友 </p>
            <img src="@/assets/img/activeImgTop.svg" alt="" class="item-qrcode" />
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
            <h3 class="item-title">
              {{ $t('麥樂寶  (泰國分店)') }}
            </h3>
            <p class="item-desc"> 您的好友 张总，邀请你成为商家会员，共享商家多样优惠您的好友 </p>
            <img src="@/assets/img/activeImgTop.svg" alt="" class="item-qrcode" />
            <p class="item-qrcode-desc"> {{ $t('到店消费时出示此二维码') }} </p>
            <img src="@/assets/img/activeImgTop.svg" alt="" class="item-img" />
          </div>
        </div>
      </div>
    </div>
    <div class="image-preview-btn-wrapper">
      <el-button class="image-preview-btn" @click="previewImage"> {{ $t('预览') }} </el-button>
      <el-button class="image-preview-btn" type="primary" @click="downloadImage"> {{ $t('保存图片') }} </el-button>
    </div>
    <el-dialog v-model="previewImageVisible" :title="$t('预览')" width="560">
      <img :src="previewImageUrl" alt="" class="preview-image" />
    </el-dialog>
  </div>
</template>
<script setup>
  import { useUserStore } from '@/store';
  import html2canvas from 'html2canvas';
  import { ref } from 'vue';

  const { userInfo } = useUserStore();
  const previewImageVisible = ref(false);
  const previewImageUrl = ref('');
  const downloadImage = () => {
    const imagePreviewItem = document.getElementById('image-preview-item');
    html2canvas(imagePreviewItem, {
      useCORS: true,
      allowTaint: true,
      scale: 1,
    })
      .then((canvas) => {
        const imgUrl = canvas.toDataURL('image/png');
        const a = document.createElement('a');
        a.href = imgUrl;
        a.download = 'image.png';
        a.click();
      })
      .catch((err) => {
        console.log(err);
      });
  };

  const previewImage = () => {
    const imagePreviewItem = document.getElementById('image-preview-item');
    html2canvas(imagePreviewItem, {
      useCORS: true,
      allowTaint: true,
      scale: 1,
    })
      .then((canvas) => {
        previewImageUrl.value = canvas.toDataURL('image/png');
        previewImageVisible.value = true;
      })
      .catch((err) => {
        console.log(err);
      });
  };
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
          min-height: 320px;
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
            margin-bottom: 16px;
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
            margin: 0 auto;
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
            margin-bottom: 36px;
          }
          .item-desc {
            color: #100a05;
            text-align: center;
            font-size: 15px;
            font-style: normal;
            font-weight: 400;
            margin-bottom: 16px;
            padding: 0 18px;
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
