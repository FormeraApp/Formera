<script lang="ts" setup>
const props = withDefaults(defineProps<{
	modelValue?: string[];
	options?: string[];
	required?: boolean;
}>(), {
	modelValue: () => [],
	options: () => [],
	required: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string[]];
	change: [];
}>();

const handleChange = (option: string, checked: boolean) => {
	const current = props.modelValue || [];
	const updated = checked ? [...current, option] : current.filter((v) => v !== option);
	emit("update:modelValue", updated);
	emit("change");
};
</script>

<template>
	<div class="checkbox-group">
		<label v-for="(option, i) in options" :key="i" class="checkbox-label">
			<input
				:checked="modelValue.includes(option)"
				type="checkbox"
				:value="option"
				@change="handleChange(option, ($event.target as HTMLInputElement).checked)"
			/>
			<span>{{ option }}</span>
		</label>
	</div>
</template>

<style scoped>
.checkbox-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.checkbox-label {
	display: flex;
	gap: 0.625rem;
	align-items: center;
	padding: 0.5rem;
	cursor: pointer;
	border-radius: var(--radius);
	transition: background-color 0.2s;
}

.checkbox-label:hover {
	background-color: var(--background);
}

.checkbox-label input {
	width: 1rem;
	height: 1rem;
	cursor: pointer;
}
</style>
