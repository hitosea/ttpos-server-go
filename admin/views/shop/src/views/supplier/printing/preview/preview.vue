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
              v-if="items.name"
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
              {{ items.name }}
            </div>
            <!-- 右边的字段 -->
            <div v-if="items.label" class="text-box text-box-r" :class="items.num ? 'flex-end' : ''">
              <p
                class="text-box-r-p1"
                :class="[
                  items.bold ? 'font-bold' : '',
                  items.big ? 'font-big' : '',
                  items.flexWidth ? 'flexWidth' : '',
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
                  items.flexWidth ? 'flexWidth' : '',
                  items.font18 ? 'font18' : '',
                  items.font500 ? 'font-500' : '',
                  items.font700 ? 'font-700' : '',
                  items.textCenter ? 'text-center' : '',
                  items.font16Bold ? 'font16Bold' : '',
                ]"
              >
                {{ items.label }}
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
          <p class="font24"> {{ $t('桌位: A01 (4人)') }} </p>
          <p class="mb-14"> <span class="span1">2024/05/04 14:15:12</span> </p>
          <p>
            <span class="span3">{{ $t('商品名称商品名称') }}</span> <span class="span3">X1</span>
          </p>
          <p class="mb-8">
            <span class="span3">{{ $t('规格名称') }}</span>
          </p>
          <p>
            <span class="span4">{{ $t('少冰') }}</span>
          </p>
          <p>
            <span class="span4">{{ $t('加珍珠') }}</span>
          </p>
          <p class="mb-8">
            <span class="span4">{{ $t('这是备注这是备注这是备注') }}</span>
          </p>
        </template>
        <template v-if="mode == 2">
          <h3>
            {{ $t('桌位: A01 (4人)') }}
          </h3>
          <p>
            <span class="span3">{{ $t('商品名称商品名称') }}</span> <span class="span3">X1</span>
          </p>
          <p class="mb-8">
            <span class="span3">{{ $t('规格名称') }}</span>
          </p>
          <p>
            <span class="span4">{{ $t('少冰') }}</span>
          </p>
          <p>
            <span class="span4">{{ $t('加珍珠') }}</span>
          </p>
          <p class="mb-8">
            <span class="span4">{{ $t('这是备注这是备注这是备注') }}</span>
          </p>
          <h2 class="border-top">2024/05/04 14:15:12</h2>
        </template>
      </div>
    </template>

    <template v-else>
      <div class="tabs-box" v-if="title == $t('结账单') || title == $t('预结账单') || title == $t('交班单') || title == $t('整单打印')">
        <div @click="modeChange(1)" class="tabs-button" :class="mode == 1 ? 'tabs-active' : ''">
          {{ $t('模板1') }}
        </div>
        <div @click="modeChange(2)" class="tabs-button" :class="mode == 2 ? 'tabs-active' : ''">
          {{ $t('模板2') }}
        </div>
        <div @click="modeChange(3)" v-if="title != $t('整单打印')" class="tabs-button" :class="mode == 3 ? 'tabs-active' : ''">
          {{ $t('模板3') }}
        </div>
      </div>

      <div class="box-border">
        <p class="title-name" v-if="(title == $t('结账单') || title == $t('预结账单')) && mode != 3">
          {{ title }}
        </p>
        <h2 class="font24" :class="mode != 1 ? 'mb-8' : 'mb-24'" v-if="mode != 3 && storeShow && (title == $t('结账单') || title == $t('预结账单'))">
          {{ $t('店铺名称') }}
        </h2>

        <template v-if="mode == 3 && storeShow && (title == $t('结账单') || title == $t('预结账单'))">
          <h2>
            {{ $t('店铺名称') }}
          </h2>
          <img v-if="mode == 3" :src="userInfo.logoUrl" alt="" class="logo" />
          <h2 class="font24" :class="mode != 1 ? 'mb-8' : 'mb-24'">
            {{ title }}
          </h2>
        </template>

        <template v-if="title == $t('充值单')">
          <h4>
            {{ $t('店铺名称') }}
          </h4>
        </template>

        <p
          v-if="(title == $t('结账单') && mode == 3) || (title == $t('预结账单') && mode == 3)"
          class="Invoice-p font14"
          v-html="$t('公司：') + $t('公司名称公司名称公司名称')"
        ></p>
        <p
          v-if="(title == $t('结账单') && mode == 3) || (title == $t('预结账单') && mode == 3)"
          class="Invoice-p font14"
          v-html="$t('商家地址：') + $t('商家地址商家地址商家地址商家地址商家地址商家地址商家地址商家地址')"
        ></p>
        <p v-if="(title == $t('结账单') && mode == 3) || (title == $t('预结账单') && mode == 3)" class="Invoice-p font14" v-html="$t('电话：') + '02-15-1441414'"></p>
        <p
          v-if="(title == $t('结账单') && mode == 3) || (title == $t('预结账单') && mode == 3)"
          class="Invoice-p font14"
          :class="mode == 3 ? 'bold-bottom-24' : ''"
          v-html="$t('税号：') + '252452524144'"
        >
        </p>
        <p v-if="(title == $t('结账单') && mode == 2) || (title == $t('预结账单') && mode == 2)" class="Invoice-p" v-html="$t('非常感谢您今天的到来，我们期待您的再次光临')"></p>
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
            {{ $t('桌位: A01 (4人)') }}
          </h4>
          <h3 v-if="mode == 2">
            {{ $t('桌位: A01 (4人)') }}
          </h3>
        </template>
        <template v-if="title == $t('退菜单')">
          <h4 class="mb-24" v-if="mode == 1"> {{ $t('退菜单') }}</h4>
          <h4 class="Invoice-h4 mb-8" v-if="mode == 1">
            {{ $t('桌位: A01 (4人)') }}
          </h4>
        </template>
        <!-- 小字的数据 -->
        <div
          class="box-main"
          :class="[
            (index == 0 && title != $t('充值单')) || (title == $t('预结账单') && index == 3 && mode == 2) || (title == $t('预结账单') && index == 4 && mode == 3)
              ? 'box-main-border'
              : '',
            title == $t('交班单') || title == $t('营业数据') ? 'bold-bottom' : '',
          ]"
          v-for="(item, index) in details"
          :key="index"
          v-show="(item[0]?.allHide && item[0].allHide.includes(mode)) || !item[0]?.allHide"
        >
          <div class="box-text-box" v-for="(items, indexs) in item" :key="indexs" v-show="(items.typeShow && items.typeShow.includes(mode)) || !items.typeShow">
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
              ]"
              v-html="items.name"
            >
            </div>
            <!-- 右边的字段 -->
            <div v-if="items.right != false && mode != items.hide" class="text-box text-box-r" :class="items.num ? 'flex-end' : ''">
              <p
                class="text-box-r-p1"
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
        <div class="brand-box" :class="title == $t('预结账单') ? 'border-top' : ''" v-if="title == $t('结账单') || title == $t('预结账单')"
          >{{ $t('感谢您的光临！本店由') }}{{ brand }}{{ $t('系统提供支持') }}</div
        >
      </div>
    </template>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose" :loading="loading"> {{ $t('关闭') }}</el-button>
        <el-button
          v-if="title == $t('结账单') || title == $t('预结账单') || title == $t('发票') || title == $t('交班单') || title == $t('整单打印') || title == $t('一菜一单')"
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
  const cloudBasic = languageStore().getCloudBasic().cloudBasic;
  const { userInfo } = useUserStore();
  export default {
    data() {
      return {
        userInfo,
        cloudBasic: cloudBasic,
        brand: '',
        dialogWidth: '',
        loading: false,
        dialogVisible: false,
        storeShow: true,
        details: [],
        titleName: '',
        mode: 1,
        one: [
          [
            {
              name: $t('当班编号'),
              label: 202401253695842521,
            },
            {
              name: $t('交班人'),
              label: $t('张三'),
            },
            {
              name: $t('当班时间'),
              label: '2023/12/15 14:00:21' + $t('至') + '2023/12/15 14:00:21',
            },
          ],
          [
            {
              name: $t('总销售额'),
              label: '￥150,125.00',
            },
            {
              name: $t('原商品金额'),
              label: '￥500.00',
            },
            {
              name: $t('支付手续费'),
              label: '￥300.00',
            },
            {
              name: $t('服务费'),
              label: '￥500.00',
            },
            {
              name: $t('税费'),
              label: '￥50.00',
            },
            {
              name: $t('商品数量'),
              label: '128',
            },
            {
              name: $t('优惠折扣'),
              label: '￥839.00',
            },
            {
              name: $t('会员折扣'),
              label: '￥597.00',
            },
            {
              name: $t('退款金额'),
              label: '￥50',
            },
            {
              name: $t('免单金额'),
              label: '￥50',
            },
          ],
          [
            {
              name: $t('实收金额'),
              label: '￥150,125.00',
              big: true,
            },
            {
              name: $t('VAT（10%）'),
              label: '',
              font700: true,
            },
            {
              name: $t('合计'),
              label: '￥68,181.00',
            },
            {
              name: '',
              label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
            {
              name: $t('VAT（8%）'),
              label: '',
              font700: true,
            },
            {
              name: $t('合计'),
              label: '￥68,295.45',
            },
            {
              name: '',
              label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
          ],
          [
            {
              name: $t('合计'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('所有订单数'),
              label: '52',
            },
            {
              name: $t('桌数'),
              label: '52',
            },
            {
              name: $t('人数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
            {
              name: $t('桌台方式'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('订单数（桌数）'),
              label: '52',
            },
            {
              name: $t('人数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
            {
              name: $t('人均'),
              label: '￥689.12',
            },
            {
              name: $t('点餐方式'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('订单数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
          ],
          [
            {
              name: $t('支付方式'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: $t('现金'),
              num: '68',
              label: '￥520.00',
            },
            {
              name: 'Krungsri Mobile',
              num: '50',
              label: '￥520.00',
            },
            {
              name: 'Cross-Border QR',
              num: '55',
              label: '￥520.00',
            },
            {
              name: 'TrueMoney',
              num: '35',
              label: '￥520.00',
            },
            {
              name: 'LINE Pay',
              num: '47',
              label: '￥520.00',
            },
            {
              name: 'Alipay',
              num: '32',
              label: '￥520.00',
            },
            {
              name: 'WeChat Pay',
              num: '28',
              label: '￥520.00',
            },
            {
              name: $t('总金额'),
              label: '￥150,1.00',
            },
          ],
          [
            {
              name: $t('高峰时间'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: '5/31 17:00-18:00',
              num: '68',
              label: '￥520.00',
            },
            {
              name: '5/31 17:00-18:00',
              num: '50',
              label: '￥520.00',
            },
          ],
          [
            {
              name: $t('分类'),
              num: $t('数量'),
              label: $t('小计'),
              font500: true,
            },
            {
              name: $t('肉类'),
              num: '68',
              label: '￥50.00',
            },
            {
              name: $t('蔬菜类'),
              num: '50',
              label: '￥50.00',
            },
            {
              name: $t('小计'),
              num: '50',
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('本班实收金额'),
              label: '￥150,125.00',
            },
            {
              name: $t('上一班遗留备用金'),
              label: '￥250.00',
            },
            {
              name: $t('中途存入现金'),
              label: '￥50.00',
            },
            {
              name: $t('中途取出现金'),
              label: '￥50.00',
            },
            {
              name: $t('本班取出现金'),
              label: '￥387.00',
            },
            {
              name: $t('本班遗留备用金'),
              label: '￥672.00',
            },
          ],
        ],
        two: [
          [
            {
              name: $t('桌位: A01 (4人)'),
              label: '',
              hide: 2,
              font24: true,
            },
            {
              name: $t('订单号'),
              label: 202401253695842521,
            },
            {
              name: $t('收银员'),
              label: $t('张三'),
            },
            {
              name: $t('时间'),
              label: '2023/12/15 14:00:21',
              hide: 2,
            },
          ],
          [
            {
              name: $t('商品'),
              num: $t('单价 | 数量'),
              label: $t('小计'),
              bold: true,
              typeShow: '1,2',
            },
          ],
          [
            {
              name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
              num: '24*12',
              label: '￥350',
            },

            {
              name: $t('商品名称商品名称'),
              num: '55*2',
              label: '￥50',
            },
          ],
          [
            {
              name: '',
              label: $t('商品数量') + $t('：') + '14',
              typeShow: '1,2',
            },
            {
              name: '',
              label: $t('商品金额') + $t('：') + '￥410',
              typeShow: '1,2',
            },
            {
              name: $t('商品数量') + $t('：') + '14',
              label: $t('商品金额') + $t('：') + '￥410',
              typeShow: '3',
            },
            {
              name: '',
              label: $t('服务费') + $t('：') + '￥50',
            },
            {
              name: '',
              label: $t('VAT（10%）') + $t('：') + '￥50',
              typeShow: '2',
            },
            {
              name: '',
              label: $t('VAT（8%）') + $t('：') + '￥50',
              typeShow: '2',
            },
            {
              name: '',
              label: $t('优惠折扣') + $t('：') + '￥50',
              typeShow: '1,2',
            },
            {
              name: '',
              label: $t('优惠折扣') + $t('：') + '￥50（4.28% OFF）',
              typeShow: '3',
              left: false,
            },
            {
              name: '',
              label: $t('会员优惠') + $t('：') + '￥3',
              typeShow: '1,2,3',
            },
            {
              name: '',
              label: $t('会员折扣') + $t('：') + $t('3.8折'),
              typeShow: '3',
            },
            {
              name: '',
              label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
              typeShow: '3',
            },
            {
              name: '',
              label: $t('免单') + $t('：') + '￥84.89',
              typeShow: '2',
            },
            {
              name: '',
              label: $t('退款金额') + $t('：') + '￥3',
              typeShow: '1,3',
            },
            {
              name: '',
              label: $t('支付手续费') + $t('：') + '￥3',
              typeShow: '1,3',
            },
            {
              name: '',
              label: $t('手动抹零') + $t('：') + '￥1.15',
              typeShow: '1,2,3',
            },
            {
              typeShow: '1',
              name: $t('合计应收'),
              label: '￥410',
              big: true,
            },
            {
              typeShow: '2',
              name: $t('合计应收'),
              label: '￥0',
              big: true,
            },
          ],
          [
            {
              allHide: '3',
              name: $t('合计应收'),
              label: '￥410',
              big: true,
            },
          ],
          [
            {
              name: '',
              label: $t('合计(其中VAT)'),
              allHide: '1,3',
            },
            {
              name: $t('VAT（10%）'),
              label: '100.00 (9.09)',
              typeShow: '1,3',
            },
            {
              name: $t('VAT（8%）'),
              label: '98.90 (7.36)',
              typeShow: '1,3',
            },
            {
              name: $t('VAT（10%）'),
              label: '￥100.00 (￥9.09)',
              typeShow: '2',
            },
            {
              name: $t('VAT（8%）'),
              label: '￥98.90 (￥7.36)',
              typeShow: '2',
            },
          ],
          [
            {
              name: $t('支付方式'),
              label: $t('现金'),
              typeShow: '1',
            },
            {
              name: $t('实收金额'),
              label: '￥31',
              typeShow: '1',
            },
            {
              name: $t('找零'),
              label: '0.02',
              typeShow: '1',
            },
            {
              name: $t('支付方式'),
              label: 'WeChat',
              typeShow: '1',
            },
            {
              name: $t('实收金额'),
              label: '￥40',
              typeShow: '1',
            },
            {
              name: $t('支付方式'),
              label: $t('免单'),
              typeShow: '2',
            },
            {
              name: $t('实收金额'),
              label: '￥0',
              typeShow: '2',
            },
            {
              name: $t('支付方式'),
              label: $t('微信支付'),
              typeShow: '3',
            },
            {
              name: $t('实收金额'),
              label: '￥410',
              typeShow: '3',
            },
            {
              name: $t('找零'),
              label: '0.1',
              typeShow: '3',
            },
          ],
          [
            {
              name: $t('会员剩余余额'),
              label: '￥100',
            },
            {
              name: $t('本次积分'),
              label: '410',
              typeShow: '1,3',
            },
            {
              name: $t('本次积分'),
              label: '0',
              typeShow: '2',
            },
          ],
        ],
        three: [
          [
            {
              name: $t('桌位: A01 (4人)'),
              label: '',
              hide: 2,
              font24: true,
            },
            {
              name: $t('订单号'),
              label: 202401253695842521,
              typeShow: '1,2',
            },
            {
              name: $t('收银员'),
              label: $t('张三'),
              typeShow: '1,2',
            },
            {
              name: $t('收银员：') + $t('张三'),
              label: '',
              typeShow: '3',
            },
            {
              name: $t('订单号：') + 202401253695842521,
              label: '',
              typeShow: '3',
              right: false,
            },
          ],
          [
            {
              name: $t('商品'),
              num: $t('单价 | 数量'),
              label: $t('小计'),
              bold: true,
              typeShow: '1,2',
            },
          ],
          [
            {
              name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
              num: '24*12',
              label: '￥50',
            },
            {
              name: $t('商品名称商品名称'),
              num: '55*2',
              label: '￥50',
            },
          ],
          [
            {
              name: '',
              label: $t('商品数量') + $t('：') + '14',
              typeShow: '1,2',
            },
            {
              name: '',
              label: $t('商品金额') + $t('：') + '￥410',
              typeShow: '1,2',
            },
            {
              name: $t('商品数量') + $t('：') + '14',
              label: $t('商品金额') + $t('：') + '￥410',
              typeShow: '3',
            },
            {
              name: '',
              label: $t('服务费') + $t('：') + '￥50',
            },
            {
              name: '',
              label: $t('VAT（10%）') + $t('：') + '￥50',
              typeShow: '2,3',
            },
            {
              name: '',
              label: $t('VAT（8%）') + $t('：') + '￥50',
              typeShow: '2,3',
            },
            {
              name: '',
              label: $t('优惠折扣') + $t('：') + '￥50',
              typeShow: '1,2',
            },
            {
              name: '',
              label: $t('优惠折扣') + $t('：') + '￥6.99（9.9% OFF）',
              typeShow: '3',
              left: false,
            },
            {
              name: '',
              label: $t('会员优惠') + $t('：') + '￥3',
              typeShow: '1,2,3',
            },
            {
              name: '',
              label: $t('会员折扣') + $t('：') + $t('3.8折'),
              typeShow: '3',
            },
            {
              name: '',
              label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
              typeShow: '3',
            },
            {
              name: '',
              label: $t('手动抹零') + $t('：') + '￥1.15',
              typeShow: '1,2,3',
            },
            {
              typeShow: '1,2',
              name: $t('合计应收'),
              label: '￥410',
              big: true,
            },
          ],
          [
            {
              allHide: '3',
              name: $t('合计应收'),
              label: '￥410',
              big: true,
            },
          ],
          [
            {
              name: '',
              label: $t('合计(其中VAT)'),
              allHide: '1',
            },
            {
              name: $t('VAT（10%）'),
              label: '100.00 (9.09)',
              typeShow: '1',
            },
            {
              name: $t('VAT（8%）'),
              label: '98.90 (7.36)',
              typeShow: '1',
            },
            {
              name: $t('VAT（10%）'),
              label: '￥100.00 (￥9.09)',
              typeShow: '2,3',
            },
            {
              name: $t('VAT（8%）'),
              label: '￥98.90 (￥7.36)',
              typeShow: '2,3',
            },
          ],
        ],
        four: [],
        five: [
          [
            {
              name: $t('时间'),
              label: '2023/12/15 14:00:21' + $t('至') + '2023/12/15 14:00:21',
            },
            {
              name: $t('总销售额'),
              label: '￥150,125.00',
            },
            {
              name: $t('原商品金额'),
              label: '￥500.00',
            },
            {
              name: $t('服务费'),
              label: '￥500.00',
            },
            {
              name: $t('支付手续费'),
              label: '￥300.00',
            },
            {
              name: $t('税费'),
              label: '￥500.00',
            },

            {
              name: $t('商品数量'),
              label: '128',
            },
            {
              name: $t('优惠折扣'),
              label: '￥839.00',
            },
            {
              name: $t('会员折扣'),
              label: '￥597.00',
            },
            {
              name: $t('退款金额'),
              label: '￥50',
            },
            {
              name: $t('免单金额'),
              label: '￥500',
            },
            {
              name: $t('实收金额'),
              label: '￥150,125.00',
              big: true,
            },
          ],
          [
            {
              name: $t('VAT（10%）'),
              label: '',
              font700: true,
            },
            {
              name: $t('合计'),
              label: '￥68,181.00',
            },
            {
              name: '',
              label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
            {
              name: $t('VAT（8%）'),
              label: '',
              font700: true,
            },
            {
              name: $t('合计'),
              label: '￥68,295.45',
            },
            {
              name: '',
              label: '(' + $t('其中VAT') + '￥1,524.00)',
            },
          ],
          [
            {
              name: $t('会员数据'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('充值金额'),
              label: '￥500',
            },
            {
              name: $t('赠送金额'),
              label: '￥100',
            },
            {
              name: $t('赠送积分'),
              label: '100',
            },
          ],
          [
            {
              name: $t('未结账数据'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('订单数'),
              label: '52',
            },
            {
              name: $t('金额'),
              label: '￥34',
            },
          ],
          [
            {
              name: $t('合计'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('所有订单数'),
              label: '52',
            },
            {
              name: $t('桌数'),
              label: '52',
            },
            {
              name: $t('人数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
            {
              name: $t('桌台方式'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('订单数（桌数）'),
              label: '52',
            },
            {
              name: $t('人数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
            {
              name: $t('人均'),
              label: '￥689.12',
            },
            {
              name: $t('点餐方式'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('订单数'),
              label: '52',
            },
            {
              name: $t('最小/大订单金额'),
              label: '￥28.15 / ￥5246.12',
            },
            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
          ],
          [
            {
              name: $t('支付方式'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: $t('现金'),
              num: '68',
              label: '￥520.00',
            },
            {
              name: 'Krungsri Mobile',
              num: '50',
              label: '￥520.00',
            },
            {
              name: 'Cross-Border QR',
              num: '55',
              label: '￥520.00',
            },
            {
              name: 'TrueMoney',
              num: '35',
              label: '￥520.00',
            },
            {
              name: 'LINE Pay',
              num: '47',
              label: '￥520.00',
            },
            {
              name: 'Alipay',
              num: '32',
              label: '￥520.00',
            },
            {
              name: 'WeChat Pay',
              num: '28',
              label: '￥520.00',
            },
            {
              name: $t('总金额'),
              label: '￥520.00',
            },
          ],
          [
            {
              name: $t('高峰时间'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: '5/31 17:00-18:00',
              num: '68',
              label: '￥520.00',
            },
            {
              name: '5/31 17:00-18:00',
              num: '50',
              label: '￥520.00',
            },
          ],
        ],
        six: [
          [
            {
              name: $t('订单号'),
              label: 202401253695842521,
              font16Small: true,
            },
            {
              name: $t('时间'),
              label: '2023/12/15 14:00:21',
              font16Small: true,
            },
          ],
          [
            {
              name: $t('商品'),
              label: $t('数量'),
              font17Normal: true,
            },
          ],
          [
            {
              name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
              label: 'X2',
              typeShow: '1',
              flexWidth: true,
              font16Normal: true,
            },
            {
              name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + `<span class="font24-400">${$t('这是备注这是备注这是备注')}</span>`,
              typeShow: '1',
              flexWidth: true,
            },
            {
              name: $t('商品名称商品名称'),
              label: 'X2',
              typeShow: '1',
              flexWidth: true,
              font16Normal: true,
            },
            {
              name: $t('商品名称商品名称品名称商品名称商品名称品名称'),
              label: 'X2',
              font24: true,
              typeShow: '2',
              flexWidth: true,
            },
            {
              name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + `<span class="font24-400">${$t('这是备注这是备注这是备注')}</span>`,
              typeShow: '2',
              flexWidth: true,
              font18Small: true,
            },
            {
              name: $t('商品名称商品名称'),
              label: 'X2',
              font24: true,
              typeShow: '2',
              flexWidth: true,
            },
          ],
        ],
        seven: [
          [
            {
              name: $t('合计金额'),
              label: '¥4502',
              big: true,
            },
            {
              name: $t('其中服务费'),
              label: '￥20.00)',
            },
            {
              name: $t('其中VAT'),
              label: '￥200.00)',
            },
          ],
          [
            {
              name: $t('仅作为餐饮费收取以上金额'),
              label: '',
            },
            {
              name: '',
              label: $t('合计 (其中VAT)'),
            },
            {
              name: $t('VAT（10%）'),
              label: '4,000.00 (363.65)',
            },
            {
              name: $t('VAT（8%）'),
              label: '501.00 (45.55)',
            },
            {
              name: $t('不包含退款金額￥3.00'),
              label: '',
            },
          ],
          [
            {
              name: $t('现金'),
              label: '¥4502',
              big: true,
            },
          ],
          [
            {
              name: $t('发票信息'),
              label: '',
            },
            {
              name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
              label: '',
            },
            {
              name: $t('公司地址：') + $t('公司地址公司地址公司地址公司地址'),
              label: '',
            },
            {
              name: $t('税号：') + '252452524144',
              label: '',
            },
            {
              name: $t('联系电话：') + '02-15-1441414',
              label: '',
            },
          ],
          [
            {
              name: $t('收银员：张三'),
              label: '',
            },
            {
              name: $t('订单号：') + 'NO.252452524144',
              label: '',
            },
            {
              name: $t('打印次数：1'),
              label: '',
            },
            {
              name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
              label: '',
            },
            {
              name: $t('商家地址：') + $t('商家地址商家地址商家地址商家地址商家地址商家地址商家地址商家地址'),
              label: '',
            },
            {
              name: $t('税号：') + '252452524144',
              label: '',
            },
            {
              name: $t('电话：') + '02-15-1441414',
              label: '',
            },
            {
              name: ' ',
              label: ' ',
            },

            {
              name: $t('*保管注意事項'),
              label: '',
            },
            {
              name: $t('如需保管时请将印刷页面朝内摺叠'),
              label: '',
            },
          ],
        ],
        eight: [
          [
            {
              name: $t('桌位: A01 (4人)'),
              label: '',
              font24: true,
            },
            {
              name: $t('订单号'),
              label: 202401253695842521,
            },
            {
              name: $t('收银员'),
              label: $t('张三'),
            },
            {
              name: $t('时间'),
              label: '2023/12/15 14:00:21',
            },
          ],

          [
            {
              name: $t('（赠）') + $t('商品名称商品名称品名称商品名称商品名称品名称') + `(${$t('少冰、')}${$t('加珍珠')})`,
              num: '24*12',
              label: '￥350',
            },

            {
              name: $t('商品名称商品名称'),
              num: '55*2',
              label: '￥50',
            },
          ],
          [
            {
              name: $t('商品数量') + $t('：') + '14',
              label: $t('商品金额') + $t('：') + '￥410',
            },
            {
              name: '',
              label: $t('服务费') + $t('：') + '￥50',
            },

            {
              name: '',
              label: $t('优惠折扣') + $t('：') + '￥6.99（9.9% OFF）',
            },
            {
              name: '',
              label: $t('会员优惠') + $t('：') + '￥3',
            },
            {
              name: '',
              label: $t('会员折扣') + $t('：') + $t('3.8折'),
            },
            {
              name: '',
              label: $t('会员卡折扣') + $t('：') + $t('3.8折'),
            },
            {
              name: '',
              label: $t('免单') + $t('：') + '￥73.98',
            },
            {
              name: '',
              label: $t('手动抹零') + $t('：') + '￥1.15',
            },
          ],
          [
            {
              name: $t('合计应收'),
              label: '￥410',
              big: true,
            },
          ],
          [
            {
              name: '',
              label: $t('合计(其中VAT)'),
            },
            {
              name: $t('VAT（10%）'),
              label: '100.00 (9.09)',
            },
            {
              name: $t('VAT（8%）'),
              label: '98.90 (7.36)',
            },
          ],
          [
            {
              name: $t('支付方式'),
              label: $t('免单'),
            },
            {
              name: $t('实收金额'),
              label: '￥0',
            },
          ],
          [
            {
              name: $t('会员剩余余额'),
              label: '￥100',
            },
            {
              name: $t('本次积分'),
              label: '410',
            },
          ],
          [
            {
              name: $t('发票信息'),
              label: '',
            },
            {
              name: $t('公司名称：') + $t('公司名称公司名称公司名称'),
              label: '',
            },
            {
              name: $t('公司地址：') + $t('公司地址公司地址公司地址公司地址'),
              label: '',
            },
            {
              name: $t('税号：') + '252452524144',
              label: '',
            },
            {
              name: $t('联系电话：') + '02-15-1441414',
              label: '',
            },
          ],
        ],
        nine: [
          [
            {
              name: $t('当班编号'),
              label: 202401253695842521,
            },
            {
              name: $t('交班人'),
              label: $t('张三'),
            },
            {
              name: $t('总销售额'),
              label: '￥150,125.00',
            },
            {
              name: $t('实收金额'),
              label: '￥150,125.00',
            },
          ],
          [
            {
              name: $t('支付方式'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: $t('现金'),
              num: '68',
              label: '￥520.00',
            },
            {
              name: 'Krungsri Mobile',
              num: '50',
              label: '￥520.00',
            },
            {
              name: 'Cross-Border QR',
              num: '55',
              label: '￥520.00',
            },
            {
              name: 'TrueMoney',
              num: '35',
              label: '￥520.00',
            },
            {
              name: 'LINE Pay',
              num: '47',
              label: '￥520.00',
            },
            {
              name: 'Alipay',
              num: '32',
              label: '￥520.00',
            },
            {
              name: 'WeChat Pay',
              num: '28',
              label: '￥520.00',
            },
            {
              name: $t('总金额'),
              label: '￥150,1.00',
            },
          ],
          [
            {
              name: $t('原商品金额'),
              label: '￥500.00',
            },
            {
              name: $t('支付手续费'),
              label: '￥300.00',
            },
            {
              name: $t('服务费'),
              label: '￥500.00',
            },
            {
              name: $t('税费'),
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('优惠折扣'),
              label: '￥839.00',
            },
            {
              name: $t('会员折扣'),
              label: '￥597.00',
            },
            {
              name: $t('免单金额'),
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('退款金额'),
              label: '￥50',
            },
          ],
          [
            {
              name: $t('会员数据'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('充值金额'),
              label: '￥500',
            },
            {
              name: $t('赠送金额'),
              label: '￥100',
            },
            {
              name: $t('赠送积分'),
              label: '100',
            },
          ],
          [
            {
              name: $t('所有订单数'),
              label: '52',
            },

            {
              name: $t('人数'),
              label: '52',
            },

            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
          ],
          [
            {
              name: $t('高峰时间'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: '5/31 17:00-18:00',
              num: '68',
              label: '￥520.00',
            },
            {
              name: '5/31 17:00-18:00',
              num: '50',
              label: '￥520.00',
            },
          ],
          [
            {
              name: $t('分类'),
              num: $t('数量'),
              label: $t('小计'),
              font500: true,
            },
            {
              name: $t('肉类'),
              num: '68',
              label: '￥50.00',
            },
            {
              name: $t('蔬菜类'),
              num: '50',
              label: '￥50.00',
            },
            {
              name: $t('小计'),
              num: '50',
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('上一班遗留备用金'),
              label: '￥250.00',
            },
            {
              name: $t('中途存入现金'),
              label: '￥50.00',
            },
            {
              name: $t('中途取出现金'),
              label: '￥50.00',
            },
            {
              name: $t('本班取出现金'),
              label: '￥387.00',
            },
            {
              name: $t('本班遗留备用金'),
              label: '￥672.00',
            },
          ],
        ],
        ten: [
          [
            {
              name: $t('收银员'),
              label: $t('张三'),
            },
            {
              name: $t('时间'),
              label: '2023-12-15 14:45:21',
            },

            {
              name: $t('充值前'),
              label: '￥50',
            },
            {
              name: $t('本次充值'),
              label: '￥500',
            },
            {
              name: $t('赠送金额'),
              label: '￥50',
            },
            {
              name: $t('赠送积分'),
              label: '500',
            },
            {
              name: $t('充值后'),
              label: '￥550',
            },
          ],
          [
            {
              name: $t('退款'),
              label: '￥1500',
            },
          ],
          [
            {
              name: $t('合计应收'),
              label: '￥500',
              big: true,
            },
          ],
          [
            {
              name: $t('支付方式'),
              label: $t('现金'),
            },
            {
              name: $t('实收金额'),
              label: '￥100',
            },
            {
              name: $t('找零'),
              label: '5',
            },
            {
              name: $t('支付方式'),
              label: 'Wechat',
            },
            {
              name: $t('实收金额'),
              label: '￥400',
            },
          ],
        ],
        eleven: [
          [
            {
              name: $t('订单号'),
              label: 202401253695842521,
              font16Small: true,
            },
            {
              name: $t('时间'),
              label: '2023/12/15 14:00:21',
              font16Small: true,
            },
          ],
          [
            {
              name: $t('商品'),
              label: $t('数量'),
              font17Normal: true,
            },
          ],
          [
            {
              name: `(${$t('退')})` + $t('商品名称商品名称品名称商品名称商品名称品名称'),
              label: 'X5',
              font22: true,
              flexWidth: true,
            },
            {
              name: `${$t('少冰')}` + '\n' + `${$t('加珍珠')}` + '\n' + $t('这是备注这是备注这是备注'),
              font18Small: true,
              flexWidth: true,
            },
            {
              name: `(${$t('退')})` + $t('商品名称商品名称'),
              label: 'X2',
              font22: true,
              flexWidth: true,
            },
          ],
          [
            {
              name: $t('退菜原因：等待时间长，口味不好'),
              label: '',
              right: false,
              font18Small: true,
              flexWidth: 3,
            },
          ],
        ],
        twelve: [
          [
            {
              name: $t('当班编号'),
              label: 202401253695842521,
            },
            {
              name: $t('交班人'),
              label: $t('张三'),
            },
            {
              name: $t('总销售额'),
              label: '￥150,125.00',
            },
            {
              name: $t('实收金额'),
              label: '￥150,125.00',
            },
          ],
          [
            {
              name: $t('支付方式'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: $t('现金'),
              num: '68',
              label: '￥520.00',
            },
            {
              name: 'Krungsri Mobile',
              num: '50',
              label: '￥520.00',
            },
            {
              name: 'Cross-Border QR',
              num: '55',
              label: '￥520.00',
            },
            {
              name: 'TrueMoney',
              num: '35',
              label: '￥520.00',
            },
            {
              name: 'LINE Pay',
              num: '47',
              label: '￥520.00',
            },
            {
              name: 'Alipay',
              num: '32',
              label: '￥520.00',
            },
            {
              name: 'WeChat Pay',
              num: '28',
              label: '￥520.00',
            },
            {
              name: $t('总金额'),
              label: '￥150,1.00',
            },
          ],
          [
            {
              name: $t('原商品金额'),
              label: '￥500.00',
            },
            {
              name: $t('支付手续费'),
              label: '￥300.00',
            },
            {
              name: $t('服务费'),
              label: '￥500.00',
            },
            {
              name: $t('税费'),
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('优惠折扣'),
              label: '￥839.00',
            },
            {
              name: $t('会员折扣'),
              label: '￥597.00',
            },
            {
              name: $t('免单金额'),
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('退款金额'),
              label: '￥50',
            },
          ],
          [
            {
              name: $t('退菜次数'),
              label: '5',
            },
            {
              name: $t('退款次数'),
              label: '5',
            },
            {
              name: $t('反结账次数'),
              label: '5',
            },
            {
              name: $t('赠菜次数'),
              label: '5',
            },
            {
              name: $t('免单次数'),
              label: '5',
            },
            {
              name: $t('转菜次数'),
              label: '5',
            },
            {
              name: $t('单品改价次数'),
              label: '5',
            },
            {
              name: $t('整单改价次数'),
              label: '5',
            },
            {
              name: $t('整单折扣次数'),
              label: '5',
            },
            {
              name: $t('整单抹零次数'),
              label: '5',
            },
          ],
          [
            {
              name: $t('会员数据'),
              label: '',
              font700: true,
              textCenter: true,
              right: false,
            },
            {
              name: $t('充值金额'),
              label: '￥500',
            },
            {
              name: $t('赠送金额'),
              label: '￥100',
            },
            {
              name: $t('赠送积分'),
              label: '100',
            },
          ],
          [
            {
              name: $t('所有订单数'),
              label: '52',
            },

            {
              name: $t('人数'),
              label: '52',
            },

            {
              name: $t('平均订单金额'),
              label: '￥689.12',
            },
          ],
          [
            {
              name: $t('高峰时间'),
              num: $t('订单数'),
              label: $t('金额'),
              font500: true,
            },
            {
              name: '5/31 17:00-18:00',
              num: '68',
              label: '￥520.00',
            },
            {
              name: '5/31 17:00-18:00',
              num: '50',
              label: '￥520.00',
            },
          ],
          [
            {
              name: $t('分类'),
              num: $t('数量'),
              label: $t('小计'),
              font500: true,
            },
            {
              name: $t('肉类'),
              num: '68',
              label: '￥50.00',
            },
            {
              name: $t('蔬菜类'),
              num: '50',
              label: '￥50.00',
            },
            {
              name: $t('小计'),
              num: '50',
              label: '￥50.00',
            },
          ],
          [
            {
              name: $t('上一班遗留备用金'),
              label: '￥250.00',
            },
            {
              name: $t('中途存入现金'),
              label: '￥50.00',
            },
            {
              name: $t('中途取出现金'),
              label: '￥50.00',
            },
            {
              name: $t('本班取出现金'),
              label: '￥387.00',
            },
            {
              name: $t('本班遗留备用金'),
              label: '￥672.00',
            },
          ],
        ],
      };
    },
    props: ['open', 'title', 'template', 'editId', 'print_method'],
    created() {
      this.dialogVisible = this.open;
      this.brand = this.cloudBasic.base.brand_name;
      this.mode = this.template;
      if (this.title == $t('交班单')) {
        if (this.mode == 1) {
          this.details = this.one;
        }
        if (this.mode == 2) {
          this.details = this.nine;
        }
        if (this.mode == 3) {
          this.details = this.twelve;
        }
        this.titleName = this.title;
      }
      if (this.title == $t('结账单')) {
        this.details = this.two;
        this.titleName = $t('桌位: ') + 'A01' + $t('（4人）');
      }
      if (this.title == $t('预结账单')) {
        this.details = this.three;
        this.titleName = $t('桌位: ') + 'A01' + $t('（4人）');
      }
      if (this.title == $t('一菜一单')) {
        this.details = this.four;
        this.dialogWidth = 2;
        this.storeShow = false;
      }
      if (this.title == $t('营业数据')) {
        this.details = this.five;
        this.titleName = this.title;
      }
      if (this.title == $t('整单打印')) {
        this.details = this.six;
        this.storeShow = false;
        this.titleName = $t('桌位: ') + 'A01';
      }
      if (this.title == $t('发票')) {
        if (this.mode == 1) {
          this.details = this.seven;
        }
        if (this.mode == 2) {
          this.details = this.eight;
        }
        this.dialogWidth = 1;
        this.titleName = $t('发票 ');
      }
      if (this.title == $t('充值单')) {
        this.details = this.ten;
        this.titleName = this.title;
      }
      if (this.title == $t('退菜单')) {
        this.details = this.eleven;
        this.storeShow = false;
        this.titleName = $t('桌位: ') + 'A01';
      }
    },
    methods: {
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

      modeChange(e) {
        this.mode = e;
        if (this.title == $t('发票') && e == 1) {
          this.details = this.seven;
        }
        if (this.title == $t('发票') && e == 2) {
          this.details = this.eight;
        }

        if (this.title == $t('交班单') && e == 1) {
          this.details = this.one;
        }
        if (this.title == $t('交班单') && e == 2) {
          this.details = this.nine;
        }
        if (this.title == $t('交班单') && e == 3) {
          this.details = this.twelve;
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
  .mb-8 {
    margin-bottom: 8px;
  }
  .mb-12 {
    margin-bottom: 12px;
  }
  .mb-14 {
    margin-bottom: 16px;
  }

  .mb-16 {
    margin-bottom: 16px;
  }

  .mb-24 {
    margin-bottom: 24px;
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

  .font-big {
    font-size: 18px;
    font-weight: 500;
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
  * {
    overflow-wrap: anywhere;
    white-space: pre-line;
  }
  .logo {
    width: 80px;
    height: 80px;
    margin: 12px auto 12px;
    border-radius: 50%;
  }
</style>
