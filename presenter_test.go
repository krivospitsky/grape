// Package grape provides tests for presenter.go functionality.
//
// Test Functions:
// - TestEntities: Tests presenter creation
// - TestField: Tests field addition
// - TestFieldAs: Tests JSON key customization
// - TestFieldWithSchema: Tests nested presenter setup
// - TestFieldFunc: Tests custom field transformation functions
// - TestFieldIf: Tests conditional field rendering
// - TestFieldDefaultValue: Tests default values for missing fields
// - TestFieldDescText: Tests field description setting
// - TestFieldExampleVal: Tests field example value setting
// - TestPresentBasicStruct: Tests basic struct presentation
// - TestPresentWithNilObject: Tests nil object handling
// - TestPresentWithPointer: Tests pointer object handling
// - TestPresentWithCustomJSONKeys: Tests custom JSON key mapping
// - TestPresentWithFieldFunc: Tests field transformation
// - TestPresentWithCondition: Tests conditional field inclusion
// - TestPresentWithDefault: Tests default value application
// - TestPresentNestedStruct: Tests nested struct presentation
// - TestPresentSliceOfStructs: Tests array of structs presentation
// - TestPresentMapOfStructs: Tests map with struct values presentation
// - TestPresentMixedTypes: Tests various data type handling
// - TestSerializeNestedWithSlice: Tests slice serialization
// - TestSerializeNestedWithMap: Tests map serialization
// - TestSerializeNestedWithPointer: Tests pointer serialization
package grape

import (
	"testing"
)

// === Presenter and Field Builder Tests ===

func TestEntities(t *testing.T) {
	p := NewEntity()
	if p == nil {
		t.Fatal("NewEntities() returned nil")
	}
	if len(p.Fields) != 0 {
		t.Errorf("Expected empty fields, got %d", len(p.Fields))
	}
}

func TestField(t *testing.T) {
	p := NewEntity()
	f := p.Field("name")
	if f == nil {
		t.Fatal("Field() returned nil")
	}
	if f.Name != "name" {
		t.Errorf("Expected name 'name', got %s", f.Name)
	}
	if f.JSONKey != "name" {
		t.Errorf("Expected JSONKey 'name', got %s", f.JSONKey)
	}
	if len(p.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(p.Fields))
	}
}

func TestFieldAs(t *testing.T) {
	p := NewEntity()
	f := p.Field("firstName").As("first_name")
	if f.JSONKey != "first_name" {
		t.Errorf("Expected JSONKey 'first_name', got %s", f.JSONKey)
	}
}

func TestFieldWithSchema(t *testing.T) {
	p := NewEntity()
	subP := NewEntity()
	f := p.Field("address").WithSchema(subP)
	if f.Presenter != subP {
		t.Error("Expected presenter to be set")
	}
}

func TestFieldFunc(t *testing.T) {
	p := NewEntity()
	f := p.Field("fullName").FieldFunc(func(obj any) any {
		return "test"
	})
	if f.Func == nil {
		t.Error("Expected func to be set")
	}
}

func TestFieldIf(t *testing.T) {
	p := NewEntity()
	f := p.Field("adminOnly").If(func(obj any, options H) bool {
		return true
	})
	if f.Condition == nil {
		t.Error("Expected condition to be set")
	}
}

func TestFieldDefaultValue(t *testing.T) {
	p := NewEntity()
	f := p.Field("status").DefaultValue("active")
	if f.Default != "active" {
		t.Errorf("Expected default 'active', got %v", f.Default)
	}
}

func TestFieldDescText(t *testing.T) {
	p := NewEntity()
	f := p.Field("email").DescText("User email address")
	if f.Desc != "User email address" {
		t.Errorf("Expected desc 'User email address', got %s", f.Desc)
	}
}

func TestFieldExampleVal(t *testing.T) {
	p := NewEntity()
	f := p.Field("age").ExampleVal(25)
	if f.Example != 25 {
		t.Errorf("Expected example 25, got %v", f.Example)
	}
}

// === Present Function Tests ===

type TestUser struct {
	Name  string
	Age   int
	Email string
}

func TestPresentBasicStruct(t *testing.T) {
	user := TestUser{Name: "John", Age: 30, Email: "john@example.com"}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age")
	p.Field("Email")

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if result["Age"] != 30 {
		t.Errorf("Expected Age 30, got %v", result["Age"])
	}
	if result["Email"] != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %v", result["Email"])
	}
}

func TestPresentWithNilObject(t *testing.T) {
	p := NewEntity()
	p.Field("Name")
	result := Present(nil, p)
	if len(result) != 0 {
		t.Errorf("Expected empty map, got %v", result)
	}
}

func TestPresentWithPointer(t *testing.T) {
	user := &TestUser{Name: "John", Age: 30}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age")

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if result["Age"] != 30 {
		t.Errorf("Expected Age 30, got %v", result["Age"])
	}
}

