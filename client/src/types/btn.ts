export interface Btn {
  style?: {
    backgroundColor?: string
    outlineColor?: string
    outlineWidth?: string
    color?: string
  }
}

export interface DestroyEmits {
  (e: "destroy"): void
}