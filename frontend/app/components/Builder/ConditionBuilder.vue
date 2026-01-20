<script lang="ts" setup>
import type { ConditionOperator, ConditionRule, FieldConditions, FormField, LayoutFieldType } from "~~/shared/types";

const props = defineProps<{
	field: FormField;
	availableFields: FormField[];
}>();

const emit = defineEmits<{
	"update:conditions": [conditions: FieldConditions | undefined];
}>();

const { t } = useI18n();

// Local state for conditions
const enabled = ref(!!props.field.conditions);
const showWhen = ref(props.field.conditions?.show ?? true);
const logic = ref<"and" | "or">(props.field.conditions?.logic ?? "and");
const rules = ref<ConditionRule[]>(props.field.conditions?.rules ?? []);

// Filter available fields (exclude self and layout fields)
const selectableFields = computed(() => {
	const layoutTypes: LayoutFieldType[] = ["section", "pagebreak", "divider", "heading", "paragraph", "image"];
	return props.availableFields.filter(
		(f) => f.id !== props.field.id && !layoutTypes.includes(f.type as LayoutFieldType)
	);
});

// Get operators for a specific field type
const getOperatorsForFieldType = (fieldType: string): ConditionOperator[] => {
	const choiceTypes: string[] = ["select", "radio", "checkbox", "dropdown"];
	const numericTypes: string[] = ["number"];
	const dateTypes: string[] = ["date", "time"];

	if (numericTypes.includes(fieldType) || dateTypes.includes(fieldType)) {
		return [
			"equals",
			"not_equals",
			"greater_than",
			"less_than",
			"greater_or_equal",
			"less_or_equal",
			"is_empty",
			"is_not_empty",
		];
	}

	if (choiceTypes.includes(fieldType)) {
		return ["equals", "not_equals", "is_empty", "is_not_empty"];
	}

	// Text fields and others
	return [
		"equals",
		"not_equals",
		"contains",
		"not_contains",
		"starts_with",
		"ends_with",
		"is_empty",
		"is_not_empty",
	];
};

// Get the field object for a rule
const getFieldForRule = (rule: ConditionRule): FormField | undefined => {
	return props.availableFields.find((f) => f.id === rule.fieldId);
};

// Check if operator needs a value input
const operatorNeedsValue = (operator: ConditionOperator): boolean => {
	return !["is_empty", "is_not_empty"].includes(operator);
};

// Check if field is a choice field
const isChoiceField = (fieldType: string): boolean => {
	const choiceTypes: string[] = ["select", "radio", "checkbox", "dropdown"];
	return choiceTypes.includes(fieldType);
};

// Check if field is numeric
const isNumericField = (fieldType: string): boolean => {
	return fieldType === "number";
};

// Emit updated conditions
const emitConditions = () => {
	if (!enabled.value) {
		emit("update:conditions", undefined);
		return;
	}

	const conditions: FieldConditions = {
		show: showWhen.value,
		logic: logic.value,
		rules: rules.value,
	};

	emit("update:conditions", conditions);
};

// Toggle enabled
const toggleEnabled = (value: boolean) => {
	enabled.value = value;
	if (!value) {
		rules.value = [];
	} else if (rules.value.length === 0) {
		addRule();
	}
	emitConditions();
};

// Add a new rule
const addRule = () => {
	const firstField = selectableFields.value[0];
	if (!firstField) return;

	const operators = getOperatorsForFieldType(firstField.type);
	rules.value.push({
		fieldId: firstField.id,
		operator: operators[0] || "equals",
		value: "",
	});
	emitConditions();
};

// Remove a rule
const removeRule = (index: number) => {
	rules.value.splice(index, 1);
	emitConditions();
};

// Update rule field
const updateRuleField = (index: number, fieldId: string) => {
	const field = props.availableFields.find((f) => f.id === fieldId);
	if (!field) return;

	const operators = getOperatorsForFieldType(field.type);
	rules.value[index] = {
		fieldId,
		operator: operators[0] || "equals",
		value: "",
	};
	emitConditions();
};

// Update rule operator
const updateRuleOperator = (index: number, operator: ConditionOperator) => {
	const rule = rules.value[index];
	if (!rule) return;

	rule.operator = operator;
	if (!operatorNeedsValue(operator)) {
		rule.value = "";
	}
	emitConditions();
};

// Update rule value
const updateRuleValue = (index: number, value: unknown) => {
	const rule = rules.value[index];
	if (!rule) return;

	rule.value = value;
	emitConditions();
};

// Watch for prop changes
watch(
	() => props.field.conditions,
	(newConditions) => {
		if (newConditions) {
			enabled.value = true;
			showWhen.value = newConditions.show;
			logic.value = newConditions.logic;
			rules.value = [...newConditions.rules];
		} else {
			enabled.value = false;
			rules.value = [];
		}
	},
	{ deep: true }
);
</script>

