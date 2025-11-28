<template>
	<div class="board-view-page">
		<div class="board-view-header">
			<div class="board-view-header-content">
				<router-link to="/" class="board-back-link">&larr; Back to boards</router-link>
				<h1 v-if="board">{{ board.name }}</h1>
			</div>
		</div>

		<div class="board-view-container">
			<div v-if="loading" class="board-loading">Loading board...</div>
			<div v-if="error" class="board-error">Error: {{ error }}</div>

			<div v-if="!loading && board" class="board-content">
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
					/>
				</div>
				<div v-else class="no-columns">No columns yet. Create one!</div>
			</div>
		</div>
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
				// Handle 404 error - board not found (user doesn't own it)
				if (err.response?.status === 404) {
					this.error = 'Board not found. You may not have access to this board.'
					setTimeout(() => {
						this.$router.push('/boards')
					}, 2000)
				} else if (err.response?.status === 401) {
					this.error = 'Unauthorized. Please log in again.'
					setTimeout(() => {
						this.$router.push('/login')
					}, 2000)
				} else {
					this.error = err.message || 'Failed to load board'
				}
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
/* All styles are defined in global styles.css */
</style>
