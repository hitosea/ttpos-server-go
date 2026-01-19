<template>
  <div class="login-bg" :style="'background-image:url(' + bgimg_url + ');'">
    <div class="login-main">
      <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-position="left" label-width="0px" class="demo-ruleForm login-container d-b-c">
        <div class="flex-1 login-box">
          <h3 class="title title-pr" style="margin-bottom: 16px">TTPOS · {{ $t('后台管理系统') }}</h3>

          <!--用户名-->
          <el-form-item prop="account">
            <div class="left-img-input">
              <img class="l-img" src="@/assets/img/user.svg" />
              <el-input class="l-input" type="text" v-model="ruleForm.account" auto-complete="off" :disabled="logining" @focus="handleFocus" :placeholder="$t('请输入邮箱/手机号')">
              </el-input>
            </div>
          </el-form-item>
          <!--密码-->
          <el-form-item prop="checkPass">
            <div class="left-img-input">
              <img class="l-img" src="@/assets/img/lock.svg" />
              <el-input
                type="password"
                class="l-input"
                v-model="ruleForm.checkPass"
                auto-complete="off"
                :disabled="logining"
                @focus="handleFocus"
                :placeholder="$t('请输入登录密码')"
              >
              </el-input>
            </div>
          </el-form-item>
          <!--  验证码 -->
          <el-form-item prop="verifycode" style="line-height: 0px">
            <div class="flex-1 verifycode">
              <div class="left-img-input" style="max-width: 264px; float: left">
                <el-input
                  v-model="ruleForm.code"
                  ref="code"
                  :disabled="logining"
                  :placeholder="$t('验证码')"
                  @focus="handleFocus"
                  @input="handleInput"
                  class="l-input"
                  style="max-width: 230px"
                ></el-input>
              </div>
              <div class="identifybox" @click="getCode">
                <img v-if="captchaImg && !captchaImgLoading" ref="imgRef" :src="captchaImg" data-time="" />
                <div class="reload-img" v-else>
                  {{ $t('获取验证码') }}
                </div>
              </div>
            </div>
          </el-form-item>
          <el-button type="primary" :disabled="disabledC" style="width: 100%; height: 48px; font-size: 18px; margin-top: 6px" @click.native.prevent="SubmitFunc" :loading="logining"
            >{{ logining ? $t('跳转中...') : $t('登录') }}
          </el-button>

          <!--登录-->
        </div>
      </el-form>
    </div>

    <el-dialog class="" @close="handleClose" v-model="dialogVisible" :close-on-click-modal="false" :close-on-press-escape="false" :title="$t('首次登录需修改密码')">
      <el-form size="small" ref="form" :rules="formRules" :model="form" label-position="top">
        <el-form-item for="no_click" :label="$t('原密码')" prop="old_password">
          <el-input type="password" :placeholder="$t('请输入原密码')" v-model="form.old_password"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('新密码')" prop="password">
          <el-input v-model="form.password" type="password" :placeholder="$t('请输入密码')" @input="changePassword"></el-input>
        </el-form-item>
        <el-form-item for="no_click" :label="$t('确认密码')" prop="confirm_password">
          <el-input v-model="form.confirm_password" type="password" :placeholder="$t('请确认密码')"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleClose">{{ $t('取消') }}</el-button>
          <el-button type="primary" @click="onSubmit" :loading="loading">{{ $t('确定') }}</el-button>
        </span>
      </template>
    </el-dialog>

    <div class="language-box">
      <el-dropdown trigger="click" @command="setLanguage" @visible-change="handleVisibleChange">
        <span class="el-dropdown-link">
          <SvgIcon class="data-box-icon" name="language"></SvgIcon>{{ languageNow }}<el-icon class="el-icon--right"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <template v-for="item in languageList">
              <el-dropdown-item v-if="item.name" :disabled="item.name == languageTag" :command="item.name">
                <div class="language-div"> {{ item.value }}<img v-if="item.name == languageTag" src="../../assets/img/Check.svg" /> </div>
              </el-dropdown-item>
            </template>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script>
  import UserApi from '@/api/user.js';
  import IndexApi from '@/api/index.js';
  import { useUserStore } from '@/store';
  import { getSessionStorage, setSessionStorage } from '@/utils/base.js';
  import { getStorage } from '@/utils/storageData';
  import { createdAuth } from '@/utils/createdAuth.js';
  import { useLockscreenStore } from '@/store/model/lockscreen.js';
  import { languageStore } from '@/store/model/language.js';
  import configObj from '@/config';
  import autoTips from '@/components/autoTips/autoTips.vue';
  import SvgIcon from '@/components/svg-icon/SvgIcon.vue';
  import { EEUIRELOAD } from '@/utils/platform.js';

  const { menu } = configObj;
  const useLockscreen = useLockscreenStore();
  const accredit = useLockscreen.accredit;
  const { afterLogin, computedUserInfo, changeUserInfo } = useUserStore();
  const language = languageStore();
  const languageNow = language.getLanguage().language;
  const cloudBasic = language.getCloudBasic().cloudBasic;
  const isCloudDeploy = languageStore().getIsCloudDeploy().isCloudDeploy;
  const macData = language.getMacData().macData;
  const languageTag = languageStore().language;
  const languageList = language.getLanguageList().languageList;
  const languageListOrigin = language.getLanguageListOrigin().languageListOrigin;
  const userInfo = computedUserInfo().userInfo;
  export default {
    components: {
      SvgIcon,
      autoTips,
    },

    computed: {
      disabledC() {
        return this.ruleForm.account == '' && this.ruleForm.checkPass == '';
      },
    },
    data() {
      // 验证码自定义验证规则
      const validateVerifycode = (rule, value, callback) => {
        if (value === '') {
          this.getCode();
          callback(new Error($t('请输入验证码')));
        } else {
          callback();
        }
      };
      let validatePass1 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入密码')));
        } else if (!/^(?:(?=.*\d)(?=.*[a-zA-Z])|(?=.*\d)(?=.*[\W_])|(?=.*[a-zA-Z])(?=.*[\W_]))(?!.*\s).{8,16}$/.test(value) && value) {
          callback(new Error($t('不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种')));
        } else {
          callback();
        }
      };
      let validatePass2 = (rule, value, callback) => {
        if (!value) {
          callback(new Error($t('请输入确认密码')));
        } else if (value != this.form.password && this.form.password) {
          callback(new Error($t('两次密码不一致！')));
        } else {
          callback();
        }
      };
      return {
        loginForm: {},
        identifyCodes: '1234567890', //验证码的数字库
        identifyCode: '', // 验证码组件传值
        /*是否正在加载*/
        loading: false,
        /*商城名称*/
        shop_name: '',
        /*背景图片*/
        bgimg_url: '',
        log_url: '',
        /*是否正在提交*/
        logining: false,
        /*表单对象*/
        ruleForm: {
          /*用户名*/
          account: '',
          /*密码*/
          checkPass: '',
          code: '',
          AuthorizationCode: '',
        },
        /*验证规则*/
        rules: {
          /*用户名*/
          account: [
            {
              required: true,
              message: $t('请输入用户名'),
              trigger: 'blur',
            },
          ],
          /*密码*/
          checkPass: [
            {
              required: true,
              message: $t('请输入登录密码'),
              trigger: 'blur',
            },
          ],
          /*密码*/
          AuthorizationCode: [
            {
              required: true,
              message: $t('请输入授权码'),
              trigger: 'blur',
            },
          ],
          code: [
            {
              required: true,
              trigger: 'blur',
              validator: validateVerifycode,
            },
          ],
        },
        /*基础配置*/
        baseData: {},
        language: language,
        languageNow: languageNow,
        languageList: languageList,
        languageListOrigin: languageListOrigin,
        languageTag: languageTag,
        accredit: accredit,
        userInfo: userInfo,
        cloudBasic: cloudBasic,
        isCloudDeploy: isCloudDeploy,
        accredit: accredit,
        macData: macData,
        brand: 'TTPOS',
        captchaImg: '',
        captchaImgLoading: false,
        captchaSign: '',
        dialogVisible: false,
        firstToken: '',
        form: {
          old_password: '',
          password: '',
          confirm_password: '',
        },
        formRules: {
          old_password: [
            {
              required: true,
              message: $t('请输入原密码'),
              trigger: 'blur',
            },
          ],
          password: [
            {
              required: true,
              validator: validatePass1,
              trigger: ['blur', 'change'],
            },
          ],
          confirm_password: [
            {
              required: true,
              validator: validatePass2,
              trigger: ['blur', 'change'],
            },
          ],
        },
      };
    },

    created() {},
    mounted() {
      // 清空验证码
      this.identifyCode = '';
      // 获取验证码
      this.getCode();
      // 监听键盘按下事件
      document.addEventListener('keyup', this.onEnter);
      window.addEventListener('focus', this.handleFocus);

      // 如果window.config是一个函数
      if (window.config instanceof Function) {
        // 将函数返回的值赋给this.brand，并将其转换为大写
        this.brand = window.config().brand.toUpperCase();
      }
    },

    destroyed() {
      document.removeEventListener('keyup', this.onEnter);
      window.removeEventListener('focus', this.handleFocus);
    },
    
    methods: {
      onEnter(event) {
        if (event.key === 'Enter' && !this.logining) {
          // 处理回车事件的逻辑
          this.SubmitFunc();
        }
      },

      /*登录方法*/
      SubmitFunc(ev) {
        var _this = this;
        this.$refs.ruleForm.validate((valid) => {
          if (valid) {
            this.logining = true;
            var Params = {
              username: this.ruleForm.account,
              password: this.ruleForm.checkPass,
              code: this.ruleForm.code,
              sign: this.captchaSign,
            };
            /*调用登录接口*/
            UserApi.login(Params, true)
              .then(async (data) => {
                this.ruleForm.code = '';
                this.getCode();
                if (data.code == -102) {
                  _this.logining = false;
                  return;
                }
                if (data.code == -201) {
                  _this.logining = false;
                  this.firstToken = data.data.token;
                  this.dialogVisible = true;
                  return;
                }

                try {
                  // 确保登录状态更新完成
                  await afterLogin(data);
                  console.log('afterLogin 完成');

                  await _this.getBase();
                  console.log('getBase 完成');

                  // 在所有异步操作完成后再验证状态
                  this.$nextTick(() => {
                    const userStore = useUserStore();
                    const currentUser = userStore.userInfo;

                    if (currentUser && currentUser.token) {
                      console.log('登录状态验证成功，用户已正确登录');
                    } else {
                      console.warn('登录流程完成但用户状态不完整，可能存在状态同步问题');
                      // 不阻止跳转，因为getBase已经成功执行
                    }
                  });
                } catch (error) {
                  console.error('登录后处理失败:', error);
                  this.$ElMessage({
                    message: this.$t('登录处理失败，请重试'),
                    type: 'error',
                  });
                  _this.logining = false;
                  this.getCode();
                }
              })
              .catch((error) => {
                //接口调用方法统一处理
                console.error('登录接口调用失败:', error);
                this.getCode();
                this.ruleForm.code = '';
                _this.logining = false;
              });
          }
        });
      },

      /*授权*/
      handleSubmit() {
        var _this = this;
        this.$refs.ruleForm.validate((valid) => {
          if (valid) {
            this.logining = true;
            var Params = {
              auth_code: this.ruleForm.AuthorizationCode,
            };
            /*调用登录接口*/
            UserApi.authCode(Params, true)
              .then((data) => {
                _this.logining = false;
                if (data.code == 1) {
                  this.$ElMessage({
                    message: $t('授权成功'),
                    type: 'success',
                  });
                  setTimeout(() => {
                    EEUIRELOAD();
                  }, 2000);
                } else {
                  this.$ElMessage({
                    message: $t('授权码错误'),
                    type: 'error',
                  });
                }
              })
              .catch((error) => {
                //接口调用方法统一处理
                _this.logining = false;
              });
          }
        });
      },

      // 获取基础信息
      async getBase() {
        this.logining = true;
        IndexApi.base(true)
          .then(async (res) => {
            languageStore().setLanguageList(res.data.language);
            const data = {};
            res.data.language.map((item) => {
              data[item.key] = '';
            });
            languageStore().setLanguageData(data);
            //设置logo
            languageStore().setCloudBasic(res.data.cloudBasic);
            //刷新
            let language = JSON.parse(localStorage.getItem('Language'));
            if (!language) {
              EEUIRELOAD();
            }
            //判断默认语言
            if (language && language.language == '' && language.languageList[0]?.name) {
              languageStore().setLanguage(language.languageList[0]?.name);
            }
            /*获取基础配置*/
            const dataInfo = {
              data: {
                shop_name: res.data.settings.shop_name,
                logoUrl: res.data.settings.shop_bg_img,
                is_open_tax: res.data.settings.is_open_tax,
              },
            };
            //设置授权数据
            setSessionStorage('supplier', res.data.supplier);
            // 设置erp数据
            setSessionStorage('erp', res.data.erp);
            //
            await changeUserInfo(dataInfo);
            let auth = getSessionStorage('authlist');
            let authlist = {};
            auth = getStorage(menu);
            createdAuth(auth, authlist);
            setSessionStorage('authlist', authlist);
            auth = authlist;
            //获取完再跳转
            try {
              // 清理事件监听器
              document.removeEventListener('keyup', this.onEnter);
              window.removeEventListener('focus', this.handleFocus);

              // 确保 supplier 数据已设置
              const supplier = getSessionStorage('supplier');
              if (!supplier || !supplier.app_id) {
                console.error('supplier 数据未正确设置，延迟重试');
                setTimeout(() => {
                  this.getBase();
                }, 500);
                return;
              }

              const app_id = supplier.app_id;
              
              // 构建带app_id的路径（hash 路由模式下，Vue Router 会自动添加 #）
              const homePath = `/${app_id}/home`;

              // 等待路由注册完成（dealWithRoute 是异步的）
              // 使用 nextTick 确保路由状态已更新
              await this.$nextTick();

              // 等待路由注册完成（dealWithRoute 在路由守卫中异步执行）
              // 给路由注册一些时间
              await new Promise((resolve) => setTimeout(resolve, 300));

              // 先尝试使用 router.push
              try {
                // 使用 replace 避免在历史记录中留下登录页
                await this.$router.push({ path: homePath, replace: true });
                
                // 等待路由跳转完成并验证
                await new Promise((resolve) => setTimeout(resolve, 300));
                
                // 验证路由是否成功跳转
                const currentPath = this.$route.path;
                if (currentPath && currentPath.includes('/home')) {
                  // 路由跳转成功，延迟执行 reload 以确保路由状态已稳定
                  this.logining = false;
                  // 使用 setTimeout 延迟 reload，确保路由跳转完全完成
                  setTimeout(() => {
                    // 再次验证路径，确保不会在错误的路径上刷新
                    const finalPath = window.location.hash || window.location.pathname;
                    if (finalPath.includes('/home') || finalPath.includes(app_id)) {
                      location.reload();
                    } else {
                      // 如果路径不对，使用 window.location.href 强制跳转
                      window.location.href = `#${homePath}`;
                    }
                  }, 100);
                  return;
                } else {
                  throw new Error(`路由跳转失败，当前路径: ${currentPath}`);
                }
              } catch (routerError) {
                console.warn('路由跳转失败，使用 window.location.href:', routerError);
                // 路由跳转失败，使用 window.location.href 强制跳转（hash 路由需要加 #）
                this.logining = false;
                // 使用完整的 URL，确保跳转正确
                const hashPath = `#${homePath}`;
                window.location.href = hashPath;
                return;
              }
            } catch (error) {
              console.error('跳转过程出错:', error);
              // 出错时直接使用强制跳转
              const supplier = getSessionStorage('supplier');
              const app_id = supplier?.app_id;
              if (app_id) {
                this.logining = false;
                window.location.href = `#/${app_id}/home`;
              } else {
                this.logining = false;
                console.error('无法获取 app_id，跳转失败');
              }
            }
          })
          .catch((error) => {
            this.logining = false;
          });
      },

      // 切换语言
      setLanguage(e) {
        if (e == this.languageTag) return;
        ElMessageBox.confirm($t('切换语言需要刷新后生效，是否确定刷新?'), $t('提示'), {
          confirmButtonText: $t('确定'),
          cancelButtonText: $t('取消'),
          type: 'warning',
        })
          .then(() => {
            this.language.setLanguage(e);
            EEUIRELOAD();
          })
          .catch(() => {
            this.$ElMessage({
              type: 'info',
              message: $t('已取消'),
            });
          });
      },

      handleVisibleChange(visible) {
        if (visible && this.languageList.length === 0) {
          languageStore().setLanguageList(this.languageListOrigin);
        }
      },

      handleFocus() {
        // 获取当前时间戳
        let timestamp = new Date().getTime();
        // Check if imgRef exists before accessing its properties
        if (this.$refs.imgRef && this.$refs.imgRef.getAttribute('data-time') && timestamp - Number(this.$refs.imgRef.getAttribute('data-time')) > 1500 * 1000) {
          // 如果大于1500s，重新获取验证码
          this.getCode();
        }
      },

      handleInput() {
        this.$nextTick(() => {
          //过滤验证码中的空间符号
          this.ruleForm.code = this.ruleForm.code.replace(/\s/g, '');
        });
      },

      // 异步获取验证码
      getCode(e) {
        // 定义表单数据
        if (e == 'return') return;
        let form = { v: 1 };
        // 调用接口获取验证码
        this.captchaImgLoading = true;
        IndexApi.getCaptcha(form, true)
          .then((data) => {
            // 将验证码签名和图片赋值给组件的data
            this.captchaSign = data.data.sign;
            this.captchaImg = data.data.base64;
            this.captchaImgLoading = false;
            this.$nextTick(() => {
              if (this.$refs.imgRef) {
                //当前时间戳放入data-time
                this.$refs.imgRef.setAttribute('data-time', new Date().getTime());
              }
            });
          })
          .catch((error) => {
            this.captchaImgLoading = false;
          });
      },

      // 复制数据
      copyData(e) {
        // 调用$copyText方法复制数据
        this.$copyText(e).then(
          () => {
            // 复制成功，显示成功提示
            this.$ElMessage({
              message: $t('复制成功'),
              type: 'success',
            });
          },
          () => {
            // 复制失败，显示失败提示
            this.$ElMessage({
              message: $t('复制失败'),
              type: 'error',
            });
          }
        );
      },

      onSubmit() {
        // 定义一个变量self，指向当前组件实例
        let self = this;
        // 获取表单数据
        let form = self.form;
        // 将firstToken赋值给form的token属性
        form.token = this.firstToken;

        // 对表单进行验证
        self.$refs.form.validate((valid) => {
          // 如果验证通过
          if (valid) {
            // 设置loading状态为true
            self.loading = true;
            // 调用IndexApi的saasEditPassword方法，传入表单数据和true
            IndexApi.saasEditPassword(form, true)
              // 如果请求成功
              .then((data) => {
                // 设置loading状态为false
                self.loading = false;
                // 弹出保存成功的提示框
                this.$ElMessage({
                  message: $t('保存成功'),
                  type: 'success',
                });
                // 关闭对话框
                this.handleClose();
              })
              // 如果请求失败
              .catch((error) => {
                // 设置loading状态为false
                self.loading = false;
              });
          }
        });
      },

      // 改变密码时，对确认密码进行验证
      changePassword() {
        if (this.form.confirm_password) {
          this.$refs.form.validateField('confirm_password');
        }
      },

      // 关闭对话框时，清空表单数据，并重新获取验证码
      handleClose() {
        this.form = {
          old_password: '',
          password: '',
          confirm_password: '',
        };
        this.dialogVisible = false;
        this.getCode();
      },
    },
  };
