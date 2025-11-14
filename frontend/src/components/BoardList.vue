<template>
	<div>
		<h1>Kanban Boards</h1>
		<form @submit.prevent="createBoard">
			<input v-model="newBoardName" placeholder="New board name" required />
			<button type="submit">Create board</button>
		</form>
		<div v-if="boards.length" class="board-container">
			<div v-for="board in boards" :key="board.id" class="board-card">
				<button @click.stop="deleteBoard(board.id)" class="delete-button">Delete</button>
				<router-link :to="`/board/${board.id}`" class="board-link">
					<span class="board-name">{{ board.name }}</span>
				</router-link>
			</div>
		</div>
		<p v-else>No boards available. Add a new board to get started!</p>
	</div>
</template>

<script>
export default {
	name: 'BoardList',
	data() {
		return {
			boards: [
				{ id: 1, name: 'Project Alpha' },
				{ id: 2, name: 'Project Beta' },
				{ id: 3, name: 'Project Gamma' },
			], // Replace with API data later
			newBoardName: '',
		}
	},
	methods: {
		createBoard() {
			// Placeholder for board creation logic
			console.log('Creating board:', this.newBoardName)
		},
		deleteBoard(id) {
			// Placeholder for board deletion logic
			console.log('Deleting board:', id)
		},
	},
}
</script>

<style scoped>
.board-container {
	display: flex;
	flex-wrap: wrap;
	gap: 16px;
	padding: 16px;
}

.board-card {
	position: relative;
	flex: 1 1 calc(33.333% - 16px);
	max-width: calc(33.333% - 16px);
	background: #ffffff;
	border: 1px solid #ddd;
	border-radius: 8px;
	padding: 16px;
	box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	transition:
		transform 0.2s,
		box-shadow 0.2s;
}

.board-card:hover {
	transform: translateY(-4px);
	box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.board-link {
	text-decoration: none;
	color: inherit;
	display: block;
	height: 100%;
}

.board-name {
	font-size: 1.25rem;
	font-weight: bold;
	color: #333;
}
</style>
