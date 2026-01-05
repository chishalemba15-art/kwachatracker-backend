import axios, { AxiosInstance } from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://kwachatracker-api.onrender.com'

class ApiClient {
    private client: AxiosInstance

    constructor() {
        this.client = axios.create({
            baseURL: API_URL,
            headers: {
                'Content-Type': 'application/json',
            },
        })

        // Add auth token to requests
        this.client.interceptors.request.use((config) => {
            if (typeof window !== 'undefined') {
                const token = localStorage.getItem('admin_token')
                if (token) {
                    config.headers.Authorization = `Bearer ${token}`
                }
            }
            return config
        })
    }

    // Auth
    async login(username: string, password: string) {
        const response = await this.client.post('/api/v1/admin/login', {
            username,
            password,
        })
        return response.data
    }

    // Stats
    async getStats() {
        const response = await this.client.get('/api/v1/admin/stats')
        return response.data
    }

    // Users
    async getUsers(params?: {
        page?: number
        limit?: number
        filter?: string
    }) {
        const response = await this.client.get('/api/v1/admin/users', { params })
        return response.data
    }

    async getUser(id: string) {
        const response = await this.client.get(`/api/v1/admin/users/${id}`)
        return response.data
    }

    // Insights
    async getInsights(params?: {
        page?: number
        limit?: number
        user_id?: string
        date_from?: string
        date_to?: string
    }) {
        const response = await this.client.get('/api/v1/admin/insights', { params })
        return response.data
    }

    async triggerInsights(userId?: string) {
        const response = await this.client.post('/api/v1/admin/insights/trigger', {
            user_id: userId,
        })
        return response.data
    }

    // Notifications
    async broadcast(data: {
        title: string
        body: string
        target: 'all' | 'active' | 'specific'
        user_ids?: string[]
        scheduled_for?: string
    }) {
        const response = await this.client.post('/api/v1/admin/broadcast', data)
        return response.data
    }

    async getNotificationHistory(params?: {
        page?: number
        limit?: number
    }) {
        const response = await this.client.get('/api/v1/admin/notifications', { params })
        return response.data
    }

    // Transactions
    async getTransactions(params?: {
        page?: number
        limit?: number
        user_id?: string
        category?: string
        date_from?: string
        date_to?: string
    }) {
        const response = await this.client.get('/api/v1/admin/transactions', { params })
        return response.data
    }
}

export const api = new ApiClient()
