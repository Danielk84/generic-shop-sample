<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'

import { useNotificationStore } from '@/store/notification'

const props = defineProps<{
  remainingMinutes: number
}>()

const router = useRouter()
const notification = useNotificationStore()

const minutes = ref<number>(props.remainingMinutes - 1)
const seconds = ref<number>(59)

const timer = setInterval(() => {
  if (seconds.value > 0) {
    seconds.value -= 1
  } else {
    if (minutes.value > 0) {
      minutes.value -= 1
      seconds.value = 59
    }
  }
}, 1000)

watch(
  () => [minutes.value, seconds.value],
  (v) => {
    if (v[0] === 0 && v[1] === 0) {
      clearInterval(timer)
      notification.error('end of remaining time, back to auth after 20s.')
      setTimeout(() => {
        router.push({ name: 'auth' })
      }, 20000)
    }
  },
  {
    immediate: true,
  },
)

onBeforeUnmount(() => {
  clearInterval(timer)
})
</script>

<template>
  <div class="timer c-flex-all-center">
    <span>{{ minutes }}</span>
    <span> : </span>
    <span>{{ seconds }}</span>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.timer {
  @apply w-fit h-fit text-2xl font-bold gap-5 underline
    text-timer-text;
}
</style>
