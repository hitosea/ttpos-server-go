<template>
  <batchCompoent v-if="type != '5'"></batchCompoent>
  <importProduct v-if="type == '5'"></importProduct>
</template>
<script>
  import batchCompoent from './batchCompoent.vue';
  import importProduct from './importProduct.vue';
  export default {
    components: {
      batchCompoent,
      importProduct,
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
