<template>
	<div>
		<router-link to="/">&larr; Back to boards</router-link>
		<h1>{{ board.name }}</h1>
		<form class="add-column-form" @submit.prevent="addColumn">
			<input v-model="newColumnName" placeholder="New column name" required />
			<button type="submit">Add column</button>
		</form>
		<VueDraggable
			v-model="columns"
			group="columns"
			item-key="id"
			class="board"
			:animation="200"
			:tag="'div'"
		>
			<template v-for="column in columns">
				<Column
					:column="column"
					:all-columns="columns"
					@delete-column="deleteColumn"
					@add-card="addCard"
					@delete-card="deleteCard"
					style="flex: 1; margin: 0 8px"
				/>
			</template>
		</VueDraggable>
	</div>
</template>

<script>
import Column from './Column.vue'
import { VueDraggable } from 'vue-draggable-plus'

export default {
	name: 'BoardView',
	components: { Column, VueDraggable },
	data() {
		return {
			board: { id: 3, name: 'Project Gamma' }, // Replace with API data later
			columns: [
				{
					id: 1,
					name: 'To Do',
					cards: [
						{
							id: 1,
							title: 'Identify key stakeholders',
							description: 'Some more elaborate explanation',
						},
						{
							id: 2,
							title: 'Draft initial project plan',
							description: 'Some more elaborate explanation',
						},
					],
				},
				{
					id: 2,
					name: 'In Progress',
					cards: [
						{
							id: 3,
							title: 'Develop core features',
							description: 'Some more elaborate explanation',
						},
					],
				},
				{
					id: 3,
					name: 'Done',
					cards: [
						{ id: 4, title: 'Set up testing framework' },
						{ id: 5, title: 'Conduct initial tests' },
					],
				},
			], // Replace with API data later
			newColumnName: '',
		}
	},
	methods: {
		addColumn() {
			// Placeholder for column creation logic
			console.log('Adding column:', this.newColumnName)
		},
		deleteColumn(id) {
			// Placeholder for column deletion logic
			console.log('Deleting column:', id)
		},
		addCard(columnId, card) {
			// Placeholder for card creation logic
			console.log('Adding card:', card)
		},
		deleteCard(cardId) {
			// Placeholder for card deletion logic
			console.log('Deleting card:', cardId)
		},
	},
	mounted() {
		console.log('BoardView mounted. Columns data:', this.columns)
	},
}
</script>

<style scoped>
.board {
	display: flex;
	flex-direction: row;
	gap: 16px;
	padding: 16px;
	overflow-x: auto;
}
.add-column-form {
	flex-direction: row;
}
</style>
