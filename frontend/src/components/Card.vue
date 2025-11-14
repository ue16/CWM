<template>
	<div class="card">
		<button class="delete-button" @click="$emit('delete-card', card.id)">Delete</button>
		<div class="title">{{ card.title }}</div>
		<div class="desc" v-if="card.description">{{ card.description }}</div>
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
.card {
	border: 1px solid #ddd;
	border-radius: 6px;
	padding: 12px;
	margin-bottom: 8px;
	background: #fff;
	box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.title {
	font-weight: bold;
	margin-bottom: 4px;
}
.desc {
	color: #555;
	font-size: 0.9em;
	margin-bottom: 8px;
}
</style>
