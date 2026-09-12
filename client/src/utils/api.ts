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

    if (
      error.response?.status !== 401 ||
      !originalRequest ||
      originalRequest._retry ||
      originalRequest.url?.includes('auth/refresh')
    ) {
      return Promise.reject(error)
    }

    originalRequest._retry = true

    const store = useStore()
    if (!(await store.refreshAccessKey())) {
      return Promise.reject(error)
    }

    const headers = AxiosHeaders.from(originalRequest.headers)
    headers.set('Authorization', `Bearer ${store.getAccessToken}`)

    originalRequest.headers = headers

    return api(originalRequest)
  },
)

export default api
