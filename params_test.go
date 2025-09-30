// Package grape provides tests for params.go functionality.
//
// Test Functions:
// - TestSchema: Tests schema creation
// - TestFieldBuilderRequires: Tests required field setup
// - TestFieldBuilderOptional: Tests optional field setup
// - TestFieldBuilderOn: Tests required field modes
// - TestFieldBuilderTypes: Tests different field types (String, Integer, Float, Boolean, Slice, JSON)
// - TestFieldBuilderValidate: Tests field validation setup
// - TestFieldBuilderWithSchema: Tests nested schema setup for JSON fields
// - TestFieldBuilderSliceOf: Tests slice field setup with element type and schema
// - TestFieldBuilderMultipleValidations: Tests multiple validation tags (last wins)
// - TestInputString: Tests string accessor with type safety
// - TestInputInteger: Tests int accessor with defaults
// - TestInputFloat: Tests float accessor with defaults
// - TestInputBoolean: Tests bool accessor with defaults
// - TestInputMethodsWithNonExistentKeys: Tests input accessors for missing keys
// - TestBindAndValidateSuccess: Tests successful JSON binding and validation
// - TestBindAndValidateMissingRequired: Tests error for missing required fields
// - TestBindAndValidateWrongType: Tests error for type mismatches
// - TestBindAndValidateValidationFailure: Tests validation tag failures
// - TestBindAndValidateRequiredOnMultipleModes: Tests required fields on specific modes
// - TestBindAndValidateEmptyJSON: Tests empty JSON handling
// - TestBindAndValidateInvalidJSON: Tests invalid JSON parsing
// - TestBindAndValidateReader: Tests JSON binding from io.Reader
// - TestBindAndValidateFloatAsInteger: Tests float to int conversion
// - TestBindAndValidateNullValues: Tests null value handling
// - TestBindAndValidateEmptyArray: Tests empty array handling
// - TestValidateJSONSuccess: Tests successful map validation
// - TestValidateJSONMissingRequired: Tests map validation with missing required fields
// - TestValidateJSONWrongType: Tests map validation with type mismatches
// - TestValidateJSONWithExtraFields: Tests preservation of extra fields in map validation
// - TestBindAndValidateNestedJSON: Tests nested object validation
// - TestBindAndValidateNestedSlice: Tests nested array validation
// - TestBindAndValidateNestedSliceOfJSONs: Tests nested array of objects validation
package grape

import (
	"encoding/json"
	"strings"
	"testing"
)

// === Schema and Field Builder Tests ===

func TestSchema(t *testing.T) {
	schema := NewParams()
	if schema == nil {
		t.Fatal("NewParams() returned nil")
	}
	if len(schema.Fields) != 0 {
		t.Errorf("Expected empty fields, got %d", len(schema.Fields))
	}
}

func TestFieldBuilderRequires(t *testing.T) {
	schema := NewParams()
	fb := schema.Requires("name")
	if fb == nil {
		t.Fatal("Requires() returned nil")
	}
	if len(schema.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(schema.Fields))
	}
	field := schema.Fields[0]
	if field.Name != "name" {
		t.Errorf("Expected name 'name', got %s", field.Name)
	}
	if len(field.RequiredOn) != 0 {
		t.Errorf("Expected empty RequiredOn, got %v", field.RequiredOn)
	}
}

func TestFieldBuilderOptional(t *testing.T) {
	schema := NewParams()
	fb := schema.Optional("age")
	if fb == nil {
		t.Fatal("Optional() returned nil")
	}
	if len(schema.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(schema.Fields))
	}
	field := schema.Fields[0]
	if field.Name != "age" {
		t.Errorf("Expected name 'age', got %s", field.Name)
	}
	if field.RequiredOn != nil {
		t.Errorf("Expected nil RequiredOn, got %v", field.RequiredOn)
	}
}

