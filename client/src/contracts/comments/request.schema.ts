export interface CommentResponse {
  id: string
  name: string
  pub_date: string
  children_amount: number
  body: string
}

export interface RelatedCommentResponse extends CommentResponse {
  parent: string
  referrer: string
  is_active: string
}