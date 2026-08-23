import type { ProductProperty } from "../products/response.interface"

export interface VendorOrderDelivere {
	user_id: string
	order_id: string
	product_id: string
	property: ProductProperty
  is_delivered: boolean
}

export interface VendorOrder extends VendorOrderDelivere {
	quantity: string
  total_bill: number
}


