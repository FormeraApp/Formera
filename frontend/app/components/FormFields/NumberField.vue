<script lang="ts" setup>
withDefaults(
	defineProps<{
		id?: string;
		modelValue?: string | number;
		placeholder?: string;
		required?: boolean;
		min?: number;
		max?: number;
	}>(),
	{
		id: "",
		modelValue: "",
		placeholder: "",
		required: false,
		min: undefined,
		max: undefined,
	}
);

const emit = defineEmits<{
	"update:modelValue": [value: string];
	blur: [];
}>();

const handleInput = (event: Event) => {
	const target = event.target as HTMLInputElement;
	emit("update:modelValue", target.value);
};

const handleBlur = () => {
	emit("blur");
};
</script>

<template>
	<input
		:id="id"
		:value="modelValue"
		class="input"
		:placeholder="placeholder"
		:required="required"
		type="number"
		:min="min"
		:max="max"
		@input="handleInput"
		@blur="handleBlur"
	/>
</template>
