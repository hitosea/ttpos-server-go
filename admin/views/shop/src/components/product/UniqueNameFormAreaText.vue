<template>
  <div class="unique-name-form" :style="{ width: props.width ?? '100%' }">
    <el-form
      :model="form"
      ref="uniqueNameFormRef"
      label-position="top"
      label-width="240px"
      size="small"
      scroll-to-error
      @validate="handleValidate"
      v-bind="props.overridesElFormProps"
    >
      <template v-if="props.singleLanguage">
        <el-form-item
          :label="`${props.labelPrefix}`"
          :error="formErrors.SINGLE"
          :required="required"
          :rules="[
            { required: true, message: props.placeholder ?? $t('请输入名称') },
            {
              validator: (rule, value, callback) => {
                if (value && value.trim() === '') {
                  callback(new Error(props.placeholder ?? $t('请输入名称')));
                } else {
                  callback();
                }
              },
            },
          ]"
          prop="SINGLE"
          :validate-status="formErrors.SINGLE ? 'error' : translateLoading || validateLoading ? 'validating' : ''"
          for="no_click"
        >
          <div class="unique-name-form-item-wrapper">
            <el-input
              class="unique-name-form-item-input"
              v-model="form.SINGLE"
              type="textarea"
              :readonly="translateLoading || validateLoading"
              :maxlength="props.maxlength ?? 255"
              :placeholder="props.placeholder ?? $t('请输入名称')"
              :disabled="props.disabled"
              :rows="3"
              show-word-limit
            >
            </el-input>
          </div>
        </el-form-item>
      </template>
      <template v-else>
        <el-form-item
          v-for="languageKey in props.singleInputMode ? [languageKeys] : languageStore.languageKeys"
          :key="languageKey"
          :label="`${props.labelPrefix}${props.singleInputMode ? '' : `(${languageStore.getLanguageValueByKey(languageKey)})`}`"
          :error="formErrors[languageKey]"
          :required="required"
          :rules="[
            { required: true, message: props.placeholder ?? $t('请输入名称') },
            {
              validator: (rule, value, callback) => {
                if (value && value.trim() === '') {
                  callback(new Error(props.placeholder ?? $t('请输入名称')));
                } else {
                  callback();
                }
              },
            },
          ]"
          :prop="languageKey"
          :validate-status="formErrors[languageKey] ? 'error' : translateLoading || validateLoading ? 'validating' : ''"
          for="no_click"
        >
          <div class="unique-name-form-item-wrapper">
            <el-input
              class="unique-name-form-item-input"
              v-model="form[languageKey]"
              type="textarea"
              :rows="3"
              :readonly="translateLoading || validateLoading"
              @focus="() => handleFocus(languageKey)"
              @blur="() => handleBlur(languageKey)"
              :maxlength="props.maxlength ?? 255"
              :placeholder="props.placeholder ?? $t('请输入名称')"
              :disabled="props.disabled"
            >
            </el-input>
            <template v-if="focusLanguageKey === languageKey && form[languageKey]">
              <el-button
                class="unique-name-form-item-translate-button"
                type="primary"
                @click="handleTranslate(languageKey)"
                :disabled="!form[languageKey] || translateLoading"
                :loading="translateLoading"
              >
                {{ $t('点击翻译') }}
              </el-button>
            </template>
          </div>
        </el-form-item>
      </template>
    </el-form>
  </div>
</template>

