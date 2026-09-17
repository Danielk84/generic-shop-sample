import type { InternalAxiosRequestConfig } from 'axios'

export interface StatusHandlers {
  notAxiosError?: () => void
  badRequest?: () => void
  notFound?: () => void
  unprocessableEntity?: () => void
  tooManyRequests?: () => void
  internalServerError?: () => void
}

export type RetryConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
}
