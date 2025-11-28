<template>
	<div
		class="card"
		:data-card-id="card.id"
		draggable="true"
		@dragstart="handleDragStart"
		@dragend="handleDragEnd"
	>
		<div class="card-actions">
			<button class="card-edit" @click="$emit('edit-card', card)" title="Edit card">
				Edit
			</button>
			<button class="card-delete" @click="$emit('delete-card', card.id)" title="Delete card">
				Delete
			</button>
		</div>
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
	methods: {
		handleDragStart(event) {
			event.dataTransfer.effectAllowed = 'move'
			this.$emit('dragstart', event)
		},
		handleDragEnd(event) {
			this.$emit('dragend', event)
		},
	},
}
</script>

<style scoped>
/* All styles are defined in global styles.css */
</style>