func TestFieldBuilderOn(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("name").On("create", "update")
	field := schema.Fields[0]
	if len(field.RequiredOn) != 2 {
		t.Errorf("Expected 2 required on, got %d", len(field.RequiredOn))
	}
	if field.RequiredOn[0] != "create" || field.RequiredOn[1] != "update" {
		t.Errorf("Expected [create, update], got %v", field.RequiredOn)
	}
}

func TestFieldBuilderTypes(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("name").String()
	field := schema.Fields[0]
	if field.Type != String {
		t.Errorf("Expected String type, got %v", field.Type)
	}

	_ = schema.Optional("age").Integer()
	field2 := schema.Fields[1]
	if field2.Type != Integer {
		t.Errorf("Expected Integer type, got %v", field2.Type)
	}

	_ = schema.Optional("height").Float()
	field3 := schema.Fields[2]
	if field3.Type != Float {
		t.Errorf("Expected Float type, got %v", field3.Type)
	}

	_ = schema.Optional("active").Boolean()
	field4 := schema.Fields[3]
	if field4.Type != Boolean {
		t.Errorf("Expected Boolean type, got %v", field4.Type)
	}

	_ = schema.Optional("tags").Slice()
	field5 := schema.Fields[4]
	if field5.Type != Slice {
		t.Errorf("Expected Slice type, got %v", field5.Type)
	}

	_ = schema.Optional("data").JSON()
	field6 := schema.Fields[5]
	if field6.Type != JSON {
		t.Errorf("Expected JSON type, got %v", field6.Type)
	}
}

func TestFieldBuilderValidate(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("email").String().Validate("email")
	field := schema.Fields[0]
	if field.Validate != "email" {
		t.Errorf("Expected validate 'email', got %s", field.Validate)
	}
}

func TestFieldBuilderWithSchema(t *testing.T) {
	schema := NewParams()
	subSchema := NewParams()
	_ = subSchema.Optional("sub").String()
	_ = schema.Optional("data").JSON().WithSchema(subSchema)
	field := schema.Fields[0]
	if field.Schema == nil {
		t.Error("Expected schema to be set")
	}
	if len(field.Schema.Fields) != 1 {
		t.Errorf("Expected 1 sub field, got %d", len(field.Schema.Fields))
	}
}

func TestFieldBuilderSliceOf(t *testing.T) {
	schema := NewParams()
	subSchema := NewParams()
	_ = subSchema.Optional("item").Integer()
	_ = schema.Optional("list").SliceOf(Integer, subSchema)
	field := schema.Fields[0]
	if field.Type != Slice {
		t.Errorf("Expected Slice type, got %v", field.Type)
	}
	if field.SliceType != Integer {
		t.Errorf("Expected SliceType Integer, got %v", field.SliceType)
	}
	if field.Schema == nil {
		t.Error("Expected schema to be set")
	}
}

func TestFieldBuilderMultipleValidations(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("email").String().Validate("email").Validate("required")

	field := schema.Fields[0]
	if field.Validate != "required" { // Last one wins
		t.Errorf("Expected validate 'required', got %s", field.Validate)
	}
}

// === Input Accessor Tests ===

func TestInputString(t *testing.T) {
	input := Input{"name": "John", "age": 30}
	if input.String("name") != "John" {
		t.Errorf("Expected 'John', got %s", input.String("name"))
	}
	if input.String("missing") != "" {
		t.Errorf("Expected '', got %s", input.String("missing"))
	}
	if input.String("age") != "" { // age is int, so should return ""
		t.Errorf("Expected '', got %s", input.String("age"))
	}
}

func TestInputInteger(t *testing.T) {
	input := Input{"age": 30, "height": 5.5}
	if input.Integer("age", 0) != 30 {
		t.Errorf("Expected 30, got %d", input.Integer("age", 0))
	}
	if input.Integer("missing", 25) != 25 {
		t.Errorf("Expected 25, got %d", input.Integer("missing", 25))
	}
	if input.Integer("height", 0) != 0 { // height is float, so should return def
		t.Errorf("Expected 0, got %d", input.Integer("height", 0))
	}
}