<template>
	<div class="condition-builder">
		<!-- Enable Toggle -->
		<div class="enable-section">
			<label class="checkbox-label">
				<input type="checkbox" :checked="enabled" @change="toggleEnabled(($event.target as HTMLInputElement).checked)" />
				<span>{{ t("conditions.enableConditionalLogic") }}</span>
			</label>
			<p class="hint">{{ t("conditions.enableHint") }}</p>
		</div>

		<!-- Condition Configuration (only shown when enabled) -->
		<div v-if="enabled" class="condition-config">
			<div v-if="selectableFields.length === 0" class="no-fields">
				<UISysIcon icon="fa-solid fa-circle-info" />
				<span>{{ t("conditions.noFieldsAvailable") }}</span>
			</div>

			<template v-else>
				<!-- Show/Hide Selector -->
				<div class="form-group">
					<label class="label">{{ t("conditions.action") }}</label>
					<select
						v-model="showWhen"
						class="input"
						@change="emitConditions"
					>
						<option :value="true">{{ t("conditions.show") }}</option>
						<option :value="false">{{ t("conditions.hide") }}</option>
					</select>
				</div>

				<!-- Logic Selector (only show if multiple rules) -->
				<div v-if="rules.length > 1" class="form-group">
					<label class="label">{{ t("conditions.matchLogic") }}</label>
					<select
						v-model="logic"
						class="input"
						@change="emitConditions"
					>
						<option value="and">{{ t("conditions.matchAll") }}</option>
						<option value="or">{{ t("conditions.matchAny") }}</option>
					</select>
				</div>

				<!-- Rules -->
				<div class="rules-section">
					<label class="label">{{ t("conditions.rules") }}</label>

					<div v-for="(rule, index) in rules" :key="index" class="rule">
						<!-- Field Selector -->
						<select
							:value="rule.fieldId"
							class="input rule-field"
							@change="updateRuleField(index, ($event.target as HTMLSelectElement).value)"
						>
							<option v-for="field in selectableFields" :key="field.id" :value="field.id">
								{{ field.label }}
							</option>
						</select>

						<!-- Operator Selector -->
						<select
							:value="rule.operator"
							class="input rule-operator"
							@change="updateRuleOperator(index, ($event.target as HTMLSelectElement).value as ConditionOperator)"
						>
							<option
								v-for="op in getOperatorsForFieldType(getFieldForRule(rule)?.type || 'text')"
								:key="op"
								:value="op"
							>
								{{ t(`conditions.operators.${op}`) }}
							</option>
						</select>

						<!-- Value Input (conditional based on operator and field type) -->
						<template v-if="operatorNeedsValue(rule.operator)">
							<!-- Choice field: dropdown -->
							<select
								v-if="getFieldForRule(rule) && isChoiceField(getFieldForRule(rule)!.type)"
								:value="rule.value as string"
								class="input rule-value"
								@change="updateRuleValue(index, ($event.target as HTMLSelectElement).value)"
							>
								<option value="">{{ t("conditions.selectValue") }}</option>
								<option
									v-for="option in getFieldForRule(rule)?.options || []"
									:key="option"
									:value="option"
								>
									{{ option }}
								</option>
							</select>

							<!-- Number field: number input -->
							<input
								v-else-if="getFieldForRule(rule) && isNumericField(getFieldForRule(rule)!.type)"
								type="number"
								:value="rule.value as number"
								class="input rule-value"
								:placeholder="t('conditions.enterValue')"
								@input="updateRuleValue(index, ($event.target as HTMLInputElement).value)"
							/>

							<!-- Text field: text input -->
							<input
								v-else
								type="text"
								:value="rule.value as string"
								class="input rule-value"
								:placeholder="t('conditions.enterValue')"
								@input="updateRuleValue(index, ($event.target as HTMLInputElement).value)"
							/>
						</template>

						<!-- Remove Rule Button -->
						<button
							type="button"
							class="btn-icon btn-danger"
							:title="t('conditions.removeRule')"
							@click="removeRule(index)"
						>
							<UISysIcon icon="fa-solid fa-trash" />
						</button>
					</div>

					<!-- Add Rule Button -->
					<button type="button" class="btn btn-secondary btn-sm" @click="addRule">
						<UISysIcon icon="fa-solid fa-plus" />
						{{ t("conditions.addRule") }}
					</button>
				</div>
			</template>
		</div>
	</div>
</template>

<style scoped>
.condition-builder {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.enable-section {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.checkbox-label {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	font-weight: 500;
	cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
	cursor: pointer;
}

.hint {
	margin: 0;
	font-size: 0.875rem;
	color: var(--text-secondary);
}

.condition-config {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	padding: 1rem;
	background: var(--surface-secondary);
	border: 1px solid var(--border);
	border-radius: var(--radius);
}

.no-fields {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	padding: 1rem;
	font-size: 0.875rem;
	color: var(--text-secondary);
	text-align: center;
	background: var(--surface);
	border-radius: var(--radius);
}

.form-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.label {
	font-size: 0.875rem;
	font-weight: 500;
	color: var(--text);
}

.rules-section {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.rule {
	display: grid;
	grid-template-columns: 1fr 1fr 1fr auto;
	gap: 0.5rem;
	align-items: start;
}

.rule-field,
.rule-operator,
.rule-value {
	min-width: 0;
}

.btn-icon {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2.5rem;
	height: 2.5rem;
	padding: 0;
	color: var(--text-secondary);
	background: transparent;
	border: 1px solid var(--border);
	border-radius: var(--radius);
	cursor: pointer;
	transition: all 0.2s;
}

.btn-icon:hover {
	color: var(--text);
	background: var(--surface-secondary);
}

.btn-danger:hover {
	color: var(--error);
	border-color: var(--error);
	background: rgba(239, 68, 68, 0.1);
}

.btn-sm {
	padding: 0.5rem 1rem;
	font-size: 0.875rem;
}

@media (width <= 768px) {
	.rule {
		grid-template-columns: 1fr;
	}

	.btn-icon {
		width: 100%;
	}
}
</style>
