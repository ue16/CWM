<template>
	<div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
		<div class="modal-content">
			<div class="modal-header">
				<h2>Edit Card</h2>
				<button class="modal-close" @click="closeModal">×</button>
			</div>
			<form @submit.prevent="handleSubmit" class="modal-form">
				<div class="form-group">
					<label for="card-title">Title</label>
					<input
						id="card-title"
						v-model="editedCard.title"
						type="text"
						placeholder="Card title"
						required
					/>
				</div>
				<div class="form-group">
					<label for="card-description">Description</label>
					<textarea
						id="card-description"
						v-model="editedCard.description"
						placeholder="Card description"
					></textarea>
				</div>

				<!-- Attachments Section -->
				<div class="attachments-section">
					<div class="attachments-section-header">
						<h3>Attachments</h3>
					</div>

					<!-- Existing Attachments List -->
					<div v-if="hasAttachments" class="existing-attachments">
						<div
							v-for="attachment in editedCard.attachments"
							:key="attachment.id"
							class="attachment-item"
							:title="`${attachment.filename} (${formatFileSize(attachment.size)})`"
						>
							<span class="attachment-icon">{{ getFileIcon(attachment.type) }}</span>
							<div class="attachment-info">
								<span class="attachment-name">{{
									truncateFilename(attachment.filename)
								}}</span>
								<span class="attachment-size">{{
									formatFileSize(attachment.size)
								}}</span>
							</div>
							<button
								type="button"
								class="attachment-delete-btn"
								@click="handleDeleteAttachment(attachment.id, attachment.filename)"
								:disabled="isDeletingId === attachment.id"
								title="Delete attachment"
							>
								{{ isDeletingId === attachment.id ? '⏳' : '✕' }}
							</button>
						</div>
					</div>

					<!-- No Attachments Message -->
					<div v-else class="no-attachments">No attachments yet</div>

					<!-- Upload Error -->
					<div v-if="uploadError" class="upload-error">
						{{ uploadError }}
					</div>

					<!-- Upload Progress -->
					<div v-if="uploadProgress" class="upload-progress">
						<div class="progress-bar">
							<div
								class="progress-fill"
								:style="{ width: uploadProgress + '%' }"
							></div>
						</div>
						<span class="progress-text">{{ uploadProgress }}%</span>
					</div>

					<!-- Upload Input -->
					<div class="upload-input-wrapper">
						<input
							ref="fileInput"
							type="file"
							class="file-input"
							@change="handleFileSelected"
							:disabled="isUploading || isSubmitting"
						/>
						<button
							type="button"
							class="upload-button"
							@click="$refs.fileInput.click()"
							:disabled="isUploading || isSubmitting"
						>
							{{ isUploading ? 'Uploading...' : '📤 Upload Attachment' }}
						</button>
					</div>
				</div>

				<div v-if="error" class="error-message">{{ error }}</div>
				<div class="modal-actions">
					<button type="button" @click="closeModal" class="btn-cancel">Cancel</button>
					<button
						type="submit"
						:disabled="isSubmitting || isUploading"
						class="btn-submit"
					>
						{{ isSubmitting ? 'Saving...' : 'Save Changes' }}
					</button>
				</div>
			</form>
		</div>
	</div>
</template>

<script>
import { updateCard } from '../api/cards'
import {
	getFileIcon,
	formatFileSize,
	truncateFilename,
	uploadAttachmentFile,
	deleteAttachmentFile,
} from '../utils/fileHelpers'

