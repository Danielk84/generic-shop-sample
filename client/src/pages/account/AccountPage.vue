<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'

import api from '@/utils/api'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler } from '@/utils/helper'
import { useValidator } from '@/utils/validator'
import {
  EmailAddrRequest,
  PhoneNumberRequest,
  PermissionType,
  UpsertShopRequest,
  type UpsertShopInput,
} from '@/contracts/users/request.schema'
import type {
  ShopInfoResponse,
  UserDetailResponse,
} from '@/contracts/users/response.interface'

const VerifiedStatusIcon = defineAsyncComponent(
  () => import('@/pages/account/VerifiedStatusIcon.vue'),
)
const ImageFrameCard = defineAsyncComponent(
  () => import('@/components/card/ImageFrameCard.vue'),
)

const router = useRouter()
const store = useStore()
const notification = useNotificationStore()
const isShopUser = computed(
  () =>
    store.getClaims.permission_type === PermissionType.Vendor ||
    store.getClaims.permission_type === PermissionType.Admin,
)

const isShopError = ref<boolean>(false)
const isUserError = ref<boolean>(false)

const isEmailVerifyOpen = ref<boolean>(false)
const isDeleteUserOpen = ref<boolean>(false)

const {
  formData: shopForm,
  errors: shopErrors,
  validate: shopValidate,
} = useValidator(UpsertShopRequest)
const {
  formData: phoneForm,
  errors: phoneErrors,
  validate: phoneValidate,
} = useValidator(PhoneNumberRequest)
const {
  formData: userPhoneForm,
  errors: userPhoneErrors,
  validate: userPhoneValidate,
} = useValidator(PhoneNumberRequest)
const {
  formData: emailForm,
  errors: emailErrors,
  validate: emailValidate,
} = useValidator(EmailAddrRequest)

const IsEmailVerifyReq = ref<boolean>(false)
const emailVerifyCode = ref<string>('')

const userQuery = useQuery({
  queryKey: ['get-user-data', store.claims.id],
  queryFn: async () =>
    api.get<UserDetailResponse>(`users/${store.claims.id}`, {
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => res.data,
})

watch(
  userQuery.error,
  (err) => {
    if (err != null) {
      errorStatusHandler(err, router)
    }
  },
  { immediate: true },
)

watch(
  userQuery.data,
  (value) => {
    if (value === undefined) return
    emailForm.value.email = value.email
    userPhoneForm.value.phone_number = value.phone_number
  },
  { immediate: true },
)

const shopQuery = useQuery({
  queryKey: ['get-shop-data', store.claims.id],
  enabled: computed(() => isShopUser.value),
  queryFn: async () =>
    api.get<ShopInfoResponse>(`users/shop/${store.claims.id}`, {
      headers: { Authorization: store.getAccessToken },
    }),
  select: (res) => res.data,
})

watch(
  shopQuery.error,
  (err) => {
    if (err != null) {
      errorStatusHandler(err, router, {
        notFound() {
          // empty block
          return
        },
      })
    }
  },
  { immediate: true },
)

watch(
  shopQuery.data,
  (value) => {
    if (value === undefined) return
    shopForm.value.brand = value.brand
    shopForm.value.shop_addr = value.shop_addr
    shopForm.value.zip_code = value.zip_code
    shopForm.value.business_code = value.business_code
    shopForm.value.bio = value.bio
    phoneForm.value.phone_number = value.shop_phone_number
  },
  { immediate: true },
)

const shopMutation = useMutation({
  mutationKey: ['upsert-shop-data'],
  mutationFn: async () => {
    const { input, isValid } = await shopValidate()
    if (!isValid || input.data === undefined) return null
    return await api.post('users/shop', input.data, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Shop profile saved.')
    isShopError.value = false
    await shopQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      badRequest() {
        notification.error('Invalid shoping profile.')
        isShopError.value = true
        return
      },
    })
  },
})

const phoneMutation = useMutation({
  mutationKey: ['update-shop-phone-number'],
  mutationFn: async () => {
    const { input, isValid } = await phoneValidate()
    if (!isValid || input.data === undefined) return null
    return api.put(`users/shop/${store.getClaims.id}`, input.data, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Shop phone number updated.')
    isUserError.value = false
    await shopQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      badRequest() {
        notification.error('Invalid shop phone number')
        isUserError.value = true
        return
      },
    })
  },
})

