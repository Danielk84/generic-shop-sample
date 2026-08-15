export interface SVGIcon {
  size?: string
  strokeColor?: string
  fillColor?: string
}

export interface Icon extends SVGIcon {
  icon: string
}