</script>

<style lang="scss" scoped>
  .login-bg {
    width: 100%;
    min-height: 100%;
    background: no-repeat;
    background-color: #fff6de;
    background-position: right bottom;
    display: -webkit-box;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-orient: vertical;
    -webkit-box-direction: normal;
    -ms-flex-direction: column;
    flex-direction: column;
    -webkit-box-align: center;
    -ms-flex-align: center;
    align-items: center;
    justify-content: center;
  }

  .login-main {
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .login-container {
    -webkit-border-radius: 16px;
    border-radius: 16px;
    -moz-border-radius: 16px;
    background-clip: padding-box;
    width: 510px;
    margin: auto;
    background-color: #ffffff;

    .title {
      margin: 0px auto 40px auto;
      text-align: center;
      font-size: 28px;
      font-family: Microsoft YaHei;
      font-weight: bold;
      color: #333333;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;

      img {
        width: 40px;
      }
    }
    .title-pr {
      color: var(--el-color-primary);
    }
    .remember {
      margin: 0px 0px 35px 0px;
    }
  }

  .log_img {
    img {
      width: 514px;
      height: 408px;
    }
  }

  .login-box {
    padding: 56px 40px;
    box-sizing: border-box;
    border-radius: 16px;
    width: 100%;
  }

  .left-img-input {
    width: 100%;
    height: 46px;
    line-height: 46px;
    background: #ffffff;
    border: 1px solid #eeeeee;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 0 16px;
    border-radius: 4px;

    .l-img {
      width: 20px;
      height: 20px;
      margin-right: 5px;
      flex-shrink: 0;
    }

    .l-input {
      flex: 1;
      border: none;
      background: none;
      font-size: 14px;
      color: #666666;

      :deep(.el-input__wrapper) {
        box-shadow: none;
        border: none;
      }
    }

    .ll-input {
      :deep(.el-input__wrapper) {
        padding: 0px;
      }
    }

    .el-input__inner {
      border: none;
      padding: 0;
    }
  }

  .mac-box {
    width: 100%;
    overflow: hidden;
  }

  .left-img-mac {
    display: flex;
    margin-bottom: 8px;
    overflow: hidden;
    width: 100%;
    background: #f3f5f6;
    border: none;
    height: 35px;
    line-height: 35px;
    p {
      color: var(--el-color-black);
      flex: 0 0 auto;
      font-size: 14px;
      flex-shrink: 0;
      overflow: hidden;
    }

    .p1 {
      flex: 1 1 auto;
      position: relative;
      margin-right: 8px;
    }

    img {
      width: 16px;
      margin-left: auto;
      cursor: pointer;
    }
  }
  .left-img-mac:nth-child(2) {
    margin-bottom: 16px;
  }
  .language-box {
    position: fixed;
    margin: auto;
    left: 0;
    right: 0;
    bottom: 32px;
    display: flex;
    align-items: center;
    justify-content: center;

    .el-dropdown-link {
      color: var(--el-color-black);
      font-size: 14px;
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
    }

    .data-box-icon {
      width: 16px;
      height: 16px;
      color: var(--el-color-black);
    }
  }

  :deep(.language-div) {
    width: 90px;
    display: flex;
    align-items: center;
    justify-content: space-between;

    img {
      width: 18px;
    }
  }

  .verifycode {
    overflow: hidden;

    .identifybox {
      float: left;
      position: relative;
      height: 46px;
      width: 165px;
      img {
        display: block;
        width: 100%;
        height: 100%;
        position: absolute;
        left: 0;
        top: 0;
        right: 0;
        bottom: 0;
      }
      .reload-img {
        background: var(--el-border-color);
        display: flex;
        width: 100%;
        height: 100%;
        position: absolute;
        left: 0;
        top: 0;
        right: 0;
        bottom: 0;
        justify-content: center;
        align-items: center;
        cursor: pointer;
      }
    }
  }
</style>
