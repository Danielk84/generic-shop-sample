<script setup lang="ts">
import { ref, watch, toRaw } from 'vue'
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { useValidator } from '@/utils/validator'
import { deepEqual, errorStatusHandler } from '@/utils/helper'
import type { ProductResponse } from '@/contracts/products/response.interface'
import {
  ProductVariantDetailRequest,
  UpdateProductVendor,
  type ProductVariantDetailInput,
  type ProductVendorInput,
  type UpdateProductVendorInput,
} from '@/contracts/products/request.schema'
import type { ProductProperty } from '@/contracts/products/response.interface'

const store = useStore()
const notification = useNotificationStore()
const router = useRouter()

const props = defineProps<{ data: ProductResponse }>()
const emits = defineEmits<{
  (e: 'afterMutation'): void
}>()

const isSetVariantOpen = ref<boolean>(false)

const vendorProperty = ref<ProductProperty | null>(null)
const vendorEdit = ref<ProductVendorInput | null>(null)

const variantDetail = {
  state: ref(structuredClone(toRaw(props.data?.variant_detail ?? []))),
  add(property: ProductProperty) {
    if (deepEqual(property, {})) {
      return
    }
    for (let i = 0; i < this.state.value.length; ++i) {
      if (deepEqual(this.state.value[i].property, property)) {
        notification.warning('property already exists.')
        return
      }
    }
    this.state.value.push({
      property: property,
      price: 0,
      vendors: [],
    })
  },
  remove(property: ProductProperty) {
    for (let i = 0; i < this.state.value.length; ++i) {
      if (deepEqual(this.state.value[i].property, property)) {
        this.state.value.splice(i, 1)
      }
    }
  },
}

const createProperty = {
  state: ref<
    Array<{
      key: string
      value: string
    }>
  >([]),
  get(): ProductProperty {
    let property: ProductProperty = {}
    this.state.value.forEach((v) => {
      property[v.key] = v.value
    })
    return property
  },
  add() {
    this.state.value.push({
      key: '',
      value: '',
    })
  },
  remove(index: number) {
    this.state.value.splice(index, 1)
  },
  clear() {
    this.state.value = []
  },
}

const setVendorValidator = useValidator(UpdateProductVendor)
const setVariantDetailValidator = useValidator(ProductVariantDetailRequest)

watch(vendorEdit, (v) => {
  if (v === null) {
    return null
  }
  setVendorValidator.formData.value.user_id = v.user_id
  setVendorValidator.formData.value.quantity = v.quantity
})

const setVendorMutation = useMutation({
  mutationKey: ['admin-set-vendor-product', props.data.id],
  mutationFn: async (input: UpdateProductVendorInput) =>
    api.put('products/set-vendor', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  onSuccess() {
    notification.success('Vendor changes saved.')
    emits('afterMutation')
    vendorProperty.value = null
    if (vendorEdit.value !== null) {
      vendorEdit.value = null
    }
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        notification.error('Invalid vendor changes.')
        return
      },
      notFound() {
        notification.error('Not found!?')
        return
      },
    })
  },
})

const setVariantDetailMutation = useMutation({
  mutationKey: ['admin-set-variant-detail-product', props.data.id],
  mutationFn: async (input: ProductVariantDetailInput) =>
    api.put(`products/set-variant-detail/${props.data.id}`, input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  onSuccess() {
    notification.success('Variant detail added.')
    isSetVariantOpen.value = false
    createProperty.clear()
    emits('afterMutation')
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        notification.error('Invalid variant detail.')
        return
      },
    })
  },
})

async function saveVendor() {
  if (vendorProperty.value === null) {
    notification.warning('First select property then manage vendor.')
    return
  }
  setVendorValidator.formData.value.id = props.data.id
  setVendorValidator.formData.value.property = vendorProperty.value
  const { input, isValid } = await setVendorValidator.validate()
  if (!isValid) {
    notification.error('Invalid data.')
    return
  }
  if (input.data === undefined) {
    notification.error('Unexpected error.')
    return
  }
  await setVendorMutation.mutateAsync(input.data)
}

async function saveVariantDetail() {
  variantDetail.add(createProperty.get())
  setVariantDetailValidator.formData.value.items = variantDetail.state.value

  const { input, isValid } = await setVariantDetailValidator.validate()
  if (!isValid) {
    notification.error('Invalid data set.')
    return
  }
  if (input.data === undefined) {
    notification.error('Unexpected error.')
    return
  }
  await setVariantDetailMutation.mutateAsync(input.data)
}
</script>

