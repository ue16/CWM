<template>
	<div class="column">
		<div class="column-header">
			<span class="column-name">{{ column.name }}</span>
			<button class="delete-button" @click="$emit('delete-column', column.id)">Delete</button>
		</div>
		<VueDraggable v-model="column.cards" group="cards" item-key="id" :animation="200">
			<template v-for="card in column.cards">
				<Card :card="card" @delete-card="$emit('delete-card', card.id)" />
			</template>
		</VueDraggable>
		<div v-if="!column.cards.length">No cards</div>
		<form @submit.prevent="addCard">
			<input v-model="newCardTitle" placeholder="Card title" required />
			<textarea v-model="newCardDescription" placeholder="Description"></textarea>
			<button type="submit">Add card</button>
		</form>
	</div>
</template>

<script>
import Card from './Card.vue'
import { VueDraggable } from 'vue-draggable-plus'

export default {
	name: 'Column',
	components: { Card, VueDraggable },
	props: {
		column: Object,
		allColumns: Array,
	},
	data() {
		return {
			newCardTitle: '',
			newCardDescription: '',
		}
	},
	methods: {
		addCard() {
			const card = {
				title: this.newCardTitle,
				description: this.newCardDescription,
			}
			this.$emit('add-card', this.column.id, card)
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
