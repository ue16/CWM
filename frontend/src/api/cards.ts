import { client } from './client'

const API_BASE = '/api/v2'

export interface Card {
	id: number
	title: string
	description?: string
	column_id: number
	position: number
}

export const fetchCards = async (columnId: number): Promise<Card[]> => {
	const response = await client.get(`${API_BASE}/card/${columnId}`)
	return response.data
}

export const createCard = async (
	columnId: number,
	title: string,
	description?: string,
): Promise<Card> => {
	const response = await client.post(`${API_BASE}/card/`, {
		column_id: columnId,
		title,
		description,
	})
	return response.data
}

export const deleteCard = async (id: number): Promise<void> => {
	await client.delete(`${API_BASE}/card/`, { data: { id } })
}
