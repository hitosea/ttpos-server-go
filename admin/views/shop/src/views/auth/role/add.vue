<template>
  <div v-loading="loading" class="add-box-role">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" label-position="top" label-width="180px">
      <!--添加门店-->
      <div class="common-form">{{ $t('添加角色') }}</div>

      <el-form-item for="no_click" :label="$t('角色名称：')" prop="role_name" :rules="[{ required: true, message: $t('请输入角色名称') }]">
        <el-input v-model="form.role_name" :placeholder="$t('请输入角色名称')" :maxlength="50" class="max-w460"></el-input>
      </el-form-item>

      <el-form-item for="no_click" class="role-list" :label="$t('权限列表：')" :rules="[{ required: true, message: '' }]">
        <div class="flex">
          <!-- 权限主要列表 -->
          <div class="role-menu">
            <template v-for="(item, index) in data" :key="index">
              <div class="role-menu-title" :class="{ active: active == index }" @click="handleClick(index)" v-if="item.path != 'management_app'">{{ item.name }}</div>
            </template>
          </div>
          <div class="flex-1">
            <el-tree
              :data="data.slice(active, active + 1)"
              show-checkbox
              node-key="access_id"
              :default-expand-all="true"
              :default-checked-keys="checkedKeysMap[active]"
              :props="defaultProps"
              @check="handleCheckChange"
            ></el-tree>
          </div>
        </div>
      </el-form-item>

      <!-- <el-form-item for="no_click"  :label="$t('排序：')"><el-input type="number" v-model="form.sort" placeholder="$t('接近0，排序等级越高')"
                    class="max-w460"></el-input></el-form-item> -->

      <!--提交-->
      <div class="common-button-wrapper">
        <el-button size="small" @click="cancelFunc">{{ $t('取消') }}</el-button>
        <el-button type="primary" size="small" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
      </div>
    </el-form>
  </div>
</template>

<script>
  import AuthApi from '@/api/auth.js';
  import { useUserStore } from '@/store/index';
  const { computedSupplier } = useUserStore();
  const supplier = computedSupplier().supplier;
  const app_id = supplier.value?.app_id || 0;
  export default {
    data() {
      return {
        app_id: app_id,
        /*是否正在加载*/
        loading: true,
        /*表单数据对象*/
        form: {
          access_id: [],
          role_name: '',
          sort: 1,
        },
        data: [],
        rawMenuData: [], // 保存原始菜单数据
        appManagePermissionIds: [], // 管理APP的所有权限ID
        defaultProps: {
          children: 'children',
          label: 'name',
        },
        active: 0,
        checkedKeysMap: [],
        upCheckedKeysMap: [],
      };
    },
    created() {
      /*获取列表*/
      this.getData();
    },

    methods: {
      /*添加角色*/
      onSubmit() {
        let self = this;
        let form = self.form;
        form.access_id = [];
        self.upCheckedKeysMap.map((item) => {
          form.access_id = form.access_id.concat(item);
        });
        // 这个时候需要加上管理APP的权限和他子级权限
        if (self.appManagePermissionIds && self.appManagePermissionIds.length > 0) {
          form.access_id = form.access_id.concat(self.appManagePermissionIds);
        }

        self.$refs.form.validate((valid) => {
          if (valid) {
            if (self.form.access_id.length == 0) {
              this.$ElMessage({
                message: this.$t('请选择权限'),
                type: 'error',
              });
              return;
            }
            self.loading = true;
            AuthApi.roleAdd(
              {
                params: JSON.stringify(form),
              },
              true
            )
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: '添加成功',
                  type: 'success',
                });
                self.$router.push('/' + this.app_id + '/auth/role/index');
              })
              .catch((error) => {
                self.loading = false;
              });
          } else {
            const divElement = document.querySelector('.main-div');
            divElement.scrollTop = 0;
          }
        });
      },

      // 收集权限ID的递归函数
      collectPermissionIds(items) {
        let ids = [];
        items.forEach(item => {
          ids.push(item.access_id);
          if (item.children && item.children.length > 0) {
            ids = ids.concat(this.collectPermissionIds(item.children));
          }
        });
        return ids;
      },

      /*获取数据*/
      getData() {
        let self = this;
        AuthApi.roleAddInfo()
          .then((data) => {
            // 保存原始数据用于后续查找管理APP权限
            self.rawMenuData = data.data.menu;

            // 查找并保存管理APP的所有权限ID
            const appManageItem = data.data.menu.find(item => item.path === 'management_app');
            if (appManageItem) {
              self.appManagePermissionIds = this.collectPermissionIds([appManageItem]);
            } else {
              self.appManagePermissionIds = [];
            }

            self.data = data.data.menu;
            self.checkedKeysMap = [];
            self.upCheckedKeysMap = [];
            data.data.menu.map((item, index) => {
              self.checkedKeysMap.push([]);
              self.upCheckedKeysMap.push([]);
              self.data[index].name = $t(item.name);
              item.children.map((items, indexs) => {
                self.data[index].children[indexs].name = $t(items.name);
                items.children.map((itemThree, indexThree) => {
                  self.data[index].children[indexs].children[indexThree].name = $t(itemThree.name);
                  itemThree.children.map((itemFour, indexFour) => {
                    self.data[index].children[indexs].children[indexThree].children[indexFour].name = $t(itemFour.name);
                    itemFour.children.map((itemFive, indexFive) => {
                      self.data[index].children[indexs].children[indexThree].children[indexFour].children[indexFive].name = $t(itemFive.name);
                      itemFive.children.map((itemSix, indexSix) => {
                        self.data[index].children[indexs].children[indexThree].children[indexFour].children[indexFive].children[indexSix].name = $t(itemSix.name);
                      });
                    });
                  });
                });
              });
            });

            self.loading = false;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      handleClick(index) {
        this.active = index;
      },

      //监听选中
      handleCheckChange(data, checked) {
        this.$nextTick(() => {
          this.checkedKeysMap[this.active] = [];
          this.upCheckedKeysMap[this.active] = [];
          checked.checkedKeys.map((item) => {
            this.checkedKeysMap[this.active].push(item);
          });
          checked.checkedKeys.concat(checked.halfCheckedKeys).map((item) => {
            this.upCheckedKeysMap[this.active].push(item);
          });
        });
      },

      /*取消*/
      cancelFunc() {
        this.$router.back(-1);
      },
    },
  };
</script>

<style lang="scss" scoped>
  .img {
    margin-top: 10px;
  }

  .add-box-role {
    height: calc(100% - 14px);
    overflow: hidden;

    .el-form {
      display: flex;
      flex-direction: column;
      height: 100%;

      .role-list {
        flex: 1 1 auto;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        :deep(.el-form-item__content) {
          flex: 1 1 auto;
          overflow: hidden;
          position: relative;
        }
      }
    }
  }

  :deep(.el-tree-node__label) {
    text-transform: capitalize;
  }
  .flex {
    display: flex;
    position: absolute;
    gap: 16px;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    .role-menu {
      display: flex;
      flex-direction: column;
      gap: 4px;
      .role-menu-title {
        padding: 0 12px;
        border-radius: 4px;
        text-align: center;
        cursor: pointer;
        min-width: 160px;
      }
      .role-menu-title:hover {
        background-color: #f5f5f5;
        color: var(--el-color-primary);
      }
      .active {
        background-color: var(--el-color-primary);
        color: #fff;
        font-weight: bold;
      }
    }
    .flex-1 {
      flex: 1 1 auto;
      overflow-y: auto;
    }
  }
</style>
