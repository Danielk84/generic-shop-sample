export interface BaseBtn {
  click?: (event: MouseEvent) => void,
  style?: {
    backgroundColor?: string
    outlineColor?: string
    outlineWidth?: string
    color?: string
  }
}