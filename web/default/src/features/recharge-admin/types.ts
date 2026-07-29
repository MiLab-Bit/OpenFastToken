export interface AdminTopUp {
  id: number
  user_id: number
  amount: number
  money: number
  trade_no: string
  payment_method: string
  payment_provider: string
  create_time: number
  complete_time: number
  status: string
}

export interface AdminTopUpsData {
  page: number
  page_size: number
  total: number
  items: AdminTopUp[]
}

export interface AdminTopUpsResponse {
  success: boolean
  message?: string
  data: AdminTopUpsData
}
