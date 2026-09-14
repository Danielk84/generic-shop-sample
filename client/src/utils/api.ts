import axios, { AxiosError, AxiosHeaders } from 'axios'

import { useStore } from '@/store'
import type { RetryConfig } from '@/utils/types'

const api = axios.create({
  baseURL: `${import.meta.env.VITE_BACKEND_URL}/api/`,
  withCredentials: true,
})

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as RetryConfig | undefined

    const store = useStore()
    if (
      error.response?.status !== 401 ||
      !originalRequest ||
      originalRequest._retry ||
      originalRequest.url?.includes('auth/refresh')
    ) {
      store.setIsAuth(false)
      return Promise.reject(error)
    }

    originalRequest._retry = true

    if (!(await store.refreshAccessKey())) {
      store.setIsAuth(false)
      return Promise.reject(error)
    }

    const headers = AxiosHeaders.from(originalRequest.headers)
    headers.set('Authorization', `${store.getAccessToken}`)

    originalRequest.headers = headers

    store.setIsAuth(true)
    return api(originalRequest)
  },
)

export default api