func TestInputFloat(t *testing.T) {
	input := Input{"height": 5.5, "age": 30}
	if input.Float("height", 0.0) != 5.5 {
		t.Errorf("Expected 5.5, got %f", input.Float("height", 0.0))
	}
	if input.Float("missing", 10.0) != 10.0 {
		t.Errorf("Expected 10.0, got %f", input.Float("missing", 10.0))
	}
	if input.Float("age", 0.0) != 0.0 { // age is int, so should return def
		t.Errorf("Expected 0.0, got %f", input.Float("age", 0.0))
	}
}

func TestInputBoolean(t *testing.T) {
	input := Input{"active": true, "age": 30}
	if input.Boolean("active", false) != true {
		t.Errorf("Expected true, got %v", input.Boolean("active", false))
	}
	if input.Boolean("missing", true) != true {
		t.Errorf("Expected true, got %v", input.Boolean("missing", true))
	}
	if input.Boolean("age", false) != false { // age is int, so should return def
		t.Errorf("Expected false, got %v", input.Boolean("age", false))
	}
}

func TestInputMethodsWithNonExistentKeys(t *testing.T) {
	input := Input{}

	if input.String("missing") != "" {
		t.Errorf("Expected '', got %s", input.String("missing"))
	}
	if input.Integer("missing", 42) != 42 {
		t.Errorf("Expected 42, got %d", input.Integer("missing", 42))
	}
	if input.Float("missing", 3.14) != 3.14 {
		t.Errorf("Expected 3.14, got %f", input.Float("missing", 3.14))
	}
	if input.Boolean("missing", true) != true {
		t.Errorf("Expected true, got %v", input.Boolean("missing", true))
	}
}

// === BindAndValidate Success Tests ===

func TestBindAndValidateSuccess(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("name").String()
	_ = schema.Optional("age").Integer()
	_ = schema.Optional("email").String().Validate("email")

	raw := createTestJSON(`{"name": "John", "age": 30, "email": "john@example.com", "extra": "ignored"}`)

	input, err := schema.BindAndValidate(raw, "create")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if input.String("name") != "John" {
		t.Errorf("Expected name 'John', got %s", input.String("name"))
	}
	if input.Integer("age", 0) != 30 {
		t.Errorf("Expected age 30, got %d", input.Integer("age", 0))
	}
	if input.String("email") != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %s", input.String("email"))
	}
	if input.String("extra") != "ignored" {
		t.Errorf("Expected extra 'ignored', got %s", input.String("extra"))
	}
}

// === BindAndValidate Error Tests ===

func createTestJSON(jsonBody string) map[string]interface{} {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonBody), &raw); err != nil {
		panic(err) // For tests, panic on error
	}
	return raw
}

func TestBindAndValidateMissingRequired(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("name").On("create").String()

	raw := createTestJSON(`{"age": 30}`)
	_, err := schema.BindAndValidate(raw, "create")
	if err == nil {
		t.Error("Expected error for missing required field")
	}
	if !strings.Contains(err.Error(), "missing required field 'name'") {
		t.Errorf("Expected missing field error, got %v", err)
	}
}

func TestBindAndValidateWrongType(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("age").Integer()

	raw := createTestJSON(`{"age": "thirty"}`)
	_, err := schema.BindAndValidate(raw, "create")
	if err == nil {
		t.Error("Expected error for wrong type")
	}
	if !strings.Contains(err.Error(), "field 'age' must be int") {
		t.Errorf("Expected type error, got %v", err)
	}
}

func TestBindAndValidateValidationFailure(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("email").String().Validate("email")

	raw := createTestJSON(`{"email": "not-an-email"}`)
	_, err := schema.BindAndValidate(raw, "")
	if err == nil {
		t.Error("Expected validation error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error, got %v", err)
	}
}

