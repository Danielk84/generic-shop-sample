<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useMutation } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useValidator } from '@/utils/validator'
import { range } from '@/utils/helper'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'
import {
  PermissionNames,
  PermissionType,
  UserPermissionRequest,
  type UserPermissionInput,
} from '@/contracts/users/request.schema'

const store = useStore()
const route = useRoute()
const router = useRouter()
const notification = useNotificationStore()
const id = route.params.id

const { formData, errors, validate } = useValidator(UserPermissionRequest)
formData.value.is_active = true

const { mutate, isPending } = useMutation({
  mutationKey: ['update-user-permission', id],
  mutationFn: async (input: UserPermissionInput) => {
    return api.put(`users/${id}`, input, {
      headers: {
        Authorization: store.getAccessToken,
      },
    })
  },
  onSuccess: () => {
    notification.success('User permission updated.')
  },
  onError: (error) => {
    return errorStatusHandler(error, router)
  },
})

const onClick = async (event: Event) => {
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
  <div class="update-permission c-form-bg">
    <form class="c-form">
      <div>
        <label class="c-form-label" for="permission_type">
          Permission type:
        </label>
        <select v-model="formData.permission_type">
          <option
            v-for="i of range(
              PermissionType.Admin,
              PermissionType.BlockUser + 1,
            )"
            :value="i"
          >
            {{ PermissionNames[i] }}
          </option>
        </select>
        <p class="c-form-error" v-if="errors['permission_type'] !== undefined">
          {{ errors['permission_type'] }}
        </p>
      </div>
      <div class="c-form-item p-4 m-4 text-center">
        <label class="text-2xl font-bold" for="is_active"> Active </label>
        <input class="is-active" type="checkbox" v-model="formData.is_active" />
        <p v-if="errors['is_active'] !== undefined">
          {{ errors['is_active'] }}
        </p>
      </div>
      <button
        class="c-form-btn"
        :class="{ 'is-pending': isPending }"
        @click="onClick($event)"
      >
        Set
      </button>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.update-permission {
  @apply rounded-2xl p-4;
}

.update-permission .is-active {
  @apply accent-(--c-v-15) size-6;
}

.update-permission select {
  @apply bg-(--c-v-4) border-2 border-(--c-v-7) p-2 rounded-xl;
}
</style>
