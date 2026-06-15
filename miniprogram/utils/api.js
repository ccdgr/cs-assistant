// api.js — 微信小程序 API 工具层
const BASE_URL = 'http://localhost:8080/api/v1'

/**
 * 微信静默登录
 * 1. 调用 wx.login 获取临时 code
 * 2. 发送 code 到后端换取 token
 * 3. 持久化 token 到本地存储
 *
 * @returns {Promise<{success: boolean, token?: string, error?: string}>}
 */
function login () {
  return new Promise((resolve, reject) => {
    // Step 1: 获取微信临时凭证
    wx.login({
      success (res) {
        if (!res.code) {
          resolve({ success: false, error: 'wx.login 未返回 code' })
          return
        }

        // Step 2: 发送 code 到后端
        wx.request({
          url: `${BASE_URL}/auth/login`,
          method: 'POST',
          header: { 'Content-Type': 'application/json' },
          data: { code: res.code },
          success (resp) {
            const body = resp.data
            if (resp.statusCode === 200 && body.code === 0 && body.data && body.data.token) {
              const token = body.data.token
              // Step 3: 持久化 token
              wx.setStorageSync('token', token)
              wx.setStorageSync('expires_at', body.data.expires_at)
              resolve({ success: true, token })
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

/**
 * 读取本地存储的 token
 * @returns {string|null}
 */
function getToken () {
  return wx.getStorageSync('token') || null
}

/**
 * 检查是否已登录 (token 是否存在)
 * @returns {boolean}
 */
function isLoggedIn () {
  return !!getToken()
}

/**
 * 清除登录态
 */
function logout () {
  wx.removeStorageSync('token')
  wx.removeStorageSync('expires_at')
}

/**
 * 带认证头的通用请求
 * @param {object} options - 同 wx.request，自动注入 Authorization
 */
function authedRequest (options) {
  const token = getToken()
  const header = Object.assign(
    { 'Content-Type': 'application/json' },
    options.header || {},
    token ? { Authorization: `Bearer ${token}` } : {}
  )

  return new Promise((resolve, reject) => {
    wx.request({
      ...options,
      header,
      success (resp) {
        if (resp.statusCode === 401) {
          logout() // token 过期，清除
          resolve({ success: false, error: '登录已过期，请重新登录', code: 401 })
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

module.exports = { login, getToken, isLoggedIn, logout, authedRequest, BASE_URL }
