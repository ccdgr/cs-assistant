const BASE_URL = 'http://localhost:8080/api/v1'

/**
 * 微信静默登录
 * 1. wx.login 获取临时 code
 * 2. POST /auth/login 换取 JWT token（30天有效）
 * 3. 持久化 token
 */
function login () {
  return new Promise((resolve) => {
    wx.login({
      success (res) {
        if (!res.code) {
          resolve({ success: false, error: 'wx.login 未返回 code' })
          return
        }
        wx.request({
          url: `${BASE_URL}/auth/login`,
          method: 'POST',
          header: { 'Content-Type': 'application/json' },
          data: { code: res.code },
          success (resp) {
            const body = resp.data
            if (resp.statusCode === 200 && body.code === 0 && body.data && body.data.token) {
              wx.setStorageSync('token', body.data.token)
              resolve({ success: true, token: body.data.token })
            } else {
              resolve({ success: false, error: body.message || '登录失败' })
            }
          },
          fail (err) {
            resolve({ success: false, error: `网络请求失败: ${err.errMsg}` })
          }
        })
      },
      fail (err) {
        resolve({ success: false, error: `wx.login 失败: ${err.errMsg}` })
      }
    })
  })
}

/** 重新登录（401 时自动调用） */
function reLogin () {
  return login()
}

function getToken () {
  return wx.getStorageSync('token') || null
}

function isLoggedIn () {
  return !!getToken()
}

function logout () {
  wx.removeStorageSync('token')
}

/**
 * 带认证头的通用请求
 * 401 时自动清除 token 并触发重新登录
 */
function authedRequest (options) {
  const token = getToken()
  const header = Object.assign(
    { 'Content-Type': 'application/json' },
    options.header || {},
    token ? { Authorization: `Bearer ${token}` } : {}
  )

  return new Promise((resolve) => {
    wx.request({
      ...options,
      header,
      success (resp) {
        if (resp.statusCode === 401) {
          logout()
          resolve({ success: false, error: '登录已过期', code: 401, needReLogin: true })
        } else {
          resolve({ success: true, data: resp.data })
        }
      },
      fail (err) {
        resolve({ success: false, error: err.errMsg })
      }
    })
  })
}

module.exports = { login, reLogin, getToken, isLoggedIn, logout, authedRequest, BASE_URL }
