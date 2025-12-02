<template>
  <div v-loading="loading" class="add-box-role">
    <!--form表单-->
    <el-form size="small" ref="form" :model="form" label-position="top" label-width="180px">
      <!--编辑角色-->
      <div class="common-form">{{ $t('编辑角色') }}</div>

      <el-form-item for="no_click" :label="$t('角色名称：')" :rules="[{ required: true, message: $t('请输入角色名称') }]">
        <el-input v-model="form.role_name" :placeholder="$t('请输入角色名称')" class="max-w460" :maxlength="50"></el-input>
      </el-form-item>
      <el-form-item for="no_click" class="role-list" :label="$t('权限列表：')" :rules="[{ required: true, message: ' ' }]">
        <div class="flex">
          <!-- 权限主要列表 -->
          <div class="role-menu">
            <template v-for="(item, index) in data" :key="index">  
                <div v-if="item.path != 'management_app'" class="role-menu-title" :class="{ active: active == index }" @click="handleClick(index)">{{ item.name }}</div> 
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

      <!-- <el-form-item for="no_click"  :label="$t('排序：')"><el-input type="number" v-model="form.sort" :placeholder="$t('接近0，排序等级越高')"
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
        },
        data: [],

        /*权限树菜单重新自定义字段*/
        defaultProps: {
          children: 'children',
          label: 'name',
        },
        role_id: 0,
        active: 0,
        checkedKeysMap: [],
        upCheckedKeysMap: [],
        validAccesses: [],
      };
    },
    created() {
      this.role_id = this.$route.query.role_id;
      /*获取列表*/
      this.getData();
    },
    methods: {
      /*修改角色*/
      onSubmit() {
        let self = this;
        let form = self.form;
        //去掉form的access
        delete form.access;
        delete form.create_time;
        delete form.update_time;
        delete form.app_id;
        form.access_id = [];
        self.upCheckedKeysMap.map((item) => {
          form.access_id = form.access_id.concat(item);
        });
        self.$refs.form.validate((valid) => {
          if (valid) {
            // 如果access_id中所有元素长度都大于10，则提示请选择权限
            if (self.form.access_id?.length > 0) {  
              // 过滤掉self.form.access_id中长度大于10的元素
              self.validAccesses = self.form.access_id.filter((item) => item.toString().length < 11);  
              if (self.validAccesses.length == 0) {
                this.$ElMessage({
                  message: this.$t('请选择权限'),
                  type: 'error',
                });
                return;
              }
            }
            if (self.form.access_id?.length == 0) {
              this.$ElMessage({
                message: this.$t('请选择权限'),
                type: 'error',
              });
              return;
            }
            self.loading = true;
            AuthApi.roleEdit(
              {
                role_id: self.role_id,
                params: JSON.stringify(form),
              },
              true
            )
              .then((data) => {
                self.loading = false;
                this.$ElMessage({
                  message: $t('保存成功'),
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

      /*获取所有的数据*/
      getData() {
        let self = this;
        AuthApi.roleEditInfo({
          role_id: self.role_id,
        })
          .then((data) => {
            self.form = data.data.model;
            self.data = data.data.menu;
            let access_id = [];
            data.data.model.access.map((item) => {
              access_id.push(item.access_id);
            });
            self.checkedKeysMap = [];
            self.upCheckedKeysMap = [];
            data.data.menu.map((item, index) => {
              self.checkedKeysMap.push([]);
              self.upCheckedKeysMap.push([]);
              self.data[index].name = $t(item.name);
              if (data.data.select_menu.indexOf(item.access_id) != -1) {
                self.checkedKeysMap[index].push(item.access_id);
              }
              if (access_id.indexOf(item.access_id) != -1) {
                self.upCheckedKeysMap[index].push(item.access_id);
              }
              item.children.map((items, indexs) => {
                self.data[index].children[indexs].name = $t(items.name);
                if (data.data.select_menu.indexOf(items.access_id) != -1) {
                  self.checkedKeysMap[index].push(items.access_id);
                }
                if (access_id.indexOf(items.access_id) != -1) {
                  self.upCheckedKeysMap[index].push(items.access_id);
                }
                items.children.map((itemThree, indexThree) => {
                  self.data[index].children[indexs].children[indexThree].name = $t(itemThree.name);
                  if (data.data.select_menu.indexOf(itemThree.access_id) != -1) {
                    self.checkedKeysMap[index].push(itemThree.access_id);
                  }
                  if (access_id.indexOf(itemThree.access_id) != -1) {
                    self.upCheckedKeysMap[index].push(itemThree.access_id);
                  }
                  itemThree.children.map((itemFour, indexFour) => {
                    self.data[index].children[indexs].children[indexThree].children[indexFour].name = $t(itemFour.name);
                    if (data.data.select_menu.indexOf(itemFour.access_id) != -1) {
                      self.checkedKeysMap[index].push(itemFour.access_id);
                    }
                    if (access_id.indexOf(itemFour.access_id) != -1) {
                      self.upCheckedKeysMap[index].push(itemFour.access_id);
                    }
                    itemFour.children.map((itemFive, indexFive) => {
                      self.data[index].children[indexs].children[indexThree].children[indexFour].children[indexFive].name = $t(itemFive.name);
                      if (data.data.select_menu.indexOf(itemFive.access_id) != -1) {
                        self.checkedKeysMap[index].push(itemFive.access_id);
                      }
                      if (access_id.indexOf(itemFive.access_id) != -1) {
                        self.upCheckedKeysMap[index].push(itemFive.access_id);
                      }
                      itemFive.children.map((itemSix, indexSix) => {
                        if (data.data.select_menu.indexOf(itemSix.access_id) != -1) {
                          self.checkedKeysMap[index].push(itemSix.access_id);
                        }
                        if (access_id.indexOf(itemSix.access_id) != -1) {
                          self.upCheckedKeysMap[index].push(itemSix.access_id);
                        }
                        self.data[index].children[indexs].children[indexThree].children[indexFour].children[indexFive].children[indexSix].name = $t(itemSix.name);
                      });
                    });
                  });
                });
              });
            });

            if (self.form.parent_id == 0) {
              self.form.parent_id = 0 + '';
            }
            self.loading = false;
          })
          .catch((error) => {
            self.loading = false;
          });
      },

      /*监听选中*/
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
  .basic-setting-content {
  }

  .product-add {
  }

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
