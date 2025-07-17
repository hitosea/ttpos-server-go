<template>
  <batchCompoent v-if="type != '5'"></batchCompoent>
  <importProduct v-if="type == '5'"></importProduct>
  <batchDiscount v-if="type == '6'"></batchDiscount>
</template>
<script>
  import batchCompoent from './batchCompoent.vue';
  import importProduct from './importProduct.vue';
  import batchDiscount from './batchDiscount.vue';
  export default {
    components: {
      batchCompoent,
      importProduct,
      batchDiscount,
    },
    data() {
      return {
        type: '',
        title: '',
      };
    },
    mounted() {
      this.type = this.$route.query.type;
      this.title = this.$route.query.title;
    },
  };
</script>
<style lang=""></style>
