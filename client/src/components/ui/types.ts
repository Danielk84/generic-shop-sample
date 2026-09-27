export interface SVGIcon {
  size?: string

  strokeColor?: string
  strokeDarkColor?: string

  fillColor?: string
  fillDarkColor?: string
}

export interface Icon extends SVGIcon {
  icon: string
}
