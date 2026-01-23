/**
 * Get a visual icon for a file based on its MIME type
 */
export function getFileIcon(mimeType: string | null | undefined): string {
	if (!mimeType) return '📄'

	const type = (mimeType || '').toLowerCase()

	// Images
	if (
		type.startsWith('image/') ||
		type.includes('jpeg') ||
		type.includes('jpg') ||
		type.includes('png') ||
		type.includes('gif') ||
		type.includes('webp')
	) {
		return '🖼️'
	}

	// PDF documents
	if (type.includes('pdf')) {
		return '📕'
	}

	// Word documents
	if (type.includes('word') || type.includes('document')) {
		return '📝'
	}

	// Spreadsheets
	if (type.includes('spreadsheet') || type.includes('sheet') || type.includes('excel')) {
		return '📊'
	}

	// Presentations
	if (type.includes('presentation') || type.includes('powerpoint')) {
		return '🎨'
	}

	// Video types
	if (
		type.startsWith('video/') ||
		type.includes('mp4') ||
		type.includes('webm') ||
		type.includes('avi') ||
		type.includes('mov') ||
		type.includes('mkv')
	) {
		return '🎬'
	}

	// Audio types
	if (
		type.startsWith('audio/') ||
		type.includes('mp3') ||
		type.includes('wav') ||
		type.includes('flac') ||
		type.includes('ogg') ||
		type.includes('aac')
	) {
		return '🎵'
	}

	// Archive types
	if (
		type.includes('zip') ||
		type.includes('rar') ||
		type.includes('7z') ||
		type.includes('archive') ||
		type.includes('gzip') ||
		type.includes('tar')
	) {
		return '📦'
	}

	// Code/Text types
	if (
		type.includes('json') ||
		type.includes('xml') ||
		type.includes('javascript') ||
		type.includes('typescript') ||
		type.includes('python') ||
		type.includes('code') ||
		type.includes('text') ||
		type.includes('plain')
	) {
		return '💻'
	}

	// Default fallback
	return '📄'
}

/**
 * Format file size in human-readable format
 * @param sizeStr - File size as string or number (in bytes)
 * @returns Formatted file size string (e.g., "1.5 MB")
 */
export function formatFileSize(sizeStr: string | number | null | undefined): string {
	if (!sizeStr) return '0 B'

	// If it's already a formatted string (e.g., "1.5 MB"), return it as-is
	if (typeof sizeStr === 'string' && /[A-Za-z]/.test(sizeStr)) {
		return sizeStr
	}

	// Convert to number and check if valid
	const bytes = typeof sizeStr === 'number' ? sizeStr : parseInt(sizeStr, 10)
	if (isNaN(bytes)) return String(sizeStr)

	const units = ['B', 'KB', 'MB', 'GB', 'TB']
	let size = bytes
	let unitIndex = 0

	while (size >= 1024 && unitIndex < units.length - 1) {
		size /= 1024
		unitIndex++
	}

	return size.toFixed(1) + ' ' + units[unitIndex]
}

/**
 * Truncate a filename while preserving the extension
 * @param filename - The filename to truncate
 * @param maxLength - Maximum length of the result (default: 20)
 * @returns Truncated filename with extension preserved
 */
export function truncateFilename(filename: string | null | undefined, maxLength: number = 20): string {
	if (!filename) return 'Unknown'
	if (filename.length <= maxLength) return filename

	const lastDotIndex = filename.lastIndexOf('.')
	if (lastDotIndex === -1) {
		// No extension found
		return filename.substring(0, maxLength) + '...'
	}

	const name = filename.substring(0, lastDotIndex)
	const ext = filename.substring(lastDotIndex)

	// Calculate available space for the name part
	const availableLength = maxLength - ext.length - 3 // 3 for "..."
	if (availableLength <= 0) {
		// Extension is too long, just truncate the whole thing
		return filename.substring(0, maxLength) + '...'
	}

	return name.substring(0, availableLength) + '...' + ext
}
