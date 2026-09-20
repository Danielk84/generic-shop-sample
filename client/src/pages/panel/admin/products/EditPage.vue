<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuery, useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useValidator } from '@/utils/validator'
import {
  UpdateProductRequest,
  type UpdateProductInput,
} from '@/contracts/products/request.schema'
import type {
  ProductResponse,
  ProductProperty,
} from '@/contracts/products/response.interface'

const BackBtn = defineAsyncComponent(
  () => import('@/components/ui/button/BackBtn.vue'),
)

const route = useRoute()
const router = useRouter()
const store = useStore()

const callback_page = route.query.page
const productID = route.params.productID
if (productID === undefined || productID === '') {
  router.push('/404')
}

const { formData, errors, validate } = useValidator(UpdateProductRequest)

const errorMsg = ref<string>('')
const isError = ref<boolean>(false)

const properties = ref<
  {
    key: string
    value: string
  }[]
>([])
function addProperty() {
  properties.value.push({
    key: '',
    value: '',
  })
}
function deleteProperty(index: number) {
  properties.value.splice(index, 1)
}

const query = useQuery({
  queryKey: ['admin-product-edit-query', productID],
  queryFn: async () =>
    api.get<ProductResponse>(`products/overview/${productID}`, {
      headers: {
        Authorization: store.getAccessToken,
      },
    }),
  select: (res) => res.data,
})

watch(
  query.data,
  (data) => {
    if (data === undefined) {
      return
    }
    formData.value['id'] = productID
    formData.value['name'] = data.name
    formData.value['description'] = data.description
    properties.value = []
    for (const [key, value] of Object.entries(data.common_detail)) {
      properties.value.push({
        key: key,
        value: value,
      })
    }
  },
  {
    immediate: true,
  },
)

watch(
  query.error,
  async (err) => {
    if (err != null) {
      errorStatusHandler(err, router)
    }
  },
  {
    immediate: true,
  },
)

const mutation = useMutation({
  mutationKey: ['admin-product-edit-mutation', productID],
  mutationFn: async (input: UpdateProductInput) => {
    return api.put(`/products/`, input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  async onSuccess() {
    errorMsg.value = ''
    isError.value = false
    await query.refetch()
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        errorMsg.value = 'invalid data.'
        isError.value = true
      },
    })
  },
})

async function onClick(event: MouseEvent) {
  event.preventDefault()

  formData.value.common_detail = {}
  properties.value.forEach((item) => {
    ;(formData.value.common_detail as ProductProperty)[item.key] = item.value
  })
  const { input, isValid } = await validate()
  if (!isValid) {
    isError.value = true
    return
  }
  if (input.data !== undefined) {
    mutation.mutate(input.data)
  }
}
</script>

<template>
  <div class="product-edit-page c-flex-all-center">
    <div class="edit-box c-form-bg" :class="{ 'c-form-error-shadow': isError }">
      <div class="py-4">
        <BackBtn
          :icon="{
            strokeColor: '--color-panel-products-back-icon',
            fillColor: '--color-panel-products-back-icon',
          }"
        />
      </div>
      <form class="c-form">
        <div class="c-form-item">
          <label class="c-form-label" for="name">
            <span>Name: </span>
          </label>
          <input class="c-form-input" v-model="formData.name" />
          <p v-if="errors['name'] !== undefined" class="c-form-error">
            {{ errors['name'] }}
          </p>
        </div>
        <div class="c-form-item base-item">
          <label class="c-form-label" for="description">
            <span>Description: </span>
          </label>
          <textarea
            class="c-form-input base-input"
            v-model="(formData as UpdateProductInput).description"
            rows="4"
            cols="5"
          ></textarea>
          <p v-if="errors['description'] !== undefined" class="c-form-error">
            {{ errors['description'] }}
          </p>
        </div>
        <div class="c-flex-all-center flex-col">
          <button
            class="add-property c-flex-all-center"
            type="button"
            @click="addProperty()"
          >
            Add property
          </button>
          <div v-if="properties.length !== 0" class="properties-box">
            <div class="properties-head">
              <h2>Key</h2>
              <h2>Value</h2>
            </div>
            <div
              v-for="(item, index) of properties"
              :key="index"
              class="properties"
            >
              <input
                class="property-item"
                v-model="item.key"
                :placeholder="item.key"
              />
              <input
                class="property-item"
                v-model="item.value"
                :placeholder="item.value"
              />
              <button
                type="button"
                class="delete-property"
                @click="deleteProperty(index)"
              >
                X
              </button>
            </div>
          </div>
        </div>
        <button
          class="c-form-btn"
          :class="{ 'c-is-pending': !query.isPending }"
          type="button"
          @click="onClick($event)"
        >
          <span>Save</span>
        </button>
        <p v-if="errorMsg !== ''" class="c-form-error">
          {{ errorMsg }}
        </p>
      </form>
      <div class="next-box">
        <RouterLink
          class="next-btn"
          :to="{
            name: 'admin-product-upload-img',
            params: { productID: productID },
            query: { page: callback_page },
          }"
        >
          <span>Next</span>
        </RouterLink>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.product-edit-page {
  @apply w-screen min-h-screen p-10;
}

.product-edit-page .edit-box {
  @apply rounded-2xl p-10;
}

.product-edit-page .base-item {
  @apply flex flex-col gap-4;
}

.product-edit-page textarea {
  @apply resize-y;
}

.product-edit-page .properties-box {
  @apply my-10 flex flex-col gap-5;
}

.product-edit-page .properties-head {
  @apply flex flex-row items-center justify-evenly
    text-2xl font-bold;
}

.product-edit-page .properties {
  @apply flex flex-row gap-4;
}

.product-edit-page .add-property {
  @apply m-10 rounded-2xl
    px-10 py-5
    hover:brightness-110
    text-panel-products-add-property-btn-text
    bg-panel-products-add-property-btn
    cursor-pointer;
}

.product-edit-page .delete-property {
  @apply text-2xl font-bold cursor-pointer;
}

.product-edit-page .property-item {
  @apply outline-4
    focus:brightness-120
    p-4 outline-c-form-outline
    rounded-xl h-15 w-full text-xl;
}

.product-edit-page .next-box {
  @apply w-full flex items-center justify-center;
}

.product-edit-page .next-btn {
  @apply py-5 px-20 rounded-2xl m-5
    w-fit h-fit text-xl font-bold
    bg-panel-products-next-btn
    text-panel-products-next-text
    hover:brightness-95;
}
</style>
