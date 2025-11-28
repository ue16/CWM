<template>
	<div class="column">
		<div class="column-header">
			<span class="column-name">{{ column.name }}</span>
			<button class="delete-button" @click="$emit('delete-column', column.id)">×</button>
		</div>
		<div class="cards-container">
			<div v-if="column.cards && column.cards.length">
				<Card
					v-for="card in column.cards"
					:key="card.id"
					:card="card"
					@delete-card="handleDeleteCard"
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
	</div>
</template>

<script>
import Card from './Card.vue'
import { createCard } from '../api/cards'

export default {
	name: 'Column',
	components: { Card },
	props: {
		column: {
			type: Object,
			required: true,
		},
		allColumns: Array,
	},
	data() {
		return {
			newCardTitle: '',
			newCardDescription: '',
			isCreating: false,
			error: '',
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
	},
}
</script>

<style scoped>
/* All styles are defined in global styles.css */
</style>
