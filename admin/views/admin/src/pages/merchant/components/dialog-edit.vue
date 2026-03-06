<template>
  <el-dialog
    width="960"
    :title="hasEdit ? $t('編輯商家') : $t('添加商家')"
    :modelValue="props.show"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    @close="handleClose()"
  >
    <div class="max-h-[75vh] overflow-auto pr-4" v-if="showDiv">
      <el-form :model="formData" :rules="formRules" ref="formElement" label-position="top" label-width="auto">
        <el-form-item :label="$t('商家名称')" prop="name">
          <el-input v-model="formData.name" type="text" maxlength="100" clearable :placeholder="$t('请输入商家名称')"></el-input>
        </el-form-item>
        <!--
        NOTE sass暂时没一二级商家 2024年08月22日10:11:01
        <el-form-item :label="$t('部署方式')" prop="deploy_mode">
          <el-radio-group v-model="formData.deploy_mode" :disabled="hasEdit" @change="handleChangeDeployMode">
            <el-radio :value="0">{{ $t('局域网部署') }}</el-radio>
            <el-radio :value="1">{{ $t('SAAS') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('商家等级')" prop="level" v-if="formData.deploy_mode == 0">
          <el-radio-group v-model="formData.level" :disabled="hasEdit">
            <el-radio :value="1">{{ $t('一级') }}</el-radio>
            <el-radio :value="2">{{ $t('二级') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="formData.level == 2">
          <el-form-item :label="$t('上级商家')" prop="parent_id">
            <el-select v-model="formData.parent_id" :placeholder="$t('请选择上级商家')" @change="handleSelectChange" clearable style="min-width: 200px">
              <el-option v-for="item in selectList" :key="item.id" :value="item.id" :label="item.name" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('经营模式')" prop="store_type">
            <el-radio-group v-model="formData.store_type" :disabled="hasEdit">
              <el-radio :value="20">{{ $t('直营店') }}</el-radio>
              <el-radio :value="10">{{ $t('加盟店') }}</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item :label="$t('商品分类')" prop="category_set">
            <el-radio-group v-model="formData.category_set" :disabled="hasEdit">
              <el-radio :value="10">{{ $t('同步总店') }}</el-radio>
              <el-radio :value="20">{{ $t('不同步总店') }}</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>
        <el-form-item :label="$t('连锁编号')" v-if="(props.detail?.suppliers_total && props.detail?.suppliers_total > 0) || formData.level == 2">
          <el-input v-model="formData.chain_number" type="text" maxlength="50" clearable :placeholder="$t('请输入连锁编号')"></el-input>
        </el-form-item>
        -->

        <el-form-item label="LOGO" prop="logo">
          <el-image v-if="hasEdit" :src="formData.logo" :preview-src-list="[formData.logo]" style="width: 48px; height: 48px; background-color: #f6f8fb" fit="cover" />
          <template v-else>
            <ti-upload :model-value="formData.logo" @change="handleUpload" />
            <div class="text-[#ccc] w-full">{{ $t('支持JPG、JPEG、PNG格式，小於15MB，尺寸：120*120px') }}</div>
          </template>
        </el-form-item>

        <el-form-item :label="$t('进销存')" prop="sale_stock">
          <el-radio-group v-model="formData.sale_stock">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('优惠券')" prop="is_open_coupon">
          <el-radio-group v-model="formData.is_open_coupon">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('营销活动')" prop="is_open_marketing">
          <el-radio-group v-model="formData.is_open_marketing">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('会员')" prop="is_open_member">
          <el-radio-group v-model="formData.is_open_member">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('平板点餐')" prop="is_open_tablet">
          <el-radio-group v-model="formData.is_open_tablet">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('扫码H5')" prop="is_open_scan">
          <el-radio-group v-model="formData.is_open_scan">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('点餐助手')" prop="is_open_assistant">
          <el-radio-group v-model="formData.is_open_assistant">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('后厨KDS')" prop="is_open_kitchen_kds">
          <el-radio-group v-model="formData.is_open_kitchen_kds">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('自助点餐机')" prop="enable_kiosk">
          <el-radio-group v-model="formData.enable_kiosk">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('自助餐')" prop="is_open_buffet">
          <el-radio-group v-model="formData.is_open_buffet">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- 扫码点餐到店自取 -->
        <el-form-item :label="$t('扫码点餐到店自取')" prop="is_open_member_instant">
          <el-radio-group v-model="formData.is_open_member_instant">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('短信服务')" prop="enable_sms">
          <el-radio-group v-model="formData.enable_sms" @click="handleChangeSms">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('扫码点餐接单')" prop="is_accept_scan_order">
          <el-radio-group v-model="formData.is_accept_scan_order">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <template #label>
            <span class="flex items-center">
              <b class="text-[#f56c6c] mr-1">*</b>
              {{ $t('本地打印服务') }}
              <el-tooltip :content="$t('如使用商米云打印，建议关闭，否则将影响收银机性能')"
                ><el-icon class="ml-1 text-[16px]"><QuestionFilled /></el-icon
              ></el-tooltip>
            </span>
          </template>
          <el-radio-group v-model="formData.is_open_local_print">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('高级票据打印')" prop="is_open_advanced_ticket_print">
          <el-radio-group v-model="formData.is_open_advanced_ticket_print">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('桌台地图')" prop="enable_table_map">
          <el-radio-group v-model="formData.enable_table_map">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- Grab外卖 -->
        <el-form-item :label="$t('Grab外卖')" prop="enable_grab_delivery">
          <el-radio-group v-model="formData.enable_grab_delivery">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- LINE MAN外卖 -->
        <el-form-item :label="$t('LINE MAN外卖')" prop="enable_lineman_delivery">
          <el-radio-group v-model="formData.enable_lineman_delivery">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('数据管理')" prop="enable_data_management">
          <el-radio-group v-model="formData.enable_data_management">
            <el-radio :value="1">{{ $t('开启') }}</el-radio>
            <el-radio :value="0">{{ $t('关闭') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item :label="$t('授权开始时间')" prop="auth_start_time">
          <el-date-picker v-model="formData.auth_start_time" type="datetime" @change="handleDateChange" :placeholder="$t('选择日期时间')"></el-date-picker>
        </el-form-item>
        <el-form-item :label="$t('授权时间')" prop="auth_day">
          <el-input-number v-model="formData.auth_day" :controls="false" :precision="0" :min="0" :max="999" clearable :placeholder="$t('请输入授权时间')"></el-input-number>
          <div class="px-2 text-[#495060]">{{ $t('天') }}</div>
          <div class="text-[#ccc] w-full">{{ $t('0为永不过期') }}</div>
        </el-form-item>

        <el-form-item :label="$t('联系电话')" prop="link_phone" :rules="[{ required: true, message: $t('请输入联系电话') }]">
          <el-input v-model="formData.link_phone" type="text" maxlength="20" :placeholder="$t('请输入联系电话')"></el-input>
        </el-form-item>

        <el-form-item :label="$t('支持语言')" prop="languages">
          <el-select v-model="formData.languages" :placeholder="$t('请选择支持语言')" multiple @change="handleLangChange" clearable style="min-width: 200px">
            <el-option v-for="item in languageList" :key="item.key" :value="item.key" :label="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('桌台上限')" prop="table_limit">
          <div class="flex items-center gap-2 w-full">
            <el-input-number
              v-model="formData.table_limit"
              :controls="false"
              :precision="0"
              :min="0"
              :max="999"
              clearable
              :placeholder="limitTable ? $t('不限制') : $t('请输入桌台上限')"
              style="width: 100%"
              :disabled="limitTable"
            ></el-input-number>
            <el-checkbox v-model="limitTable" @change="handleChangeTable">{{ $t('不限制') }}</el-checkbox>
          </div>
        </el-form-item>
        <el-form-item :label="$t('打印机上限')" prop="printer_limit">
          <div class="flex items-center gap-2 w-full">
            <el-input-number
              v-model="formData.printer_limit"
              :controls="false"
              :precision="0"
              :min="0"
              :max="999"
              clearable
              :placeholder="limitPrinter ? $t('不限制') : $t('请输入打印机上限')"
              style="width: 100%"
              :disabled="limitPrinter"
            ></el-input-number>
            <el-checkbox v-model="limitPrinter" @change="handleChangePrinter">{{ $t('不限制') }}</el-checkbox>
          </div>
        </el-form-item>
        <el-form-item :label="$t('收银机登录上限')" prop="cash_limit">
          <el-input-number
            v-model="formData.cash_limit"
            :controls="false"
            :precision="0"
            :min="0"
            :max="999"
            clearable
            :placeholder="$t('请输入收银机登录上限')"
            style="width: 100%"
          ></el-input-number>
          <!-- <div class=" text-[#ccc] w-full">{{ $t('最大999') }}</div> -->
        </el-form-item>
        <el-form-item :label="$t('平板登录上限')" prop="tablet_limit">
          <el-input-number
            v-model="formData.tablet_limit"
            :controls="false"
            :precision="0"
            :min="0"
            :max="999"
            clearable
            :placeholder="$t('请输入平板登录上限')"
            style="width: 100%"
          ></el-input-number>
          <!-- <div class=" text-[#ccc] w-full">{{ $t('最大999') }}</div> -->
        </el-form-item>
        <el-form-item :label="$t('点餐助手登录上限')" prop="assistant_limit">
          <el-input-number
            v-model="formData.assistant_limit"
            :controls="false"
            :precision="0"
            :min="0"
            :max="999"
            clearable
            :placeholder="$t('请输入点餐助手登录上限')"
            style="width: 100%"
          ></el-input-number>
          <!-- <div class=" text-[#ccc] w-full">{{ $t('最大999') }}</div> -->
        </el-form-item>
        <el-form-item :label="$t('厨显登录上限')" prop="kitchen_limit">
          <el-input-number
            v-model="formData.kitchen_limit"
            :controls="false"
            :precision="0"
            :min="0"
            :max="999"
            clearable
            :placeholder="$t('请输入厨显登录上限')"
            style="width: 100%"
          ></el-input-number>
          <!-- <div class=" text-[#ccc] w-full">{{ $t('最大999') }}</div> -->
        </el-form-item>
        <el-form-item :label="$t('短信额度')" prop="sms_quota" v-if="formData.enable_sms == 1">
          <el-input-number
            v-model="formData.sms_quota"
            :controls="false"
            clearable
            :precision="0"
            :placeholder="$t('请输入短信额度')"
            style="width: 100%"
            @blur="handleSmsQuotaInput"
          ></el-input-number>
          <div class="text-[#ccc] w-full">{{ $t('最小额度为0，短信额度的最大值为9,999,999，默认为200') }}</div>
        </el-form-item>
        <el-form-item :label="$t('时区')" prop="timezone">
          <el-select v-model="formData.timezone" :placeholder="$t('请选择时区')" @change="handleZoneChange" clearable style="min-width: 200px">
            <el-option v-for="item in timeZoneList" :key="item.key" :value="item.key" :label="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('超管邮箱')" prop="user_name">
          <el-input v-model="formData.user_name" type="text" maxlength="50" clearable :placeholder="$t('请输入邮箱')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('超管密码')" prop="password">
          <el-input
            v-model="formData.password"
            type="password"
            maxlength="16"
            clearable
            show-password
            :placeholder="hasEdit ? $t('不修改则使用原密码') : $t('请输入超管密码')"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('确认超管密码')" prop="confirm_password">
          <el-input v-model="formData.confirm_password" type="password" maxlength="16" clearable show-password :placeholder="$t('请再次输入超管密码')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('所在地')">
          <el-input v-model="formData.address" type="textarea" rows="5" maxlength="500" clearable :placeholder="$t('请输入所在地')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('经纬度')">
          <el-input v-model="formData.coordinates" type="textarea" rows="5" maxlength="500" clearable :placeholder="$t('如:13.716412789763694, 100.52312952599786')"></el-input>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleClose()">{{ $t('取消') }}</el-button>
      <el-button :loading="formLoading" type="primary" @click="handleSubmit()">{{ $t('确定') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { ref, watch, computed, reactive } from 'vue';
  import { ShowAddType, fetchShowAdd, fetchShowEdit, getLangList } from '@/api/merchant';
  import { message } from '@/utils/feedback';
  import { $t } from '@/i18n';
  import dayjs from 'dayjs';

  const emits = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'close', value: boolean): void;
    (e: 'change', value: any): void;
  }>();
  const props = withDefaults(
    defineProps<{
      show?: boolean;
      detail?: ShowAddType & { app_id?: number; shop_supplier_name?: string };
    }>(),
    {
      show: false,
      detail: () => ({}),
    },
  );

  const formData = ref<ShowAddType>({
    name: '', // 商家名称
    level: 1, // 商家等级: 从1开始
    parent_id: undefined, // 上级商家ID
    store_type: 20, // 经营模式: 20直营店, 10加盟店
    category_set: 10, // 商品分类设置10同步 主店20分店创建
    logo: '', // Logo
    sale_stock: 1, // 进销存: 0不开启, 1开启
    is_open_tax: 0, // 税务对接: 0不开启, 1开启
    reserve: 1, // 预订: 0不开启, 1开启
    auth_day: 0, // 授权时间(天) 0为永不过期
    auth_start_time: '', // 授权开始时间
    link_phone: '', // 联系电话
    cash_limit: undefined, // 收银机上限
    tablet_limit: undefined, // 平板上限
    kitchen_limit: undefined, // 厨显上限
    assistant_limit: undefined, // 点餐上限
    user_name: '', // 邮箱
    password: '', // 超管密码
    confirm_password: '', // 再次确认超管密码
    address: '', // 所在地
    deploy_mode: 1, //部署方式
    chain_number: '', //部署方式
    languages: [], // 支持语言
    timezone: '', // 时区
    table_limit: undefined, // 桌台限制
    printer_limit: undefined, // 打印机限制
    is_open_member: 0, // 是否开启会员: 0不开启, 1开启
    is_open_tablet: 0, // 是否开启平板: 0不开启, 1开启
    is_open_scan: 0, // 是否开启扫码H5: 0不开启, 1开启
    is_open_assistant: 0, // 是否开启点餐助手: 0不开启, 1开启
    is_open_kitchen_kds: 0, // 是否开启后厨KDS: 0不开启, 1开启
    is_open_buffet: 0, // 是否开启自助餐: 0不开启, 1开启
    is_accept_scan_order: 0, // 是否开启扫码点餐接单: 0不开启, 1开启
    is_open_local_print: 0, // 是否开启本地打印服务: 0不开启, 1开启（v1.1.0）
    enable_sms: 0, // 是否开启短信服务: 0不开启, 1开启
    sms_quota: 200, // 短信额度
    is_open_coupon: 0, // 是否开启优惠券: 0不开启, 1开启
    is_open_marketing: 0, // 是否开启营销活动: 0不开启, 1开启
    is_open_advanced_ticket_print: 0, // 是否开启高级票据打印: 0不开启, 1开启
    coordinates: '', // 经纬度
    enable_table_map: 0, // 是否启用桌台地图能力: 0不开启, 1开启
    enable_grab_delivery: 0, // 是否开启Grab外卖: 0不开启, 1开启
    enable_lineman_delivery: 0, // 是否开启LINE MAN外卖: 0不开启, 1开启
    enable_data_management: 0, // 是否启用数据管理能力: 0不开启, 1开启
    enable_kiosk: 0, // 是否开启自助点餐机: 0不开启, 1开启
    is_open_member_instant: 0, // 是否开启扫码点餐到店自取: 0不开启, 1开启
  });

  const limitTable = ref(false);
  const limitPrinter = ref(false);
  const showDiv = ref(false);

  const formRules = reactive({
    name: [{ required: true, message: $t('请输入商家名称'), trigger: 'blur' }],
    level: [{ required: true, message: $t('请输入商家等级'), trigger: 'blur' }],
    parent_id: [{ required: true, message: $t('请选择上级商家'), trigger: 'blur' }],
    store_type: [{ required: true, message: $t('请选择经营模式'), trigger: 'blur' }],
    category_set: [{ required: true, message: $t('请选择商品分类'), trigger: 'blur' }],
    logo: [{ required: true, message: $t('请上传LOGO'), trigger: ['change', 'blur'] }],
    sale_stock: [{ required: true, message: $t('请选择进销存'), trigger: 'blur' }],
    is_open_tax: [{ required: true, message: $t('请选择税务对接'), trigger: 'blur' }],
    reserve: [{ required: true, message: $t('请选择预订'), trigger: 'blur' }],
    auth_day: [{ required: true, message: $t('请输入授权时间'), trigger: 'blur' }],
    auth_start_time: [{ required: true, message: $t('请选择授权开始时间'), trigger: 'blur' }],
    cash_limit: [{ required: true, message: $t('请输入收银机登录上限'), trigger: 'blur' }],
    tablet_limit: [{ required: true, message: $t('请输入平板登录上限'), trigger: 'blur' }],
    kitchen_limit: [{ required: true, message: $t('请输入厨显登录上限'), trigger: 'blur' }],
    assistant_limit: [{ required: true, message: $t('请输入点餐助手登录上限'), trigger: 'blur' }],
    timezone: [{ required: true, message: $t('请选择时区'), trigger: 'blur' }],
    languages: [{ required: true, message: $t('请选择支持语言'), trigger: 'blur' }],
    is_open_member: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_tablet: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_scan: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_assistant: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_kitchen_kds: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_buffet: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_accept_scan_order: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_local_print: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_sms: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    sms_quota: [{ required: true, message: $t('请输入短信额度'), trigger: 'blur' }],
    is_open_coupon: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_marketing: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    is_open_advanced_ticket_print: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_table_map: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_grab_delivery: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_lineman_delivery: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_data_management: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    enable_kiosk: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
    printer_limit: [
      {
        required: true,
        trigger: 'blur',
        validator: (_: object, value: string, callback: any) => {
          if (!value && !limitPrinter.value) {
            callback(new Error($t('请输入打印机上限')));
          }
          callback();
        },
      },
    ],
    table_limit: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (!value && !limitTable.value) {
            callback(new Error($t('请输入桌台上限')));
          }
          callback();
        },
      },
    ],
    user_name: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          const re = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
          if (!value) {
            callback(new Error($t('请输入邮箱')));
          } else if (!re.test(value)) {
            callback(new Error($t('请确认邮箱格式')));
          }
          callback();
        },
      },
    ],
    password: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (hasEdit.value && !value) callback();
          if (!value) {
            callback(new Error($t('请输入超管密码')));
          } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value)) {
            callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
          }
          callback();
        },
      },
    ],
    confirm_password: [
      {
        required: true,
        trigger: ['change', 'blur'],
        validator: (_: object, value: string, callback: any) => {
          if (hasEdit.value && !value) callback();
          if (!value) {
            callback(new Error($t('请输入确认密码')));
          } else if (value !== formData.value.password) {
            callback(new Error($t('两次密码不一致！')));
          }
          callback();
        },
      },
    ],
  });
  const formLoading = ref(false);
  const formElement = ref();
  const languageList = ref<any>([]);
  const timeZoneList = ref<any>([]);
  const hasEdit = computed(() => !!props.detail?.app_id);

  const handleUpload = (url: string) => {
    formData.value.logo = url;
    formElement.value?.validateField('logo').catch(() => {});
  };

  const handleLangChange = () => {
    formElement.value?.validateField('languages').catch(() => {});
  };

  const handleZoneChange = () => {
    formElement.value?.validateField('timezone').catch(() => {});
  };

  const handleDateChange = () => {
    formElement.value?.validateField('parent_id').catch(() => {});
  };

  const handleChangeTable = () => {
    formData.value.table_limit = undefined;
    formElement.value?.validateField('table_limit').catch(() => {});
  };

  const handleChangePrinter = () => {
    formData.value.printer_limit = undefined;
    formElement.value?.validateField('printer_limit').catch(() => {});
  };

  const handleChangeSms = () => {
    formData.value.sms_quota = 200;
  };

  const handleSmsQuotaInput = (value: number | undefined) => {
    if (value === undefined) return;
    setTimeout(() => {
      if (value < 0) {
        formData.value.sms_quota = 0;
      } else if (value > 9999999) {
        formData.value.sms_quota = 9999999;
      }
    }, 300);
  };

  const handleSubmit = () => {
    if (formData.value.enable_sms === 1 && formData.value.sms_quota != null && (formData.value.sms_quota < 0 || formData.value.sms_quota > 9999999)) {
      message.error($t('输入内容不合规，请重新输入'));
      formData.value.sms_quota = 0;
      return;
    }
    formElement.value?.validate(async (valid: boolean) => {
      if (!valid) {
        setTimeout(() => {
          const errorItems = document.querySelectorAll('.el-form-item__error');
          if (errorItems.length > 0) {
            const firstErrorItem = errorItems[0];
            firstErrorItem.scrollIntoView({
              behavior: 'smooth',
              block: 'center',
            });
          }
        }, 200);
        return;
      }
      try {
        formLoading.value = true;
        let res = null;
        // formData.name 去掉前后空格
        formData.value.name = formData.value.name?.trim();
        formData.value.auth_start_time = dayjs(formData.value.auth_start_time).format('YYYY-MM-DD HH:mm:ss');
        if (limitPrinter.value) {
          formData.value.printer_limit = -1;
        }
        if (limitTable.value) {
          formData.value.table_limit = -1;
        }
        if (hasEdit.value) {
          res = await fetchShowEdit({ ...formData.value, app_id: props.detail?.app_id });
        } else {
          res = await fetchShowAdd({ ...formData.value });
        }
        message.success(res.msg);
        handleClose();
        emits('change', res);
      } catch (error) {
        //
      } finally {
        formLoading.value = false;
      }
    });
  };

  const handleClose = () => {
    emits('update:show', false);
    emits('close', false);
  };

  watch(
    () => props.show,
    (val) => {
      if (!val) {
        setTimeout(() => {
          showDiv.value = false;
        }, 300);
        return;
      }
      showDiv.value = true;
      formElement.value?.resetFields();
      // 重置校验
      formRules.password[0].required = !hasEdit.value;
      formRules.confirm_password[0].required = !hasEdit.value;
      formRules.logo[0].required = !hasEdit.value;
      //
      getLangList({ page: 1, list_rows: 1000 }).then((res) => {
        languageList.value = res.data?.language_list || [];
        timeZoneList.value = res.data?.time_zone_list || [];
      });
      // 默认值
      formData.value = {
        name: props.detail?.shop_supplier_name || '',
        level: props.detail?.level || 1,
        parent_id: props.detail?.parent_id || undefined,
        store_type: props.detail?.store_type || 20,
        category_set: props.detail?.category_set || 10,
        logo: props.detail?.logo || '',
        sale_stock: props.detail?.sale_stock || 0,
        is_open_tax: props.detail?.is_open_tax || 0,
        reserve: props.detail?.reserve || 0,
        auth_day: props.detail?.auth_day || 0,
        auth_start_time: props.detail?.auth_start_time || dayjs().format('YYYY-MM-DD HH:mm:ss'),
        link_phone: props.detail?.link_phone || '',
        cash_limit: hasEdit.value ? props.detail?.cash_limit : undefined,
        kitchen_limit: hasEdit.value ? props.detail?.kitchen_limit : undefined,
        tablet_limit: hasEdit.value ? props.detail?.tablet_limit : undefined,
        assistant_limit: hasEdit.value ? props.detail?.assistant_limit : undefined,
        user_name: props.detail?.user_name || '',
        password: props.detail?.password || '',
        confirm_password: props.detail?.confirm_password || '',
        address: props.detail?.address || '',
        deploy_mode: Number(props.detail?.deploy_mode) || 0, //部署方式
        chain_number: props.detail?.chain_number || '', //连锁编号
        timezone: props.detail?.timezone || '', //连锁编号
        languages: JSON.parse(props.detail?.languages || '[]') || [], //连锁编号
        is_open_member: props.detail?.is_open_member || 0, //
        is_open_tablet: props.detail?.is_open_tablet || 0, //
        is_open_scan: props.detail?.is_open_scan || 0, //
        is_open_assistant: props.detail?.is_open_assistant || 0, //
        is_open_kitchen_kds: props.detail?.is_open_kitchen_kds || 0, //
        is_open_buffet: props.detail?.is_open_buffet || 0, //
        is_accept_scan_order: props.detail?.is_accept_scan_order || 0, //
        is_open_local_print: props.detail?.is_open_local_print || 0, //
        enable_sms: props.detail?.enable_sms || 0, //
        sms_quota: props.detail?.sms_quota || 0, //
        table_limit: props.detail?.table_limit == -1 ? undefined : props.detail?.table_limit || undefined, //
        printer_limit: props.detail?.printer_limit == -1 ? undefined : props.detail?.printer_limit || undefined, //
        is_open_coupon: props.detail?.is_open_coupon || 0, //
        is_open_marketing: props.detail?.is_open_marketing || 0, //
        is_open_advanced_ticket_print: props.detail?.is_open_advanced_ticket_print || 0, //
        coordinates: props.detail?.coordinates || '', //
        enable_table_map: props.detail?.enable_table_map || 0, //
        enable_grab_delivery: props.detail?.enable_grab_delivery || 0, //
        enable_lineman_delivery: props.detail?.enable_lineman_delivery || 0, //
        enable_data_management: props.detail?.enable_data_management || 0, //
        enable_kiosk: props.detail?.enable_kiosk || 0, //
        is_open_member_instant: props.detail?.is_open_member_instant || 0, //
      };
      props.detail?.printer_limit == -1 ? (limitPrinter.value = true) : (limitPrinter.value = false);
      props.detail?.table_limit == -1 ? (limitTable.value = true) : (limitTable.value = false);
      // 清空
      setTimeout(() => {
        formElement.value?.clearValidate();
      }, 10);
    },
  );
</script>

<style lang="scss" scoped></style>