export default {
	name: 'EditCardModal',
	props: {
		isOpen: {
			type: Boolean,
			required: true,
		},
		card: {
			type: Object,
			default: null,
		},
	},
	data() {
		return {
			editedCard: {
				id: null,
				title: '',
				description: '',
				column_id: null,
				attachments: [],
			},
			isSubmitting: false,
			isUploading: false,
			isDeletingId: null,
			error: '',
			uploadError: '',
			uploadProgress: 0,
		}
	},
	computed: {
		hasAttachments() {
			return this.editedCard.attachments && this.editedCard.attachments.length > 0
		},
	},
	watch: {
		card(newCard) {
			if (newCard) {
				this.editedCard = {
					id: newCard.id,
					title: newCard.title,
					description: newCard.description || '',
					column_id: newCard.column_id,
					attachments: newCard.attachments || [],
				}
			}
		},
		isOpen(newVal) {
			if (newVal && this.card) {
				this.editedCard = {
					id: this.card.id,
					title: this.card.title,
					description: this.card.description || '',
					column_id: this.card.column_id,
					attachments: this.card.attachments || [],
				}
				this.error = ''
				this.uploadError = ''
				this.uploadProgress = 0
			}
		},
	},
	methods: {
		async handleSubmit() {
			if (!this.editedCard.title.trim()) {
				this.error = 'Title is required'
				return
			}

			this.isSubmitting = true
			this.error = ''

			try {
				await updateCard(this.editedCard)
				this.$emit('card-updated', this.editedCard)
				this.closeModal()
			} catch (err) {
				console.error('Error updating card:', err)
				if (err.response?.status === 401) {
					this.error = 'Unauthorized. Please log in again.'
				} else if (err.response?.status === 404) {
					this.error = 'Card not found.'
				} else {
					this.error = 'Failed to update card'
				}
			} finally {
				this.isSubmitting = false
			}
		},

		async handleFileSelected(event) {
			const file = event.target.files?.[0]
			if (!file) return

			// Reset file input
			this.$refs.fileInput.value = ''

			try {
				this.isUploading = true
				this.uploadError = ''
				this.uploadProgress = 0

				const finalizedAttachment = await uploadAttachmentFile(
					this.editedCard.id,
					file,
					(progress) => {
						this.uploadProgress = progress
					},
				)

				// Add to attachments list
				this.editedCard.attachments.push(finalizedAttachment)

				// Clear progress after a brief moment
				setTimeout(() => {
					this.uploadProgress = 0
				}, 500)
			} catch (err) {
				console.error('Error uploading attachment:', err)
				this.uploadError = err.message || 'Failed to upload attachment. Please try again.'
				this.uploadProgress = 0
			} finally {
				this.isUploading = false
			}
		},

		getFileIcon,
		formatFileSize,
		truncateFilename,

		async handleDeleteAttachment(attachmentId, filename) {
			if (!confirm(`Are you sure you want to delete "${filename}"?`)) {
				return
			}

			try {
				this.isDeletingId = attachmentId
				this.uploadError = ''

				await deleteAttachmentFile(this.editedCard.id, attachmentId)

				// Remove from local list
				this.editedCard.attachments = this.editedCard.attachments.filter(
					(att) => att.id !== attachmentId,
				)
			} catch (err) {
				console.error('Error deleting attachment:', err)
				this.uploadError = err.message || 'Failed to delete attachment. Please try again.'
			} finally {
				this.isDeletingId = null
			}
		},

		closeModal() {
			this.$emit('close')
		},
	},
}
</script>

<style scoped>
.modal-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
}

.modal-content {
	background-color: white;
	border-radius: 8px;
	box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
	width: 90%;
	max-width: 500px;
	max-height: 90vh;
	overflow-y: auto;
}

.modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 20px;
	border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
	margin: 0;
	font-size: 1.5rem;
	color: #333;
}

.modal-close {
	background: none;
	border: none;
	font-size: 1.5rem;
	cursor: pointer;
	color: #666;
	padding: 0;
	width: 32px;
	height: 32px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 4px;
	transition: background-color 0.2s;
	font-family: var(--font-family-emoji);
}

.modal-close:hover {
	background-color: #f0f0f0;
}

.modal-form {
	padding: 20px;
}

.form-group {
	margin-bottom: 16px;
}

.form-group label {
	display: block;
	margin-bottom: 8px;
	font-weight: 600;
	color: #333;
}

.form-group input,
.form-group textarea {
	width: 100%;
	padding: 8px 12px;
	border: 1px solid #ddd;
	border-radius: 4px;
	font-family: inherit;
	font-size: 1rem;
	box-sizing: border-box;
}

.form-group input:focus,
.form-group textarea:focus {
	outline: none;
	border-color: #4caf50;
	box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}

.form-group textarea {
	resize: vertical;
	min-height: 100px;
}

