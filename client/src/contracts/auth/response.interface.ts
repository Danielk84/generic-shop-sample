export interface AccessTokenResponse {
  token: string
}

export interface AuthClaims {
  id: string
  email: string
  permission_type: number
}
