<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import { errorStatusHandler } from '@/utils/helper'
import {
  CreateProductRequest,
  type CreateProductInput,
} from '@/contracts/products/request.schema'
import type { ProductProperty } from '@/contracts/products/response.interface'

const store = useStore()
const router = useRouter()

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
const { formData, errors, validate } = useValidator(CreateProductRequest)

const isError = ref<boolean>(false)
const errorMsg = ref<string>('')
const { mutateAsync, isPending } = useMutation({
  mutationKey: ['create-product'],
  mutationFn: async (input: CreateProductInput) => {
    return api.post('products/', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess(res) {
    router.push({
      name: 'admin-product-upload-img',
      params: { productID: res.data.productID },
      replace: true,
    })
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        errorMsg.value = 'invalid request'
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
    await mutateAsync(input.data)
  }
}
</script>

<template>
  <div class="create-product-page c-flex-all-center">
    <form class="c-form c-form-bg min-h-fit c-flex-all-center">
      <div class="c-form-item base-item">
        <label class="c-form-label" for="name"> Name : </label>
        <input class="c-form-input base-input" v-model="formData.name" />
        <p v-if="errors['name'] !== undefined" class="c-form-error">
          {{ errors['name'] }}
        </p>
      </div>
      <div class="c-form-item base-item">
        <label class="c-form-label" for="describtion"> Describtion : </label>
        <textarea
          class="c-form-input base-input"
          v-model="(formData as CreateProductInput).description"
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
        :class="{ 'c-is-pending': isPending }"
        @click="onClick($event)"
      >
        Create
      </button>
      <p class="c-form-error" v-if="errorMsg !== ''">
        {{ errorMsg }}
      </p>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.create-product-page {
  @apply w-full min-h-screen p-10;
}

.create-product-page form {
  @apply rounded-2xl py-8 w-70/100 px-10;
}

.create-product-page .base-input {
  @apply w-150;
}

.create-product-page .base-item {
  @apply flex flex-col gap-4;
}

.create-product-page textarea {
  @apply resize-y;
}

.create-product-page .add-property {
  @apply m-10 rounded-2xl
    px-10 py-5
    hover:brightness-110
    text-panel-products-add-property-btn-text
    bg-panel-products-add-property-btn
    cursor-pointer;
}

.create-product-page .properties-box {
  @apply my-10 flex flex-col gap-5;
}

.create-product-page .properties-head {
  @apply flex flex-row items-center justify-evenly
    text-2xl font-bold;
}

.create-product-page .properties {
  @apply flex flex-row gap-4;
}

.create-product-page .delete-property {
  @apply text-2xl font-bold cursor-pointer;
}

.create-product-page .property-item {
  @apply outline-4
    focus:brightness-120
    p-4 outline-c-form-outline
    rounded-xl h-15 w-full text-xl;
}
</style>
