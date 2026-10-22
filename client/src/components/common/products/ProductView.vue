<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMutation, useQuery } from '@tanstack/vue-query'
import axios from 'axios'

import api from '@/utils/api'
import icons from '@/utils/icons'
import { useStore } from '@/store'
import { useNotificationStore } from '@/store/notification'
import { errorStatusHandler, setCallbackURL } from '@/utils/helper'
import { isValidID } from '@/utils/validator'
import type { ProductProperty } from '@/contracts/products/response.interface'
import type {
  ProductResponse,
  ProductImageResponse,
} from '@/contracts/products/response.interface'

const BaseIcon = defineAsyncComponent(
  () => import('@/components/ui/BaseIcon.vue'),
)
const ImageFrameList = defineAsyncComponent(
  () => import('@/components/common/products/ImageFrameList.vue'),
)
const CommentsList = defineAsyncComponent(
  () => import('@/components/common/comments/CommentsList.vue'),
)

const props = defineProps<{
  queryKey: string
  queryPath: string
}>()

const route = useRoute()
const router = useRouter()
const store = useStore()
const notification = useNotificationStore()
const id = String(route.params.productID ?? '')

const selectedVariantIndex = ref(0)

const { data: product, error } = useQuery({
  queryKey: [props.queryKey, id],
  queryFn: async () =>
    api.get<ProductResponse>(`${props.queryPath}/${id}`, {
      headers: {
        Authorization: store.hasAccessToken ? store.getAccessToken : undefined,
      },
    }),
  select: (res) => res.data,
  enabled: () => isValidID(id),
})

watch(
  error,
  (err) => {
    if (err !== null) {
      errorStatusHandler(err, router)
    }
  },
  {
    immediate: true,
  },
)

const { data: images } = useQuery({
  queryKey: ['product-images', id],
  queryFn: async () => api.get<ProductImageResponse[]>(`products/images/${id}`),
  select: (res) => res.data,
  enabled: () => isValidID(id),
})

const selectedVariant = computed(
  () => product.value?.variant_detail[selectedVariantIndex.value] ?? null,
)

const selectedProperty = computed<ProductProperty>(
  () => selectedVariant.value?.property ?? {},
)

const selectedPrice = computed(
  () => selectedVariant.value?.price ?? product.value?.price ?? 0,
)

const addToBasketMutation = useMutation({
  mutationKey: ['product-add-to-basket'],
  mutationFn: async () => {
    if (product.value === undefined) {
      throw new Error('Product is not loaded')
    }

    // Reuse the open (unpaid) order if the backend already has one, otherwise
    // it creates a new order. Its id is what the basket page is routed with.
    const order = await api.post<{ order_id: string }>(
      'orders/',
      {},
      { headers: { Authorization: store.getAccessToken } },
    )
    const orderId = order.data.order_id

    try {
      await api.post(
        'orders/items/',
        {
          order_id: orderId,
          product_id: product.value.id,
          property: selectedProperty.value,
          price: selectedPrice.value,
        },
        { headers: { Authorization: store.getAccessToken } },
      )
    } catch (err) {
      // The order_items table is keyed by (order, product), so adding a product
      // that is already in the basket is rejected with a 404. Treat it as
      // "already added" and still send the user to the basket.
      if (axios.isAxiosError(err) && err.response?.status === 404) {
        return { orderId, alreadyAdded: true }
      }
      throw err
    }

    return { orderId, alreadyAdded: false }
  },
  onSuccess(res) {
    if (res.alreadyAdded) {
      notification.info('Product is already in your basket.')
    } else {
      notification.success('Product added to basket.')
    }
    router.push({ name: 'basket', params: { id: res.orderId } })
  },
  onError(err) {
    errorStatusHandler(err, router, {
      badRequest() {
        notification.error('Unable to add product to basket.')
        return
      },
      notFound() {
        notification.error('Product is no longer available.')
        return
      },
    })
  },
})

async function addToBasket() {
  if (selectedVariant.value === null) {
    notification.error('Choose a product variant.')
    return
  }
  if (!store.hasAccessToken) {
    // Keep the current product in the callback url so the user returns here
    // after logging in and can retry adding the product.
    router.push({
      name: 'auth',
      query: { callback_url: setCallbackURL(route.fullPath) },
    })
    return
  }
  await addToBasketMutation.mutateAsync()
}
</script>

