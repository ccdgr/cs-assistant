const api = require('./utils/api')

App({
  onLaunch () {
    // 启动时自动静默登录
    this.autoLogin()
  },

  async autoLogin () {
    // 已有有效 token 则跳过
    if (api.isLoggedIn()) {
      console.log('已有登录态, token:', api.getToken().substring(0, 8) + '...')
      return
    }

    console.log('开始微信静默登录...')
    const result = await api.login()
    if (result.success) {
      console.log('登录成功, token:', result.token.substring(0, 8) + '...')
    } else {
      console.error('登录失败:', result.error)
    }
  },

  globalData: {
    userInfo: null
  }
})
