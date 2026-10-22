<script setup lang="ts">
import { ref, watch, defineAsyncComponent, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuery, useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { useValidator } from '@/utils/validator'
import { errorStatusHandler } from '@/utils/helper'
import type { OrderResponse } from '@/contracts/orders/response.interface'
import {
  OrderUserInfoRequest,
  type OrderUserInfoInput,
} from '@/contracts/orders/request.schema'

const PaymentPage = defineAsyncComponent(
  () => import('@/pages/panel/orders/PaymentPage.vue'),
)

const OrderListView = defineAsyncComponent(
  () => import('@/components/common/orders/OrderListView.vue'),
)

const store = useStore()
const notification = useNotificationStore()
const router = useRouter()
const route = useRoute()

const order_id = String(route.params.id ?? '')

const isEditBoxOpen = ref<boolean>(false)
const isVerifyBoxOpen = ref<boolean>(false)
const isPayBoxOpen = ref<boolean>(false)

const {
  formData: OrderInfoForm,
  errors: OrderInfoErrors,
  validate: OrderInfoValidate,
} = useValidator(OrderUserInfoRequest)

const getOrderQuery = useQuery({
  queryKey: ['user-get-query', order_id],
  queryFn: async () =>
    api.get<OrderResponse>(`orders/${order_id}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  select: (res) => res.data,
})

watch(getOrderQuery.error, (err) => {
  if (err != undefined) {
    errorStatusHandler(err, router)
  }
})

const isEditable = computed(() => {
  const order = getOrderQuery.data.value
  return order !== undefined && !order.is_paid && !order.is_confirmed
})

const hasItems = computed(
  () =>
    Number(getOrderQuery.data.value?.total_bill) > 0 &&
    Number(getOrderQuery.data.value?.items_total) > 0,
)

const isPayable = computed(
  () =>
    isEditable.value &&
    hasItems.value &&
    getOrderQuery.data.value?.is_verified === true,
)

function openEdit() {
  const order = getOrderQuery.data.value
  if (order === undefined) {
    return
  }
  OrderInfoForm.value.address = order.address
  OrderInfoForm.value.zip_code = order.zip_code
  isEditBoxOpen.value = true
}

const setUserInfoMutation = useMutation({
  mutationKey: ['set-user-info-mutation'],
  mutationFn: async (input: OrderUserInfoInput) =>
    api.put(`orders/set-user-info/${order_id}`, input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  async onSuccess() {
    notification.success('User information saved.')
    await getOrderQuery.refetch()
  },
  onError(error) {
    if (error !== undefined) {
      errorStatusHandler(error, router, {
        badRequest() {
          notification.error('Invalid data.')
          return
        },
      })
    }
  },
})

const verifyUserInfoMutation = useMutation({
  mutationKey: ['verify-user-info-mutation'],
  mutationFn: async () =>
    api.put(
      `orders/verify/${order_id}`,
      { accepted: true },
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    ),
  onSuccess() {
    notification.success('Verified.')
    return
  },
  onError(error) {
    if (error !== undefined) {
      errorStatusHandler(error, router, {
        badRequest() {
          notification.error('Invalid request.')
          return
        },
      })
    }
  },
})

async function saveUserInfo() {
  const { input, isValid } = await OrderInfoValidate()
  if (!isValid) {
    notification.error('Invalid data.')
    return
  }
  if (input.data === undefined) {
    notification.error('Unexpected error!?')
    return
  }
  try {
    await setUserInfoMutation.mutateAsync(input.data)
  } catch {
    return
  }
  isEditBoxOpen.value = false
  isVerifyBoxOpen.value = true
}

async function saveVerify() {
  try {
    await verifyUserInfoMutation.mutateAsync()
  } catch {
    return
  }
  await getOrderQuery.refetch()
  isVerifyBoxOpen.value = false
}
</script>

<template>
  <section class="order-page flex items-center gap-5">
    <h1 class="w-full text-center text-2xl sm:text-3xl font-bold">
      Your Order
    </h1>
    <section class="order-info">
      <table class="w-full table-fixed">
        <tr>
          <td>ID:</td>
          <td>{{ getOrderQuery.data.value?.id }}</td>
        </tr>
        <tr>
          <td>Address:</td>
          <td>{{ getOrderQuery.data.value?.address ?? 'Unknown!?' }}</td>
        </tr>
        <tr>
          <td>Zip Code:</td>
          <td>{{ getOrderQuery.data.value?.zip_code ?? 'Unknown!?' }}</td>
        </tr>
        <tr>
          <td>Total Bill:</td>
          <td>{{ getOrderQuery.data.value?.total_bill }}</td>
        </tr>
        <tr>
          <td>Items Total:</td>
          <td>{{ getOrderQuery.data.value?.items_total }}</td>
        </tr>
      </table>
      <div class="status-box">
        <div
          class="is-done"
          :class="[getOrderQuery.data.value?.is_paid ? 'done' : 'not-done']"
        >
          Paid
        </div>
        <div
          class="is-done"
          :class="[getOrderQuery.data.value?.is_verified ? 'done' : 'not-done']"
        >
          Verified
        </div>
        <div
          class="is-done"
          :class="[
            getOrderQuery.data.value?.is_delivered ? 'done' : 'not-done',
          ]"
        >
          Delivered
        </div>
        <div
          class="is-done"
          :class="[
            getOrderQuery.data.value?.is_confirmed ? 'done' : 'not-done',
          ]"
        >
          Confirmed
        </div>
      </div>
      <div v-if="isEditable" class="w-full c-flex-all-center">
        <button class="edit-btn" @click="openEdit">Edit</button>
      </div>
    </section>
    <button
      v-if="isPayable"
      class="c-flex-all-center text-2xl bg-(--c-v-14) text-(--c-v-1) px-10 py-4 cursor-pointer rounded-2xl"
      @click="isPayBoxOpen = true"
    >
      Next
    </button>
    <section class="w-full">
      <OrderListView :orderID="order_id" target="customer" />
    </section>
    <section v-if="isEditBoxOpen" class="c-floating-window c-flex-all-center">
      <div class="c-floating-box c-form-bg">
        <div class="w-full p-2">
          <button
            class="text-2xl font-bold text-(--c-v-11)"
            @click="isEditBoxOpen = false"
          >
            <span>X</span>
          </button>
        </div>
        <form class="c-form">
          <div class="c-form-item">
            <label class="c-form-label" for="address">Address</label>
            <input
              v-model="(OrderInfoForm as OrderUserInfoInput).address"
              class="c-form-input"
              type="text"
            />
            <p
              v-if="OrderInfoErrors['address'] !== undefined"
              class="c-form-error"
            >
              {{ OrderInfoErrors['address'] }}
            </p>
          </div>
          <div class="c-form-item">
            <label class="c-form-label" for="zip_code">Zip Code</label>
            <input
              v-model="(OrderInfoForm as OrderUserInfoInput).zip_code"
              class="c-form-input"
              type="text"
            />
            <p
              v-if="OrderInfoErrors['zip_code'] !== undefined"
              class="c-form-error"
            >
              {{ OrderInfoErrors['zip_code'] }}
            </p>
          </div>
          <button
            class="c-form-btn"
            :disabled="setUserInfoMutation.isPending.value"
            @click="saveUserInfo"
          >
            Save
          </button>
        </form>
      </div>
    </section>
    <section v-if="isVerifyBoxOpen" class="c-floating-window c-flex-all-center">
      <div
        class="c-floating-box c-form-bg p-4 gap-4 c-flex-all-center flex-col"
      >
        <h2 class="text-xl font-bold">Are You Sure?</h2>
        <ul class="w-full">
          <li>Address: {{ getOrderQuery.data.value?.address }}</li>
          <li>Zip Code: {{ getOrderQuery.data.value?.zip_code }}</li>
        </ul>
        <div class="flex flex-wrap justify-between w-full gap-4">
          <button
            type="button"
            @click="saveVerify"
            :disabled="verifyUserInfoMutation.isPending.value"
            class="base-btn verify-btn"
          >
            Verify
          </button>
          <button
            class="base-btn cancel-btn"
            @click="
              () => {
                isVerifyBoxOpen = false
                isEditBoxOpen = true
              }
            "
          >
            Cancel
          </button>
        </div>
      </div>
    </section>
    <PaymentPage v-if="isPayBoxOpen" :order_id="order_id" />
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.order-page {
  @apply w-full min-h-screen flex-col p-4;
}

.order-page .order-info {
  @apply w-90/100 sm:w-80/100 h-fit p-4 overflow-hidden
    border-2 border-(--c-v-9) rounded-xl
    shadow-[0_0_10px_5px] shadow-(color:--c-v-10);
}

.order-page tr {
  @apply flex items-center justify-between flex-wrap gap-4
    text-xl font-bold w-full h-fit
    text-wrap wrap-break-word
    border-b border-(--c-v-10);
}

.order-page td {
  @apply text-wrap wrap-break-word max-w-full;
}

.order-page .status-box {
  @apply w-full h-fit flex gap-5 p-4 items-center justify-center flex-wrap;
}

.order-page .is-done {
  @apply w-30 h-10 text-sm font-bold rounded-2xl
    flex items-center justify-center;
}

.order-page .done {
  @apply bg-(--c-v-14) text-(--c-v-1);
}

.order-page .not-done {
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.order-page .edit-btn {
  @apply rounded-2xl bg-(--c-v-10) text-(--c-v-7) text-2xl font-bold cursor-pointer
    w-full sm:w-50/100 h-20;
}

.order-page li {
  @apply text-wrap wrap-break-word w-full;
}

.order-page .base-btn {
  @apply text-2xl font-bold cursor-pointer hover:brightness-110
    flex items-center justify-center
    w-30 h-15 rounded-2xl;
}

.order-page .verify-btn {
  @apply bg-(--c-v-13) text-(--c-v-1);
}

.order-page .cancel-btn {
  @apply text-(--c-v-11) border-4 border-(--c-v-11);
}
</style>
