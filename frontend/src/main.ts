import { createApp } from 'vue'
import App from './App.vue'
import { createRouter, createWebHistory } from 'vue-router'
import './assets/styles.css'

// Import components for routing
import Login from './components/Login.vue'
import BoardList from './components/BoardList.vue'
import BoardView from './components/BoardView.vue'
import { authAPI } from './api/auth'

// Define routes
const routes = [
	{ path: '/login', component: Login },
	{
		path: '/',
		redirect: () => {
			const user = authAPI.getUser()
			return user ? '/boards' : '/login'
		},
	},
	{ path: '/boards', component: BoardList, meta: { requiresAuth: true } },
	{ path: '/board/:id', component: BoardView, props: true, meta: { requiresAuth: true } },
]

// Create router instance
const router = createRouter({
	history: createWebHistory(),
	routes,
})

// Navigation guard for authentication
router.beforeEach((to, from, next) => {
	const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)
	const isAuthenticated = authAPI.isAuthenticated()

	if (requiresAuth && !isAuthenticated) {
		// Redirect to login if trying to access protected route without auth
		next('/login')
	} else if (to.path === '/login' && isAuthenticated) {
		// Redirect to boards if already logged in and trying to access login
		next('/boards')
	} else {
		next()
	}
})

// Create and mount the app
const app = createApp(App)
app.use(router)
app.mount('#app')
