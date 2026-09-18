import { defineStore } from 'pinia'

import { randKey } from '@/utils/helper'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  message: string
  duration?: number
}

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    notifications: [] as Notification[],
  }),
  persist: {
    storage: sessionStorage,
  },
  getters: {
    Notify(): Notification[] {
      return this.notifications.reverse()
    },
  },
  actions: {
    show(
      message: string,
      type: NotificationType = 'info',
      duration = 3000,
    ): string {
      const id = randKey()
      this.notifications.push({
        id,
        type,
        message,
        duration,
      })

      if (duration > 0) {
        setTimeout(() => {
          this.remove(id)
        }, duration)
      }
      return id
    },

    success(message: string, duration = 3000): string {
      return this.show(message, 'success', duration)
    },

    error(message: string, duration = 5000): string {
      return this.show(message, 'error', duration)
    },

    warning(message: string, duration = 4000): string {
      return this.show(message, 'warning', duration)
    },

    info(message: string, duration = 3000): string {
      return this.show(message, 'info', duration)
    },

    remove(id: string) {
      this.notifications = this.notifications.filter(
        (notification) => notification.id !== id,
      )
    },

    clear() {
      this.notifications = []
    },
  },
})
