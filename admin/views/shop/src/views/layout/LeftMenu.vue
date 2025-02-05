<template>
  <div class="left-menu-wrapper">
    <!--主菜单-->
    <div class="menu-wrapper">
      <div class="home-login">
        <div :class="active_menu != null ? 'home-icon' : 'home-icon router-link-active'" @click="choseMenu(1, null, null)">
          <!-- <span class="icon iconfont icon-tubiaozhizuomoban-" v-if="!userInfo.logoUrl"></span> -->
          <img :src="userInfo.logoUrl" class="logoImg" />
        </div>
      </div>
      <div class="d-c-c">
        <span class="fb tc home-title">{{ userInfo.shopName || '点餐系统连锁总店' }}</span>
      </div>
      <div class="nav-wrapper mt10">
        <div class="first-menu-content">
          <ul class="nav-ul">
            <template v-for="(item, index) in menuList" :key="index">
              <li :class="active_menu == index ? 'menu-item router-link-active' : 'menu-item'" @click="choseMenu(2, item, index)" v-if="item.is_menu == 1">
                <div class="item-box">
                  <span :class="'icon iconfont menu-item-icon ' + item.icon"></span>
                  <span>{{ $t(item.name) }}</span>
                </div>
              </li>
            </template>
          </ul>
        </div>
      </div>
    </div>
    <!--子菜单-->
    <div class="child-menu-wrapper">
      <div class="child-menu right-animation">
        <ul v-if="active_menu != null">
          <template v-for="(item, index) in menuList[active_menu]['children']" :key="index">
            <li :class="active_child == index ? 'routre-link router-link-active' : 'router-link'" @click="choseMenu(3, item, index)" v-if="item.is_menu == 1">
              <span class="name">{{ $t(item.name) }}</span>
            </li>
          </template>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
  import { useUserStore } from '@/store';
  import { reactive, toRefs, defineComponent, nextTick, handleError } from 'vue';
  import { languageStore } from '@/store/model/language.js';
  import { useRoute } from 'vue-router';

  // import { Value } from 'sass';
  export default defineComponent({
    components: {},
    setup(props, { emit }) {
      const { userInfo, bus_emit, menus, computedRenderMenus } = useUserStore();
      const renderMenus = computedRenderMenus().renderMenus;
      const route = useRoute();
      const language = languageStore();
      const cloudBasic = language.getCloudBasic().cloudBasic;
      const { computedSupplier } = useUserStore();
      const supplier = computedSupplier().supplier;
      const state = reactive({
        route,
        /*传到顶部的标题*/
        munu_name: $t('首页'),
        /*选中的菜单*/
        active_menu: null,
        /*子菜单选择*/
        active_child: 0,
        /*菜单数据*/
        menuList: renderMenus,
        /*商城名称*/
        shop_name: '',
        menus,
        app_id: supplier.value?.app_id || 0,
      });

      /*菜单*/
      const selectMenu = (to) => {
        let menupath = to.path;
        let active_menu = null;
        let active_child = null;
        if (state.menuList && state.menuList.length > 0) {
          for (let i = 0; i < state.menuList.length; i++) {
            if (state.menuList[i].path == menupath) {
              active_menu = i;
              break;
            }
            if (state.menuList[i].children) {
              for (let j = 0; j < state.menuList[i].children.length; j++) {
                if (state.menuList[i].children[j].path == menupath) {
                  active_menu = i;
                  active_child = j;
                  break;
                }
              }
            }
            if (!active_menu && !active_child) {
              if (state.menuList[i].childrenList.includes(menupath)) {
                active_menu = i;
                break;
              }
            }
          }
          state.active_menu = active_menu;
          state.active_child = active_child;
          emit('selectMenu', active_menu);
        }
        nextTick(() => {
          bus_emit('MenuName', (to.meta && to.meta.showMenuTitle) || to.meta.title);
        });
      };

      selectMenu(route);
      return {
        ...toRefs(state),
        userInfo,
        selectMenu,
        bus_emit,
        cloudBasic,
      };
    },
    mounted() {
      if (this.route.path.includes('/home')) {
        this.$emit('selectMenu', null);
      }
    },
    watch: {
      $route: {
        handler(newVal) {
          if (
            newVal.meta.topTree == '/product/store/product/add' ||
            newVal.meta.topTree == '/product/store/product/edit' ||
            newVal.meta.topTree == '/product/store/product/batch' ||
            newVal.meta.topTree == '/product/store/product/importProduct'
          ) {
            this.menuList.map((item, index) => {
              if (item.name == '商品管理') {
                this.active_menu = index;
                this.active_child = 0;
                this.$emit('selectMenu', false);
              }
            });
          }
          if (newVal.meta.topTree == '/supplier/table/table/importQrcode') {
            this.menuList.map((item, index) => {
              if (item.name == '门店管理') {
                this.active_menu = index;
                item.children.map((child, i) => {
                  if (child.name == '桌码管理') this.active_child = i;
                });

                this.$emit('selectMenu', false);
              }
            });
          }
          if (newVal.meta.topTree == '/store/order/detail') {
            this.menuList.map((item, index) => {
              if (item.name == '订单管理') {
                this.active_menu = index;
                item.children.map((child, i) => {
                  if (child.name == '用餐订单') this.active_child = i;
                });

                this.$emit('selectMenu', false);
              }
            });
          }
          if (newVal.meta.topTree == '/store/recharge/detail') {
            this.menuList.map((item, index) => {
              if (item.name == '订单管理') {
                this.active_menu = index;
                item.children.map((child, i) => {
                  if (child.name == '充值订单') this.active_child = i;
                });

                this.$emit('selectMenu', false);
              }
            });
          }
          if (newVal.meta.topTree == '/auth/role/edit' || newVal.meta.topTree == '/auth/role/add') {
            this.menuList.map((item, index) => {
              if (item.name == '用户管理') {
                this.active_menu = index;
                item.children.map((child, i) => {
                  if (child.name == '角色管理') this.active_child = i;
                });

                this.$emit('selectMenu', false);
              }
            });
          }
        },
        deep: true,
        immediate: true,
      },
      menuList: {
        handler(newVal) {
          if (newVal && newVal.length > 0) {
            this.menuList.map((item) => {
              if (item.path.includes('undefined')) {
                item.path = item.path.replace('undefined', this.app_id);
              }
              if (item.children && item.children.length > 0) {
                item.children.map((child) => {
                  if (child.path.includes('undefined')) {
                    child.path = child.path.replace('undefined', this.app_id);
                  }
                });
              }
            });
          }
        },
        deep: true,
        immediate: true,
      },
    },
    methods: {
      /*点击菜单跳转*/
      choseMenu(type, item, index, query) {
        if (type == 1) {
          this.active_menu = null;
          this.active_child = null;
          this.$router.push('/');
          this.$emit('selectMenu', null);
          // this.bus_emit('MenuName', '首页');
        } else if (type == 2) {
          this.active_menu = index;
          this.active_child = 0;
          // this.bus_emit('MenuName', item.name);
          if (item.children) {
            this.$emit('selectMenu', false);
            query ? this.$router.push({ path: item.children[0].path, query }) : this.$router.push(item.children[0].path);
          } else {
            this.$router.push(item.redirect_name);
            this.$emit('selectMenu', null);
          }
        } else if (type == 3) {
          let path = item.path;
          if (item.redirect_name) {
            path = item.redirect_name;
          }
          this.active_child = index;
          // this.bus_emit('MenuName', item.name);
          this.$router.push(path);
        }
      },
    },
  });
</script>
<style scoped>
  .home-login .icon-tubiaozhizuomoban- {
    color: #3a8ee6;
    font-size: 28px;
  }

  .logoImg {
    width: 100%;
  }

  .menu-item-icon.icon.iconfont {
    font-size: 18px;
  }

  .menu-item .item-box {
    display: flex;
    align-items: center;
  }
</style>
