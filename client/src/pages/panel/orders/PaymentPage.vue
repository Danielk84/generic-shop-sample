<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'
import { errorStatusHandler } from '@/utils/helper'

import api from '@/utils/api'
import { useStore } from '@/store'

const props = defineProps<{
  order_id: string
}>()

const store = useStore()
const router = useRouter()

const initPaymentMutation = useMutation({
  mutationKey: ['user-init-payment'],
  retry: false,
  mutationFn: async () =>
    (
      await api.post<{
        payment_url: string
      }>(`payment/${props.order_id}`, null, {
        headers: {
          Authorization: store.getAccessToken,
        },
      })
    ).data,
  onSuccess(data) {
    window.location.assign(data.payment_url)
  },
  onError(error) {
    if (error !== undefined) {
      errorStatusHandler(error, router)
    }
  },
})
</script>

<template>
  <section class="c-floating-window c-flex-all-center bg-(--c-v-1)/70">
    <button
      class="text-2xl font-bold bg-(--c-v-15) text-(--c-v-7) py-4 px-6 rounded-2xl hover:brightness-120"
      @click="
        () => {
          initPaymentMutation.mutate()
        }
      "
    >
      Let's Pay
    </button>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
