<template>
  <div class="date-box" :class="computeDate">
    <i class="el-icon el-input__icon el-range__icon">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
        <path
          fill="currentColor"
          d="M128 384v512h768V192H768v32a32 32 0 1 1-64 0v-32H320v32a32 32 0 0 1-64 0v-32H128v128h768v64zm192-256h384V96a32 32 0 1 1 64 0v32h160a32 32 0 0 1 32 32v768a32 32 0 0 1-32 32H96a32 32 0 0 1-32-32V160a32 32 0 0 1 32-32h160V96a32 32 0 0 1 64 0zm-32 384h64a32 32 0 0 1 0 64h-64a32 32 0 0 1 0-64m0 192h64a32 32 0 1 1 0 64h-64a32 32 0 1 1 0-64m192-192h64a32 32 0 0 1 0 64h-64a32 32 0 0 1 0-64m0 192h64a32 32 0 1 1 0 64h-64a32 32 0 1 1 0-64m192-192h64a32 32 0 1 1 0 64h-64a32 32 0 1 1 0-64m0 192h64a32 32 0 1 1 0 64h-64a32 32 0 1 1 0-64"
        ></path>
      </svg>
    </i>
    <el-date-picker size="small" v-model="dateStar" :disabled-date="starDate" :clearable="false" value-format="YYYY-MM-DD" type="date" :placeholder="$t('开始日期')">
    </el-date-picker>
    ~
    <el-date-picker size="small" v-model="dateEnd" :disabled-date="endDate" :clearable="false" value-format="YYYY-MM-DD" type="date" :placeholder="$t('结束日期')">
    </el-date-picker>
    <div class="clear-box" @click="clearDate">
      <i class="el-icon el-input__icon el-range__close-icon">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
          <path
            fill="currentColor"
            d="m466.752 512-90.496-90.496a32 32 0 0 1 45.248-45.248L512 466.752l90.496-90.496a32 32 0 1 1 45.248 45.248L557.248 512l90.496 90.496a32 32 0 1 1-45.248 45.248L512 557.248l-90.496 90.496a32 32 0 0 1-45.248-45.248z"
          ></path>
          <path fill="currentColor" d="M512 896a384 384 0 1 0 0-768 384 384 0 0 0 0 768m0 64a448 448 0 1 1 0-896 448 448 0 0 1 0 896"></path>
        </svg>
      </i>
    </div>
  </div>
</template>
<script>
  export default {
    data() {
      return {
        dateStar: '',
        dateEnd: '',
      };
    },
    computed: {
      computeDate() {
        if (this.dateStar && this.dateEnd) {
          this.$emit('onchange', this.dateStar, this.dateEnd);
        }
        if (this.dateStar || this.dateEnd) {
          return 'clear-show';
        }
      },
    },
    methods: {
      starDate(time) {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        if (this.dateEnd) {
          const endDate = new Date(this.dateEnd);
          endDate.setHours(0, 0, 0, 0);
          const thirtyDaysBeforeEnd = new Date(endDate.getTime() - 29 * 24 * 60 * 60 * 1000);
          return time.getTime() > today.getTime() || time.getTime() < thirtyDaysBeforeEnd.getTime();
        }
        return time.getTime() > today.getTime();
      },

      endDate(time) {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        if (this.dateStar) {
          const startDate = new Date(this.dateStar);
          startDate.setHours(0, 0, 0, 0);
          const thirtyDaysAfterStart = new Date(startDate.getTime() + 29 * 24 * 60 * 60 * 1000);
          return time.getTime() < startDate.getTime() || time.getTime() > thirtyDaysAfterStart.getTime() || time.getTime() > today.getTime();
        }
        return time.getTime() > today.getTime();
      },

      clearDate() {
        this.dateStar = '';
        this.dateEnd = '';
        this.$emit('onchange', this.dateStar, this.dateEnd);
      },
    },
  };
</script>
<style lang="scss" scoped>
  .date-box {
    display: flex;
    align-items: center;
    gap: 8px;
    border: solid 1px var(--el-border-color);
    border-radius: 4px;
    padding: 0 12px;

    :deep(.el-input__prefix-inner) {
      display: none;
    }

    :deep(.el-input__wrapper) {
      box-shadow: none !important;
      width: 120px;
    }

    :deep(.el-input__inner) {
      text-align: center;
    }

    .el-input__icon {
      color: var(--el-text-color-placeholder);
    }

    .clear-box {
      display: flex;
      width: 14px;
      cursor: pointer;
    }

    .clear-box i {
      display: none;
    }

    .el-range__icon {
      margin-right: 14px;
    }
  }

  .clear-show:hover .clear-box i {
    display: block;
  }
</style>
