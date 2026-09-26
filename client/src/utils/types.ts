import type { InternalAxiosRequestConfig } from 'axios'

export interface StatusHandlers {
  notAxiosError?: () => void
  badRequest?: () => void
  unauthorized?: () => void
  forbidden?: () => void
  notFound?: () => void
  contentTooLarge?: () => void
  unprocessableEntity?: () => void
  tooManyRequests?: () => void
  internalServerError?: () => void
}

export type RetryConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
}
