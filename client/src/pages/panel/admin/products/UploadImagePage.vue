<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useMutation, useQuery } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { errorStatusHandler } from '@/utils/helper'
import { useNotificationStore } from '@/store/notification'
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
const notification = useNotificationStore()
const productID = route.params.productID
const callback_page = route.query.page
if (productID === undefined || productID === '') {
  router.push('/404')
}

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
        notFound() {
          // empty block
          return
        },
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
    notification.success('Image added.')
    await refetch()
  },
  onError(error) {
    errorStatusHandler(error, router, {
      badRequest() {
        notification.warning('Invalid photo.')
      },
      notFound() {
        notification.error('Photo not acceptable.')
        return
      },
    })
  },
})

const deleteImageMutation = useMutation({
  mutationFn: async (image: ProductImageResponse) => {
    return api.delete(`products/images/${productID}/${image.id}`, {
      headers: { Authorization: store.getAccessToken },
    })
  },
  onSuccess: async () => {
    notification.success('Image deleted.')
    await refetch()
  },
  onError: (error) => {
    errorStatusHandler(error, router, {
      notFound() {
        notification.error('Your photo not founded.')
        return
      },
      badRequest() {
        notification.warning('Invalid photo')
        return
      },
    })
  },
})

const onChange = async (event: Event) => {
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

window.scroll(0, 0)
</script>

<template>
  <div class="upload-image-page">
    <div class="w-full px-4">
      <BackBtn
        page-name="admin-products-list"
        :query="{ page: callback_page }"
      />
    </div>
    <div class="show-box">
      <div class="empty-box c-flex-all-center" v-if="data === undefined">
        <h2>There are not any images.</h2>
      </div>
      <div v-else class="image-controls c-flex-all-center">
        <ImageFrameList :data="data" />
        <div class="delete-list">
          <button
            v-for="image in data"
            :key="image.id"
            type="button"
            :disabled="deleteImageMutation.isPending.value"
            @click="deleteImageMutation.mutate(image)"
          >
            Delete {{ image.img_path }}
          </button>
        </div>
      </div>
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
          fill-color="--c-v-8-text"
        />
      </label>
    </form>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.upload-image-page {
  @apply w-screen h-screen flex flex-col
    items-center;
}

.upload-image-page .image-controls {
  @apply flex-row gap-4 px-10;
}

.upload-image-page .delete-list {
  @apply flex flex-wrap justify-center gap-2;
}

.upload-image-page .delete-list button {
  @apply px-3 py-2 rounded-lg font-bold cursor-pointer
    bg-(--c-v-11) text-(--c-v-7) disabled:opacity-50
    hover:brightness-110
    max-w-120 truncate;
}

.upload-image-page .show-box {
  @apply w-full border-b-2 p-2;
}

.upload-image-page .add-img-btn {
  @apply m-10 py-10 px-5 hover:brightness-90 cursor-pointer
      font-bold rounded-xl text-2xl
      bg-(--c-v-8)
      text-(--c-v-7);
}

.show-box .empty-box {
  @apply p-10 text-2xl font-bold;
}
</style>
