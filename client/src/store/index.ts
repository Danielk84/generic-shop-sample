import { defineStore } from 'pinia'
import axios from 'axios'

import api from '@/utils/api'
import { parseJwt } from '@/utils/helper'
import type { AccessTokenResponse } from '@/contracts/auth/response.interface'
import type { AuthClaims } from '@/contracts/auth/response.interface'

interface State {
  user: {
    email: string
    accessToken: string
    isAuth: boolean
  }
  claims: AuthClaims
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
  },
  actions: {
    setEmail(email: string) {
      this.user.email = email
    },
    setAccessToken(token: string) {
      this.user.accessToken = token
      this.setClaims(token)
    },
    setIsAuth(status: boolean) {
      this.user.isAuth = status
    },
    async refreshAccessKey(): Promise<boolean> {
      const res = await api.post<AccessTokenResponse>('auth/refresh')
      if (res.status === axios.HttpStatusCode.Ok) {
        this.user.accessToken = res.data.token
        this.setClaims(res.data.token)
        return true
      }
      return false
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
      this.claims.permission_type = 4
    },
  },
})
