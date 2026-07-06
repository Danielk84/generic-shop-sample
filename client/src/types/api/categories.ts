export interface Category {
  tag: string
  backgroundImage?: string
}

export interface OfferCategory extends Category {
  percent: number
}