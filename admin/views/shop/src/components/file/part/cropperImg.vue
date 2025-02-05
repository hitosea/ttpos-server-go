<template>
    <div>
        <vue-cropper ref="cropper" :src="img" :aspectRatio="aspectRatio" alt=" "></vue-cropper>
        <span class="dialog-footer">
            <el-button :loading="loading" @click="handleClose"> {{ $t('取消') }}</el-button>
            <el-button type="primary" :loading="loading" @click="handelCropper"> {{ $t('确定') }} </el-button>
        </span>
    </div>
</template>
<script>
import VueCropper from 'vue-cropperjs';
import 'cropperjs/dist/cropper.css';
export default {
    components: { VueCropper },
    props: {
        imgSrc: {
            type: String,
            default: '',
        },
        aspectRatio: {
            type: Number,
            default: 1,
        },
    },
    data() {
        return {
            img: '',
            afterImg: '',
            loading: false,
            okLoading: '',
        }
    },
    created() {
        this.img = this.imgSrc
    },
    methods: {
        sureSava() {
            this.afterImg = this.$refs.cropper
                .getCroppedCanvas({
                    imageSmoothingQuality: 'high',
                    maxWidth: 1024,
                    maxHeight: 1024,
                })
                .toDataURL('image/png')
            this.$emit('upload', this.base64ToBlob(this.afterImg))
        },

        base64ToBlob(code) {
            const parts = code.split(';base64,')
            const contentType = parts[0].split(':')[1]
            const raw = window.atob(parts[1])
            const rawLength = raw.length
            const uInt8Array = new Uint8Array(rawLength)
            for (let i = 0; i < rawLength; ++i) {
                uInt8Array[i] = raw.charCodeAt(i)
            }
            return new Blob([uInt8Array], {
                type: contentType
            })
        },
        handelCropper() {
            if (!this.$refs.cropper.getCroppedCanvas()){
              this.$ElMessage({
                message: $t("图片无法显示，请重新上传"),
                type: "warning",
              });
              return
            }
            this.okLoading = ElLoading.service({
                lock: true,
                text: $t("图片裁剪中,请等待"),
                background: "rgba(0, 0, 0, 0.7)",
            });
            this.loading = true;
            this.sureSava();
        },

        handleClose() {
            this.$emit('handleClose')
        },
    },
}
</script>
<style scoped>
.dialog-footer {
    margin-top: 20px;
}
</style>
