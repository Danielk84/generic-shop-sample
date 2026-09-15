import { type Router } from 'vue-router'
import axios from 'axios'

import type { StatusHandlers } from '@/utils/types'

export const getCssVar = (cssVar: string) =>
  cssVar === 'none' ? 'none' : `var(${cssVar})`

export const getCssUrl = (cssUrl?: string) =>
  cssUrl ? `url(${cssUrl})` : undefined

export function* range(start: number, end: number, step: number = 1) {
  for (let i = start; i < end; i += step) {
    yield i
  }
}

export function errorStatusHandler(
  err: Error | null,
  router: Router,
  handlers?: StatusHandlers,
) {
  if (err === null) {
    return
  }
  if (!axios.isAxiosError(err)) {
    if (typeof handlers?.notAxiosError === 'function') {
      handlers.notAxiosError()
    }
    return
  }

  switch (err.request.status) {
    case axios.HttpStatusCode.BadRequest:
      if (typeof handlers?.badRequest === 'function') {
        handlers.badRequest()
      } else {
        router.push('/')
      }
      break
    case axios.HttpStatusCode.Unauthorized:
      router.push('/')
      break
    case axios.HttpStatusCode.Forbidden:
      router.push('/')
      break
    case axios.HttpStatusCode.NotFound:
      if (typeof handlers?.notFound === 'function') {
        handlers.notFound()
      } else {
        router.push('/')
      }
      break
    case axios.HttpStatusCode.UnprocessableContent:
      if (typeof handlers?.unprocessableEntity === 'function') {
        handlers.unprocessableEntity()
      } else {
        router.push('/')
      }
  }
}

export function randKey() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2)
}

const ignoreCallbackUrl: string[] = [
  '/404',
  '/401',
  '/403',
  '/422',
  '/register',
  '/login',
] as const

export function setCallbackURL(fullPath: string): string {
  for (let i = 0; i < ignoreCallbackUrl.length; i++) {
    if (fullPath.includes(ignoreCallbackUrl[i])) {
      fullPath = ''
      break
    }
  }
  return encodeURIComponent(fullPath)
}

export function parseJwt<T>(token: string): T | undefined {
  if (!token) {
    return
  }
  const base64Url = token.split('.')[1]
  const base64 = base64Url.replace('-', '+').replace('_', '/')
  return JSON.parse(window.atob(base64)) as T
}

export function formatDate(input: string) {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(input))
}

export function useTimer(
  fn: () => void,
  timout?: number,
) {
  let timer: number | undefined = undefined;
  return {
    start() {
      if (timer === undefined) {
        timer = setInterval(fn, timout)
      }
    },
    end() {
      clearInterval(timer)
      timer = undefined
    }
  }
}