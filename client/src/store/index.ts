import { defineStore } from 'pinia'
import axios from 'axios'

import api from '@/utils/api'
import { parseJwt } from '@/utils/helper'
import type { AccessTokenResponse } from '@/contracts/auth/response.interface'
import type { AuthClaims } from '@/contracts/auth/response.interface'

// Shared by every caller so concurrent 401s trigger a single refresh request.
let refreshPromise: Promise<boolean> | null = null

interface State {
  user: {
    email: string
    accessToken: string
    isAuth: boolean
  }
  claims: AuthClaims
  theme: {
    isDark: boolean
  }
  loading: boolean
}

export const useStore = defineStore('store', {
  state: (): State => ({
    user: {
      email: '',
      accessToken: '',
      isAuth: false,
    },
    claims: {
      id: '',
      email: '',
      permission_type: 3,
    },
    theme: {
      isDark: false,
    },
    loading: false,
  }),
  persist: {
    storage: sessionStorage,
  },
  getters: {
    getEmail: (state) => state.user.email,
    getAccessToken: (state) => {
      return 'Bearer ' + state.user.accessToken
    },
    hasAccessToken: (state) => {
      return state.user.accessToken !== ''
    },
    getClaims: (state) => {
      return state.claims
    },
    isDarkTheme: (state) => {
      return state.theme.isDark
    },
  },
  actions: {
    setEmail(email: string) {
      this.user.email = email
    },
    setAccessToken(token: string) {
      this.user.accessToken = token
      this.setClaims(token)
      this.setIsAuth(true)
    },
    setIsAuth(status: boolean) {
      this.user.isAuth = status
    },
    async refreshAccessKey(): Promise<boolean> {
      if (refreshPromise !== null) {
        return refreshPromise
      }
      refreshPromise = (async () => {
        try {
          const res = await api.post<AccessTokenResponse>('auth/refresh')
          if (res.status === axios.HttpStatusCode.Ok && res.data.token) {
            this.setAccessToken(res.data.token)
            return true
          }
          return false
        } catch {
          return false
        } finally {
          refreshPromise = null
        }
      })()
      return refreshPromise
    },
    setClaims(token: string) {
      const claims = parseJwt<AuthClaims>(token)
      if (claims === undefined) {
        this.claims.id = ''
        this.claims.email = ''
        this.claims.permission_type = 3
      } else {
        this.claims = claims
      }
    },
    setLoadingStatus(status: boolean) {
      this.loading = status
    },
    logout() {
      this.user.email = ''
      this.user.accessToken = ''
      this.user.isAuth = false
      this.claims.id = ''
      this.claims.email = ''
      this.claims.permission_type = 3
    },
    setDarkTheme(status: boolean) {
      this.theme.isDark = status
    },
  },
})
