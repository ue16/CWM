import axios from 'axios'
import { authAPI } from './auth'

const baseURL = import.meta.env.BACKEND_BASE_URL || ''

export const client = axios.create({
	baseURL: baseURL,
})

// Add request interceptor to include Authorization header with Bearer token
client.interceptors.request.use((config) => {
	const token = authAPI.getToken()
	if (token) {
		config.headers['Authorization'] = `${token}`
	}
	return config
})

// Add response interceptor to handle 401 errors
client.interceptors.response.use(
	(response) => response,
	(error) => {
		if (error.response?.status === 401) {
			// Token is invalid or expired, clear stored data and redirect to login
			authAPI.logout()
			window.location.href = '/login'
		}
		return Promise.reject(error)
	},
)
