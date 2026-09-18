<script setup lang="ts">
import { useNotificationStore } from '@/store/notification'

const notifications = useNotificationStore()
</script>

<template>
  <div class="notifications-container">
    <TransitionGroup name="notification">
      <div
        v-for="notification in notifications.Notify"
        :key="notification.id"
        class="item c-flex-all-center"
        :class="`item-${notification.type}`"
      >
        <span>{{ notification.message }}</span>

        <button @click="notifications.remove(notification.id)">x</button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.notifications-container {
  @apply fixed z-70 w-screen h-fit overflow-hidden
    flex flex-col items-end gap-5 p-10;
}

.notifications-container .item {
  @apply rounded-2xl w-100 h-20 p-4
    bg-notification-bg
    text-notification-text
    flex items-center justify-between
    border-t-6 border
    hover:brightness-95;
}

.notifications-container span {
  @apply truncate font-bold;
}

.notifications-container button {
  @apply p-2 w-fit h-full flex
    items-start cursor-pointer;
}

.notifications-container .item-success {
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
