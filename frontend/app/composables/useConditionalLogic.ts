import type { ConditionRule, FieldConditions, FormField } from "~~/shared/types";

export function useConditionalLogic() {
	/**
	 * Evaluates a single condition rule
	 */
	const evaluateRule = (rule: ConditionRule, fieldValue: unknown): boolean => {
		const { operator, value: ruleValue } = rule;

		// Handle empty/not empty checks first
		if (operator === "is_empty") {
			return (
				!fieldValue ||
				(Array.isArray(fieldValue) && fieldValue.length === 0) ||
				(typeof fieldValue === "string" && fieldValue.trim() === "")
			);
		}

		if (operator === "is_not_empty") {
			return (
				!!fieldValue &&
				(!Array.isArray(fieldValue) || fieldValue.length > 0) &&
				(typeof fieldValue !== "string" || fieldValue.trim() !== "")
			);
		}

		// Convert values for comparison
		const normalizedFieldValue = normalizeValue(fieldValue);
		const normalizedRuleValue = normalizeValue(ruleValue);

		switch (operator) {
			case "equals":
				// For arrays (checkbox, dropdown), check if array contains the value
				if (Array.isArray(normalizedFieldValue)) {
					return normalizedFieldValue.includes(normalizedRuleValue);
				}
				return normalizedFieldValue === normalizedRuleValue;

			case "not_equals":
				if (Array.isArray(normalizedFieldValue)) {
					return !normalizedFieldValue.includes(normalizedRuleValue);
				}
				return normalizedFieldValue !== normalizedRuleValue;

			case "contains":
				if (Array.isArray(normalizedFieldValue)) {
					return normalizedFieldValue.some((v) =>
						String(v).toLowerCase().includes(String(normalizedRuleValue).toLowerCase())
					);
				}
				return String(normalizedFieldValue || "")
					.toLowerCase()
					.includes(String(normalizedRuleValue).toLowerCase());

			case "not_contains":
				if (Array.isArray(normalizedFieldValue)) {
					return !normalizedFieldValue.some((v) =>
						String(v).toLowerCase().includes(String(normalizedRuleValue).toLowerCase())
					);
				}
				return !String(normalizedFieldValue || "")
					.toLowerCase()
					.includes(String(normalizedRuleValue).toLowerCase());

			case "starts_with":
				return String(normalizedFieldValue || "")
					.toLowerCase()
					.startsWith(String(normalizedRuleValue).toLowerCase());

			case "ends_with":
				return String(normalizedFieldValue || "")
					.toLowerCase()
					.endsWith(String(normalizedRuleValue).toLowerCase());

			case "greater_than":
				return Number(normalizedFieldValue) > Number(normalizedRuleValue);

			case "less_than":
				return Number(normalizedFieldValue) < Number(normalizedRuleValue);

			case "greater_or_equal":
				return Number(normalizedFieldValue) >= Number(normalizedRuleValue);

			case "less_or_equal":
				return Number(normalizedFieldValue) <= Number(normalizedRuleValue);

			default:
				console.warn(`Unknown operator: ${operator}`);
				return false;
		}
	};

	/**
	 * Normalize value for comparison
	 */
	const normalizeValue = (value: unknown): unknown => {
		if (value === null || value === undefined) return "";
		if (typeof value === "string") return value.trim();
		return value;
	};

	/**
	 * Evaluates all conditions for a field
	 */
	const evaluateConditions = (conditions: FieldConditions, formData: Record<string, unknown>): boolean => {
		if (!conditions || !conditions.rules || conditions.rules.length === 0) {
			return true; // No conditions = always visible
		}

		const results = conditions.rules.map((rule) => {
			const fieldValue = formData[rule.fieldId];
			return evaluateRule(rule, fieldValue);
		});

		// Combine results based on logic operator
		const conditionsMet = conditions.logic === "and" ? results.every((r) => r) : results.some((r) => r);

		// Return based on show/hide flag
		return conditions.show ? conditionsMet : !conditionsMet;
	};

	/**
	 * Checks if a field should be visible
	 */
	const isFieldVisible = (field: FormField, formData: Record<string, unknown>): boolean => {
		if (!field.conditions) {
			return true; // No conditions = always visible
		}
		return evaluateConditions(field.conditions, formData);
	};

	/**
	 * Get all visible fields from a list
	 */
	const getVisibleFields = (fields: FormField[], formData: Record<string, unknown>): FormField[] => {
		return fields.filter((field) => isFieldVisible(field, formData));
	};

	/**
	 * Get all hidden fields from a list
	 */
	const getHiddenFields = (fields: FormField[], formData: Record<string, unknown>): FormField[] => {
		return fields.filter((field) => !isFieldVisible(field, formData));
	};

	return {
		evaluateRule,
		evaluateConditions,
		isFieldVisible,
		getVisibleFields,
		getHiddenFields,
	};
}