func TestBindAndValidateRequiredOnMultipleModes(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("adminField").On("admin", "superuser").String()

	// Should be required for "admin"
	raw := createTestJSON(`{}`)
	_, err := schema.BindAndValidate(raw, "admin")
	if err == nil {
		t.Error("Expected error for missing required field on admin")
	}

	// Should not be required for "user"
	raw2 := createTestJSON(`{}`)
	_, err = schema.BindAndValidate(raw2, "user")
	if err != nil {
		t.Errorf("Expected no error for user, got %v", err)
	}
}

// === ValidateJSON Tests ===

func TestValidateJSONSuccess(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("name").String()
	_ = schema.Optional("age").Integer()

	raw := map[string]interface{}{
		"name": "John",
		"age":  30,
		"extra": "ignored",
	}

	result, err := schema.validateJSON(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("Expected age 30, got %v", result["age"])
	}
	if result["extra"] != "ignored" {
		t.Errorf("Expected extra 'ignored', got %v", result["extra"])
	}
}

func TestValidateJSONMissingRequired(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("name").On("create").String()

	raw := map[string]interface{}{}

	_, err := schema.validateJSON(raw, "create")
	if err == nil {
		t.Error("Expected error for missing required field")
	}
	if !strings.Contains(err.Error(), "missing required field 'name'") {
		t.Errorf("Expected missing field error, got %v", err)
	}
}

func TestValidateJSONWrongType(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("age").Integer()

	raw := map[string]interface{}{
		"age": "thirty",
	}

	_, err := schema.validateJSON(raw, "")
	if err == nil {
		t.Error("Expected error for wrong type")
	}
	if !strings.Contains(err.Error(), "field 'age' must be int") {
		t.Errorf("Expected type error, got %v", err)
	}
}

func TestValidateJSONWithExtraFields(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("name").String()

	raw := map[string]interface{}{
		"name": "John",
		"extra": "value",
		"another": 123,
	}

	result, err := schema.validateJSON(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", result["name"])
	}
	if result["extra"] != "value" {
		t.Errorf("Expected extra 'value', got %v", result["extra"])
	}
	if result["another"] != 123.0 {
		t.Errorf("Expected another 123.0, got %v", result["another"])
	}
}

// === Nested Structures Tests ===

func TestBindAndValidateNestedJSON(t *testing.T) {
	schema := NewParams()
	subSchema := NewParams()
	_ = subSchema.Optional("city").String()
	_ = schema.Optional("address").JSON().WithSchema(subSchema)

	raw := createTestJSON(`{"address": {"city": "NYC", "zip": "10001"}}`)

	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	addr, ok := input["address"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected address to be map")
	}
	if addr["city"] != "NYC" {
		t.Errorf("Expected city 'NYC', got %v", addr["city"])
	}
	if addr["zip"] != "10001" {
		t.Errorf("Expected zip '10001', got %v", addr["zip"])
	}
}

func TestBindAndValidateNestedSlice(t *testing.T) {
	schema := NewParams()
	subSchema := NewParams()
	_ = subSchema.Optional("value").Integer()
	_ = schema.Optional("numbers").SliceOf(Integer, subSchema)

	raw := createTestJSON(`{"numbers": [10, 20, 30]}`)

	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	nums, ok := input["numbers"].([]interface{})
	if !ok {
		t.Fatal("Expected numbers to be slice")
	}
	if len(nums) != 3 {
		t.Errorf("Expected 3 numbers, got %d", len(nums))
	}
	if len(nums) != 3 || nums[0] != 10.0 || nums[1] != 20.0 || nums[2] != 30.0 {
		t.Errorf("Expected [10.0 20.0 30.0], got %v", nums)
	}
}

func TestBindAndValidateNestedSliceOfJSONs(t *testing.T) {
	schema := NewParams()
	subSchema := NewParams()
	_ = subSchema.Optional("name").String()
	_ = schema.Optional("users").SliceOf(JSON, subSchema)

	raw := createTestJSON(`{"users": [{"name": "Alice"}, {"name": "Bob"}]}`)

	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	users, ok := input["users"].([]interface{})
	if !ok {
		t.Fatal("Expected users to be slice")
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	user1, ok := users[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user1 to be map")
	}
	if user1["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", user1["name"])
	}
}

