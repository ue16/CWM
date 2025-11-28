import { client } from './client'

const API_BASE = '/api/v2'

export interface Column {
	id: number
	name: string
	board_id: number
	position: number
}

export const fetchColumns = async (boardId: number): Promise<Column[]> => {
	const response = await client.get(`${API_BASE}/column/${boardId}`)
	return response.data
}

export const createColumn = async (boardId: number, name: string): Promise<Column> => {
	const response = await client.post(`${API_BASE}/column/`, { board_id: boardId, name })
	return response.data
}

export const deleteColumn = async (id: number): Promise<void> => {
	await client.delete(`${API_BASE}/column/`, { data: { id } })
}

export const moveColumn = async (column: Column, newPosition: number): Promise<Column> => {
	const response = await client.put(`${API_BASE}/column/`, {
		id: column.id,
		name: column.name,
		board_id: column.board_id,
		position: newPosition,
	})
	return response.data
}
