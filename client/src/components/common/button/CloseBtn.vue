<script setup lang="ts">

import type { Icon } from '@/types/icons';
import type { DestroyEmits } from '@/types/btn';
import { defineAsyncComponent } from 'vue';

const BaseIcon = defineAsyncComponent(() => import("@/components/common/BaseIcon.vue"))

const props = defineProps<Icon>()
const emits = defineEmits<DestroyEmits>()
</script>

<template>
  <button class="size-fit cursor-pointer" @click="emits('destroy')">
    <BaseIcon
      :icon="props.icon"
      :size="props.size"
      :fill-color="props.fillColor"
      :stroke-color="props.strokeColor" />
  </button>
</template>