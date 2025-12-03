<script lang="ts" setup>
const props = withDefaults(defineProps<{
	modelValue?: number;
	minValue?: number;
	maxValue?: number;
	minLabel?: string;
	maxLabel?: string;
}>(), {
	modelValue: 0,
	minValue: 1,
	maxValue: 10,
	minLabel: "",
	maxLabel: "",
});

const emit = defineEmits<{
	"update:modelValue": [value: number];
	change: [];
}>();

const numbers = computed(() => {
	return Array.from({ length: props.maxValue - props.minValue + 1 }, (_, i) => i + props.minValue);
});

const handleClick = (value: number) => {
	emit("update:modelValue", value);
	emit("change");
};
</script>

<template>
	<div class="scale-group">
		<span v-if="minLabel" class="scale-label">{{ minLabel }}</span>
		<div class="scale-buttons">
			<button
				v-for="num in numbers"
				:key="num"
				:class="['scale-btn', { 'scale-btn-active': modelValue === num }]"
				type="button"
				@click="handleClick(num)"
			>
				{{ num }}
			</button>
		</div>
		<span v-if="maxLabel" class="scale-label">{{ maxLabel }}</span>
	</div>
</template>

<style scoped>
.scale-group {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
	align-items: center;
}

.scale-label {
	font-size: 0.8125rem;
	color: var(--text-secondary);
}

.scale-buttons {
	display: flex;
	flex-wrap: wrap;
	gap: 0.375rem;
}

.scale-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2.5rem;
	height: 2.5rem;
	font-weight: 500;
	cursor: pointer;
	background: var(--background);
	border: 1px solid var(--border);
	border-radius: var(--radius);
	transition: all 0.2s;
}

.scale-btn:hover {
	border-color: var(--primary);
}

.scale-btn-active {
	color: white;
	background: var(--primary);
	border-color: var(--primary);
}
</style>
