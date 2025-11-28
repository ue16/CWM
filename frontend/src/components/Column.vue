<template>
	<div
		class="column"
		:data-column-id="column.id"
		draggable="true"
		@dragstart.self="handleColumnDragStart"
		@dragend="handleColumnDragEnd"
		@dragover.prevent="handleColumnDragOver"
		@dragleave="handleColumnDragLeave"
		@drop="handleColumnDrop"
		:class="{ 'drag-over': isColumnDragOver, dragging: isColumnDragging }"
	>
		<div class="column-header">
			<span class="column-name">{{ column.name }}</span>
			<button class="delete-button" @click="$emit('delete-column', column.id)">×</button>
		</div>
		<div
			class="cards-container"
			@dragover.prevent="handleCardContainerDragOver"
			@dragleave="handleCardContainerDragLeave"
			@drop="handleCardContainerDrop"
			:class="{ 'drag-over': isCardDragOver }"
		>
			<div v-if="column.cards && column.cards.length">
				<Card
					v-for="card in column.cards"
					:key="card.id"
					:card="card"
					@delete-card="handleDeleteCard"
					@edit-card="handleEditCard"
					@dragstart="handleCardDragStart"
					@dragend="handleCardDragEnd"
				/>
			</div>
			<div v-else class="no-cards">No cards</div>
		</div>
		<div v-if="error" class="error-message">{{ error }}</div>
		<form @submit.prevent="addCard" class="column-add-card-form">
			<input v-model="newCardTitle" placeholder="Card title" required />
			<textarea v-model="newCardDescription" placeholder="Description"></textarea>
			<button type="submit" :disabled="isCreating">
				{{ isCreating ? 'Adding...' : 'Add card' }}
			</button>
		</form>

		<EditCardModal
			:is-open="isEditModalOpen"
			:card="selectedCard"
			@close="closeEditModal"
			@card-updated="handleCardUpdated"
		/>
	</div>
</template>

<script>
import Card from './Card.vue'
import EditCardModal from './EditCardModal.vue'
import { createCard } from '../api/cards'
import { useDragDrop } from '../composables/useDragDrop'

export default {
	name: 'Column',
	components: { Card, EditCardModal },
	props: {
		column: {
			type: Object,
			required: true,
		},
		allColumns: Array,
	},
	setup() {
		const dragDrop = useDragDrop()
		return {
			...dragDrop,
		}
	},
	data() {
		return {
			newCardTitle: '',
			newCardDescription: '',
			isCreating: false,
			error: '',
			isColumnDragging: false,
			isColumnDragOver: false,
			isCardDragOver: false,
			isEditModalOpen: false,
			selectedCard: null,
		}
	},
	methods: {
		async addCard() {
			if (!this.newCardTitle.trim()) return

			this.isCreating = true
			this.error = ''

			try {
				if (!this.column.cards) {
					this.column.cards = []
				}
				const newCard = await createCard(
					this.column.id,
					this.newCardTitle,
					this.newCardDescription,
				)
				this.column.cards.push(newCard)
				this.newCardTitle = ''
				this.newCardDescription = ''
			} catch (err) {
				console.error('Error creating card:', err)
				if (err.response?.status === 404) {
					this.error = 'Column not found. You may not have access to this board.'
				} else if (err.response?.status === 401) {
					this.error = 'Unauthorized. Please log in again.'
				} else {
					this.error = 'Failed to create card'
				}
			} finally {
				this.isCreating = false
			}
		},

		async handleDeleteCard(cardId) {
			this.error = ''
			this.$emit('delete-card', cardId)
		},

		handleEditCard(card) {
			this.selectedCard = card
			this.isEditModalOpen = true
		},

		closeEditModal() {
			this.isEditModalOpen = false
			this.selectedCard = null
		},

		handleCardUpdated(updatedCard) {
			// Find and update the card in the column
			const cardIndex = this.column.cards.findIndex((c) => c.id === updatedCard.id)
			if (cardIndex !== -1) {
				this.column.cards[cardIndex] = updatedCard
			}
			this.closeEditModal()
		},

		// Column drag handlers - only trigger on column itself, not on cards
		handleColumnDragStart(event) {
			this.isColumnDragging = true
			this.setDragData(event, {
				type: 'column',
				payload: {
					columnId: this.column.id,
				},
			})
		},

		handleColumnDragEnd(event) {
			this.isColumnDragging = false
			this.isColumnDragOver = false
		},

		handleColumnDragOver(event) {
			const dragData = this.getDragData(event)
			// Only show drag-over for column drags on column level
			if (dragData && dragData.type === 'column') {
				event.preventDefault()
				if (event.dataTransfer) {
					event.dataTransfer.dropEffect = 'move'
				}
				this.isColumnDragOver = true
			}
		},

		handleColumnDragLeave(event) {
			if (event.target === event.currentTarget) {
				this.isColumnDragOver = false
			}
		},

		handleColumnDrop(event) {
			event.preventDefault()
			event.stopPropagation()
			this.isColumnDragOver = false

			const dragData = this.getDragData(event)
			if (!dragData || dragData.type !== 'column') return

			const { columnId } = dragData.payload
			this.$emit('move-column', {
				columnId,
				targetColumnId: this.column.id,
			})
		},

		// Card drag handlers
		handleCardDragStart(event) {
			event.stopPropagation()
			const cardElement = event.target.closest('[data-card-id]')
			if (cardElement) {
				const cardId = cardElement.getAttribute('data-card-id')
				const card = this.column.cards.find((c) => String(c.id) === String(cardId))
				if (card) {
					this.setDragData(event, {
						type: 'card',
						payload: {
							card,
							fromColumnId: this.column.id,
						},
					})
				}
			}
		},

		handleCardDragEnd(event) {
			event.stopPropagation()
			this.isCardDragOver = false
		},

		handleCardContainerDragOver(event) {
			const dragData = this.getDragData(event)
			// Only show drag-over for card drags on container level
			if (dragData && dragData.type === 'card') {
				event.preventDefault()
				if (event.dataTransfer) {
					event.dataTransfer.dropEffect = 'move'
				}
				this.isCardDragOver = true
			}
		},

		handleCardContainerDragLeave(event) {
			if (event.target === event.currentTarget) {
				this.isCardDragOver = false
			}
		},

		handleCardContainerDrop(event) {
			event.preventDefault()
			event.stopPropagation()
			this.isCardDragOver = false

			const dragData = this.getDragData(event)
			if (!dragData || dragData.type !== 'card') return

			const { card, fromColumnId } = dragData.payload
			this.$emit('move-card', {
				card,
				fromColumnId,
				toColumnId: this.column.id,
			})
		},
	},
}
</script>

<style scoped>
/* All styles are defined in global styles.css */
</style>