/* Attachments Section */
.attachments-section {
	margin-bottom: 16px;
	padding: 16px;
	background-color: #f9f9f9;
	border-radius: 6px;
	border: 1px solid #e0e0e0;
}

.attachments-section-header {
	margin-bottom: 12px;
}

.attachments-section-header h3 {
	margin: 0;
	font-size: 0.95rem;
	font-weight: 600;
	color: #333;
}

.existing-attachments {
	display: flex;
	flex-direction: column;
	gap: 8px;
	margin-bottom: 12px;
}

.attachment-item {
	display: flex;
	align-items: center;
	gap: 10px;
	padding: 8px 10px;
	background-color: white;
	border: 1px solid #e0e0e0;
	border-radius: 4px;
	transition: all 0.2s;
}

.attachment-item:hover {
	background-color: #fafafa;
	border-color: #d0d0d0;
}

.attachment-icon {
	flex-shrink: 0;
	font-size: 16px;
	font-family: var(--font-family-emoji);
}

.attachment-info {
	display: flex;
	flex-direction: column;
	flex: 1;
	min-width: 0;
	gap: 2px;
}

.attachment-name {
	font-weight: 500;
	color: #333;
	font-size: 0.9rem;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.attachment-size {
	font-size: 0.8rem;
	color: #999;
}

.attachment-delete-btn {
	flex-shrink: 0;
	background: none;
	border: none;
	color: #d32f2f;
	cursor: pointer;
	padding: 4px 8px;
	font-size: 1rem;
	border-radius: 4px;
	transition: all 0.2s;
	font-family: var(--font-family-emoji);
}

.attachment-delete-btn:hover:not(:disabled) {
	background-color: #ffebee;
}

.attachment-delete-btn:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.no-attachments {
	text-align: center;
	padding: 12px;
	color: #999;
	font-size: 0.9rem;
	font-style: italic;
	margin-bottom: 12px;
}

.upload-error {
	padding: 8px 10px;
	background-color: #ffebee;
	border: 1px solid #ef5350;
	border-radius: 4px;
	color: #d32f2f;
	font-size: 0.85rem;
	margin-bottom: 12px;
	border-left: 3px solid #d32f2f;
}

.upload-progress {
	display: flex;
	align-items: center;
	gap: 10px;
	margin-bottom: 12px;
}

.progress-bar {
	flex: 1;
	height: 6px;
	background-color: #e0e0e0;
	border-radius: 3px;
	overflow: hidden;
}

.progress-fill {
	height: 100%;
	background-color: #4caf50;
	transition: width 0.3s ease;
}

.progress-text {
	font-size: 0.8rem;
	color: #666;
	min-width: 35px;
	text-align: right;
}

.upload-input-wrapper {
	display: flex;
	gap: 8px;
}

.file-input {
	display: none;
}

.upload-button {
	flex: 1;
	padding: 8px 12px;
	background-color: #4caf50;
	color: white;
	border: none;
	border-radius: 4px;
	font-size: 0.9rem;
	cursor: pointer;
	transition: background-color 0.2s;
	font-family: inherit;
	font-weight: 500;
}

.upload-button:hover:not(:disabled) {
	background-color: #45a049;
}

.upload-button:disabled {
	background-color: #ccc;
	cursor: not-allowed;
}

.error-message {
	color: #d32f2f;
	padding: 12px;
	background-color: #ffebee;
	border-radius: 4px;
	margin-bottom: 16px;
	font-size: 0.9rem;
	border-left: 4px solid #d32f2f;
}

.modal-actions {
	display: flex;
	gap: 12px;
	justify-content: flex-end;
	margin-top: 20px;
}

.btn-cancel,
.btn-submit {
	padding: 10px 20px;
	border: none;
	border-radius: 4px;
	font-size: 1rem;
	cursor: pointer;
	transition: background-color 0.2s;
}

.btn-cancel {
	background-color: #f0f0f0;
	color: #333;
}

.btn-cancel:hover {
	background-color: #e0e0e0;
}

.btn-submit {
	background-color: #4caf50;
	color: white;
}

.btn-submit:hover:not(:disabled) {
	background-color: #45a049;
}

.btn-submit:disabled {
	background-color: #ccc;
	cursor: not-allowed;
}
</style>
