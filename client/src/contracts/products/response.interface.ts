import { ProductProperty } from './request.schema'

export type ProductProperty = Record<string, string>

export interface ProductSummaryResponse {
  id: string
  name: string
  price: number
  img_path: string
  pub_date: string
}

export interface ProductStatusResponse extends ProductSummaryResponse {
  available_quantity: number
  is_available: boolean
  is_active: boolean
}

export interface ProductVendor {
  user_id: string
  quantity: number
}

export interface ProductVariantDetail {
  property: ProductProperty
  price: number
  vendors: ProductVendor[]
}

export interface ProductResponse {
  id: string
  name: string
  price: number
  pub_date: string
  available_quantity: number
  is_available: boolean
  is_active: boolean
  description: string
  common_detail: ProductProperty
  variant_detail: ProductVariantDetail[]
  view_counter: number
}

export interface Product {
  to: string
  name: string
  price: number
  backgroundImage?: string
}

export interface ProductImageResponse {
  id: string
  img_path: string
}