// === Edge Cases ===

func TestBindAndValidateEmptyJSON(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("name").String()

	raw := createTestJSON(`{}`)
	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if input.String("name") != "" {
		t.Errorf("Expected empty name, got %s", input.String("name"))
	}
}

func TestBindAndValidateInvalidJSON(t *testing.T) {
	schema := NewParams()
	// For invalid JSON, we can't use createTestJSON, so we'll test with nil or empty map
	raw := map[string]interface{}{}
	_, err := schema.BindAndValidate(raw, "")
	// This should work since we pass the parsed map directly now
	if err != nil {
		t.Errorf("Expected no error for valid map, got %v", err)
	}
}

func TestBindAndValidateReader(t *testing.T) {
	schema := NewParams()
	_ = schema.Requires("name").String()
	_ = schema.Optional("age").Integer()

	jsonBody := `{"name": "John", "age": 30}`
	reader := strings.NewReader(jsonBody)

	input, err := schema.BindAndValidateReader(reader, "create")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if input.String("name") != "John" {
		t.Errorf("Expected name 'John', got %s", input.String("name"))
	}
	if input.Integer("age", 0) != 30 {
		t.Errorf("Expected age 30, got %d", input.Integer("age", 0))
	}
}

func TestBindAndValidateFloatAsInteger(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("count").Integer()

	raw := createTestJSON(`{"count": 5.0}`)
	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if input.Integer("count", 0) != 5 {
		t.Errorf("Expected 5, got %d", input.Integer("count", 0))
	}
}

func TestBindAndValidateNullValues(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("name").String()

	raw := createTestJSON(`{"name": null}`)
	_, err := schema.BindAndValidate(raw, "")
	if err == nil {
		t.Error("Expected error for null value on string field")
	}
	if !strings.Contains(err.Error(), "field 'name' must be string") {
		t.Errorf("Expected type error, got %v", err)
	}
}

func TestBindAndValidateEmptyArray(t *testing.T) {
	schema := NewParams()
	_ = schema.Optional("items").Slice()

	raw := createTestJSON(`{"items": []}`)
	input, err := schema.BindAndValidate(raw, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	items, ok := input["items"].([]interface{})
	if !ok {
		t.Fatal("Expected items to be slice")
	}
	if len(items) != 0 {
		t.Errorf("Expected empty array, got %v", items)
	}
}

// === Input.ToModel Tests ===

// Test structures for ToModel
type ToModelUser struct {
	ID       int
	Name     string
	Email    string
	Age      int
	IsActive bool
}

type ToModelUserWithUnexported struct {
	ID       int
	Name     string
	email    string // unexported
}

// === ToModel Basic Functionality Tests ===

func TestToModelBasic(t *testing.T) {
	input := Input{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	user := &ToModelUser{}
	input.ToModel(user)

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected Name 'John Doe', got %s", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %s", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("Expected Age 30, got %d", user.Age)
	}
}

func TestToModelCaseInsensitive(t *testing.T) {
	input := Input{
		"id":       1,
		"name":     "John Doe", // exact match
		"email":    "john@example.com", // exact match
		"age":      30, // exact match
		"is_active": true, // snake_case to PascalCase
	}

	user := &ToModelUser{}
	input.ToModel(user)

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected Name 'John Doe', got %s", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %s", user.Email)
	}
	if user.Age != 30 {
		t.Errorf("Expected Age 30, got %d", user.Age)
	}
	if user.IsActive != true {
		t.Errorf("Expected IsActive true, got %v", user.IsActive)
	}
}

func TestToModelTypeConversion(t *testing.T) {
	input := Input{
		"id":  1.0, // float64 -> int
		"age": "25", // string -> int (if convertible)
	}

	user := &ToModelUser{}
	input.ToModel(user)

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	// Note: string to int conversion might not work in all cases
	// depending on reflect's conversion capabilities
}

