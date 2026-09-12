<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'
import axios from 'axios'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import {
  RegisterRequest,
  type RegisterInput,
} from '@/contracts/auth/request.schema'
import type { StatusResponse } from '@/contracts/response.interface'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const router = useRouter()
const store = useStore()
const { formData, errors, validate } = useValidator(RegisterRequest)

const errorMsg = ref<string>('')
const showPassword = ref<boolean>(false)

const { mutate, isPending } = useMutation({
  mutationKey: ['register'],
  mutationFn: async (input: RegisterInput) => {
    return api.post<StatusResponse>('auth/register', input)
  },
  onSuccess: () => {
    errorMsg.value = ''
    router.push({ name: 'auth' })
  },
  onError: (error) => {
    if (!axios.isAxiosError(error)) {
      errorMsg.value = 'Unxpected error, try again.'
      return
    }
    switch (error.response?.status) {
      case axios.HttpStatusCode.BadRequest:
        errorMsg.value = `Invalid request, try again.`
        break
      default:
        router.push({ name: 'auth' })
    }
  },
})

const onClick = async (event: MouseEvent) => {
  event.preventDefault()

  formData.value.email = store.getEmail
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
  <div class="register-page c-flex-all-center">
    <div class="register-box c-form-bg">
      <form class="c-form">
        <div class="c-form-item">
          <label class="c-form-label" for="pass-key"> Pass Key: </label>
          <div class="pass-key c-form-input c-flex-all-center">
            <input
              class="c-clean-input"
              :type="showPassword ? 'text' : 'password'"
              v-model="formData.pass_key"
            />
            <button
              type="button"
              class="pass-show-btn"
              @mousedown.prevent
              @click="showPassword = !showPassword"
            >
              <BaseIcon
                v-if="showPassword"
                :icon="icons.pages.auth.eyeOn"
                stroke-color="--color-eye-icon"
              />
              <BaseIcon
                v-else
                :icon="icons.pages.auth.eyeOff"
                stroke-color="--color-eye-icon"
              />
            </button>
          </div>
          <p class="c-form-error" v-if="errors['pass_key'] !== undefined">
            {{ errors['pass_key'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="email"> Email: </label>
          <input class="c-form-input" v-model="formData.email" />
          <p class="c-form-error" v-if="errors['email'] !== undefined">
            {{ errors['email'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="phone_number"> Phone Number: </label>
          <input class="c-form-input" v-model="formData.phone_number" />
          <p class="c-form-error" v-if="errors['phone_number'] !== undefined">
            {{ errors['phone_number'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="email"> Fist name: </label>
          <input class="c-form-input" v-model="formData.fist_name" />
          <p class="c-form-error" v-if="errors['fist_name'] !== undefined">
            {{ errors['fist_name'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="last_name"> Last name: </label>
          <input class="c-form-input" v-model="formData.last_name" />
          <p class="c-form-error" v-if="errors['last_name'] !== undefined">
            {{ errors['last_name'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="national_code">
            National code:
          </label>
          <input class="c-form-input" v-model="formData.national_code" />
          <p class="c-form-error" v-if="errors['national_code'] !== undefined">
            {{ errors['national_code'] }}
          </p>
        </div>
        <button
          class="c-form-btn"
          :class="{ 'c-is-pending': isPending }"
          @click="onClick($event)"
        >
          Send
        </button>
        <p class="c-form-error" v-if="errorMsg !== ''">
          {{ errorMsg }}
        </p>
      </form>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.register-page {
  @apply w-screen min-h-screen p-20;
}

.register-page .register-box {
  @apply w-150 h-fit rounded-2xl p-10;
}
</style>
