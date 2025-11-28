<template>
	<div class="auth-container">
		<div class="auth-card">
			<h1>{{ isLogin ? 'Login' : 'Register' }}</h1>

			<form @submit.prevent="handleSubmit">
				<div class="form-group">
					<label for="username">Username</label>
					<input
						id="username"
						v-model="form.username"
						type="text"
						placeholder="Enter your username"
						required
					/>
				</div>

				<div class="form-group" v-if="!isLogin">
					<label for="email">Email</label>
					<input
						id="email"
						v-model="form.email"
						type="email"
						placeholder="Enter your email"
						required
					/>
				</div>

				<div class="form-group">
					<label for="password">Password</label>
					<input
						id="password"
						v-model="form.password"
						type="password"
						placeholder="Enter your password"
						required
					/>
				</div>

				<button type="submit" class="submit-btn" :disabled="isLoading">
					{{ isLoading ? 'Loading...' : isLogin ? 'Login' : 'Register' }}
				</button>
			</form>

			<p v-if="error" class="error-message">{{ error }}</p>
			<p v-if="success" class="success-message">{{ success }}</p>

			<p class="toggle-mode">
				{{ isLogin ? "Don't have an account? " : 'Already have an account? ' }}
				<button type="button" @click="toggleMode" class="link-btn">
					{{ isLogin ? 'Register' : 'Login' }}
				</button>
			</p>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../api/auth'

const router = useRouter()
const isLogin = ref(true)
const isLoading = ref(false)
const error = ref('')
const success = ref('')

const form = ref({
	username: '',
	email: '',
	password: '',
})

const toggleMode = () => {
	isLogin.value = !isLogin.value
	error.value = ''
	success.value = ''
	form.value = {
		username: '',
		email: '',
		password: '',
	}
}

const handleSubmit = async () => {
	error.value = ''
	success.value = ''
	isLoading.value = true

	try {
		if (isLogin.value) {
			// Login
			const response = await authAPI.login(form.value.username, form.value.password)
			authAPI.setToken(response.token)
			authAPI.setUser(response.user)
			success.value = 'Login successful!'
			setTimeout(() => {
				router.push(`/boards`)
			}, 500)
		} else {
			// Register
			const response = await authAPI.register(
				form.value.username,
				form.value.email,
				form.value.password,
			)
			if (response.success && response.token && response.user) {
				authAPI.setToken(response.token)
				authAPI.setUser(response.user)
				success.value = 'Registration successful!'
				const userId = response.user?.id
				setTimeout(() => {
					if (userId) {
						router.push(`/boards`)
					}
				}, 500)
			} else {
				error.value = response.message || 'Registration failed'
			}
		}
	} catch (err: any) {
		error.value =
			err.response?.data?.error ||
			err.message ||
			(isLogin.value ? 'Login failed' : 'Registration failed')
	} finally {
		isLoading.value = false
	}
}
</script>

<style scoped></style>
