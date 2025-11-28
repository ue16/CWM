<template>
	<div class="column">
		<div class="column-header">
			<span class="column-name">{{ column.name }}</span>
			<button class="delete-button" @click="$emit('delete-column', column.id)">Delete</button>
		</div>
		<div v-if="column.cards && column.cards.length">
			<Card
				v-for="card in column.cards"
				:key="card.id"
				:card="card"
				@delete-card="$emit('delete-card', card.id)"
			/>
		</div>
		<div v-else>No cards</div>
		<form @submit.prevent="addCard">
			<input v-model="newCardTitle" placeholder="Card title" required />
			<textarea v-model="newCardDescription" placeholder="Description"></textarea>
			<button type="submit">Add card</button>
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
		}
	},
	methods: {
		async addCard() {
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
		},
	},
}
</script>

<style scoped>
.column {
	border: 1px solid #ddd;
	border-radius: 6px;
	padding: 12px;
	margin: 8px;
	background: #f9f9f9;
}
.column-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 12px;
}
.column-name {
	font-size: 1.25rem;
	font-weight: bold;
	color: #333;
}
</style>
