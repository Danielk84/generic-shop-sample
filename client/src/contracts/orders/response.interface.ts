import type {
  ProductProperty,
  ProductVendor,
} from '../products/response.interface'

export interface OrderID {
  id: string
  user_id: string
}

export interface OrderSummaryResponse extends OrderID {
  started_at: string
  is_paid: boolean
  is_delivered: boolean
}

export interface OrderUserInfo {
  address: string
  zip_code: string
}

export interface OrderResponse extends OrderSummaryResponse, OrderUserInfo {
  items_total: number
  total_bill: number
  is_verified: boolean
  is_confirmed: boolean
  payment_summary: ProductProperty[]
}

export interface OrderItem {
  order_id: string
  product_id: string
}

export interface OwnedOrderItemResponse extends OrderItem {
  items_total: number
  processed_items: number
  price: number
  confirmed_vendors: Array<ProductVendor>
  name: string
}

export interface OrderItemResponse extends OwnedOrderItemResponse {
  property: ProductProperty
}

export interface DailySalesResponse {
  date: string
  count: number
}