const userPhoneMutation = useMutation({
  mutationKey: ['update-user-phone-number'],
  mutationFn: async () => {
    const { input, isValid } = await userPhoneValidate()
    if (!isValid || input.data === undefined) return null
    return api.put('users/set-phone-number', input.data, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Verification code sent to your phone.')
    isUserError.value = false
    await userQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      badRequest() {
        notification.error('Invalid user phone number.')
        isUserError.value = true
        return
      },
    })
  },
})

const emailMutation = useMutation({
  mutationKey: ['update-user-email'],
  mutationFn: async () => {
    const { input, isValid } = await emailValidate()
    if (!isValid || input.data === undefined) return null
    return api.put('users/set-email', input.data, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Verification code sent to the new email.')
    isUserError.value = false
    isEmailVerifyOpen.value = true
    await userQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      notFound() {
        notification.error('Invalid user email')
        isUserError.value = true
        return
      },
    })
  },
})

const verifyEmailMutation = useMutation({
  mutationKey: ['verify-user-email'],
  mutationFn: async () =>
    api.put(
      'users/verify-email',
      { key: emailVerifyCode.value },
      {
        headers: { Authorization: store.getAccessToken },
      },
    ),
  async onSuccess() {
    notification.success('Email verified.')
    emailVerifyCode.value = ''
    isEmailVerifyOpen.value = false
  },
  onError(err) {
    errorStatusHandler(err, router, {
      badRequest() {
        notification.error('Invalid user email verify code.')
      },
    })
  },
})

const deleteAccountMutation = useMutation({
  mutationKey: ['delete-user-account'],
  mutationFn: async () => {
    return api.delete('users', {
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: () => {
    store.logout()
    notification.success('Account deleted.')
    router.push({ name: 'home' })
  },
  onError: (err) => {
    errorStatusHandler(err, router)
  },
})

const imageMutation = useMutation({
  mutationKey: ['upload-shop-profile'],
  mutationFn: async (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.post('users/shop/upload', form, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Shop image uploaded.')
    await shopQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      notFound() {
        notification.error('First create shop, context not found.')
        return
      },
      badRequest() {
        notification.error('Invalid file request.')
        return
      },
    })
  },
})

const removeImageMutation = useMutation({
  mutationKey: ['remove-shop-profile'],
  mutationFn: async () => {
    return api.delete('users/shop', {
      headers: { Authorization: store.getAccessToken },
    })
  },
  async onSuccess() {
    notification.success('Shop image removed.')
    await shopQuery.refetch()
  },
  onError(err) {
    errorStatusHandler(err, router, {
      notFound() {
        notification.error('Profile photo not found!')
        return
      },
    })
  },
})

async function onImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file !== undefined) await imageMutation.mutateAsync(file)
  IsEmailVerifyReq.value = false
}
</script>

