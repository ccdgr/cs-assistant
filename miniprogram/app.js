const api = require('./utils/api')

App({
  onLaunch () {
    this.autoLogin()
  },

  async autoLogin () {
    if (api.isLoggedIn()) {
      console.log('已有登录态')
      return
    }
    console.log('开始微信静默登录...')
    const result = await api.login()
    if (result.success) {
      console.log('登录成功')
    } else {
      console.error('登录失败:', result.error)
    }
  },

  /** 全局 401 处理 —— 各页面可直接调用 app.reLogin() */
  async reLogin () {
    console.log('401 触发重新登录...')
    const result = await api.reLogin()
    if (result.success) {
      console.log('重新登录成功')
      return true
    }
    console.error('重新登录失败:', result.error)
    return false
  },

  globalData: {
    userInfo: null
  }
})
