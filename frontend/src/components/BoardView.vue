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
						@move-card="moveCard"
						@move-column="moveColumn"
					/>
				</div>
				<div v-else class="no-columns">No columns yet. Create one!</div>
			</div>
		</div>
	</div>
</template>

<script>
import Column from './Column.vue'
import { fetchBoard } from '@/api/boards'
import { fetchColumns, createColumn, deleteColumn } from '@/api/columns'
import { fetchCards, createCard, deleteCard, moveCard as moveCardAPI } from '@/api/cards'
import { moveColumn as moveColumnAPI } from '@/api/columns'

export default {
	name: 'BoardView',
	components: { Column },
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
		async moveCard(moveData) {
			try {
				const { card, fromColumnId, toColumnId } = moveData

				// If moving within the same column, do nothing
				if (fromColumnId === toColumnId) {
					return
				}

				// Remove card from source column
				const fromColumn = this.columns.find((c) => c.id === fromColumnId)
				if (fromColumn && fromColumn.cards) {
					fromColumn.cards = fromColumn.cards.filter((c) => c.id !== card.id)
				}

				// Add card to target column
				const toColumn = this.columns.find((c) => c.id === toColumnId)
				if (toColumn) {
					if (!toColumn.cards) {
						toColumn.cards = []
					}
					toColumn.cards.push(card)
				}

				// Call API to move the card
				await moveCardAPI(card, toColumnId)
			} catch (err) {
				console.error('Error moving card:', err)
				this.error = 'Failed to move card'
				// Reload board data to ensure consistency
				setTimeout(() => {
					this.loadBoardData()
				}, 1000)
			}
		},
		async moveColumn(moveData) {
			try {
				const { columnId, targetColumnId } = moveData

				// Find the column being moved and the target column
				const fromIndex = this.columns.findIndex((c) => c.id === columnId)
				const toIndex = this.columns.findIndex((c) => c.id === targetColumnId)

				if (fromIndex === -1 || toIndex === -1) return

				// If moving to the same position, do nothing
				if (fromIndex === toIndex) {
					return
				}

				// Move the column in the array
				const [movedColumn] = this.columns.splice(fromIndex, 1)
				this.columns.splice(toIndex, 0, movedColumn)

				// Update positions for all columns
				const positionUpdates = this.columns.map((col, index) => ({
					id: col.id,
					name: col.name,
					board_id: col.board_id,
					position: index + 1,
				}))

				// Call API to update positions for all affected columns
				const movedCol = positionUpdates.find((c) => c.id === columnId)
				if (movedCol) {
					await moveColumnAPI(movedCol, movedCol.position)
				}

				// Update local columns with new positions
				this.columns = this.columns.map((col) => {
					const updated = positionUpdates.find((c) => c.id === col.id)
					return { ...col, position: updated?.position || col.position }
				})
			} catch (err) {
				console.error('Error moving column:', err)
				this.error = 'Failed to move column'
				// Reload board data to ensure consistency
				setTimeout(() => {
					this.loadBoardData()
				}, 1000)
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
