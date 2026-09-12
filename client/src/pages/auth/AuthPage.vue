<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'
import axios from 'axios'

import api from '@/utils/api'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useValidator } from '@/utils/validator'
import {
  EmailAddrRequest,
  type EmailAddrInput,
} from '@/contracts/users/request.schema'
import type {
  MsgResponse,
  StatusResponse,
} from '@/contracts/response.interface'

const BackBtn = defineAsyncComponent(
  () => import('@/components/ui/button/BackBtn.vue'),
)

const router = useRouter()
const route = useRoute()
const store = useStore()
const callbackUrl = route.query.callback_url

if (store.getEmail === '') {
  router.push({ name: 'auth' })
}

const { formData, errors, validate } = useValidator(EmailAddrRequest)

const errorMsg = ref<string>('')
const isError = ref<boolean>(false)

const { mutate, isPending } = useMutation({
  mutationKey: ['is_user_exists'],
  mutationFn: async (input: EmailAddrInput) => {
    return api.post<StatusResponse | MsgResponse>('auth/is-user-exists', input)
  },
  onSuccess: (res) => {
    if ((res.data as StatusResponse).status === 'OK') {
      errorMsg.value = ''
      router.push({
        name: 'login',
        query: {
          callback_url: typeof callbackUrl === 'string' ? callbackUrl : '',
        },
      })
    } else {
      errorMsg.value = 'invalid response'
    }
  },
  onError: (error) => {
    errorStatusHandler(error, router, {
      notAxiosError() {
        errorMsg.value = 'Unexpected error, try again.'
      },
      badRequest() {
        errorMsg.value = 'invalid email, fix and try again.'
      },
      notFound() {
        if (!axios.isAxiosError(error)) {
          return
        }
        if (error.response?.data.msg === 'register') {
          router.push({ name: 'register' })
        } else {
          router.push('/')
        }
      },
      unprocessableEntity() {
        errorMsg.value = 'Unprocessable request, please try later.'
      },
    })
  },
})

const onClick = async (event: MouseEvent) => {
  event.preventDefault()

  const { input, isValid } = await validate()
  if (!isValid) {
    errorMsg.value = 'invalid input.'
    isError.value = true
    return
  }
  if (input.data !== undefined) {
    store.setEmail(input.data.email)
    mutate(input.data)
  }
}
</script>

<template>
  <div class="auth-page c-flex-all-center">
    <div class="auth-box c-form-bg" :class="{ 'c-form-error-shadow': isError }">
      <div class="info">
        <BackBtn
          :icon="{
            strokeColor: '--color-auth-p-icon',
            fillColor: '--color-auth-p-icon',
          }"
        />
        <h1>
          <span>Login or Register</span>
        </h1>
      </div>
      <form class="auth-form c-form">
        <div class="c-form-item">
          <label class="c-form-label" for="email">
            Please Enter your email address
          </label>
          <input
            v-model="formData.email"
            class="c-form-input"
            placeholder="person@email.com"
          />
          <p class="c-form-error" v-if="errors['email'] !== undefined">
            {{ errors['email'] }}
          </p>
        </div>
        <p class="c-form-error" v-if="errorMsg !== ''">
          {{ errorMsg }}
        </p>
        <button
          @click="onClick($event)"
          class="c-form-btn"
          :class="{ 'c-is-pending': isPending }"
        >
          Login / Register
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.auth-page {
  @apply w-screen h-screen;
}

.auth-page > .auth-box {
  @apply w-100 h-120 p-6 rounded-2xl
    flex flex-col justify-between items-center;
}

.auth-box .info {
  @apply w-full h-fit
    flex flex-row justify-between items-center;
}

.auth-box h1 {
  @apply text-2xl font-bold;
}

.auth-box .pass-key {
  @apply flex-row gap-3;
}

.auth-box .pass-show-btn {
  @apply cursor-pointer size-6;
}
</style>
