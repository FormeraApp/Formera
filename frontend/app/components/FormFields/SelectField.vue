<script lang="ts" setup>
const props = withDefaults(defineProps<{
	id?: string;
	modelValue?: string;
	options?: string[];
	placeholder?: string;
	required?: boolean;
}>(), {
	id: "",
	modelValue: "",
	options: () => [],
	placeholder: "",
	required: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string];
	blur: [];
}>();

const { t } = useI18n();

const handleChange = (event: Event) => {
	const target = event.target as HTMLSelectElement;
	emit("update:modelValue", target.value);
};

const handleBlur = () => {
	emit("blur");
};

const placeholderText = computed(() => props.placeholder || t("select.placeholder"));
</script>

<template>
	<select
		:id="id"
		:value="modelValue"
		class="input"
		:required="required"
		@change="handleChange"
		@blur="handleBlur"
	>
		<option value="">{{ placeholderText }}</option>
		<option v-for="(option, i) in options" :key="i" :value="option">
			{{ option }}
		</option>
	</select>
</template>
