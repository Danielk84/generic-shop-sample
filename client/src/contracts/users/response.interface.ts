export interface UserResponse {
  id: string
  name: string
  permission_type: number
  is_active: boolean
  is_verified: boolean
}

export interface UserDetailResponse extends UserResponse {
  email: string
  is_v_email: string
  phone_number: string
  is_v_phone_number: string
  national_code: string
}

export interface ShopInfoResponse extends UserResponse {
	// user_s.users table
	email: string
	is_v_email: boolean
	phone_number: string
	is_v_phone_number: boolean

	// user_s.shop table
	brand: string
	shop_addr: string
	zip_code: string

	business_code: string
	shop_phone_number: string

	img_path: string
	bio: string
	is_shop: boolean
}

export interface ShopResponse {
	user_id: string
	brand: string
	phone_number: string
	is_shop: boolean
}
