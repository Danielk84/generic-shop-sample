export type ProductProperty = Record<string, string>

export interface ProductSummaryResponse {
  id: string
  name: string
  price: number
  pubDate: string
}

export interface ProductStatusResponse extends ProductSummaryResponse {
  available_quantity: number
  is_available: boolean
  is_active: boolean
}

export interface ProductResponse extends ProductStatusResponse{
  descriptions: string
  common_detail: ProductProperty
  variant_detail: Array<ProductProperty>
}

export interface ProductVendor {
  user_id: string
  quantity: number
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
