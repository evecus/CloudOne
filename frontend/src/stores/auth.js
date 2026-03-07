import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token'))

  // token 是否存在（不验证有效性，有效性由后端返回 401 决定）
  const isLoggedIn = computed(() => !!token.value)

  /**
   * 解析 JWT payload，获取过期时间（本地预判，不可替代后端校验）
   */
  function getTokenExpiry() {
    if (!token.value) return null
    try {
      const payload = JSON.parse(atob(token.value.split('.')[1]))
      return payload.exp ? new Date(payload.exp * 1000) : null
    } catch {
      return null
    }
  }

  /**
   * 是否即将过期（距过期不足 1 小时时提示用户，但不强制登出）
   */
  const isTokenExpiringSoon = computed(() => {
    const expiry = getTokenExpiry()
    if (!expiry) return false
    return expiry.getTime() - Date.now() < 60 * 60 * 1000
  })

  /**
   * 本地判断 token 是否已过期（用于页面初始化时快速判断，减少无效请求）
   */
  function isTokenExpired() {
    const expiry = getTokenExpiry()
    if (!expiry) return false
    return Date.now() > expiry.getTime()
  }

  async function fetchUser() {
    if (!token.value) return
    // 本地快速判断，过期则直接登出，不发请求
    if (isTokenExpired()) {
      logout()
      return
    }
    try {
      const { data } = await api.get('/user')
      user.value = data
    } catch {
      // 401 由 api.js 拦截器处理
    }
  }

  async function login(username, password) {
    const { data } = await api.post('/auth/login', { username, password })
    _setSession(data.token, data.user)
  }

  async function setup(username, password) {
    const { data } = await api.post('/auth/setup', { username, password })
    _setSession(data.token, data.user)
  }

  /**
   * 修改密码后后端会下发新 token，前端需要更新存储
   */
  function updateToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  function _setSession(tok, usr) {
    token.value = tok
    user.value = usr
    localStorage.setItem('token', tok)
  }

  return {
    user,
    token,
    isLoggedIn,
    isTokenExpiringSoon,
    isTokenExpired,
    getTokenExpiry,
    login,
    setup,
    logout,
    fetchUser,
    updateToken,
  }
})
