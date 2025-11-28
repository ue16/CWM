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
				<div v-if="error" class="error-message">{{ error }}</div>
				<div class="modal-actions">
					<button type="button" @click="closeModal" class="btn-cancel">Cancel</button>
					<button type="submit" :disabled="isSubmitting" class="btn-submit">
						{{ isSubmitting ? 'Saving...' : 'Save Changes' }}
					</button>
				</div>
			</form>
		</div>
	</div>
</template>

<script>
import { updateCard } from '../api/cards'

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
			},
			isSubmitting: false,
			error: '',
		}
	},
	watch: {
		card(newCard) {
			if (newCard) {
				this.editedCard = {
					id: newCard.id,
					title: newCard.title,
					description: newCard.description || '',
					column_id: newCard.column_id,
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
				}
				this.error = ''
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
	border-color: #4CAF50;
	box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}

.form-group textarea {
	resize: vertical;
	min-height: 100px;
}

.error-message {
	color: #d32f2f;
	padding: 12px;
	background-color: #ffebee;
	border-radius: 4px;
	margin-bottom: 16px;
	font-size: 0.9rem;
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
	background-color: #4CAF50;
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
