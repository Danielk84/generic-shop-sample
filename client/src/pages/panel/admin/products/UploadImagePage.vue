<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import type { ProductImageResponse } from '@/contracts/products/response.interface'

const ImageFrameList = defineAsyncComponent(
  () => import('@/components/common/products/ImageFrameList.vue'),
)
const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const BackBtn = defineAsyncComponent(
  () => import('@/components/ui/button/BackBtn.vue'),
)

const store = useStore()
const router = useRouter()
const route = useRoute()
const productID = route.params.productID
const callback_page = route.query.page
if (productID === undefined || productID === '') {
  router.push('/404')
}

const errorMsg = ref<string>('')

const { data, error, refetch } = useQuery({
  queryKey: ['products-upload-images-query', productID],
  queryFn: async () =>
    api.get<ProductImageResponse[]>(`/products/images/${productID}`),
  select: (res) => res.data,
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router, {
        notFound() {},
      })
    }
  },
  {
    immediate: true,
  },
)

const { mutateAsync, isPending } = useMutation({
  mutationFn: async (input: File) => {
    return api.postForm(
      `products/images/${productID}`,
      { file: input },
      {
        headers: {
          Authorization: store.getAccessToken,
        },
      },
    )
  },
  async onSuccess() {
    await refetch()
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        errorMsg.value = 'Invalid image.'
      },
    })
  },
})

const onChange = async (event: Event) => {
  console.log('hello')
  const input = event.target as HTMLInputElement
  if (input.files === null) {
    return
  }
  const file = input.files[0]
  if (file === undefined) {
    return
  }
  await mutateAsync(file)
}
</script>

<template>
  <div class="upload-image-page">
    <BackBtn page-name="admin-products-list" :query="{ page: callback_page }" />
    <div class="show-box">
      <div class="empty-box c-flex-all-center" v-if="data === undefined">
        <h2>There are not any images.</h2>
      </div>
      <ImageFrameList v-else :data="data" />
    </div>
    <form>
      <label class="add-img-btn c-flex-all-center">
        <input
          class="cursor-pointer"
          ref="ImgRef"
          type="file"
          accept="image/*"
          @change="onChange($event)"
          :disabled="isPending"
        />
        <BaseIcon
          :icon="icons.pages.panel.products.upload"
          size="32px"
          stroke-color="none"
          fill-color="--color-c-form-btn-text"
        />
      </label>
      <p class="" v-if="errorMsg !== ''">
        {{ errorMsg }}
      </p>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.upload-image-page {
  @apply w-screen h-screen flex flex-col
    items-center;
}

.upload-image-page .show-box {
  @apply w-full border-b-2 p-2;
}

.upload-image-page .add-img-btn {
  @apply m-10 py-10 px-5 hover:brightness-90 cursor-pointer
      font-bold rounded-xl text-2xl
      bg-c-form-btn
      text-c-form-btn-text;
}

.show-box .empty-box {
  @apply p-10 text-2xl font-bold;
}
</style>
