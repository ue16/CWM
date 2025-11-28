import { ref } from 'vue'

export interface DragData {
	type: 'card' | 'column'
	payload: unknown
}

export function useDragDrop() {
	const isDragOver = ref(false)

	function setDragData(event: DragEvent, data: DragData) {
		if (!event.dataTransfer) return
		event.dataTransfer.effectAllowed = 'move'
		event.dataTransfer.setData('application/json', JSON.stringify(data))
	}

	function getDragData(event: DragEvent): DragData | null {
		if (!event.dataTransfer) return null
		try {
			const data = event.dataTransfer.getData('application/json')
			return JSON.parse(data) as DragData
		} catch {
			return null
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault()
		if (event.dataTransfer) {
			event.dataTransfer.dropEffect = 'move'
		}
		isDragOver.value = true
	}

	function handleDragLeave(event: DragEvent, element: HTMLElement) {
		if (event.target === element) {
			isDragOver.value = false
		}
	}

	function handleDragEnd() {
		isDragOver.value = false
	}

	return {
		isDragOver,
		setDragData,
		getDragData,
		handleDragOver,
		handleDragLeave,
		handleDragEnd,
	}
}
