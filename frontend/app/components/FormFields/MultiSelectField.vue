<script lang="ts" setup>
withDefaults(defineProps<{
	id?: string;
	modelValue?: string[];
	options?: string[];
	required?: boolean;
}>(), {
	id: "",
	modelValue: () => [],
	options: () => [],
	required: false,
});

const emit = defineEmits<{
	"update:modelValue": [value: string[]];
	blur: [];
}>();

const handleChange = (event: Event) => {
	const target = event.target as HTMLSelectElement;
	const selected = Array.from(target.selectedOptions).map((opt) => opt.value);
	emit("update:modelValue", selected);
};

const handleBlur = () => {
	emit("blur");
};
</script>

<template>
	<select
		:id="id"
		class="input input-multiselect"
		:required="required"
		multiple
		@change="handleChange"
		@blur="handleBlur"
	>
		<option
			v-for="(option, i) in options"
			:key="i"
			:value="option"
			:selected="modelValue.includes(option)"
		>
			{{ option }}
		</option>
	</select>
</template>

<style scoped>
.input-multiselect {
	min-height: 100px;
}
</style>
