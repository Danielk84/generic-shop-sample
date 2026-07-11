<script setup lang="ts">
import { defineAsyncComponent } from 'vue';

import type { DestroyEmits } from '@/types/btn';
import icons from '@/utils/icons-list';

const BaseBtn = defineAsyncComponent(() => import("@/components/common/button/BaseBtn.vue"))
const CloseBtn = defineAsyncComponent(() => import("@/components/common/button/CloseBtn.vue"))

const emits = defineEmits<DestroyEmits>()
</script>

<template>
  <div class="new-comment c-flex-all-center">
    <div class="box">
      <div class="close-btn">
        <CloseBtn @destroy="emits('destroy')"
          :icon="icons.closeCircle"
          size="30px"
          stroke-color="--color-secondary-theme"
          fill-color="none" />
      </div>
      <h3>comment your experience</h3>
      <div class="item-list c-flex-all-center">
        <div class="item">
          <label for="title">Title:</label>
          <input class="input-method" type="text" />
        </div>
        <div class="item">
          <label for="content">Experience:</label>
          <textarea class="input-method default-scrollbar" name="content" id="content"></textarea>
        </div>
      </div>
      <div class="btn">
        <BaseBtn>
          <span>Submit</span>
        </BaseBtn>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "@/styles/index.css";

.new-comment {
  @apply absolute z-60 inset-0 w-full h-full text-xl
    text-secondary-text backdrop-blur-lg;
}

.new-comment h3 {
  @apply text-3xl font-bold;
}

.new-comment .box {
  @apply flex flex-col items-center justify-between py-10 relative
    size-150 bg-primary-text rounded-2xl gap-10
    outline-5 outline-primary-theme;
}

.new-comment .box .close-btn {
  @apply absolute inset-0 top-3 left-3;
}

.new-comment .box .item-list {
  @apply flex-col gap-4;
}

.new-comment .box .item {
  @apply flex flex-col gap-1 hover:brightness-125;
}

.new-comment .box .item .input-method {
  @apply outline-4 text-xl outline-primary-theme hover:outline-4   
    w-100 p-2 max-h-35 rounded-lg;
}

.new-comment .box .btn {
  @apply w-60 h-15;
}
</style>