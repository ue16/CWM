import { client } from './client'

const API_BASE = '/api/v2'

export interface Board {
	id: number
	name: string
}

export const fetchBoards = async (): Promise<Board[]> => {
	const response = await client.get(`${API_BASE}/boards`)
	return response.data
}

export const fetchBoard = async (id: number): Promise<Board> => {
	const response = await client.get(`${API_BASE}/board/${id}`)
	return response.data
}

export const createBoard = async (name: string): Promise<Board> => {
	const response = await client.post(`${API_BASE}/board/`, { name })
	return response.data
}

export const deleteBoard = async (id: number): Promise<void> => {
	await client.delete(`${API_BASE}/board/`, { data: { id } })
}
