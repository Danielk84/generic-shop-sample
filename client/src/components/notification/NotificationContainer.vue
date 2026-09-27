<script setup lang="ts">
import { useNotificationStore } from '@/store/notification'

const notifications = useNotificationStore()
</script>

<template>
  <div v-if="notifications.Notify.length !== 0" class="notifications-container">
    <TransitionGroup name="notification">
      <div
        v-for="notification in notifications.Notify"
        :key="notification.id"
        class="item c-flex-all-center"
        :class="`item-${notification.type}`"
      >
        <button @click="notifications.remove(notification.id)">x</button>
        <p>{{ notification.message }}</p>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.notifications-container {
  @apply fixed z-70 w-fit h-fit overflow-hidden
    flex flex-col items-end gap-5 p-10;
}

.notifications-container .item {
  @apply rounded-2xl w-100 min-h-20 max-h-30 p-2
    bg-(--c-v-6) dark:bg-(--c-v-3)
    flex items-center justify-between
    border-t-6 border
    hover:brightness-95;
}

.notifications-container p {
  overflow: hidden;
  display: -webkit-box;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2; /* number of lines to show */
  line-clamp: 2;
  -webkit-box-orient: vertical;
  font: bold;
  text-wrap: balance;
}

.notifications-container button {
  @apply p-2 w-fit h-full flex
    items-center justify-center cursor-pointer;
}

.notifications-container .item-success {
  @apply border-(--nord-aurora-4);
}

.notifications-container .item-error {
  @apply border-(--c-v-11);
}

.notifications-container .item-warning {
  @apply border-(--c-v-12);
}

.notifications-container .item-info {
  @apply border-(--c-v-15);
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
