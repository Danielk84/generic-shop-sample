<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'
import axios from 'axios'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import { LoginRequest, type LoginInput } from '@/contracts/auth/request.schema'
import type { AccessTokenResponse } from '@/contracts/auth/response.interface'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)

const router = useRouter()
const route = useRoute()
const store = useStore()
const callbackUrl = route.query.callback_url

if (store.getEmail === '') {
  router.push({ name: 'auth' })
}

const { formData, errors, validate } = useValidator(LoginRequest)

const errorMsg = ref<string>('')
const isError = ref<boolean>(false)
const showPassword = ref<boolean>(false)

const { mutate, isPending } = useMutation({
  mutationKey: ['login'],
  mutationFn: async (input: LoginInput) => {
    return api.post<AccessTokenResponse>('auth/login', input)
  },
  onSuccess: ({ data }) => {
    errorMsg.value = ''
    store.setAccessToken(data.token)
    if (
      callbackUrl === undefined ||
      typeof callbackUrl !== 'string' ||
      callbackUrl === ''
    ) {
      router.push('/')
    } else {
      router.push(decodeURIComponent(callbackUrl))
    }
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
    isError.value = true
    errorMsg.value = 'invalid pass-key, try again.'
    return
  }
  if (input.data !== undefined) {
    mutate(input.data)
  }
}
</script>

<template>
  <div class="login-page c-flex-all-center">
    <div
      class="login-box c-form-bg"
      :class="{ 'to-c-form-shadow-error': isError }"
    >
      <form class="c-form">
        <div class="c-form-item">
          <label class="c-form-label" for="pass-key"> pass key: </label>
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
          <p class="c-form-error" v-if="errors['email'] !== undefined">
            {{ errors['email'] }}
          </p>
        </div>
        <button
          class="c-form-btn"
          :class="{ 'c-is-pending': isPending }"
          @click="onClick($event)"
        >
          Send
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.login-page {
  @apply w-screen h-screen;
}

.login-page .login-box {
  @apply rounded-2xl p-4;
}

.login-box .pass-key {
  @apply flex-row gap-3;
}

.login-box .pass-show-btn {
  @apply cursor-pointer size-6;
}
</style>
