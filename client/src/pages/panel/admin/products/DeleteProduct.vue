<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'

const props = defineProps<{
  id: string
}>()

const store = useStore()
const router = useRouter()
const notification = useNotificationStore()
const emit = defineEmits<{
  (event: 'success'): void
}>()

const { mutate, isPending } = useMutation({
  mutationFn: async () => {
    return api.delete(`products/${props.id}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {
    notification.success('Product deleted.')
    emit('success')
  },
  onError: (error) => {
    errorStatusHandler(error, router)
  },
})
</script>

<template>
  <button :class="{ 'is-pending': isPending }" @click="mutate()">Delete</button>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
