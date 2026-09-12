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
  }
  claims: AuthClaims
}

export const useStore = defineStore('store', {
  state: (): State => ({
    user: {
      email: '',
      accessToken: '',
    },
    claims: {
      id: '',
      email: '',
      permission_type: 4, // blocked user
    },
  }),
  getters: {
    getEmail: (state) => state.user.email,
    getAccessToken: (state) => {
      return state.user.accessToken
    },
    hasAccessToken: (state) => {
      console.log(state.user.accessToken !== '')
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
        this.claims.permission_type = 4
      } else {
        this.claims = claims
      }
    },
  },
})
