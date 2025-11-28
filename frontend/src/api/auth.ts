import { client } from './client'

export interface LoginRequest {
	username: string
	password: string
}

export interface RegisterRequest {
	username: string
	email: string
	password: string
}

export interface User {
	id: number
	username: string
	email: string
	created_at: string
}

export interface LoginResponse {
	token: string
	user: User
}

export interface AuthResponse {
	success: boolean
	message: string
	token?: string
	user?: User
}

export const authAPI = {
	login: async (username: string, password: string): Promise<LoginResponse> => {
		const response = await client.post('/api/v2/auth/login', {
			username,
			password,
		} as LoginRequest)
		return response.data
	},

	register: async (username: string, email: string, password: string): Promise<AuthResponse> => {
		const response = await client.post('/api/v2/auth/register', {
			username,
			email,
			password,
		} as RegisterRequest)
		return response.data
	},

	verify: async (token: string): Promise<User> => {
		const response = await client.get('/api/v2/auth/verify', {
			headers: {
				Authorization: `Bearer ${token}`,
			},
		})
		return response.data
	},

	logout: (): void => {
		localStorage.removeItem('token')
		localStorage.removeItem('user')
	},

	setToken: (token: string): void => {
		localStorage.setItem('token', token)
		client.defaults.headers.common['Authorization'] = `Bearer ${token}`
	},

	getToken: (): string | null => {
		return localStorage.getItem('token')
	},

	setUser: (user: User): void => {
		localStorage.setItem('user', JSON.stringify(user))
	},

	getUser: (): User | null => {
		const user = localStorage.getItem('user')
		return user ? JSON.parse(user) : null
	},

	isAuthenticated: (): boolean => {
		return !!localStorage.getItem('token')
	},
}
