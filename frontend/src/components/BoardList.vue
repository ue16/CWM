<template>
	<div class="board-list-page">
		<header class="header">
			<div class="header-content">
				<h1>Kanban Boards</h1>
				<div class="user-menu">
					<span class="user-name">{{ currentUser?.username }}</span>
					<button @click="handleLogout" class="logout-btn">Logout</button>
				</div>
			</div>
		</header>

		<div class="container">
			<form @submit.prevent="handleCreateBoard" class="create-board-form">
				<input v-model="newBoardName" placeholder="New board name" required />
				<button type="submit" :disabled="isCreating">
					{{ isCreating ? 'Creating...' : 'Create Board' }}
				</button>
			</form>

			<div v-if="error" class="error-message">{{ error }}</div>

			<div v-if="boards.length" class="board-container">
				<div v-for="board in boards" :key="board.id" class="board-card">
					<button @click.stop="handleDeleteBoard(board.id)" class="delete-button">
						×
					</button>
					<router-link :to="`/board/${board.id}`" class="board-link">
						<span class="board-name">{{ board.name }}</span>
					</router-link>
				</div>
			</div>
			<p v-else class="no-boards">No boards available. Add a new board to get started!</p>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
	fetchBoards,
	createBoard as createBoardAPI,
	deleteBoard as deleteBoardAPI,
} from '../api/boards'
import { authAPI, type User } from '../api/auth'

const router = useRouter()

const boards = ref<any[]>([])
const newBoardName = ref('')
const isCreating = ref(false)
const error = ref('')
const currentUser = ref<User | null>(null)

const loadBoards = async () => {
	try {
		error.value = ''
		boards.value = await fetchBoards()
	} catch (err: any) {
		console.error('Error loading boards:', err)
		error.value = 'Failed to load boards'
	}
}

const handleCreateBoard = async () => {
	if (!newBoardName.value.trim()) return

	isCreating.value = true
	try {
		const newBoard = await createBoardAPI(newBoardName.value)
		boards.value.push(newBoard)
		newBoardName.value = ''
		error.value = ''
	} catch (err: any) {
		console.error('Error creating board:', err)
		error.value = 'Failed to create board'
	} finally {
		isCreating.value = false
	}
}

const handleDeleteBoard = async (id: number) => {
	if (!confirm('Are you sure you want to delete this board?')) return

	try {
		await deleteBoardAPI(id)
		boards.value = boards.value.filter((board) => board.id !== id)
		error.value = ''
	} catch (err: any) {
		console.error('Error deleting board:', err)
		error.value = 'Failed to delete board'
	}
}

const handleLogout = () => {
	authAPI.logout()
	router.push('/login')
}

onMounted(() => {
	currentUser.value = authAPI.getUser()
	loadBoards()
})
</script>

<style scoped></style>
