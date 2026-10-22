<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useNotificationStore } from '@/store/notification'

const props = defineProps<{
  remainingMinutes: number
}>()

const router = useRouter()
const notification = useNotificationStore()

const remainingSeconds = ref<number>(props.remainingMinutes * 60)
const minutes = computed(() => Math.floor(remainingSeconds.value / 60))
const seconds = computed(() =>
  String(remainingSeconds.value % 60).padStart(2, '0'),
)

let redirectTimer: number | undefined

const timer = setInterval(() => {
  if (remainingSeconds.value === 0) {
    return
  }
  remainingSeconds.value -= 1
  if (remainingSeconds.value === 0) {
    clearInterval(timer)
    notification.error('end of remaining time, back to auth after 20s.')
    redirectTimer = window.setTimeout(() => {
      router.push({ name: 'auth' })
    }, 20000)
  }
}, 1000)

onBeforeUnmount(() => {
  clearInterval(timer)
  if (redirectTimer !== undefined) {
    clearTimeout(redirectTimer)
  }
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
    text-(--c-v-13);
}
</style>