<template>
  <section class="set-vendor-section">
    <div class="w-full c-flex-all-center">
      <button class="create-variant-btn" @click="isSetVariantOpen = true">
        Add Variant Detail
      </button>
    </div>
    <section class="show-items c-flex-all-center">
      <div
        v-for="[idx, item] of Object.entries(props.data.variant_detail)"
        :key="idx"
        class="item"
      >
        <h3>Price: {{ item.price }}</h3>
        <table>
          <tr v-for="[key, value] in Object.entries(item.property)" :key="key">
            <td>{{ key }}</td>
            <td>{{ value }}</td>
          </tr>
        </table>
        <button
          class="text-2xl font-bold cursor-pointer w-full text-center hover:brightness-125"
          @click="vendorProperty = item.property"
        >
          Add New Vendor
        </button>
        <div class="vendor-list">
          <ul v-for="(vendor, index) in item.vendors" :key="index">
            <button
              class="rounded-2xl w-full bg-(--c-v-2) p-4 cursor-pointer"
              @click="
                () => {
                  vendorProperty = item.property
                  vendorEdit = {
                    user_id: vendor.user_id,
                    quantity: vendor.quantity,
                  }
                }
              "
            >
              <li>{{ vendor.user_id }}</li>
              <li>{{ vendor.quantity }}</li>
            </button>
          </ul>
        </div>
      </div>
    </section>
    <section
      v-if="vendorProperty !== null || vendorEdit !== null"
      class="c-floating-window c-flex-all-center"
    >
      <!-- set vendor -->
      <div class="c-floating-box c-form-bg">
        <div class="w-full p-2">
          <button
            class="text-2xl font-bold cursor-pointer"
            type="button"
            @click="
              () => {
                vendorProperty = null
                if (vendorEdit !== null) {
                  vendorEdit = null
                }
              }
            "
          >
            <span>X</span>
          </button>
        </div>
        <form class="c-form">
          <div class="c-form-item">
            <label class="c-form-label" for="user_id">User ID</label>
            <input
              class="c-form-input"
              v-model="setVendorValidator.formData.value.user_id"
              :disabled="vendorEdit !== null"
              type="text"
            />
            <p
              class="c-form-error"
              v-if="setVendorValidator.errors.value['user_id'] !== undefined"
            >
              {{ setVendorValidator.errors.value['user_id'] }}
            </p>
          </div>
          <div class="c-form-item">
            <label class="c-form-label" for="quantity">Quantity</label>
            <input
              class="c-form-input"
              v-model="setVendorValidator.formData.value.quantity"
              type="number"
            />
            <p
              class="c-form-error"
              v-if="setVendorValidator.errors.value['quantity'] !== undefined"
            >
              {{ setVendorValidator.errors.value['quantity'] }}
            </p>
          </div>
          <button
            class="c-form-btn"
            @click="saveVendor"
            :disabled="setVendorMutation.isPending.value"
          >
            Save
          </button>
        </form>
      </div>
    </section>
    <section
      v-if="isSetVariantOpen"
      class="c-floating-window c-flex-all-center"
    >
      <!-- set variant property -->
      <div
        class="c-floating-box c-form-bg w-80/100 max-h-80/100 overflow-y-scroll overflow-x-hidden"
      >
        <div class="p-2 w-full">
          <button
            type="button"
            class="text-2xl font-bold text-(--c-v-7) cursor-pointer"
            @click="isSetVariantOpen = false"
          >
            <span>X</span>
          </button>
        </div>
        <form class="c-form">
          <button
            type="button"
            class="c-form-btn bg-(--c-v-11)"
            @click="createProperty.clear()"
          >
            Clear Properties
          </button>
          <button
            type="button"
            class="c-form-btn bg-(--c-v-14)"
            @click="createProperty.add()"
          >
            Add Property
          </button>
          <div class="h-auto flex flex-col gap-6">
            <label
              v-if="createProperty.state.value.length !== 0"
              class="w-full flex flex-row justify-evenly text-2xl font-bold"
            >
              <h2>Key</h2>
              <h2>Value</h2>
            </label>
            <div
              v-for="(item, index) of createProperty.state.value"
              :key="index"
              class="flex flex-col gap-4"
            >
              <div class="flex flex-wrap gap-4 w-full">
                <input class="c-form-input" v-model="item.key" type="text" />
                <button
                  class="text-xl font-bold text-(--c-v-12) w-8 cursor-pointer"
                  @click="createProperty.remove(index)"
                >
                  <span>X</span>
                </button>
                <input class="c-form-input" v-model="item.value" type="text" />
              </div>
            </div>
            <p
              class="c-form-error"
              v-for="key in Object.keys(setVariantDetailValidator.errors.value)"
              :key="key"
            >
              {{ setVariantDetailValidator.errors.value[key] }}
            </p>
          </div>
          <button
            class="c-form-btn"
            type="button"
            @click="saveVariantDetail"
            :disabled="setVariantDetailMutation.isPending.value"
          >
            Save
          </button>
        </form>
      </div>
    </section>
  </section>
</template>

<style scoped>
@reference "@/styles/index.css";

.set-vendor-section {
  @apply w-full max-w-200 rounded-2xl p-4 md:p-6
    border-4 border-(--c-v-15);
}

.set-vendor-section .create-variant-btn {
  @apply px-4 md:px-6 py-2 md:py-3 text-lg md:text-2xl font-bold rounded-2xl
    bg-(--c-v-15) text-(--c-v-1) cursor-pointer;
}

.set-vendor-section .show-items {
  @apply flex-col gap-4;
}

.set-vendor-section .item {
  @apply w-full h-fit rounded-2xl backdrop-brightness-125 px-2 md:px-4 mt-4
    border-b-2 border-(--c-v-1) dark:border-(--c-v-7)
    flex flex-col flex-wrap justify-center items-start gap-2;
}

.item h3 {
  @apply text-lg md:text-xl font-bold text-(--c-v-11) text-center rounded-xl py-2
    w-full my-2 border-b-2 border-t-2 border-(--c-v-11);
}

.item table {
  @apply flex flex-col w-full flex-wrap justify-center items-center gap-4 rounded-2xl p-4
    border-2 border-(--c-v-1) dark:border-(--c-v-7);
}

.item tr {
  @apply flex w-full flex-wrap justify-evenly items-center h-auto text-wrap;
}

.item td {
  @apply border-l-2 border-(--c-v-10) px-1 w-full sm:w-50/100 h-auto text-wrap wrap-break-word;
}

.vendor-list ul {
  @apply list-disc border-b w-full border-(--c-v-10) p-2;
}

.vendor-list li {
  @apply text-wrap;
}

@media (max-width: 640px) {
  .set-vendor-section .c-floating-box {
    @apply w-full max-w-full m-2;
  }

  .set-vendor-section .c-floating-window {
    @apply p-4;
  }
}
</style>