func TestPresentWithCustomJSONKeys(t *testing.T) {
	user := TestUser{Name: "John", Age: 30}

	p := NewEntity()
	p.Field("Name").As("full_name")
	p.Field("Age").As("user_age")

	result := Present(user, p)

	if result["full_name"] != "John" {
		t.Errorf("Expected full_name 'John', got %v", result["full_name"])
	}
	if result["user_age"] != 30 {
		t.Errorf("Expected user_age 30, got %v", result["user_age"])
	}
}

func TestPresentWithFieldFunc(t *testing.T) {
	user := TestUser{Name: "John", Age: 30}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age").FieldFunc(func(obj any) any {
		u := obj.(TestUser)
		return u.Age + 1
	})

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if result["Age"] != 31 {
		t.Errorf("Expected Age 31, got %v", result["Age"])
	}
}

func TestPresentWithCondition(t *testing.T) {
	user := TestUser{Name: "John", Age: 30}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age").If(func(obj any, options H) bool {
		u := obj.(TestUser)
		return u.Age >= 18
	})

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if result["Age"] != 30 {
		t.Errorf("Expected Age 30, got %v", result["Age"])
	}

	// Test condition that excludes field
	user2 := TestUser{Name: "Jane", Age: 16}
	result2 := Present(user2, p)

	if result2["Name"] != "Jane" {
		t.Errorf("Expected Name 'Jane', got %v", result2["Name"])
	}
	if _, exists := result2["Age"]; exists {
		t.Error("Expected Age field to be excluded")
	}
}

func TestPresentWithDefault(t *testing.T) {
	type TestUserWithOptional struct {
		Name string
		City string
	}

	user := TestUserWithOptional{Name: "John"}

	p := NewEntity()
	p.Field("Name")
	p.Field("City").DefaultValue("Unknown")

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if result["City"] != "Unknown" {
		t.Errorf("Expected City 'Unknown', got %v", result["City"])
	}
}

type TestAddress struct {
	Street string
	City   string
}

type TestUserWithAddress struct {
	Name    string
	Address *TestAddress
}

func TestPresentNestedStruct(t *testing.T) {
	addr := &TestAddress{Street: "123 Main St", City: "NYC"}
	user := TestUserWithAddress{Name: "John", Address: addr}

	addrP := NewEntity()
	addrP.Field("Street")
	addrP.Field("City")

	p := NewEntity()
	p.Field("Name")
	p.Field("Address").WithSchema(addrP)

	result := Present(user, p)

	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}

	addrResult, ok := result["Address"].(H)
	if !ok {
		t.Fatal("Expected Address to be map")
	}
	if addrResult["Street"] != "123 Main St" {
		t.Errorf("Expected Street '123 Main St', got %v", addrResult["Street"])
	}
	if addrResult["City"] != "NYC" {
		t.Errorf("Expected City 'NYC', got %v", addrResult["City"])
	}
}

func TestPresentSliceOfStructs(t *testing.T) {
	users := []TestUser{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
	}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age")

	result := serializeNested(users, p)

	usersResult, ok := result.([]any)
	if !ok {
		t.Fatalf("Expected result to be slice, got %T", result)
	}
	if len(usersResult) != 2 {
		t.Errorf("Expected 2 users, got %d", len(usersResult))
	}

	user1, ok := usersResult[0].(H)
	if !ok {
		t.Fatal("Expected user1 to be map")
	}
	if user1["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", user1["Name"])
	}
}

func TestPresentSlice(t *testing.T) {
	users := []TestUser{
		{Name: "Alice", Age: 28},
		{Name: "Bob", Age: 35},
	}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age")

	result := PresentSlice(users, p)

	if len(result) != 2 {
		t.Errorf("Expected 2 users, got %d", len(result))
	}

	user1, ok := result[0].(H)
	if !ok {
		t.Fatal("Expected user1 to be map")
	}
	if user1["Name"] != "Alice" {
		t.Errorf("Expected Name 'Alice', got %v", user1["Name"])
	}
	if user1["Age"] != 28 {
		t.Errorf("Expected Age 28, got %v", user1["Age"])
	}

	user2, ok := result[1].(H)
	if !ok {
		t.Fatal("Expected user2 to be map")
	}
	if user2["Name"] != "Bob" {
		t.Errorf("Expected Name 'Bob', got %v", user2["Name"])
	}
	if user2["Age"] != 35 {
		t.Errorf("Expected Age 35, got %v", user2["Age"])
	}
}

func TestPresentSliceWithNil(t *testing.T) {
	result := PresentSlice(nil, NewEntity())
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %d", len(result))
	}
}

func TestPresentSliceWithNonSlice(t *testing.T) {
	result := PresentSlice("not a slice", NewEntity())
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %d", len(result))
	}
}

func TestPresentMapOfStructs(t *testing.T) {
	userMap := map[string]any{
		"user1": TestUser{Name: "John", Age: 30},
		"user2": TestUser{Name: "Jane", Age: 25},
	}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age")

	result := serializeNested(userMap, p)

	resultMap, ok := result.(H)
	if !ok {
		t.Fatalf("Expected result to be map, got %T", result)
	}

	user1, ok := resultMap["user1"].(H)
	if !ok {
		t.Fatal("Expected user1 to be map")
	}
	if user1["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", user1["Name"])
	}
}

