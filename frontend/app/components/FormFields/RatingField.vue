<script lang="ts" setup>
const props = withDefaults(
	defineProps<{
		modelValue?: number;
		minValue?: number;
		maxValue?: number;
	}>(),
	{
		modelValue: 0,
		minValue: 1,
		maxValue: 5,
	}
);

const emit = defineEmits<{
	"update:modelValue": [value: number];
	change: [];
}>();

const stars = computed(() => {
	return Array.from({ length: props.maxValue - props.minValue + 1 }, (_, i) => i + props.minValue);
});

const handleClick = (value: number) => {
	emit("update:modelValue", value);
	emit("change");
};
</script>

<template>
	<div class="rating-group">
		<button
			v-for="star in stars"
			:key="star"
			:class="['rating-star', { 'rating-star-active': star <= modelValue }]"
			type="button"
			@click="handleClick(star)"
		>
			<i
				:class="star <= modelValue ? 'fa-solid fa-star' : 'fa-regular fa-star'"
				style="font-size: 24px"
			></i>
		</button>
	</div>
</template>

<style scoped>
.rating-group {
	display: flex;
	gap: 0.25rem;
}

.rating-star {
	padding: 0.25rem;
	color: var(--border);
	cursor: pointer;
	background: none;
	border: none;
	transition:
		color 0.2s,
		transform 0.2s;
}

.rating-star:hover {
	transform: scale(1.1);
}

.rating-star-active {
	color: #fbbf24;
}
</style>
