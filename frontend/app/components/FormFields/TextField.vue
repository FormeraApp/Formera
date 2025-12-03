<script lang="ts" setup>
const props = withDefaults(defineProps<{
	id?: string;
	modelValue?: string;
	type?: "text" | "email" | "phone" | "url";
	placeholder?: string;
	required?: boolean;
}>(), {
	id: "",
	modelValue: "",
	type: "text",
	placeholder: "",
	required: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string];
	blur: [];
}>();

const inputType = computed(() => {
	switch (props.type) {
		case "email":
			return "email";
		case "phone":
			return "tel";
		case "url":
			return "url";
		default:
			return "text";
	}
});

const defaultPlaceholder = computed(() => {
	if (props.placeholder) return props.placeholder;
	if (props.type === "url") return "https://";
	return "";
});

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
		:placeholder="defaultPlaceholder"
		:required="required"
		:type="inputType"
		@input="handleInput"
		@blur="handleBlur"
	/>
</template>
