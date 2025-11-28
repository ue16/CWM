<template>
	<div>
		<router-link to="/">&larr; Back to boards</router-link>
		<div v-if="loading" class="loading">Loading board...</div>
		<div v-if="error" class="error">Error: {{ error }}</div>
		<h1 v-if="board">{{ board.name }}</h1>
		<form v-if="board" class="add-column-form" @submit.prevent="addColumn">
			<input v-model="newColumnName" placeholder="New column name" required />
			<button type="submit">Add column</button>
		</form>
		<div v-if="columns && columns.length" class="board">
			<Column
				v-for="column in columns"
				:key="column.id"
				:column="column"
				:all-columns="columns"
				@delete-column="deleteColumn"
				@add-card="addCard"
				@delete-card="deleteCard"
				style="flex: 1; margin: 0 8px"
			/>
		</div>
		<div v-else class="board">No columns yet. Create one!</div>
	</div>
</template>

<script>
import Column from './Column.vue'
import { VueDraggable } from 'vue-draggable-plus'
import { fetchBoard } from '@/api/boards'
import { fetchColumns, createColumn, deleteColumn } from '@/api/columns'
import { fetchCards, createCard, deleteCard } from '@/api/cards'

export default {
	name: 'BoardView',
	components: { Column, VueDraggable },
	data() {
		return {
			board: null,
			columns: [],
			newColumnName: '',
			error: null,
			loading: true,
		}
	},
	methods: {
		async addColumn() {
			const newColumn = await createColumn(this.board.id, this.newColumnName)
			newColumn.cards = []
			this.columns.push(newColumn)
			this.newColumnName = ''
		},
		async deleteColumn(id) {
			await deleteColumn(id)
			this.columns = this.columns.filter((column) => column.id !== id)
		},
		async addCard(columnId, title) {
			const newCard = await createCard(columnId, title)
			// Card will be added through parent-child prop binding
		},
		async deleteCard(cardId) {
			await deleteCard(cardId)
			// Remove card from all columns
			for (const column of this.columns) {
				if (column.cards) {
					column.cards = column.cards.filter((card) => card.id !== cardId)
				}
			}
		},
		async loadBoardData() {
			try {
				const boardId = parseInt(this.$route.params.id)
				this.board = await fetchBoard(boardId)
				this.columns = await fetchColumns(boardId)
				if (!this.columns) {
					this.columns = []
				}
				// Initialize cards array on all columns
				for (const column of this.columns) {
					if (!column.cards) {
						column.cards = []
					}
					column.cards = await fetchCards(column.id)
				}
			} catch (err) {
				console.error('Error loading board:', err)
				this.error = err.message
			} finally {
				this.loading = false
			}
		},
	},
	async mounted() {
		this.loading = true
		await this.loadBoardData()
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
.loading {
	padding: 16px;
	color: #666;
	font-size: 16px;
}
.error {
	padding: 16px;
	color: #d32f2f;
	background-color: #ffebee;
	border-radius: 4px;
	margin: 8px;
}
</style>
