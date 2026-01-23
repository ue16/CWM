import { client } from './client'

const API_BASE = '/api/v2'

export interface Attachment {
	id: number
	filename: string
	size: string
	type: string
	url: string
}

export interface AttachmentUrlResponse {
	url: string
}

/**
 * Get a pre-signed URL for downloading an attachment
 */
export const getAttachmentUrl = async (cardId: number, attachmentId: number): Promise<string> => {
	const response = await client.get<AttachmentUrlResponse>(
		`${API_BASE}/card/${cardId}/attachment/${attachmentId}/url`,
	)
	return response.data.url
}

/**
 * Download an attachment by fetching the pre-signed URL and triggering the download
 */
export const downloadAttachment = async (
	cardId: number,
	attachmentId: number,
	filename: string,
): Promise<void> => {
	try {
		// Get the pre-signed URL from the backend
		const url = await getAttachmentUrl(cardId, attachmentId)

		// Use fetch to download the file and trigger the browser's download dialog
		const response = await fetch(url)
		if (!response.ok) {
			throw new Error(`Failed to download file: ${response.statusText}`)
		}

		// Get the file blob
		const blob = await response.blob()

		// Create a temporary anchor element with the blob to trigger the download
		const link = document.createElement('a')
		link.href = URL.createObjectURL(blob)
		link.download = filename
		link.style.display = 'none'

		// Append to body, click, and remove
		document.body.appendChild(link)
		link.click()
		document.body.removeChild(link)

		// Clean up the object URL
		URL.revokeObjectURL(link.href)
	} catch (error) {
		console.error('Failed to download attachment:', error)
		throw error
	}
}
