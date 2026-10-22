<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import { useNotificationStore } from '@/store/notification'
import {
  CategoryTag,
  type CategoryTagInput,
} from '@/contracts/categories/request.schema'
import { errorStatusHandler } from '@/utils/helper'

const emits = defineEmits<{
  (e: 'refetch'): void
}>()

const store = useStore()
const notification = useNotificationStore()
const router = useRouter()

const errorMsg = ref<string>('')
const isError = ref<boolean>(false)

const { formData, errors, validate } = useValidator(CategoryTag)

const { mutate, isPending } = useMutation({
  mutationKey: ['admin-categories-create'],
  mutationFn: async (input: CategoryTagInput) => {
    return api.post('categories/', input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {
    if (errorMsg.value !== '') errorMsg.value = ''
    if (isError.value) isError.value = false
    formData.value.tag = ''
    notification.success('new tag added.')
    emits('refetch')
  },
  onError: (error) => {
    errorStatusHandler(error, router, {
      notFound() {
        errorMsg.value = 'Failed to create new tag'
        isError.value = true
      },
    })
  },
})

const onClick = async (event: MouseEvent) => {
  event.preventDefault()

  const { input, isValid } = await validate()
  if (!isValid) {
    return
  }
  if (input.data !== undefined) {
    mutate(input.data)
  }
}
</script>

<template>
  <div
    class="create-tag-box c-form-bg"
    :class="{ 'c-form-error-shadow': isError }"
  >
    <form>
      <div class="form-item">
        <label class="" for="tag"> Tag: </label>
        <input v-model="formData.tag" />
        <button :class="{ 'c-is-pending': isPending }" @click="onClick($event)">
          Add tag
        </button>
      </div>
      <p class="form-error" v-if="errors['tag'] !== undefined">
        {{ errors['tag'] }}
      </p>
      <p class="form-error" v-if="errorMsg !== ''">
        {{ errorMsg }}
      </p>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.create-tag-box {
  @apply w-full max-w-150 h-fit
    rounded-2xl p-4 md:p-6;
}

.create-tag-box form {
  @apply w-full flex flex-col gap-4;
}

.create-tag-box .form-item {
  @apply w-full flex flex-col md:flex-row gap-4
    justify-center items-center;
}

.form-item label {
  @apply text-lg md:text-2xl font-bold;
}

.form-item input {
  @apply outline-4 hover:brightness-140
    focus:brightness-110
    p-3 md:p-4 outline-(--c-v-7)
    rounded-xl h-12 md:h-15 w-full text-base md:text-xl;
}

.form-item button {
  @apply hover:brightness-90 cursor-pointer
    font-bold rounded-xl h-12 md:h-15 w-full md:w-35 text-base md:text-xl
    bg-(--c-v-8)
    text-(--c-v-7);
}

.create-tag-box .form-error {
  @apply p-3 md:p-4 text-wrap font-bold
    rounded-xl max-h-3 w-full text-sm text-center
    text-(--c-v-11);
}
</style>
