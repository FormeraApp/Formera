<script lang="ts" setup>
withDefaults(defineProps<{
	id?: string;
	modelValue?: string;
	placeholder?: string;
	required?: boolean;
	rows?: number;
}>(), {
	id: "",
	modelValue: "",
	placeholder: "",
	required: false,
	rows: 4,
});

const emit = defineEmits<{
	"update:modelValue": [value: string];
	blur: [];
}>();

const handleInput = (event: Event) => {
	const target = event.target as HTMLTextAreaElement;
	emit("update:modelValue", target.value);
};

const handleBlur = () => {
	emit("blur");
};
</script>

<template>
	<textarea
		:id="id"
		:value="modelValue"
		class="input"
		:placeholder="placeholder"
		:required="required"
		:rows="rows"
		@input="handleInput"
		@blur="handleBlur"
	/>
</template>
