<template>
	<div class="card">
		<button class="card-delete" @click="$emit('delete-card', card.id)">Delete</button>
		<div class="card-title">{{ card.title }}</div>
		<div v-if="card.description" class="card-description">{{ card.description }}</div>
	</div>
</template>

<script>
export default {
	name: 'Card',
	props: {
		card: Object,
	},
}
</script>

<style scoped>
/* All styles are defined in global styles.css */
</style>
