import { type Ref } from "vue"

import { getCssUrl } from "@/utils/helper"

export function imageOnLoadHook(
  imgTarget: Ref<HTMLImageElement | null>,
  bgTarget: Ref<HTMLDivElement | null>,
) {
  return () => {
    if (imgTarget.value === null || bgTarget.value === null) {
      throw new TypeError("null imgTarget or bgTarget")
    }
    const img = imgTarget.value
    const src = img.getAttribute("src")
    if (src === null) {
      throw new TypeError("null src in img")
    }

    const style = bgTarget.value.style
    style.backgroundImage = getCssUrl(src) as string
    style.backgroundRepeat = "repeat"
    style.backgroundSize = "cover"
  }
}