<template>
    <div v-loading="loading">
        <div class="common-form">{{ $t('服务器存储空间') }}</div>
        <div class="memory-main">
            <p>{{ $t('总大小：') }}<span>{{ systemData.total_space }}GB</span></p>
            <p>{{ $t('可用大小：') }}<span>{{ systemData.free_space }}GB</span></p>
            <p>{{ $t('已用大小：') }}<span>{{ systemData.used_space }}GB</span></p>
            <p>{{ $t('已用百分比：') }}<span>{{ systemData.used_percentage }}%</span> </p>
        </div>
    </div>
</template>
<script>
import SettingApi from '@/api/setting.js';
export default {
    data() {
        return {
            systemData: {
                total_space: '',
                free_space: '',
                used_space: '',
                used_percentage: '',
            },
            loading: false,
        }
    },
    mounted(){
        this.getParams();
    },
    methods: {
        /*获取配置数据*/
        getParams() {
            let self = this;
            self.loading = true;
            SettingApi.getServerStorage({}, true)
                .then(res => {
                    self.systemData = res.data;
                    self.loading = false;
                })
                .catch(error => {
                    self.loading = false;
                });
        },
    },
}
</script>
<style lang="scss" scoped>
.memory-main {
    display: flex;
    gap: 16px;

    p {
        font-size: 14px;

        span {
            color: #FFBE00;
        }
    }
}
</style>