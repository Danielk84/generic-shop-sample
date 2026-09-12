import type { InternalAxiosRequestConfig } from 'axios'

export interface StatusHandlers {
  notAxiosError?: () => void
  badRequest?: () => void
  notFound?: () => void
  unprocessableEntity?: () => void
}

export type RetryConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
}
