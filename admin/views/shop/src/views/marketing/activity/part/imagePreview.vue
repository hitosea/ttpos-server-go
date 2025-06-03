<template>
  <div>
    <div class="image-preview-wrapper">
      <div class="image-preview-item" id="image-preview-item">
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
            <img src="@/assets/img/activeImgTop.png" alt="" class="item-qrcode" />
            <p class="item-qrcode-desc"> {{ $t('到店消费时出示此二维码') }} </p>
            <img src="@/assets/img/activeImgTop.png" alt="" class="item-img" />
          </div>
        </div>
      </div>
      <el-button class="image-preview-btn" @click="downloadImage"> {{ $t('下载') }} </el-button>
    </div>
  </div>
</template>
<script setup>
  import { useUserStore } from '@/store';
  import { ref } from 'vue';
  import html2canvas from 'html2canvas';

  const { userInfo } = useUserStore();
  const imagePreviewRef = ref(null);

  const downloadImage = () => {
    const imagePreviewItem = document.getElementById('image-preview-item');
    html2canvas(imagePreviewItem, {
      useCORS: true,
      allowTaint: true,
      scale: 2,
    }).then((canvas) => {
      const imgUrl = canvas.toDataURL('image/png');
      const a = document.createElement('a');
      a.href = imgUrl;
      a.download = 'image.png';
      a.click();
    });
  };
</script>
<style lang="scss" scoped>
  .image-preview-wrapper {
    width: 100%;
    padding: 36px;
    border-radius: 16px;
    background-color: #f3f5f6;
    .image-preview-item {
      padding: 63px 0 45px 0;
      border-radius: 16px;
      background: #fff;
      position: relative;
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
        background-image: url('@/assets/img/activeImgBottom.png');
        background-repeat: no-repeat;
        background-position: bottom center;
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
          padding-bottom: 117px;
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
            width: 206px;
            position: absolute;
            bottom: -5px;
            left: -5px;
            z-index: 3;
          }
        }
      }
    }
  }
</style>
