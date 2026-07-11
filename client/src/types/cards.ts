export interface Seller {
  to: string
  username: string
  totalSells: number
  followers: number
  backgroundImage?: string
}

export interface PriceOfferFrame {
  to: string
  category: string
  percent: number
  backgroundImage?: string
}

export type ImageFrame = {
  src: string
  alt?: string
}