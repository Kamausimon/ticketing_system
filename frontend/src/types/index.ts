// Base types matching your Go backend models

export interface Event {
  id: number
  title: string
  description: string
  start_date: string
  end_date: string
  venue: string
  city: string
  country: string
  category: string
  status: 'draft' | 'published' | 'cancelled' | 'completed'
  organizer_id: number
  total_capacity: number
  available_capacity: number
  image_url?: string
  created_at: string
  updated_at: string
}

export interface TicketClass {
  id: number
  event_id: number
  name: string
  description: string
  price: number
  quantity_available: number
  quantity_sold: number
  sales_start_date: string
  sales_end_date: string
  min_per_order: number
  max_per_order: number
  is_active: boolean
}

export interface Order {
  id: number
  account_id: number
  event_id: number
  first_name: string
  last_name: string
  email: string
  total_amount: number
  currency: string
  status: 'pending' | 'paid' | 'completed' | 'cancelled' | 'refunded'
  payment_status: 'pending' | 'processing' | 'completed' | 'failed' | 'refunded'
  order_date: string
  completed_at?: string
  order_items: OrderItem[]
}

export interface OrderItem {
  id: number
  order_id: number
  ticket_class_id: number
  quantity: number
  unit_price: number
  total_price: number
}

export interface Ticket {
  id: number
  ticket_number: string
  qr_code: string
  holder_name: string
  holder_email: string
  status: 'active' | 'used' | 'cancelled' | 'refunded'
  checked_in_at?: string
  event_title?: string
  ticket_class?: string
}

export interface User {
  id: number
  email: string
  first_name: string
  last_name: string
  phone_number?: string
  email_verified: boolean
  two_factor_enabled: boolean
  created_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  first_name: string
  last_name: string
  phone_number?: string
}

export interface PaymentInitiateRequest {
  order_id: number
  amount: number
  currency: string
  payment_method: 'mpesa' | 'card'
  phone_number?: string
  email: string
}

export interface PaymentResponse {
  success: boolean
  transaction_id: string
  status: string
  payment_url?: string
  message: string
}
