<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import {
  UserPermissionRequest,
  type UserPermissionInput,
} from '@/contracts/users/request.schema'
import { errorStatusHandler } from '@/utils/helper'

const store = useStore()
const route = useRoute()
const id = route.params.id

const { formData, errors, validate } = useValidator(UserPermissionRequest)

const { mutate, isPending } = useMutation({
  mutationKey: ['update-user-permission', id],
  mutationFn: async (input: UserPermissionInput) => {
    api.put(`user/${id}`, input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {},
  onError: (error) => {
    return errorStatusHandler(error)
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
  <div class="c-form-bg">
    <form class="c-form">
      <div>
        <label class="c-form-label" for="permission_type">
          Permission type:
        </label>
        <input class="c-form-input" v-model="formData.permission_type" />
        <p class="c-form-error" v-if="errors['permission_type'] !== undefined">
          {{ errors['permission_type'] }}
        </p>
      </div>
      <div class="c-form-item">
        <label for="is_active"> Active </label>
        <input type="checkbox" v-model="formData.is_active" />
        <p v-if="errors['is_active'] !== undefined">
          {{ errors['is_active'] }}
        </p>
      </div>
      <button :class="{ 'is-pending': isPending }" @click="onClick($event)">
        Set
      </button>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";
</style>
