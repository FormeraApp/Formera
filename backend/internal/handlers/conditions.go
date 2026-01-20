package handlers

import (
	"fmt"
	"formera/internal/models"
)

// ValidateFieldConditions validates that field conditions don't create circular dependencies
// and that all referenced fields exist
func ValidateFieldConditions(fields []models.FormField) error {
	// Build field ID map
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field.ID] = true
	}

	// Build dependency graph
	dependencies := make(map[string][]string)
	for _, field := range fields {
		if field.Conditions != nil && len(field.Conditions.Rules) > 0 {
			for _, rule := range field.Conditions.Rules {
				// Check if referenced field exists
				if !fieldMap[rule.FieldID] {
					return fmt.Errorf("field %s references non-existent field %s in conditions",
						field.ID, rule.FieldID)
				}

				// Check for self-reference
				if rule.FieldID == field.ID {
					return fmt.Errorf("field %s cannot have a condition on itself", field.ID)
				}

				// Check that referenced field is not a layout field
				referencedField := findField(fields, rule.FieldID)
				if referencedField != nil && isLayoutField(referencedField.Type) {
					return fmt.Errorf("field %s cannot reference layout field %s in conditions",
						field.ID, rule.FieldID)
				}

				dependencies[field.ID] = append(dependencies[field.ID], rule.FieldID)
			}
		}
	}

	// Check for circular dependencies using DFS
	for fieldID := range dependencies {
		visited := make(map[string]bool)
		if hasCycle(fieldID, dependencies, visited, make(map[string]bool)) {
			return fmt.Errorf("circular dependency detected involving field %s", fieldID)
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles in the dependency graph
func hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if hasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// findField finds a field by ID
func findField(fields []models.FormField, id string) *models.FormField {
	for i := range fields {
		if fields[i].ID == id {
			return &fields[i]
		}
	}
	return nil
}

// isLayoutField checks if a field type is a layout field
func isLayoutField(fieldType models.FieldType) bool {
	layoutTypes := []models.FieldType{
		models.FieldTypeSection,
		models.FieldTypePagebreak,
		models.FieldTypeDivider,
		models.FieldTypeHeading,
		models.FieldTypeParagraph,
		models.FieldTypeImage,
	}

	for _, t := range layoutTypes {
		if fieldType == t {
			return true
		}
	}
	return false
}
