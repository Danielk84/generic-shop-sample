<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import { errorStatusHandler } from '@/utils/helper'
import {
  UpdateProductRequest,
  type UpdateProductInput,
} from '@/contracts/products/request.schema'

const store = useStore()
const router = useRouter()
const { formData, errors, validate } = useValidator(UpdateProductRequest)

const { mutate, isPending } = useMutation({
  mutationKey: ['admin-product-set-vendor'],
  mutationFn: async (input: UpdateProductInput) => {
    return api.put('products/set-vendor', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {},
  onError: (error) => {
    errorStatusHandler(error, router)
  },
})

const props = defineProps<{
  id: string
}>()

const onClick = async (event: MouseEvent) => {
  event.preventDefault()

  formData.value.id = props.id
  const { input, isValid } = await validate()
  if (!isValid) {
    return
  }
  if (input.data !== undefined) {
    mutate(input.data)
  }
}

// common detail not handled now.
</script>

<template>
  <div>
    <form>
      <div>
        <label for=""></label>
        <input type="text " v-model="formData.name" />
        <p v-if="errors['name']">{{ errors['name'] }}</p>
      </div>

      <div>
        <label for=""></label>
        <input type="text " v-model="formData.description" />
        <p v-if="errors['description']">{{ errors['description'] }}</p>
      </div>
      <button
        :class="{ 'is-pending': isPending }"
        @click="onClick($event)"
      ></button>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
