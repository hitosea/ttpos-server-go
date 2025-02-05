<template>
    <el-dialog v-model="show" :title="turnDownRow.review_status == 0 ? $t('提示') : $t('驳回原因')" width="420" align-center @close="handleClose">
        <el-form size="small" :inline="true" ref="form" :model="form" label-position="top">
            <el-form-item v-if="turnDownRow.review_status == 0" style="width: 100%;margin-right: 0;" :label="$t('确定审核驳回吗？')" :rules="[{ required: true, message: $t('请输入驳回原因') }]"
                prop="refused">
                <el-input size="small" v-model="form.refused" :placeholder="$t('请输入驳回原因')"></el-input>
            </el-form-item>
            <p v-else>
                {{ turnDownRow.refused || $t('无')}}
            </p>
        </el-form>

        <template #footer>
            <div class="dialog-footer">
                <el-button v-if="turnDownRow.review_status == 0" @click="handleClose">{{ $t('取消') }}</el-button>
                <el-button v-if="turnDownRow.review_status == 2" @click="handleClose">{{ $t('关闭') }}</el-button>
                <el-button v-if="turnDownRow.review_status == 0" type="primary" @click="handleTurnDown"> {{ $t('确定') }}</el-button>
            </div>
        </template>
    </el-dialog>
</template>
<script>
import InventoryApi from '@/api/inventory.js';
export default {
    props: {
        dialogVisible: {
            type: Boolean,
            default: false,
        },
        turnDownRow: {
            default: '',
        },
    },
    data() {
        return {
            show: false,
            form: {
                refused: '',
            }
        }
    },
    created() {
        this.show = this.dialogVisible;

    },
    methods: {
        /*驳回*/
        handleTurnDown() {
            let self = this;
            self.$refs.form.validate((valid) => {
                if (valid) {
                    self.loading = true;
                    InventoryApi.erpDamagedProductRecordReview({
                        id: this.turnDownRow.id,
                        review_status: 2,
                        refused:this.form.refused,
                    },
                        true
                    )
                        .then(data => {
                            self.loading = false;
                            if (data.code == 1) {
                                this.$ElMessage({
                                    message: $t('操作成功'),
                                    type: 'success'
                                });
                                //刷新页面
                                this.handleClose();
                            } else {
                                self.loading = false;
                            }
                        })
                        .catch(error => {
                            self.loading = false;
                        });
                }
            })


        },

        handleClose() {
            this.$emit('closeDialog', {
                type: 'success',
                openDialog: false
            })
        },
    },
}
</script>
<style lang="">

</style>