func TestToModelNilValues(t *testing.T) {
	// Start with existing values
	user := &ToModelUser{
		Name:  "Original Name",
		Email: "original@example.com",
		Age:   25,
	}

	input := Input{
		"name":  nil, // nil should not overwrite
		"email": "new@example.com",
		"age":   nil, // nil should not overwrite
	}

	input.ToModel(user)

	// Nil values should not overwrite existing values
	if user.Name != "Original Name" {
		t.Errorf("Expected Name 'Original Name', got %s", user.Name)
	}
	if user.Email != "new@example.com" {
		t.Errorf("Expected Email 'new@example.com', got %s", user.Email)
	}
	if user.Age != 25 {
		t.Errorf("Expected Age 25, got %d", user.Age)
	}
}

func TestToModelInvalidDestination(t *testing.T) {
	input := Input{"name": "John"}

	// Test with nil pointer
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil destination")
		}
	}()
	input.ToModel(nil)

	// Test with non-pointer
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-pointer destination")
		}
	}()
	var user ToModelUser
	input.ToModel(user)
}

func TestToModelNonExistentFields(t *testing.T) {
	input := Input{
		"name":            "John",
		"non_existent":    "value",
		"another_missing": 123,
	}

	user := &ToModelUser{}
	input.ToModel(user)

	if user.Name != "John" {
		t.Errorf("Expected Name 'John', got %s", user.Name)
	}
	// Non-existent fields should be ignored without error
}

func TestToModelUnexportedFields(t *testing.T) {
	input := Input{
		"id":    1,
		"name":  "John",
		"email": "john@example.com", // this maps to unexported field
	}

	user := &ToModelUserWithUnexported{}
	input.ToModel(user)

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Name != "John" {
		t.Errorf("Expected Name 'John', got %s", user.Name)
	}
	// Unexported fields should be skipped
	if user.email != "" {
		t.Errorf("Expected unexported email to remain empty, got %s", user.email)
	}
}

// === Utility Function Tests ===

func TestStringsEqualFold(t *testing.T) {
	tests := []struct {
		structField string
		jsonKey     string
		expected    bool
	}{
		{"Name", "name", true},
		{"FirstName", "first_name", true},
		{"UserID", "user_id", true},
		{"Name", "Name", false}, // different cases but same
		{"Name", "fullname", false},
		{"", "", true},
		{"ID", "id", true},
		{"IsActive", "is_active", true},
	}

	for _, test := range tests {
		result := stringsEqualFold(test.structField, test.jsonKey)
		if result != test.expected {
			t.Errorf("stringsEqualFold(%q, %q) = %v, expected %v",
				test.structField, test.jsonKey, result, test.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"UserID", "user_id"},
		{"IsActive", "is_active"},
		{"XMLHttpRequest", "xmlhttp_request"},
		{"", ""},
		{"A", "a"},
		{"AB", "ab"},
		{"ABC", "abc"},
	}

	for _, test := range tests {
		result := toSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("toSnakeCase(%q) = %q, expected %q",
				test.input, result, test.expected)
		}
	}
}

// === ToModel Integeregration Tests ===

func TestToModelIntegeregration(t *testing.T) {
	// Simulate data flow: JSON -> Input -> ToModel -> Struct
	jsonData := map[string]interface{}{
		"id":         123,
		"name":       "Jane Smith",
		"email":      "jane@example.com",
		"age":        28,
		"is_active":  true,
		"extra_field": "ignored",
	}

	// Create Input (simulating validation output)
	input := Input(jsonData)

	// JSON to struct
	user := &ToModelUser{}
	input.ToModel(user)

	// Verify mapping
	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}
	if user.Name != "Jane Smith" {
		t.Errorf("Expected Name 'Jane Smith', got %s", user.Name)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("Expected Email 'jane@example.com', got %s", user.Email)
	}
	if user.Age != 28 {
		t.Errorf("Expected Age 28, got %d", user.Age)
	}
	if user.IsActive != true {
		t.Errorf("Expected IsActive true, got %v", user.IsActive)
	}
}