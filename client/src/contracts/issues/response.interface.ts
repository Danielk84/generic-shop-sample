export interface IssuesSummaryResponse {
  id: string
  user_id: string
  is_done: boolean
}

export interface IssuesResponse extends IssuesSummaryResponse {
  req: string
  res: string
}