<script setup>
  import { ref, reactive, nextTick, getCurrentInstance, watch, computed } from 'vue';
  import { useLanguageStore } from '@/store';
  import IndexApi from '@/api/index.js';

  defineOptions({
    name: 'UniqueNameForm',
  });

  const { proxy } = getCurrentInstance();

  const languageStore = useLanguageStore();
  const languageKeys = computed(() => languageStore.currentLanguage);

  const emit = defineEmits(['nowLangeData']);

  const props = defineProps({
    //是否必填
    required: {
      type: Boolean,
      required: false,
      default: true,
    },

    singleLanguage: {
      type: Boolean,
      required: false,
      default: false,
    },
    //单输入框模式
    singleInputMode: {
      type: Boolean,
      required: false,
      default: false,
    },
    overrideLanguages: {
      type: Object,
      required: false,
    },
    overridesElFormProps: {
      type: Object,
      required: false,
      default: () => ({
        labelWidth: '240px',
        scrollToError: true,
        size: 'small',
        labelPosition: 'top',
      }),
    },
    labelPrefix: {
      type: String,
      default: '',
    },
    maxlength: {
      type: Number,
      default: 255,
      validator: (value) => {
        return value > 0;
      },
    },
    placeholder: {
      type: String,
      required: false,
    },
    width: {
      type: String,
      required: false,
      default: '100%',
    },
    apiId: {
      type: Number,
      required: false,
      default: undefined,
    },
    parent_id: {
      type: Number,
      required: false,
      default: undefined,
    },
    //是否验证唯一性
    isUnique: {
      type: Boolean,
      required: false,
      default: true,
    },
    disabled: {
      type: Boolean,
      required: false,
      default: false,
    },
    apiSource: {
      type: String,
      required: true,
      validator: (value) => {
        //来源
        // product_barcode-产品商品条码
        // product_img-产品图片名称
        // product-产品
        // category-分类
        // sku-规格库
        // attribute-属性库
        // feed-加料库
        // unit-单位库
        // label-打印标签
        // buffet-自助餐
        // table-桌位
        // table_area-桌位区域
        // table_type-桌位类型
        // printer-打印机管理
        // supplier_printing-商品打印
        // pay_type-支付管理
        return [
          'product_barcode',
          'product_img',
          'product',
          'category',
          'sku',
          'attribute',
          'feed',
          'unit',
          'label',
          'buffet',
          'table',
          'table_area',
          'table_type',
          'printer',
          'supplier_printing',
          'pay_type',
        ].includes(value);
      },
    },
    otherGroupNames: {
      type: Array,
      required: false,
      default: () => [],
    },
  });

  const form = reactive(props.singleLanguage ? { SINGLE: '' } : languageStore.getLanguageKeyForm());
  const formErrors = reactive(props.singleLanguage ? { SINGLE: '' } : languageStore.getLanguageKeyForm());

  if (typeof props.overrideLanguages === 'object') {
    for (const key of Object.keys(form)) {
      form[key] = props.overrideLanguages[key] ?? '';
    }
  }

  const uniqueNameFormRef = ref(null);

  const translateLoading = ref(false);
  const validateLoading = ref(false);

  const handleTranslate = async (language) => {
    translateLoading.value = true;
    const data = {
      data: [
        {
          lang: language,
          content: form[language],
        },
      ],
    };
    try {
      const res = await IndexApi.aiTranslate(data, true);
      const {
        data: { data: translateDataJsonString },
      } = res;
      const translateData = JSON.parse(translateDataJsonString);
      for (const key of Object.keys(form)) {
        form[key] = key === 'zhtw' ? translateData[0]['zh-TW'].substring(0, props.maxlength) ?? '' : translateData[0][key].substring(0, props.maxlength) ?? '';
      }
      await nextTick();
      //   await handleValidateUnique();
      //   await validateOtherGroupNames();
    } catch (error) {
      console.error(error);
    } finally {
      translateLoading.value = false;
      focusLanguageKey.value = '';
    }
  };

  const validate = async () => {
    validateLoading.value = true;
    try {
      const baseValidateResult = await uniqueNameFormRef.value.validate();
      if (!baseValidateResult) return false;
      //   const uniqueValidateResult = await handleValidateUnique();
      //   if (!uniqueValidateResult) return false;
      //   const otherGroupNamesValidateResult = await validateOtherGroupNames();
      //   if (!otherGroupNamesValidateResult) return false;
      return true;
    } catch (error) {
      console.error('validate error', error);
      return false;
    } finally {
      validateLoading.value = false;
    }
  };

  const handleValidate = async (prop, isValid, message) => {
    formErrors[prop] = isValid ? '' : message;
  };
  //  TODO: 2025年11月06日09:24:06 去掉唯一性
  //   const handleValidateUnique = async () => {
  //     if (!props.isUnique) return true;
  //     const params = {
  //       name: form,
  //       source: props.apiSource,
  //       id: props.apiId,
  //     };
  //     if (props.apiSource == 'attribute') {
  //       params.parent_id = props.parent_id;
  //     }
  //     validateLoading.value = true;
  //     try {
  //       const res = await IndexApi.checkNameExist(params, true);
  //       await handleValidateUniqueResult(res.data);
  //       return Object.values(res.data).every((value) => value === false);
  //     } catch (error) {
  //       console.error('validate unique error', error);
  //       return false;
  //     } finally {
  //       validateLoading.value = false;
  //     }
  //   };

  // 校验是否与其他分组名称重复 2025年11月06日09:26:25 去掉唯一性
  //   const validateOtherGroupNames = async () => {
  //     if (!props.otherGroupNames.length) return true;

  //     // 清空之前的重复错误信息
  //     for (const key of Object.keys(formErrors)) {
  //       if (formErrors[key] === $t('此名称与其他分组重复')) {
  //         formErrors[key] = '';
  //       }
  //     }

  //     let hasRepeat = false;

  //     // 检查每个语言字段是否与其他分组对应的key值的名称重复
  //     for (const languageKey of Object.keys(form)) {
  //       const currentValue = form[languageKey];
  //       if (!currentValue || currentValue.trim() === '') continue;

  //       // 只比较相同语言键的值
  //       const isRepeat = props.otherGroupNames.some((item) => {
  //         return item[languageKey] === currentValue;
  //       });

  //       if (isRepeat) {
  //         formErrors[languageKey] = $t('此名称与其他分组重复');
  //         hasRepeat = true;
  //       }
  //     }

  //     return !hasRepeat;
  //   };

  //   const handleValidateUniqueResult = async (data) => {
  //     formErrors.value = {};
  //     await nextTick();
  //     for (const key of Object.keys(formErrors)) {
  //       formErrors[key] = data[key] ? $t('此名称已存在') : '';
  //     }
  //   };

  const focusLanguageKey = ref('');
  const handleFocus = (languageKey) => {
    focusLanguageKey.value = languageKey;
  };

  const handleBlur = async (languageKey) => {
    await nextTick();
    await new Promise((resolve) => setTimeout(resolve, 300));
    if (translateLoading.value) return;
    if (focusLanguageKey.value === languageKey) {
      focusLanguageKey.value = '';
    }
  };

  watch(
    () => form,
    (newVal) => {
      emit('nowLangeData', newVal[languageKeys]);
    },
    {
      immediate: true,
      deep: true,
    }
  );

  defineExpose({
    validate: validate,
    data: form,
  });
</script>

<style lang="scss" scoped>
  .unique-name-form {
    .unique-name-form-item-wrapper {
      width: 100%;
      display: flex;
      align-items: center;
      gap: 10px;

      .unique-name-form-item-input {
        flex-grow: 1;
      }

      .unique-name-form-item-translate-button {
        flex-shrink: 0;
      }
    }
  }
</style>
