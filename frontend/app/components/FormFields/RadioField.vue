<script lang="ts" setup>
withDefaults(defineProps<{
	modelValue?: string;
	name: string;
	options?: string[];
	required?: boolean;
}>(), {
	modelValue: "",
	options: () => [],
	required: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string];
	change: [];
}>();

const handleChange = (option: string) => {
	emit("update:modelValue", option);
	emit("change");
};
</script>

<template>
	<div class="radio-group">
		<label v-for="(option, i) in options" :key="i" class="radio-label">
			<input
				:checked="modelValue === option"
				:name="name"
				:required="required"
				type="radio"
				:value="option"
				@change="handleChange(option)"
			/>
			<span>{{ option }}</span>
		</label>
	</div>
</template>

<style scoped>
.radio-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.radio-label {
	display: flex;
	gap: 0.625rem;
	align-items: center;
	padding: 0.5rem;
	cursor: pointer;
	border-radius: var(--radius);
	transition: background-color 0.2s;
}

.radio-label:hover {
	background-color: var(--background);
}

.radio-label input {
	width: 1rem;
	height: 1rem;
	cursor: pointer;
}
</style>