<template>
  <div class="account-page">
    <h1>Account</h1>

    <section
      v-if="isShopUser"
      class="account-card c-form-bg"
      :class="{ 'c-form-error-shadow': isShopError }"
    >
      <h2>Shop profile</h2>
      <form class="c-form" @submit.prevent="shopMutation.mutate()">
        <div class="c-form-item">
          <label class="c-form-label" for="brand">Brand</label>
          <input
            class="c-form-input"
            id="brand"
            v-model="shopForm.brand"
            type="text"
          />
          <p class="c-form-error" v-if="shopErrors['brand'] !== undefined">
            {{ shopErrors['brand'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="shop_addr">Shop address</label>
          <input
            class="c-form-input"
            id="shop_addr"
            v-model="shopForm.shop_addr"
            type="text"
          />
          <p class="c-form-error" v-if="shopErrors['shop_addr'] !== undefined">
            {{ shopErrors['shop_addr'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="zip_code">Zip code</label>
          <input
            class="c-form-input"
            id="zip_code"
            v-model="shopForm.zip_code"
            type="text"
          />
          <p class="c-form-error" v-if="shopErrors['zip_code'] !== undefined">
            {{ shopErrors['zip_code'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="business_code">Business code</label>
          <input
            class="c-form-input"
            id="business_code"
            v-model="shopForm.business_code"
            type="text"
          />
          <p
            class="c-form-error"
            v-if="shopErrors['business_code'] !== undefined"
          >
            {{ shopErrors['business_code'] }}
          </p>
        </div>

        <div class="c-form-item">
          <label class="c-form-label" for="bio">Bio</label>
          <textarea
            class="c-form-input resize-none h-30"
            id="bio"
            v-model="(shopForm as UpsertShopInput).bio"
            rows="4"
          ></textarea>
          <p class="c-form-error" v-if="shopErrors['bio'] !== undefined">
            {{ shopErrors['bio'] }}
          </p>
        </div>

        <button
          class="c-form-btn"
          type="submit"
          :disabled="shopMutation.isPending.value"
        >
          Save shop
        </button>
      </form>
    </section>

    <section
      class="user-forms account-card c-form-bg"
      :class="{ 'c-form-error-shadow': isUserError }"
    >
      <h2>Contact details</h2>
      <form
        v-if="isShopUser"
        class="c-form"
        @submit.prevent="phoneMutation.mutate()"
      >
        <div class="c-form-item">
          <label class="c-form-label" for="shop_phone">Shop phone</label>
          <div class="c-form-input flex">
            <input
              class="c-clean-input"
              id="shop_phone"
              v-model="phoneForm.phone_number"
              type="text"
            />
            <VerifiedStatusIcon
              :status="shopQuery.data.value?.is_v_phone_number"
            />
          </div>
          <p
            class="c-form-error"
            v-if="phoneErrors['phone_number'] !== undefined"
          >
            {{ phoneErrors['phone_number'] }}
          </p>
        </div>
        <button
          class="c-form-btn"
          type="submit"
          :disabled="phoneMutation.isPending.value"
        >
          Update phone
        </button>
      </form>
      <form class="c-form" @submit.prevent="userPhoneMutation.mutate()">
        <div class="c-form-item">
          <label class="c-form-label" for="user_phone">Personal phone</label>
          <div class="c-form-input flex">
            <input
              class="c-clean-input"
              id="user_phone"
              v-model="userPhoneForm.phone_number"
              type="text"
            />
            <VerifiedStatusIcon
              :status="userQuery.data.value?.is_v_phone_number"
            />
          </div>
          <p
            class="c-form-error"
            v-if="userPhoneErrors['phone_number'] !== undefined"
          >
            {{ userPhoneErrors['phone_number'] }}
          </p>
        </div>
        <button
          class="c-form-btn"
          type="submit"
          :disabled="userPhoneMutation.isPending.value"
        >
          Update personal phone
        </button>
      </form>
      <form class="c-form" @submit.prevent="emailMutation.mutate()">
        <div class="c-form-item">
          <label class="c-form-label" for="email">Email</label>
          <div class="c-form-input flex">
            <input
              class="c-clean-input"
              id="email"
              v-model="emailForm.email"
              type="email"
            />
            <VerifiedStatusIcon :status="userQuery.data.value?.is_v_email" />
          </div>
          <p class="c-form-error" v-if="emailErrors['email'] !== undefined">
            {{ emailErrors['email'] }}
          </p>
        </div>
        <button
          class="c-form-btn"
          type="submit"
          :disabled="emailMutation.isPending.value"
        >
          Save & Send verification
        </button>
      </form>
    </section>

    <section v-if="isShopUser" class="account-card c-form-bg">
      <h2>Shop image</h2>
      <div class="size-100 border-4 rounded-2xl m-5 border-(--c-v-13) p-4">
        <ImageFrameCard
          :img="shopQuery.data.value?.img_path"
          alt="shop profile"
        />
      </div>
      <input
        class="upload-img"
        type="file"
        accept="image/*"
        @change="onImageChange"
      />
      <button
        v-if="shopQuery.data.value?.img_path"
        class="delete-img"
        type="button"
        :disabled="removeImageMutation.isPending.value"
        @click="removeImageMutation.mutate()"
      >
        Remove image
      </button>
    </section>

    <section class="account-card danger-card">
      <h2>Delete account</h2>
      <button type="button" @click="isDeleteUserOpen = true">
        Delete account
      </button>
    </section>
    <section
      v-if="isEmailVerifyOpen"
      class="c-floating-window c-flex-all-center"
    >
      <div class="c-floating-box account-card">
        <form class="c-form" @submit.prevent="verifyEmailMutation.mutate()">
          <label class="c-form-label" for="email-code">Verification code</label>
          <input
            class="c-form-input"
            id="email-code"
            v-model="emailVerifyCode"
            inputmode="numeric"
            type="text"
          />
          <button
            class="c-form-btn"
            type="submit"
            :disabled="verifyEmailMutation.isPending.value"
          >
            Verify email
          </button>
        </form>
      </div>
    </section>
    <section
      v-if="isDeleteUserOpen"
      class="c-floating-window c-flex-all-center"
    >
      <div class="c-floating-box delete-account">
        <h2>Are you sure to DELETING account?</h2>
        <div class="flex flex-row w-full justify-between gap-10 py-5">
          <button
            class="delete-btn"
            type="button"
            :disabled="deleteAccountMutation.isPending.value"
            @click="deleteAccountMutation.mutate()"
          >
            Delete
          </button>
          <button class="cancel-btn" @click="isDeleteUserOpen = false">
            Cancel
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.account-page {
  @apply w-full min-h-150 p-10 flex flex-col items-center gap-15;
}

.account-page > h1 {
  @apply text-4xl font-bold;
}

.account-page .account-card {
  @apply w-full max-w-200 p-6 rounded-2xl;
}

.account-card h2 {
  @apply text-2xl font-bold;
}

.user-forms form {
  @apply border-b-2 border-(--c-v-5) mb-10 rounded-xl;
}

.account-card img {
  @apply w-32 h-32 object-cover rounded-2xl;
}

.account-card .upload-img {
  @apply rounded-2xl w-full p-4 text-2xl
    bg-(--c-v-9) text-(--c-v-1)
    my-10 text-center cursor-pointer
    hover:brightness-95;
}

.account-card .delete-img {
  @apply rounded-2xl w-full p-4 text-2xl
    bg-(--c-v-11) text-(--c-v-7) cursor-pointer
    hover:brightness-95;
}

.account-page .danger-card {
  @apply border-(--c-v-11) border-4;
}

.danger-card > button {
  @apply bg-(--c-v-11) text-(--c-v-7) cursor-pointer
    text-2xl rounded-2xl w-full p-4 mt-10
    hover:brightness-95;
}

.account-page .delete-account {
  @apply p-5 flex flex-col gap-10;
}

.delete-account h2 {
  @apply font-bold text-xl;
}

.delete-account button {
  @apply cursor-pointer w-40/100 h-14
    rounded-2xl hover:brightness-125
    text-2xl font-bold;
}

.delete-account .delete-btn {
  @apply bg-(--c-v-11) text-(--c-v-7);
}

.delete-account .cancel-btn {
  @apply border-4 border-(--c-v-14);
}
</style>
