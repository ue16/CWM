import { createApp } from 'vue'
import App from './App.vue'
import { createRouter, createWebHistory } from 'vue-router'
import './assets/styles.css'

// Import components for routing
import BoardList from './components/BoardList.vue'
import BoardView from './components/BoardView.vue'

// Define routes
const routes = [
	{ path: '/', component: BoardList },
	{ path: '/board/:id', component: BoardView, props: true },
]

// Create router instance
const router = createRouter({
	history: createWebHistory(),
	routes,
})

// Create and mount the app
const app = createApp(App)
app.use(router)
app.mount('#app')
