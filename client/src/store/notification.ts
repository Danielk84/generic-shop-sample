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
  actions: {
    show(message: string, type: NotificationType = 'info', duration = 3000) {
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
    },

    success(message: string, duration = 3000) {
      this.show(message, 'success', duration)
    },

    error(message: string, duration = 5000) {
      this.show(message, 'error', duration)
    },

    warning(message: string, duration = 4000) {
      this.show(message, 'warning', duration)
    },

    info(message: string, duration = 3000) {
      this.show(message, 'info', duration)
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