<template>
  <section v-if="product !== undefined" class="product-page">
    <section class="intro">
      <section class="gallery">
        <ImageFrameList :data="images" />
      </section>
      <section class="info">
        <h1>{{ product.name }}</h1>
        <p class="price">{{ selectedPrice }}</p>
        <div class="meta-box">
          <p class="meta">Published: {{ product.pub_date }}</p>
          <p class="meta">Views: {{ product.view_counter }}</p>
        </div>
        <p class="stock" :class="{ unavailable: !product.is_available }">
          {{
            product.is_available
              ? `In stock: ${product.available_quantity}`
              : 'Unavailable'
          }}
        </p>
        <button
          class="add-to-basket"
          :disabled="
            !product.is_available ||
            selectedVariant === null ||
            addToBasketMutation.isPending.value
          "
          @click="addToBasket"
        >
          <BaseIcon
            :icon="icons.common.products.basket"
            size="42px"
            stroke-color="--c-v-7"
            stroke-dark-color="--c-v-7"
            fill-color="--c-v-7"
            fill-dark-color="--c-v-7"
          />
          <span>
            {{
              addToBasketMutation.isPending.value
                ? 'Adding...'
                : 'Add to basket'
            }}
          </span>
        </button>
      </section>
    </section>
    <section class="content-box">
      <section class="content">
        <section class="variant-selection">
          <h2>Choose Variants</h2>
          <div class="variant-options">
            <button
              v-for="(variant, idx) in product.variant_detail"
              :key="idx"
              class="variant-option"
              :class="{ selected: selectedVariantIndex === idx }"
              @click="selectedVariantIndex = idx"
              type="button"
            >
              <div class="variant-props">
                <span
                  class="wrap-break-word text-wrap"
                  v-for="(v, k) in variant.property"
                  :key="k"
                >
                  {{ k }}: {{ v }}
                </span>
              </div>
              <span class="variant-price">{{ variant.price }}</span>
            </button>
          </div>
          <p v-if="product.variant_detail.length === 0" class="no-variants">
            No variants available
          </p>
        </section>
        <section class="description">
          <h2>Description</h2>
          <p>{{ product.description }}</p>
        </section>
        <section
          v-if="Object.keys(product.common_detail).length > 0"
          class="props"
        >
          <h2>Details</h2>
          <ul>
            <li v-for="(v, k) in product.common_detail" :key="k">
              <span class="opacity-70 font-bold">{{ k }}</span>
              <span class="basis-3/4 border-b border-(--c-v-8) p-4">{{
                v
              }}</span>
            </li>
          </ul>
        </section>
      </section>
      <section class="comments">
        <h2>Comments</h2>
        <CommentsList
          :is-parent="true"
          :relation="{ parent: '', referrer: product.id }"
        />
      </section>
    </section>
  </section>
  <div v-else class="product-page c-flex-all-center">
    <p class="empty">There is not any content.</p>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.product-page {
  @apply w-full min-h-screen -mt-5
    flex flex-col gap-10 items-start
    max-md:flex-col max-md:p-4
    bg-radial
    text-(--c-v-7)
    from-(--c-v-0)
    to-(--c-v-2);
}

.product-page .content-box {
  @apply w-full h-fit
    rounded-t-4xl py-4 px-4 md:px-10
    bg-(--c-v-7) dark:bg-(--c-v-0)
    text-(--c-v-1) dark:text-(--c-v-7);
}

.product-page .intro {
  @apply w-full gap-6 md:gap-10 p-4 md:p-10
    flex flex-col md:flex-row items-center justify-between;
}

.product-page .info {
  @apply w-full md:w-1/2 min-h-120 flex flex-col justify-between gap-4;
}

.product-page .info h1 {
  @apply text-2xl md:text-4xl font-bold;
}

.product-page .price {
  @apply text-2xl md:text-3xl font-bold text-(--c-v-11)
    border-t-2 border-b-2 p-4 text-center
    rounded-2xl;
}

.product-page .meta-box {
  @apply border-t-2 border-b-2 border-(--c-v-9)
    rounded-2xl p-4;
}

.product-page .meta {
  @apply text-lg brightness-70;
}

.product-page .stock {
  @apply w-fit px-4 py-2 rounded-xl font-bold
    bg-(--c-v-14) text-(--c-v-1);
}

.product-page .stock.unavailable {
  @apply bg-(--c-v-11);
}

.product-page .content {
  @apply w-full;
}

.product-page section {
  @apply mt-4 flex flex-col gap-2;
}

.product-page section h2 {
  @apply text-xl md:text-2xl font-bold pb-2
    border-b border-(--c-v-4) dark:border-(--c-v-5);
}

.product-page .props ul {
  @apply flex flex-col gap-1;
}

.product-page .props li {
  @apply flex flex-col md:flex-row justify-between gap-2;
}

.product-page .variant-props {
  @apply flex flex-row gap-3 flex-wrap text-wrap wrap-break-word;
}

.product-page .variant-selection {
  @apply mb-6;
}

.product-page .variant-options {
  @apply flex flex-col gap-2;
}

.product-page .variant-option {
  @apply flex flex-col justify-between items-center
    p-3 rounded-xl border-2 transition-all
    border-(--c-v-5) bg-transparent cursor-pointer
    hover:border-(--c-v-8) hover:brightness-110;
}

.product-page .variant-option.selected {
  @apply border-(--c-v-14) bg-(--c-v-14)/10;
}

.product-page .variant-option .variant-props {
  @apply flex flex-row gap-3 flex-wrap;
}

.product-page .variant-price {
  @apply font-bold text-lg border-t-2 border-b-2 border-(--c-v-11) rounded-2xl w-full text-center py-2;
}

.product-page .no-variants {
  @apply text-(--c-v-4) text-center py-4;
}

.product-page .variant-select,
.product-page .variant-select-control {
  @apply block font-bold;
}

.product-page .variant-select-control {
  @apply w-full p-3 rounded-xl border border-(--c-v-5)
    bg-transparent text-(--c-v-1) cursor-pointer;
}

.product-page .gallery {
  @apply w-full md:w-1/2;
}

@media (max-width: 768px) {
  .product-page .intro {
    @apply flex-col;
  }

  .product-page .gallery {
    @apply w-full;
  }

  .product-page .info {
    @apply w-full;
  }
}

.product-page .add-to-basket {
  @apply w-full md:w-80/100 min-h-14 px-4 rounded-xl font-bold cursor-pointer
    bg-(--c-v-14) text-(--c-v-7) hover:brightness-95
    disabled:opacity-50 disabled:cursor-not-allowed
    flex flex-row items-center justify-center;
}

.product-page .comments {
  @apply w-full px-4 md:px-10;
}

.product-page .comments h2 {
  @apply text-xl md:text-2xl font-bold mb-4;
}

.product-page .empty {
  @apply w-full text-center text-xl md:text-2xl font-bold;
}
</style>
