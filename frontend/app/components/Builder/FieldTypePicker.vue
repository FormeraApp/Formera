<script lang="ts" setup>
const { t } = useI18n();
const emit = defineEmits<{ select: [fieldType: FieldType] }>();

const categories = computed(() => [
	{ key: "input", label: t("builder.categories.input"), fields: FIELD_CATEGORIES.input },
	{ key: "choice", label: t("builder.categories.choice"), fields: FIELD_CATEGORIES.choice },
	{ key: "special", label: t("builder.categories.special"), fields: FIELD_CATEGORIES.special },
	{ key: "layout", label: t("builder.categories.layout"), fields: FIELD_CATEGORIES.layout },
]);

const expandedCategories = ref<string[]>(["input", "choice"]);

const toggleCategory = (key: string) => {
	const index = expandedCategories.value.indexOf(key);
	if (index > -1) {
		expandedCategories.value.splice(index, 1);
	} else {
		expandedCategories.value.push(key);
	}
};

const selectField = (fieldType: FieldType) => {
	emit("select", fieldType);
};

</script>

<template>
	<div class="field-picker">
		<div v-for="category in categories" :key="category.key" class="category">
			<button
				class="category-header"
				:aria-expanded="expandedCategories.includes(category.key)"
				@click="toggleCategory(category.key)"
			>
				<span>{{ category.label }}</span>
				<UISysIcon :icon="expandedCategories.includes(category.key) ? 'fa-solid fa-chevron-down' : 'fa-solid fa-chevron-right'" />
			</button>
			<div v-show="expandedCategories.includes(category.key)" class="category-fields">
				<button
					v-for="fieldType in category.fields"
					:key="fieldType"
					class="field-type-btn"
					:title="$t(`fields.${fieldType}.description`)"
					@click="selectField(fieldType)"
				>
					<UISysIcon :icon="FIELD_META[fieldType].icon" />
					<span>{{ $t(`fields.${fieldType}.label`) }}</span>
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.field-picker {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.category-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	width: 100%;
	padding: 0.5rem 0.75rem;
	font-size: 0.75rem;
	font-weight: 600;
	color: var(--text-secondary);
	text-transform: uppercase;
	letter-spacing: 0.05em;
	cursor: pointer;
	background: none;
	border: none;
	border-radius: var(--radius);
	transition: all 0.2s;
}

.category-header:hover {
	color: var(--text);
	background: var(--background);
}

.category-header i {
	font-size: 0.625rem;
}

.category-fields {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
	padding: 0.25rem 0;
}

.field-type-btn {
	display: flex;
	gap: 0.625rem;
	align-items: center;
	padding: 0.5rem 0.75rem;
	font-size: 0.8125rem;
	color: var(--text);
	text-align: left;
	cursor: pointer;
	background: none;
	border: 1px solid transparent;
	border-radius: var(--radius);
	transition: all 0.2s;
}

.field-type-btn:hover {
	background: var(--background);
	border-color: var(--border);
}

.field-type-btn i {
	width: 1rem;
	color: var(--text-secondary);
	text-align: center;
}
</style>
