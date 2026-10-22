<script setup lang="ts">
import { ref, defineAsyncComponent, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { errorStatusHandler } from '@/utils/helper'

const router = useRouter()
const route = useRoute()

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const Authority = route.query['Authority']
const Status = route.query['Status']

const isSuccessPayment = ref<boolean | null>(null)

const verifyPaymentMutation = useMutation({
  mutationKey: ['user-verify-payment'],
  // The callback consumes the authority, so retrying cannot succeed.
  retry: false,
  mutationFn: async () =>
    api.get('payment/callback', {
      params: {
        Authority: Authority,
        Status: Status,
      },
    }),
  onSuccess() {
    isSuccessPayment.value = true
  },
  onError(err) {
    isSuccessPayment.value = false
    errorStatusHandler(err, router, {
      notFound() {
        return
      },
      badRequest() {
        return
      },
    })
  },
})

onMounted(async () => {
  try {
    await verifyPaymentMutation.mutateAsync()
  } catch {
    // The mutation's onError handler already updates the page state.
  }
})
</script>

<template>
  <section class="verify-payment-page c-flex-all-center">
    <div
      class="payment-status-box c-flex-all-center"
      :class="{
        'success-border': isSuccessPayment === true,
        'failed-border': isSuccessPayment === false,
        'in-process-border': isSuccessPayment === null,
      }"
    >
      <div v-if="isSuccessPayment === true" class="status-box">
        <BaseIcon
          :icon="icons.pages.panel.orders.verified"
          size="120px"
          strokeColor="--c-v-14"
          strokeDarkColor="--c-v-14"
          fillColor="--c-v-14"
          fillDarkColor="--c-v-14"
        />
        <RouterLink
          class="back-url callback-text"
          :to="{ name: 'orders-list' }"
        >
          Successfully Verified, Back To Orders.
        </RouterLink>
      </div>
      <div v-else-if="isSuccessPayment === null" class="status-box">
        <BaseIcon
          class="animate-spin [animation-direction:reverse] [animation-duration:5s]"
          :icon="icons.pages.panel.orders.inProcess"
          size="120px"
          strokeColor="--c-v-13"
          strokeDarkColor="--c-v-13"
          fillColor="--c-v-13"
          fillDarkColor="--c-v-13"
        />
        <h1 class="callback-text">Payment Verifying...</h1>
      </div>
      <div v-else class="status-box">
        <BaseIcon
          :icon="icons.pages.panel.orders.notVerified"
          size="120px"
          strokeColor="--c-v-11"
          strokeDarkColor="--c-v-11"
          fillColor="--c-v-11"
          fillDarkColor="--c-v-11"
        />
        <RouterLink class="back-url callback-text" to="/">
          Not Verified, Back To Home!
        </RouterLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.verify-payment-page {
  @apply w-full h-screen flex-col;
}

.verify-payment-page .payment-status-box {
  @apply rounded-2xl h-70/100 w-90/100 sm:w-70/100 flex-col
    border-4 shadow-[0_0_10px_5px] -mt-40;
}

.verify-payment-page .success-border {
  @apply border-(--c-v-14) shadow-(color:--c-v-14);
}

.verify-payment-page .failed-border {
  @apply border-(--c-v-11) shadow-(color:--c-v-11);
}

.verify-payment-page .in-process-border {
  @apply border-(--c-v-10) shadow-(color:--c-v-10);
}

.verify-payment-page .callback-text {
  @apply text-2xl font-bold text-wrap w-full text-center;
}

.verify-payment-page .status-box {
  @apply flex flex-col items-center justify-center gap-10
    size-full;
}

.verify-payment-page .back-url {
  @apply underline text-(--c-v-9);
}
</style>
