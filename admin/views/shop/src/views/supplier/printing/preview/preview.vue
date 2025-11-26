<template>
  <el-dialog @close="handleClose" width="480" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="title">
    <template v-if="dialogWidth == '1'">
      <div class="tabs-box" v-if="title == $t('发票')">
        <div @click="modeChange(1)" class="tabs-button" :class="mode == 1 ? 'tabs-active' : ''">
          {{ $t('模板1') }}
        </div>
        <div @click="modeChange(2)" class="tabs-button" :class="mode == 2 ? 'tabs-active' : ''">
          {{ $t('模板2') }}
        </div>
      </div>
      <div class="box-border">
        <template v-if="mode == 1">
          <p class="Invoice-p">{{ $t('店铺名称') }}</p>
          <p class="Invoice-p mb-16" v-html="$t('非常感谢您今天的到来，我们期待您的再次光临')"></p>
          <h4 class="Invoice-h4">{{ titleName }}</h4>
          <p class="Invoice-p mb-24">2024/12/15 14:45:21</p>
          <h5>{{ $t('先生/小姐') }}</h5>
        </template>
        <template v-if="mode == 2">
          <p class="Invoice-p">{{ $t('店铺名称') }}</p>
          <img :src="userInfo.logoUrl" alt="" class="logo" />
          <h4 class="Invoice-h4 mb-8">{{ titleName }}</h4>
          <p class="Invoice-p font14" v-html="$t('公司：') + $t('公司名称公司名称公司名称')"></p>
          <p class="Invoice-p font14" v-html="$t('商家地址：') + $t('商家地址商家地址商家地址商家地址商家地址商家地址商家地址商家地址')"></p>
          <p class="Invoice-p font14" v-html="$t('电话：') + '02-15-1441414'"></p>
          <p class="Invoice-p font14 bold-bottom-24" v-html="$t('税号：') + '252452524144'"></p>
        </template>
        <!-- 小字的数据 -->
        <div class="box-main" v-for="(item, index) in details" :key="index">
          <div class="box-text-box" v-for="(items, indexs) in item" :key="indexs">
            <!-- 左边的字段 -->
            <div
              v-if="items.name && items.left != false"
              class="text-box"
              :class="[
                items.bold ? 'font-bold' : '',
                items.big ? 'font-big' : '',
                items.flexWidth ? 'flexWidth' : '',
                items.font500 ? 'font-500' : '',
                items.font700 ? 'font-700' : '',
                items.textCenter ? 'text-center' : '',
                items.font16Bold ? 'font16Bold' : '',
              ]"
            >
              {{ items.name }}{{ is_show_sku == 1 && items.addLabel ? items.addLabel : '' }}
            </div>
            <!-- 右边的字段 -->
            <div v-if="items.label" class="text-box text-box-r" :class="[items.flexWidthRight ? 'flexWidthRight' : '', items.num ? 'flex-end' : '']">
              <p
                class="text-box-r-p1"
                :class="[
                  items.bold ? 'font-bold' : '',
                  items.big ? 'font-big' : '',
                  items.font500 ? 'font-500' : '',
                  items.font700 ? 'font-700' : '',
                  items.textCenter ? 'text-center' : '',
                  items.font16Bold ? 'font16Bold' : '',
                ]"
                v-if="items.num"
              >
                {{ items.num }}
              </p>
              <p
                class="text-box-r-p2"
                :class="[
                  items.bold ? 'font-bold' : '',
                  items.big ? 'font-big' : '',
                  items.font18 ? 'font18' : '',
                  items.font500 ? 'font-500' : '',
                  items.font700 ? 'font-700' : '',
                  items.textCenter ? 'text-center' : '',
                  items.font16Bold ? 'font16Bold' : '',
                ]"
              >
                {{ items.label }}
                <img v-if="items.img" src="@/assets/img/dashed.svg" alt="" class="dashed" />
              </p>
            </div>
          </div>
        </div>
        <div class="brand-box" v-if="title == $t('发票')">{{ $t('感谢您的光临！本店由') }}{{ brand }}{{ $t('系统提供支持') }} </div>
      </div>
    </template>

    <template v-else-if="dialogWidth == '2'">
      <div class="tabs-box">
        <div @click="modeChange(1)" class="tabs-button" :class="mode == 1 ? 'tabs-active' : ''">
          {{ $t('模板1') }}
        </div>
        <div @click="modeChange(2)" class="tabs-button" :class="mode == 2 ? 'tabs-active' : ''">
          {{ $t('模板2') }}
        </div>
      </div>
      <div class="box-border one-menu">
        <template v-if="mode == 1">
          <p class="font24"> {{ $t('(外卖)') }}{{ $t('桌位: A01 (4人)') }} </p>
          <div class="text-order—remark">
            {{ $t('整单备注：这是整单备注备注') }}
          </div>
          <p class="mb-14"> <span class="span1">2024/05/04 14:15:12</span> </p>
          <p>
            <span class="span3">{{ $t('（打包）') }}{{ $t('商品名称商品名称') }}</span> <span class="span3">X1</span>
          </p>
          <p class="mb-8">
            <span class="span3">{{ $t('规格名称') }}</span>
          </p>
          <p>
            <span v-if="is_show_sku == 1" class="span4">{{ $t('少冰') }}</span>
          </p>
          <p>
            <span v-if="is_show_sku == 1" class="span4">{{ $t('加珍珠') }}</span>
          </p>
          <p class="mb-8">
            <span class="span4">{{ $t('这是备注这是备注这是备注') }}</span>
          </p>
          <p>
            <span class="span3">{{ $t('（打包）') }}{{ $t('套餐') }}-{{ $t('套餐名称套餐名称') }}</span> <span class="span3">X1</span>
          </p>
        </template>
        <template v-if="mode == 2">
          <h3>
            {{ $t('(外卖)') }}{{ $t('桌位: A01 (4人)') }}
          </h3>
          <div class="text-order—remark">
            {{ $t('整单备注：这是整单备注备注') }}
          </div>
          <p>
            <span class="span3">{{ $t('（打包）') }}{{ $t('商品名称商品名称') }}</span> <span class="span3">X1</span>
          </p>
          <p class="mb-8">
            <span class="span3">{{ $t('规格名称') }}</span>
          </p>
          <p>
            <span v-if="is_show_sku == 1" class="span4">{{ $t('少冰') }}</span>
          </p>
          <p>
            <span v-if="is_show_sku == 1" class="span4">{{ $t('加珍珠') }}</span>
          </p>
          <p class="mb-8">
            <span class="span4">{{ $t('这是备注这是备注这是备注') }}</span>
          </p>
          <p>
            <span class="span3">{{ $t('（打包）') }}{{ $t('套餐') }}-{{ $t('套餐名称套餐名称') }}</span> <span class="span3">X1</span>
          </p>
          <h2 class="border-top">2024/05/04 14:15:12</h2>
        </template>
      </div>
    </template>

    <template v-else>
      <div class="tabs-box" v-if="title == $t('结账单') || title == $t('预结账单') || title == $t('交班单') || title == $t('整单打印') || title == $t('退菜单')">
        <div @click="modeChange(1)" class="tabs-button" :class="mode == 1 ? 'tabs-active' : ''">
          {{ $t('模板1') }}
        </div>
        <div @click="modeChange(2)" class="tabs-button" :class="mode == 2 ? 'tabs-active' : ''">
          {{ $t('模板2') }}
        </div>
        <div @click="modeChange(3)" v-if="title != $t('退菜单')" class="tabs-button" :class="mode == 3 ? 'tabs-active' : ''">
          {{ $t('模板3') }}
        </div>
        <div @click="modeChange(4)" v-if="title == $t('结账单') || title == $t('预结账单')" class="tabs-button" :class="mode == 4 ? 'tabs-active' : ''">
          {{ $t('模板4') }}
        </div>
        <div @click="modeChange(5)" v-if="title == $t('结账单')" class="tabs-button" :class="mode == 5 ? 'tabs-active' : ''">
          {{ $t('模板5') }}
        </div>
      </div>
      <div class="tabs-box" v-if="detail">
        <div
          @click="customizeChange(detail?.default_tpl)"
          v-if="detail?.default_tpl?.name"
          class="tabs-button"
          :class="detail?.default_tpl?.customize_uuid == customize_uuid ? 'tabs-active' : ''"
        >
          {{ $t(detail?.default_tpl?.name) }}
        </div>
        <div
          @click="customizeChange(item)"
          v-if="detail?.is_adv_receipt_tpl"
          v-for="item in detail?.adv_receipt_tpls"
          :key="item.customize_uuid"
          class="tabs-button"
          :class="item.customize_uuid == customize_uuid ? 'tabs-active' : ''"
        >
          {{ $t(item.name) }}
        </div>
      </div>
      <div v-if="customize_uuid && customize_uuid != 0" class="text-center mb-12">{{ $t('如调整自定义模板，请前往桌面端操作') }}</div>
      <div class="box-border" v-if="imgUrl">
        <p class="title-name">
          <img style="width: 100%; height: 100%" :src="imgUrl" alt="" class="logo" />
        </p>
      </div>
      <div v-else class="box-border">
        <p class="title-name" v-if="(title == $t('结账单') || title == $t('预结账单')) && mode != 3 && mode != 4 && mode != 5">
          {{ title }}
        </p>
        <h2 class="font24" :class="mode != 1 ? 'mb-8' : 'mb-24'" v-if="mode != 3 && mode != 4 && mode != 5 && storeShow && (title == $t('结账单') || title == $t('预结账单'))">
          {{ $t('店铺名称') }}
        </h2>

        <template v-if="(mode == 3 || mode == 4 || mode == 5) && storeShow && (title == $t('结账单') || title == $t('预结账单'))">
          <h2>
            {{ $t('店铺名称') }}
          </h2>
          <img v-if="mode == 3 || mode == 4 || mode == 5" :src="userInfo.logoUrl" alt="" class="logo" />
          <h2 class="font24" :class="mode != 1 ? 'mb-8' : 'mb-24'">
            {{ title }}
          </h2>
        </template>

        <template v-if="title == $t('充值单')">
          <h4>
            {{ $t('店铺名称') }}
          </h4>
        </template>
        <template v-if="title == $t('外送单')">
          <h2>
            {{ $t('店铺名称') }}
          </h2>
          <img :src="userInfo.logoUrl" alt="" class="logo" />
          <h2 class="font24" :class="mode != 1 ? 'mb-8' : 'mb-24'">
            {{ titleName }}
          </h2>
        </template>

        <p
          v-if="(title == $t('结账单') && (mode == 3 || mode == 4 || mode == 5)) || (title == $t('预结账单') && (mode == 3 || mode == 4))"
          class="Invoice-p font14"
          v-html="$t('公司：') + $t('公司名称公司名称公司名称')"
        ></p>
        <p
          v-if="(title == $t('结账单') && (mode == 3 || mode == 4 || mode == 5)) || (title == $t('预结账单') && (mode == 3 || mode == 4))"
          class="Invoice-p font14"
          v-html="$t('商家地址：') + $t('商家地址商家地址商家地址商家地址商家地址商家地址商家地址商家地址')"
        ></p>
        <p
          v-if="(title == $t('结账单') && (mode == 3 || mode == 4 || mode == 5)) || (title == $t('预结账单') && (mode == 3 || mode == 4))"
          class="Invoice-p font14"
          v-html="$t('电话：') + '02-15-1441414'"
        ></p>
        <p
          v-if="(title == $t('结账单') && (mode == 3 || mode == 4 || mode == 5)) || (title == $t('预结账单') && (mode == 3 || mode == 4))"
          class="Invoice-p font14"
          :class="mode == 3 || mode == 4 ? 'bold-bottom-24' : ''"
          v-html="$t('税号：') + '252452524144'"
        >
        </p>
        <p v-if="title == $t('结账单') && mode == 5" class="Invoice-p font14" v-html="$t('收银机SN：') + '354543543958980'"> </p>
        <p v-if="title == $t('结账单') && mode == 5" class="Invoice-p font14" :class="mode == 5 ? 'bold-bottom-24' : ''" v-html="$t('打印机SN：') + '354543543958980'"> </p
        ><p v-if="(title == $t('结账单') && mode == 2) || (title == $t('预结账单') && mode == 2)" class="Invoice-p" v-html="$t('非常感谢您今天的到来，我们期待您的再次光临')"></p>
        <p v-if="title == $t('结账单') && mode == 2" class="Invoice-p font14"> 2024/05/04 14:45:21 </p>

        <h3 v-if="(title == $t('结账单') && mode == 2) || (title == $t('预结账单') && mode == 2)" class="font24">
          {{ titleName }}
        </h3>
        <template v-if="title == $t('交班单')">
          <p class="title-name" v-if="mode == 1">
            {{ titleName }}
          </p>
          <h3 v-if="mode == 1">
            {{ $t('店铺名称') }}
          </h3>

          <h2 v-if="mode == 2 || mode == 3">
            {{ $t('店铺名称') }}
          </h2>
          <h2 v-if="mode == 2 || mode == 3" class="font24 mb-8">
            {{ title }}
          </h2>
          <p v-if="mode == 2 || mode == 3" class="Invoice-p font14 mb-24"> 2023-12-15 14:00:21{{ $t('至') }}2023-12-15 00:05:58 </p>
        </template>
        <template v-if="title == $t('营业数据')">
          <p class="title-name">
            {{ titleName }}
          </p>
          <h3>
            {{ $t('店铺名称') }}
          </h3>
        </template>
        <template v-if="title == $t('整单打印')">
          <h4 class="mb-12" v-if="mode == 1">
            {{ $t('(外卖)') }}{{ $t('桌位: A01 (4人)') }}
          </h4>
          <h3 v-if="mode == 2 || mode == 3">
            {{ $t('(外卖)') }}{{ $t('桌位: A01 (4人)') }}
          </h3>
          <div class="text-order—remark">
            {{ $t('整单备注：这是整单备注备注') }}
          </div>
        </template>
        <template v-if="title == $t('退菜单')">
          <h2 class="mb-8" v-if="mode == 2">*******************************************</h2>
          <h4 class="mb-8" v-if="mode == 1 || mode == 2"> {{ $t('退菜单') }}</h4>
          <h2 class="mb-24" v-if="mode == 2">*******************************************</h2>
          <h4 class="Invoice-h4 mb-8" v-if="mode == 1 || mode == 2">
            {{ $t('(外卖)') }}{{ $t('桌位: A01 (4人)') }}
          </h4>
          <div class="text-order—remark">
            {{ $t('整单备注：这是整单备注备注') }}
          </div>
        </template>
        <template v-if="title == $t('出菜单')">
          <h4 class="mb-8"> {{ $t('出菜单') }}</h4>
          <h4 class="Invoice-h4 mb-8">
            {{ $t('(外卖)') }}{{ $t('桌号/序号/外送: A01 (4人)') }}
          </h4>
          <div class="text-order—remark">
            {{ $t('整单备注：这是整单备注备注') }}
          </div>
        </template>
        <!-- 小字的数据 -->
        <div
          class="box-main"
          :class="[
            (index == 0 && title != $t('充值单')) ||
            (title == $t('预结账单') && index == 3 && mode == 2) ||
            (title == $t('预结账单') && index == 4 && mode == 3) ||
            (title == $t('退菜单') && index == 3 && mode == 2)
              ? 'box-main-border'
              : '',
            title == $t('交班单') || title == $t('营业数据') ? 'bold-bottom' : '',
          ]"
          v-for="(item, index) in details"
          :key="index"
          v-show="(item[0]?.allHide && item[0].allHide.includes(mode)) || !item[0]?.allHide"
        >
          <div
            class="box-text-box"
            :class="items.showSkuAttr && is_show_sku == 0 ? 'show-sku-attr' : ''"
            v-for="(items, indexs) in item"
            :key="indexs"
            v-show="(items.typeShow && items.typeShow.includes(mode)) || !items.typeShow"
          >
            <!-- 左边的字段 -->
            <div
              class="text-box"
              v-if="items.left != false && mode != items.hide"
              :class="[
                items.bold ? 'font-bold' : '',
                items.big ? 'font-big' : '',
                items.flexWidth ? 'flexWidth' : '',
                items.font24 ? 'font24' : '',
                items.font22 ? 'font22' : '',
                items.font18Small ? 'font18Small' : '',
                items.font17Normal ? 'font17Normal' : '',
                items.font16Small ? 'font16Small' : '',
                items.font16Normal ? 'font16Normal' : '',
                items.font500 ? 'font-500' : '',
                items.font700 ? 'font-700' : '',
                items.textCenter ? 'text-center' : '',
                items.font16Bold ? 'font16Bold' : '',
                items.lineHeight ? 'line-height-' + items.lineHeight : '',
              ]"
              v-html="items.name"
            >
            </div>
            <!-- 右边的字段 -->
            <div v-if="items.right != false && mode != items.hide" class="text-box text-box-r" :class="[items.flexWidthRight ? 'flexWidthRight' : '', items.num ? 'flex-end' : '']">
              <p
                class="text-box-r-p1"
                :class="[
                  items.bold ? 'font-bold' : '',
                  items.big ? 'font-big' : '',
                  items.font24 ? 'font24' : '',
                  items.font22 ? 'font22' : '',
                  items.font18Small ? 'font18Small' : '',
                  items.font17Normal ? 'font17Normal' : '',
                  items.font16Small ? 'font16Small' : '',
                  items.font16Normal ? 'font16Normal' : '',
                  items.font500 ? 'font-500' : '',
                  items.font700 ? 'font-700' : '',
                  items.textCenter ? 'text-center' : '',
                  items.lineHeight ? 'line-height-' + items.lineHeight : '',
                ]"
                v-if="items.num"
              >
                {{ items.num }}
              </p>
              <p
                class="text-box-r-p2"
                :class="[
                  items.bold ? 'font-bold' : '',
                  items.big ? 'font-big' : '',
                  items.font24 ? 'font24' : '',
                  items.font22 ? 'font22' : '',
                  items.font18Small ? 'font18Small' : '',
                  items.font17Normal ? 'font17Normal' : '',
                  items.font16Small ? 'font16Small' : '',
                  items.font16Normal ? 'font16Normal' : '',
                  items.font500 ? 'font-500' : '',
                  items.font700 ? 'font-700' : '',
                  items.textCenter ? 'text-center' : '',
                  items.lineHeight ? 'line-height-' + items.lineHeight : '',
                ]"
              >
                {{ items.label }}
              </p>
            </div>
          </div>
        </div>
        <div class="code-box" v-if="title == $t('预结账单')" :class="title == $t('预结账单') && mode != 1 ? 'border-top-12' : ''">
          <img src="@/assets/QRcode.svg" alt="" class="code" />
          <p>{{ $t('请用 Wechat 扫一扫支付') }}</p>
        </div>
        <div class="code-box" v-if="title == $t('结账单') && mode == 5" :class="title == $t('结账单') && mode == 5 ? 'border-top-12' : ''">
          <img src="@/assets/barcode.png" alt="" class="barcode" />
        </div>
        <div
          class="brand-box"
          :class="title == $t('预结账单') || (title == $t('结账单') && mode == 5) ? 'border-top' : ''"
          v-if="title == $t('结账单') || title == $t('预结账单') || title == $t('外送单')"
          >{{ $t('感谢您的光临！本店由') }}{{ brand }}{{ $t('系统提供支持') }}</div
        >
        <template v-if="title == $t('退菜单') && mode == 2">
          <h2 class="mb-4 mt-n18">*******************************************</h2>
          <div class="brand-box" :class="title == $t('退菜单') ? 'border-top-none' : ''">
            {{ $t('请停止制作以上菜品！') }}
          </div>
          <h2>*******************************************</h2>
        </template>
      </div>
    </template>
    <div class="flex-switch" v-if="title == $t('整单打印') || (title == $t('发票') && mode == 2) || title == $t('一菜一单')">
      <span class="font14">{{ $t('是否显示规格、属性') }}</span>
      <el-switch v-model="is_show_sku" :active-value="1" :inactive-value="0" />
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose" :loading="loading"> {{ $t('关闭') }}</el-button>
        <el-button
          v-if="
            title == $t('结账单') ||
            title == $t('预结账单') ||
            title == $t('发票') ||
            title == $t('交班单') ||
            title == $t('整单打印') ||
            title == $t('一菜一单') ||
            title == $t('退菜单')
          "
          type="primary"
          @click="onSubmit"
          :loading="loading"
        >
          {{ $t('确定') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>
<script>
  import SettingApi from '@/api/setting.js';
  import { languageStore } from '@/store/model/language';
  import { useUserStore } from '@/store';
  import { previewData } from './data';
  const cloudBasic = languageStore().getCloudBasic().cloudBasic;
  const languageKey = languageStore().getLanguageKey().language;
  const { userInfo } = useUserStore();

  export default {
    data() {
      return {
        userInfo,
        cloudBasic: cloudBasic,
        languageKey: languageKey,
        brand: '',
        dialogWidth: '',
        loading: false,
        dialogVisible: false,
        storeShow: true,
        details: [],
        titleName: '',
        mode: 1,
        is_show_sku: 0,
        detail: null,
        customize_uuid: 0,
        imgUrl: '',
      };
    },
    props: ['open', 'title', 'template', 'editId', 'print_method', 'isShowSku', 'tmpUuid'],
    created() {
      this.dialogVisible = this.open;
      this.brand = this.cloudBasic.base.brand_name;
      this.is_show_sku = this.isShowSku;
      this.customize_uuid = this.tmpUuid;
      this.mode = this.tmpUuid ? 0 : this.template;
      if (this.title == $t('交班单')) {
        if (this.mode == 1) {
          this.details = previewData.one;
        }
        if (this.mode == 2) {
          this.details = previewData.nine;
        }
        if (this.mode == 3) {
          this.details = previewData.twelve;
        }
        this.titleName = this.title;
      }
      if (this.title == $t('结账单')) {
        this.details = previewData.two;
        this.titleName = $t('桌位: ') + 'A01' + $t('（4人）');
      }
      if (this.title == $t('预结账单')) {
        this.details = previewData.three;
        this.titleName = $t('桌位: ') + 'A01' + $t('（4人）');
      }
      if (this.title == $t('一菜一单')) {
        this.dialogWidth = 2;
        this.storeShow = false;
      }
      if (this.title == $t('营业数据')) {
        this.details = previewData.five;
        this.titleName = this.title;
      }
      if (this.title == $t('整单打印')) {
        this.details = previewData.six;
        this.storeShow = false;
        this.titleName = $t('桌位: ') + 'A01';
      }
      if (this.title == $t('发票')) {
        if (this.mode == 1) {
          this.details = previewData.seven;
        }
        if (this.mode == 2) {
          this.details = previewData.eight;
        }
        this.dialogWidth = 1;
        this.titleName = $t('发票');
      }
      if (this.title == $t('充值单')) {
        this.details = previewData.ten;
        this.titleName = this.title;
      }
      if (this.title == $t('退菜单')) {
        if (this.mode == 1) {
          this.details = previewData.eleven;
        }
        if (this.mode == 2) {
          this.details = previewData.thirteen;
        }
        this.storeShow = false;
        this.titleName = $t('桌位: ') + 'A01';
      }
      if (this.title == $t('出菜单')) {
        this.details = previewData.fourteen;
        this.storeShow = false;
        this.titleName = $t('桌号/序号/外送: A01 (4人)');
      }
      if (this.title == $t('外送单')) {
        this.details = previewData.fifteen;
        this.storeShow = false;
        this.titleName = $t('外送: 0001');
      }
      // 获取模板数据
      this.getData();
    },
    methods: {
      getData() {
        let self = this;
        let Params = {};
        Params.id = self.editId;
        SettingApi.printerTemplateDetail(Params, true).then((data) => {
          self.detail = data.data;
          if (this.customize_uuid != 0) {
            let customize = self.detail?.adv_receipt_tpls?.find((item) => item.customize_uuid == this.customize_uuid);
            if (customize) {
              self.customizeChange(customize);
            } else if (self.detail?.default_tpl) {
              self.customizeChange(self.detail?.default_tpl);
            }
          }
        });
      },
      onSubmit() {
        if (((this.title == $t('结账单') && this.mode == '3') || (this.title == $t('发票') && this.mode == '2') || this.title == $t('预结账单')) && this.print_method == 1) {
          ElMessageBox.confirm($t('请前去打印设置调整打印方式为“图片打印”，否则图片将无法正常打印，如已设置，请忽略此步骤'), $t('提示'), {
            confirmButtonText: $t('确定'),
            cancelButtonText: $t('取消'),
            type: 'warning',
          })
            .then(() => {
              this.handleSubmit();
            })
            .catch(() => {
              this.$ElMessage({
                type: 'info',
                message: $t('已取消'),
              });
            });
        } else {
          this.handleSubmit();
        }
      },

      handleSubmit() {
        let self = this;
        let form = {};
        form.id = this.editId;
        form.template = this.mode;
        form.is_show_sku = this.is_show_sku;
        form.tmp_uuid = this.customize_uuid;
        self.loading = true;
        SettingApi.setTemplate(form, true)
          .then((data) => {
            self.loading = false;
            this.$ElMessage({
              message: $t('操作成功'),
              type: 'success',
            });
            this.$emit('close', 1);
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      handleClose() {
        this.$emit('close');
      },

      customizeChange(e) {
        this.imgUrl = e.img_url;
        this.customize_uuid = e.customize_uuid;
        this.mode = 0;
      },

      modeChange(e) {
        this.mode = e;
        this.customize_uuid = 0;
        this.imgUrl = '';
        if (this.title == $t('发票')) {
          if (e == 1) {
            this.details = previewData.seven;
          }
          if (e == 2) {
            this.details = previewData.eight;
          }
        }

        if (this.title == $t('交班单')) {
          if (e == 1) {
            this.details = previewData.one;
          }
          if (e == 2) {
            this.details = previewData.nine;
          }
          if (e == 3) {
            this.details = previewData.twelve;
          }
        }

        if (this.title == $t('退菜单')) {
          if (e == 1) {
            this.details = previewData.eleven;
          }
          if (e == 2) {
            this.details = previewData.thirteen;
          }
        }
      },
    },
  };
</script>
<style scoped>
  * {
    text-transform: capitalize;
  }
  .tabs-box {
    display: flex;
    margin-bottom: 12px;
  }

  .tabs-button {
    padding: 5px 12px;
    border: 1px solid var(--el-color-border);
    cursor: pointer;
    color: var(--el-color-black);
    font-size: 14px;
    font-weight: 400;
  }

  .tabs-button:first-child {
    border-radius: 4px 0 0 4px;
  }

  .tabs-button:last-child {
    border-radius: 0 4px 4px 0;
  }

  .tabs-active {
    border: 1px solid var(--el-color-primary);
    background: var(--el-color-primary);
  }

  .box-border {
    border: 1px solid var(--el-color-border);
    border-radius: 4px;
    padding: 24px;
  }

  .title-name {
    font-size: 18px;
    font-weight: 500;
    margin-bottom: 8px;
    color: var(--el-color-black);
  }

  .box-border h2 {
    text-align: center;
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
  }

  .box-border h2.font24 {
    text-align: center;
    font-size: 24px;
    font-weight: 700;
    color: var(--el-color-black);
  }

  .box-border .Invoice-p {
    font-size: 16px;
    font-weight: 400;
    color: var(--el-color-black);
    text-align: center;
  }

  .box-border .Invoice-p-18 {
    font-size: 18px;
    font-weight: 400;
    color: var(--el-color-black);
    text-align: center;
  }

  .box-border .font14 {
    font-size: 14px;
  }

  .flex-switch {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
  }

  .one-menu p {
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: var(--el-color-black);
    font-size: 18px;
    font-weight: 700;
  }
  .one-menu p.font24 {
    font-size: 24px;
    font-weight: 700;
  }
  .one-menu p .span1 {
    font-size: 16px;
    font-weight: 400;
  }

  .one-menu p .span2 {
    font-size: 16px;
    font-weight: 400;
  }
  .one-menu p .span3 {
    font-size: 22px;
    font-weight: 500;
  }
  .one-menu p .span4 {
    font-size: 18px;
    font-weight: 400;
  }

  .box-border h3 {
    text-align: center;
    font-size: 24px;
    font-weight: 700;
    color: var(--el-color-black);
    margin-bottom: 24px;
  }

  .box-border h3.font16 {
    text-align: center;
    font-size: 16px;
    font-weight: 700;
    color: var(--el-color-black);
    margin-bottom: 24px;
  }

  .box-border h4 {
    text-align: center;
    font-size: 24px;
    color: var(--el-color-black);
    margin-bottom: 24px;
    font-weight: 700;
  }

  .box-border .Invoice-h4 {
    margin-bottom: 0;
  }

  .text-order—remark {
    font-size: 24px;
    font-weight: 700;
    color: var(--el-color-black);
    margin-bottom: 12px;
    margin-top: 4px;
    border-top: 1px dashed var(--el-color-black);
    border-bottom: 1px dashed var(--el-color-black);
    padding: 12px 0;
  }

  .box-border h4.mb-8 {
    margin-bottom: 8px;
  }
  .box-border h4.mb-12 {
    margin-bottom: 12px;
  }

  .box-border h4 span {
    font-size: 16px;
    font-weight: 400;
  }

  .border-box {
    display: flex;
    margin-top: 8px;
  }

  .box-border h5 {
    font-size: 20px;
    color: var(--el-color-black);
    font-weight: 400;
    margin-bottom: 8px;
    border-bottom: 1px dashed var(--el-color-black);
    font-size: 16px;
    width: 100%;
    text-align: right;
    padding-bottom: 8px;
  }

  .box-border h6 {
    font-size: 18px;
    color: var(--el-color-black);
    margin-bottom: 8px;
    font-weight: 700;
    border-bottom: 1px dashed var(--el-color-black);
    width: 100%;
    text-align: right;
    margin-bottom: 16px;
    padding-bottom: 8px;
  }

  .box-main {
    border-bottom: 1px dashed var(--el-color-black);
    padding-bottom: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
  }

  .box-main-border {
    border-bottom: none;
  }

  .bold-bottom {
    border-bottom: 1px dashed var(--el-color-black);
  }

  .bold-bottom-24 {
    border-bottom: 1px dashed var(--el-color-black);
    margin-bottom: 24px;
    padding-bottom: 24px;
  }

  .box-main:nth-last-child(1) {
    margin-bottom: 0;
    border-bottom: none;
  }

  .box-text-box {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 24px;
  }

  .text-box {
    font-size: 14px;
    color: var(--el-color-black);
    flex: 1;
  }

  .flex-end {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .text-box-r-p2 {
    text-align: right;
    display: flex;
    justify-content: flex-end;
    align-items: flex-end;
  }

  .font-bold {
    font-size: 15px;
    font-weight: 500;
  }

  .font-500 {
    font-weight: 500;
  }

  .font-700 {
    font-size: 15px;
    font-weight: 700;
  }

  .font16Bold {
    font-size: 16px;
    font-weight: 700;
  }

  .text-center {
    text-align: center;
  }

  .flexWidth {
    flex: 4;
  }

  .flexWidthRight {
    flex: 4;
  }

  .font-big {
    font-size: 18px;
    font-weight: 500;
  }
  .line-height-2-5 {
    line-height: 2.5;
  }

  .show-sku-attr {
    opacity: 0;
    height: 0;
  }

  .is-show-sku-attr {
    opacity: 0;
  }

  .font24 {
    font-size: 24px;
    font-weight: 500;
  }
  :deep(.font24-400) {
    font-size: 24px;
    font-weight: 400;
  }
  .font22 {
    font-size: 22px;
    font-weight: 500;
  }
  .font18Small {
    font-size: 18px;
    font-weight: 400;
  }
  .font16Normal {
    font-size: 16px;
    font-weight: 500;
  }
  .font16Small {
    font-size: 16px;
    font-weight: 400;
  }
  .font17Normal {
    font-size: 17px;
    font-weight: 500;
  }
  .font18 {
    font-size: 18px;
    font-weight: 700;
  }
  .code-box {
    padding-bottom: 12px;
  }
  .code-box p {
    text-align: center;
    font-size: 16px;
    color: var(--el-color-black);
    font-weight: 700;
  }
  .code {
    width: 140px;
    height: 140px;
    margin: auto;
  }
  .barcode {
    margin: auto;
  }
  .brand-box {
    text-align: center;
    font-size: 16px;
    color: var(--el-color-black);
    font-weight: 700;
    padding-bottom: 12px;
  }
  .border-top {
    border-top: 1px dashed var(--el-color-black);
    padding-top: 8px;
  }
  .border-top-12 {
    border-top: 1px dashed var(--el-color-black);
    padding-top: 12px;
    margin-top: -12px;
  }
  .border-top-none {
    border-top: none;
    padding-bottom: 8px;
  }
  * {
    overflow-wrap: anywhere;
    white-space: pre-line;
  }
  .logo {
    width: 80px;
    height: 80px;
    margin: 12px auto 12px;
  }
  .mb-4 {
    margin-bottom: 4px !important;
  }
  .mb-8 {
    margin-bottom: 8px !important;
  }
  .mb-12 {
    margin-bottom: 12px !important;
  }
  .mb-14 {
    margin-bottom: 16px !important;
  }

  .mb-16 {
    margin-bottom: 16px !important;
  }

  .mb-24 {
    margin-bottom: 24px !important;
  }
  .mt-n18 {
    margin-top: -18px !important;
  }
</style>
