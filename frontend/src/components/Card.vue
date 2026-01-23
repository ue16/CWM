<template>
	<div
		class="card"
		:data-card-id="card.id"
		draggable="true"
		@dragstart="handleDragStart"
		@dragend="handleDragEnd"
	>
		<div class="card-actions">
			<button class="card-edit" @click="$emit('edit-card', card)" title="Edit card">
				Edit
			</button>
			<button class="card-delete" @click="$emit('delete-card', card.id)" title="Delete card">
				Delete
			</button>
		</div>
		<div class="card-title">{{ card.title }}</div>
		<div v-if="card.description" class="card-description">{{ card.description }}</div>

		<!-- Attachments Section -->
		<div v-if="hasAttachments" class="card-attachments">
			<div class="attachments-header">
				<span class="attachments-icon">📎</span>
				<span class="attachments-count">{{ card.attachments.length }}</span>
			</div>
			<div class="attachments-list">
				<div
					v-for="attachment in card.attachments"
					:key="attachment.id"
					class="attachment-item"
					:title="`${attachment.filename} (${formatFileSize(attachment.size)})`"
				>
					<span class="attachment-icon">{{ getFileIcon(attachment.type) }}</span>
					<span class="attachment-name">{{ truncateFilename(attachment.filename) }}</span>
					<span class="attachment-size">{{ formatFileSize(attachment.size) }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { getFileIcon, formatFileSize, truncateFilename } from '../utils'

export default {
	name: 'Card',
	props: {
		card: Object,
	},
	computed: {
		hasAttachments() {
			return this.card.attachments && this.card.attachments.length > 0
		},
	},
	methods: {
		handleDragStart(event) {
			event.dataTransfer.effectAllowed = 'move'
			this.$emit('dragstart', event)
		},
		handleDragEnd(event) {
			this.$emit('dragend', event)
		},
		getFileIcon,
		formatFileSize,
		truncateFilename,
	},
}
</script>

<style scoped>
/* All styles are defined in global styles.css */
</style>
