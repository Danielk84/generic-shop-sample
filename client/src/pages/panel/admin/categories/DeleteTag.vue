<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'

const props = defineProps<{
  id: number
}>()

const emits = defineEmits<{
  (e: 'afterDelete'): void
}>()

const notification = useNotificationStore()
const store = useStore()
const router = useRouter()

const { mutateAsync, isPending } = useMutation({
  mutationKey: ['admin-categories-delete', props.id],
  mutationFn: async () => {
    return api.delete(`categories/${props.id}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess() {
    notification.success('Tag deleted')
    emits('afterDelete')
  },
  onError: (error) => {
    errorStatusHandler(error, router, {
      notFound() {
        notification.error('Failed to delete tag.')
        emits('afterDelete')
      },
    })
  },
})
</script>

<template>
  <button
    class="delete-btn c-flex-all-center"
    :class="{ 'c-is-pending': isPending }"
    @click="
      async () => {
        await mutateAsync()
        emits('afterDelete')
      }
    "
  >
    Delete
  </button>
</template>

<style scoped>
@reference "@/styles/index.css";

.delete-btn {
  @apply rounded-2xl hover:brightness-110
    size-full cursor-pointer
    bg-(--c-v-11)
    text-(--c-v-7)
    text-2xl font-bold;
}
</style>
