<template>
  <!--
      	
      	时间：2020-06-01
      	描述：插件中心-分销-提现申请-弹窗
      -->
  <div>
    <el-dialog :title="$t('下载二维码')" width="35%" v-model="dialogVisible" @close="dialogFormVisible" :close-on-click-modal="false" :close-on-press-escape="false">
      <div v-loading="loading">
        <el-form size="small">
          <el-form-item :label-width="formLabelWidth">
            <!-- <el-radio-group v-model="source">
						<el-radio label="wx">{{ $t('微信小程序') }}</el-radio>
						<el-radio label="mp">{{ $t('公众号，H5网页') }}</el-radio>
					</el-radio-group> -->
          </el-form-item>
          <canvas ref="qrCanvas" style="display: none"></canvas>
          <img class="qr-code" :src="modifiedQRUrl" style="width: 200px; margin: auto" />
          <p v-if="type === 'menu'" class="click-button" @click="handleClick">
            {{ $t('更新二维码') }}
          </p>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="qrcodeClick">{{ $t('下载') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
  import qs from 'qs';
  import { useUserStore } from '@/store';
  import SettingApi from '@/api/setting.js';

  const { token, userInfo } = useUserStore();
  export default {
    data() {
      return {
        status: '',
        reject_reason: '',
        /*左边长度*/
        formLabelWidth: '120px',
        /*是否显示*/
        dialogVisible: false,
        loading: false,
        source: 'h5',
        token,
        QRUrl: '',
        modifiedQRUrl: '',
        userInfo,
      };
    },
    props: ['open', 'type'],
    watch: {
      open: function (n, o) {
        this.dialogVisible = this.open;
        if (n) {
          this.getCode('get');
        }
      },
    },
    created() {},
    methods: {
      getCode(e) {
        self.loading = true;
        if (this.type === 'menu') {
          SettingApi.getBusinessQrcode(
            {
              action: e,
            },
            true
          )
            .then((data) => {
              self.loading = false;
              this.QRUrl = data.data;
              this.modifyQRCode();
            })
            .catch((error) => {
              self.loading = false;
            });
        } else {
          SettingApi.getCompanyQrcode(
            {
              action: e,
            },
            true
          )
            .then((data) => {
              self.loading = false;
              this.QRUrl = data.data;
              this.modifyQRCode();
            })
            .catch((error) => {
              self.loading = false;
            });
        }
      },

      qrcodeClick() {
        const link = document.createElement('a');
        link.href = this.modifiedQRUrl;
        if (this.type === 'menu') {
          link.download = `${this.userInfo.shopName}_menu_qrcode.png`; // 设置下载文件名
        } else {
          link.download = `${this.userInfo.shopName}_company_qrcode.png`; // 设置下载文件名
        }
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);

        this.$emit('close', false);
      },

      handleClick() {
        ElMessageBox.confirm($t('确认更新?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            this.getCode('update');
          })
          .catch(() => {
            this.$ElMessage({
              type: 'info',
              message: $t('已取消更新'),
            });
          });
      },

      /*关闭弹窗*/
      dialogFormVisible(e) {
        this.QRUrl = '';
        if (e) {
          this.$emit('close', true);
        } else {
          this.$emit('close', false);
        }
      },

      modifyQRCode() {
        let self = this;
        const canvas = self.$refs.qrCanvas;
        const ctx = canvas.getContext('2d');
        const img = new Image();
        img.src = self.QRUrl;

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
          if (this.type === 'menu') {
            ctx.fillText(self.$t('电子菜单'), canvas.width / 2, 64);
          }

          // 绘制二维码
          if (this.type === 'menu') {
            ctx.drawImage(img, 80, 80);
          } else {
            ctx.drawImage(img, 80, 55);
          }

          // 绘制底部文字
          ctx.font = '54px Arial';

          // 计算最大允许宽度
          const maxWidth = canvas.width * 0.8;

          // 测量文本宽度
          const textWidth = ctx.measureText(self.userInfo.shopName).width;

          let displayText = self.userInfo.shopName;

          // 如果文本宽度超过最大允许宽度，则截断并添加省略号
          if (textWidth > maxWidth) {
            displayText = self.userInfo.shopName.slice(0, 8) + '...';
          }

          // 绘制文本
          ctx.fillText(displayText, canvas.width / 2, canvas.height - 40);

          // 将Canvas内容转换为图片URL
          self.modifiedQRUrl = canvas.toDataURL('image/png');
        };
      },
    },
  };
</script>
<style scoped>
  .click-button {
    margin-top: 16px;
    cursor: pointer;
    text-align: center;
    font-weight: bold;
  }
</style>