func TestPresentMixedTypes(t *testing.T) {
	type MixedStruct struct {
		Name   string
		Count  int
		Active bool
		Items  []string
		Meta   map[string]any
	}

	obj := MixedStruct{
		Name:   "Test",
		Count:  42,
		Active: true,
		Items:  []string{"a", "b", "c"},
		Meta:   map[string]any{"key": "value"},
	}

	p := NewEntity()
	p.Field("Name")
	p.Field("Count")
	p.Field("Active")
	p.Field("Items")
	p.Field("Meta")

	result := Present(obj, p)

	if result["Name"] != "Test" {
		t.Errorf("Expected Name 'Test', got %v", result["Name"])
	}
	if result["Count"] != 42 {
		t.Errorf("Expected Count 42, got %v", result["Count"])
	}
	if result["Active"] != true {
		t.Errorf("Expected Active true, got %v", result["Active"])
	}
	if len(result["Items"].([]string)) != 3 {
		t.Errorf("Expected Items length 3, got %v", result["Items"])
	}
	if result["Meta"].(map[string]any)["key"] != "value" {
		t.Errorf("Expected Meta key 'value', got %v", result["Meta"])
	}
}

// === SerializeNested Tests ===

func TestSerializeNestedWithSlice(t *testing.T) {
	users := []TestUser{
		{Name: "John"},
		{Name: "Jane"},
	}

	p := NewEntity()
	p.Field("Name")

	result := serializeNested(users, p)

	resultSlice, ok := result.([]any)
	if !ok {
		t.Fatal("Expected slice result")
	}
	if len(resultSlice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resultSlice))
	}
}

func TestSerializeNestedWithMap(t *testing.T) {
	userMap := map[string]any{
		"user1": TestUser{Name: "John"},
		"user2": TestUser{Name: "Jane"},
	}

	p := NewEntity()
	p.Field("Name")

	result := serializeNested(userMap, p)

	resultMap, ok := result.(H)
	if !ok {
		t.Fatal("Expected map result")
	}
	if len(resultMap) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resultMap))
	}
}

func TestSerializeNestedWithPointer(t *testing.T) {
	user := &TestUser{Name: "John"}

	p := NewEntity()
	p.Field("Name")

	result := serializeNested(user, p)

	resultMap, ok := result.(H)
	if !ok {
		t.Fatal("Expected map result")
	}
	if resultMap["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", resultMap["Name"])
	}
}

func TestPresentWithM(t *testing.T) {
	user := TestUser{Name: "John", Age: 30, Email: "john@example.com"}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age").If(func(obj any, options H) bool {
		return options["type"] == "full"
	})
	p.Field("Email").If(func(obj any, options H) bool {
		return options["include_email"] == true
	})

	// Test with no options - should only show Name
	result := Present(user, p)
	if result["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result["Name"])
	}
	if _, exists := result["Age"]; exists {
		t.Error("Expected Age field to be excluded without full type")
	}
	if _, exists := result["Email"]; exists {
		t.Error("Expected Email field to be excluded without include_email")
	}

	// Test with full type - should show Name and Age
	result2 := Present(user, p, H{"type": "full"})
	if result2["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result2["Name"])
	}
	if result2["Age"] != 30 {
		t.Errorf("Expected Age 30, got %v", result2["Age"])
	}
	if _, exists := result2["Email"]; exists {
		t.Error("Expected Email field to be excluded without include_email")
	}

	// Test with include_email - should show Name and Email
	result3 := Present(user, p, H{"include_email": true})
	if result3["Name"] != "John" {
		t.Errorf("Expected Name 'John', got %v", result3["Name"])
	}
	if _, exists := result3["Age"]; exists {
		t.Error("Expected Age field to be excluded without full type")
	}
	if result3["Email"] != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %v", result3["Email"])
	}
}

func TestPresentSliceWithM(t *testing.T) {
	users := []TestUser{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 35},
	}

	p := NewEntity()
	p.Field("Name")
	p.Field("Age").If(func(obj any, options H) bool {
		return options["show_age"] == true
	})

	// Test without options - should only show names
	result := PresentSlice(users, p)
	if len(result) != 2 {
		t.Errorf("Expected 2 users, got %d", len(result))
	}

	user1 := result[0].(H)
	if user1["Name"] != "Alice" {
		t.Errorf("Expected Name 'Alice', got %v", user1["Name"])
	}
	if _, exists := user1["Age"]; exists {
		t.Error("Expected Age field to be excluded")
	}

	// Test with options - should show names and ages
	result2 := PresentSlice(users, p, H{"show_age": true})
	user1WithAge := result2[0].(H)
	if user1WithAge["Name"] != "Alice" {
		t.Errorf("Expected Name 'Alice', got %v", user1WithAge["Name"])
	}
	if user1WithAge["Age"] != 25 {
		t.Errorf("Expected Age 25, got %v", user1WithAge["Age"])
	}
}
