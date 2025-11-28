import axios from 'axios'

const baseURL = import.meta.env.BACKEND_BASE_URL || ''

export const client = axios.create({
	baseURL: baseURL,
})
