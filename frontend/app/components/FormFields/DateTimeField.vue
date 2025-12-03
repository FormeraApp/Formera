<script lang="ts" setup>
withDefaults(
	defineProps<{
		id?: string;
		modelValue?: string;
		type?: "date" | "time";
		required?: boolean;
	}>(),
	{
		id: "",
		modelValue: "",
		type: "date",
		required: false,
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
		:required="required"
		:type="type"
		@input="handleInput"
		@blur="handleBlur"
	/>
</template>
