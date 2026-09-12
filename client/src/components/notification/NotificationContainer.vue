<script setup lang="ts">
import { useNotificationStore } from '@/store/notification'

const notifications = useNotificationStore()
</script>

<template>
  <div class="notifications-container">
    <TransitionGroup name="notification">
      <div
        v-for="notification in notifications.notifications"
        :key="notification.id"
        class="item c-flex-all-center"
        :class="`item-${notification.type}`"
      >
        <span>{{ notification.message }}</span>

        <button @click="notifications.remove(notification.id)">×</button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.notifications-container {
  @apply fixed w-screen h-fit overflow-hidden
    flex items-center justify-end gap-5;
}

.notifications-container .item {
  @apply rounded-2xl w-100 h-20
    text-ellipsis text-2xl font-bold
    bg-notification-bg
    text-notification-text
    border-4;
}

.notifications-container .item-sucess {
  @apply border-notification-success;
}

.notifications-container .item-error {
  @apply border-notification-error;
}

.notifications-container .item-warning {
  @apply border-notification-warning;
}

.notifications-container .item-info {
  @apply border-notification-info;
}

.notification-enter-active,
.notification-leave-active {
  transition: all 0.5s ease;
}
.notification-enter-from,
.notification-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
