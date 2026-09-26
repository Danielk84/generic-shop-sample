export interface CommentResponse {
  id: string
  name: string
  pub_date: string
  children_amount: number
  body: string
}

export interface RelatedCommentResponse extends CommentResponse {
  user_id: string
  parent: string
  referrer: string
  is_active: boolean
}
