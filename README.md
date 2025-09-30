# grape

A lightweight, framework-agnostic Go library for parameter validation, JSON binding, and data presentation. Designed to work seamlessly with popular HTTP frameworks like Gin, Echo, Chi, and Fiber.

**Inspired by the [Ruby on Rails Grape gem](https://github.com/ruby-grape/grape)** - bringing the simplicity and elegance of Grape's parameter validation and API building approach to the Go ecosystem.

[![Go Report Card](https://goreportcard.com/badge/github.com/krivospitsky/grape)](https://goreportcard.com/report/github.com/krivospitsky/grape)
[![Go Reference](https://pkg.go.dev/badge/github.com/krivospitsky/grape.svg)](https://pkg.go.dev/github.com/krivospitsky/grape)

## Features

- **Framework Agnostic**: Works with Gin, Echo, Chi, Fiber, or standalone
- **Schema Reuse**: Define validation rules once, apply different modes for different operations
- **Type-Safe Validation**: Strong typing with comprehensive validation rules
- **Nested Structures**: Support for complex nested objects and arrays
- **Conditional Fields**: Dynamic field inclusion based on runtime conditions
- **Custom Transformations**: Field-level data transformation functions
- **Default Values**: Automatic default value assignment
- **JSON Binding**: Direct JSON parsing with detailed error handling
- **Data Presentation**: Transform Go structs to JSON-friendly maps with presenters
- **ORM Mapping**: Map validated data to GORM-compatible model structs

## Installation

```bash
go get github.com/krivospitsky/grape
```

## Quick Start

Build a complete REST API with validation, database mapping, and presentation using Gin and GORM:

```go
import (
    "github.com/krivospitsky/grape"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// User model - GORM will create users table
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"` // Don't expose in API
    Age       *int      // Nullable field
    IsActive  bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// User validation schema - defined once, used everywhere
var userSchema = grape.NewParams().
    Requires("name").On("create").String().Validate("required,min=2,max=100"). // Required only for creation
    Requires("email").On("create").String().Validate("email").
    Optional("age").Integer().Validate("min=0,max=150").
    Requires("password").On("create").String().Validate("min=8").
    Optional("isActive").On("update").Boolean()

// Single user presenter with conditional fields based on options
var userPresenter = grape.NewEntity().
    Field("ID").As("user_id").
    Field("Name").
    Field("Email").
    Field("Age").If(func(obj any, options grape.H) bool {
        return options["view"] == "detailed" // Only show age in detailed view
    }).
    Field("IsActive").If(func(obj any, options grape.H) bool {
        return options["view"] == "detailed" // Only show active status in detailed view
    }).
    Field("CreatedAt").FieldFunc(func(obj any) any {
        return obj.(User).CreatedAt.Format(time.RFC3339)
    })

func createUser(c *gin.Context) {
    // Validate input using schema
    input, err := userSchema.BindAndValidateReader(c.Request.Body, "create")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Map validated data to GORM model
    user := User{}
    input.ToModel(&user)

    // Hash password (in real app, use proper hashing)
    user.Password = hashPassword(input.String("password"))

    // Save to database
    if err := db.Create(&user).Error; err != nil {
        c.JSON(500, gin.H{"error": "Failed to create user"})
        return
    }

    // Return created user with detailed view
    result := grape.Present(user, userPresenter, grape.H{"view": "detailed"})
    c.JSON(201, result)
}

func getUser(c *gin.Context) {
    userID := c.Param("id")

    // Get user from database
    var user User
    if err := db.First(&user, userID).Error; err != nil {
        c.JSON(404, gin.H{"error": "User not found"})
        return
    }

    // Present user data with basic view
    result := grape.Present(user, userPresenter)
    c.JSON(200, result)
}

func updateUser(c *gin.Context) {
    userID := c.Param("id")

    // Get existing user
    var user User
    if err := db.First(&user, userID).Error; err != nil {
        c.JSON(404, gin.H{"error": "User not found"})
        return
    }

    // Validate input (only provided fields)
    input, err := userSchema.BindAndValidateReader(c.Request.Body, "update")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Map only provided fields (preserves existing data)
    input.ToModel(&user)

    // Save changes
    if err := db.Save(&user).Error; err != nil {
        c.JSON(500, gin.H{"error": "Failed to update user"})
        return
    }

    // Return updated user with detailed view
    result := grape.Present(user, userPresenter, grape.H{"view": "detailed"})
    c.JSON(200, result)
}

func listUsers(c *gin.Context) {
    var users []User

    // Get all users
    if err := db.Find(&users).Error; err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch users"})
        return
    }

    // Present all users at once with basic view
    results := grape.PresentSlice(users, userPresenter)

    c.JSON(200, gin.H{"users": results})
}

func deleteUser(c *gin.Context) {
    userID := c.Param("id")

    // Delete user
    if err := db.Delete(&User{}, userID).Error; err != nil {
        c.JSON(500, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(200, gin.H{"message": "User deleted successfully"})
}
```


This example demonstrates the complete workflow:
- **Validation**: Input data is validated using schemas with different modes
- **Mapping**: Validated data is automatically mapped to GORM models
- **Database**: GORM handles database operations
- **Presentation**: Responses are formatted consistently using presenters
- **Framework Integration**: Works seamlessly with Gin web framework

The schema, presenter, and model definitions are reusable across all handlers, ensuring consistency and reducing code duplication.

## API Reference

### Parameter Validation

#### Params Creation

```go
schema := grape.NewParams()
```

#### Field Definition

```go
// Required fields (must be present based on mode)
schema.Requires("fieldName").String()

// Optional fields
schema.Optional("fieldName").Int()

// Field types
.String()      // string validation
.Integer()     // integer validation
.Float()       // float validation
.BigDecimal()  // big decimal validation (string or float)
.Numeric()     // numeric validation (float or string)
.Date()        // date string validation
.DateTime()    // datetime string validation
.Time()        // time string validation
.Boolean()     // boolean validation
.JSON()        // JSON object/array/string validation
.Slice()       // array validation

// Validation rules
.Validate("rule1,rule2")  // Uses go-playground/validator syntax

// JSON key mapping
.As("jsonKey")

// Required on specific modes
.On("create", "update")

// Default values
.DefaultValue("default")

// Descriptions and examples
.DescText("Field description")
.ExampleVal("example value")

// Nested schemas
.WithSchema(subSchema)

// Slice element types
.SliceOf(grape.Int, subSchema)  // For []int
.SliceOf(grape.H, subSchema)  // For []map
```

#### Validation Methods

```go
// Bind from map[string]interface{}
input, err := schema.BindAndValidate(data, "mode")

// Bind from io.Reader (HTTP request body)
input, err := schema.BindAndValidateReader(r.Body, "mode")
```

#### Data Access

```go
// Type-safe accessors with defaults
name := input.String("field")
count := input.Integer("count", 0)
price := input.Float("price", 0.0)
active := input.Boolean("active", false)
bigDecimal := input.BigDecimal("amount")
numeric := input.Numeric("value", 0.0)
date := input.Date("birthdate")
dateTime := input.DateTime("createdAt")
time := input.Time("scheduledTime")

// Direct map access
rawData := input["field"]
```

### Data Presentation

#### Presenter Creation

```go
presenter := grape.NewEntity()
```

#### Field Configuration

```go
presenter.Field("StructField").
    As("json_key").                                    // JSON key mapping
    WithSchema(subPresenter).                          // Nested presentation
    FieldFunc(func(obj any) any {      // Custom transformation
        return transform(obj.(MyType).Field)
    }).
    If(func(obj any, options grape.H) bool { // Conditional inclusion
        return obj.(MyType).ShouldInclude
    }).
    DefaultValue("default").                           // Default values
    DescText("Description").                           // Documentation
    ExampleVal("example")                              // Examples
```

#### Presentation

```go
result := grape.Present(myStruct, presenter)
// Returns map[string]interface{} ready for JSON encoding

// With options for conditional fields
result := grape.Present(myStruct, presenter, grape.H{"type": "full"})
```

#### Slice Presentation

```go
results := grape.PresentSlice(mySlice, presenter)
// Returns []map[string]interface{} where each element is a presented map for each item in the slice

// With options for conditional fields
results := grape.PresentSlice(mySlice, presenter, grape.H{"show_details": true})
```

#### Conditional Fields with Options

```go
presenter.Field("secretField").If(func(obj any, options grape.H) bool {
    return options["type"] == "full" // Only show for full type
})
```

## Framework Integration

Define your schemas at package level:

```go
package handlers

import "github.com/krivospitsky/grape"

var productSchema = grape.NewParams().
    Requires("name").On("create", "update").String().Validate("required,min=2").
    Optional("description").String().
    Requires("price").On("create", "update").Float().Validate("min=0").
    Optional("category").String().
    Optional("tags").SliceOf(grape.String, nil)
```

### Gin

```go
func createProduct(c *gin.Context) {
    input, err := productSchema.BindAndValidateReader(c.Request.Body, "create")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Create product...
    c.JSON(201, input)
}

func updateProduct(c *gin.Context) {
    input, err := productSchema.BindAndValidateReader(c.Request.Body, "update")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Update product...
    c.JSON(200, input)
}
```

### Echo

```go
func createProduct(c echo.Context) error {
    input, err := productSchema.BindAndValidateReader(c.Request().Body, "create")
    if err != nil {
        return c.JSON(400, map[string]string{"error": err.Error()})
    }

    // Create product...
    return c.JSON(201, input)
}
```

### Chi

```go
func createProduct(w http.ResponseWriter, r *http.Request) {
    input, err := productSchema.BindAndValidateReader(r.Body, "create")
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    // Create product...
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(input)
}
```

### Fiber

```go
func createProduct(c *fiber.Ctx) error {
    input, err := productSchema.BindAndValidateReader(c.Body(), "create")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // Create product...
    return c.JSON(input)
}
```

### Standalone Usage

```go
// Direct map validation
data := grape.H{
    "name": "John",
    "age": 30,
}

schema := grape.NewParams().
    Requires("name").String().
    Optional("age").Int()

input, err := schema.BindAndValidate(data, "create")
```

## Advanced Examples

### Complex Validation Schema

Define reusable schemas at package level:

```go
package schemas

import "github.com/krivospitsky/grape"

// Reusable profile schema
var profileSchema = grape.NewParams().
    Optional("bio").String().Validate("max=500").
    Optional("website").String().Validate("url").
    Optional("socialLinks").SliceOf(grape.String, nil)

// Reusable address schema
var addressSchema = grape.NewParams().
    Requires("street").String().
    Requires("city").String().
    Optional("country").String().DefaultValue("US").
    Optional("postalCode").String()

// Main user schema with different modes
var userSchema = grape.NewParams().
    Requires("name").On("create", "update").String().Validate("required,min=2,max=100").
    Requires("email").On("create").String().Validate("email"). // Required only for creation
    Optional("age").Int().Validate("min=0,max=150").
    Optional("profile").Map().WithSchema(profileSchema).
    Optional("addresses").On("update").SliceOf(grape.H, addressSchema) // Only updatable

// Complete registration schema
var registrationSchema = grape.NewParams().
    Requires("user").Map().WithSchema(userSchema).
    Requires("password").String().Validate("min=8").
    Optional("termsAccepted").Bool().
    Requires("preferences").Map().WithSchema(grape.NewParams().
        Optional("theme").String().DefaultValue("light").
        Optional("notifications").Bool().DefaultValue(true).
    )
```

Usage in handlers:

```go
package handlers

import (
    "myapp/schemas"
)

func registerUser(c *gin.Context) {
    input, err := schemas.registrationSchema.BindAndValidateReader(c.Request.Body, "create")
    // Handles all nested validation...
}

func updateUser(c *gin.Context) {
    input, err := schemas.userSchema.BindAndValidateReader(c.Request.Body, "update")
    // Uses same schema with different mode...
}
```

### Conditional Presentation

Define presenters with conditional logic at package level:

```go
package presenters

import "github.com/krivospitsky/grape"

type User struct {
    ID       int
    Name     string
    Email    string
    IsAdmin  bool
    Password string
    CreatedAt time.Time
}

// Single user presenter with conditional fields based on options
var userPresenter = grape.NewEntity().
    Field("ID").
    Field("Name").
    Field("Email").
    Field("Age").If(func(obj any, options grape.H) bool {
        return options["view"] == "detailed" // Show age in detailed view
    }).
    Field("IsAdmin").If(func(obj any, options grape.H) bool {
        return options["show_admin"] == true // Show admin status based on permissions
    }).
    Field("CreatedAt").FieldFunc(func(obj any) any {
        return obj.(User).CreatedAt.Format(time.RFC3339)
    })
```

Usage in handlers:

```go
func getUser(c *gin.Context) {
    user := getUserFromDB(c.Param("id"))
    currentUser := getCurrentUser(c)

    // Build options based on permissions
    options := grape.H{}
    if currentUser.IsAdmin {
        options["show_admin"] = true
        options["view"] = "detailed"
    }

    result := grape.Present(user, userPresenter, options)
    c.JSON(200, result)
}
```

### Nested Data Presentation

Define reusable presenters for complex data structures:

```go
package presenters

import "github.com/krivospitsky/grape"

type Order struct {
    ID       int
    User     User
    Items    []OrderItem
    Total    float64
    Status   string
}

type OrderItem struct {
    Product  Product
    Quantity int
    Price    float64
}

type Product struct {
    ID    int
    Name  string
    Price float64
}

// Reusable presenters
var productPresenter = grape.NewEntity().
    Field("ID").
    Field("Name").
    Field("Price")

var orderItemPresenter = grape.NewEntity().
    Field("Product").WithSchema(productPresenter).
    Field("Quantity").
    Field("Price").
    Field("Total").FieldFunc(func(obj any) any {
        item := obj.(OrderItem)
        return float64(item.Quantity) * item.Price
    })

var orderUserPresenter = grape.NewEntity().
    Field("ID").
    Field("Name")

var orderPresenter = grape.NewEntity().
    Field("ID").
    Field("User").WithSchema(orderUserPresenter).
    Field("Items").WithSchema(orderItemPresenter).
    Field("Total").
    Field("Status").
    Field("CreatedAt").If(func(obj any, options grape.H) bool {
        return options["view"] == "detailed" // Only show created date in detailed view
    }).FieldFunc(func(obj any) any {
        return obj.(Order).CreatedAt.Format(time.RFC3339)
    })
```

Usage across multiple handlers:

```go
func getOrder(c *gin.Context) {
    order := getOrderFromDB(c.Param("id"))
    currentUser := getCurrentUser(c)

    // Build options based on permissions
    options := grape.H{}
    isOwner := order.User.ID == currentUser.ID
    if isOwner || currentUser.IsAdmin {
        options["view"] = "detailed"
    }

    result := grape.Present(order, orderPresenter, options)
    c.JSON(200, result)
}

func listOrders(c *gin.Context) {
    orders := getOrdersFromDB()

    // Present all orders at once with basic view (no CreatedAt field)
    results := grape.PresentSlice(orders, orderPresenter)

    c.JSON(200, grape.H{"orders": results})
}
```

## Validation Rules

The library uses [go-playground/validator](https://github.com/go-playground/validator) for field validation. Common rules include:

- `required` - Field must be present
- `email` - Must be valid email format
- `min=5` - Minimum value/length
- `max=100` - Maximum value/length
- `len=10` - Exact length
- `url` - Must be valid URL
- `uuid` - Must be valid UUID
- `json` - Must be valid JSON

## Data Mapping

Map validated Input data directly to Go structs with automatic field name conversion and nil preservation.

### Basic Usage

```go
package main

import (
    "github.com/krivospitsky/grape"
)

// GORM model struct (PascalCase fields)
type User struct {
    ID        uint   `gorm:"primaryKey"`
    FirstName string `gorm:"column:first_name"`
    LastName  string `gorm:"column:last_name"`
    Email     string
    Age       *int   // Nullable field
}

func createUser(input grape.Input) {
    user := User{}

    // Map Input directly to struct - no separate mapper needed!
    input.ToModel(&user)

    // user.FirstName, user.LastName, user.Email are now populated
    // Age remains nil if input["age"] is not provided
}
```

### Field Name Mapping

The ToModel method automatically converts snake_case JSON keys to PascalCase struct fields:

```go
// Input data with snake_case keys
data := grape.H{
    "first_name": "John",
    "last_name":  "Doe",
    "email":      "john@example.com",
    "is_active":  true,
    "user_id":    123,
}

// Struct with PascalCase fields
type User struct {
    ID        uint
    FirstName string
    LastName  string
    Email     string
    IsActive  bool
    UserID    int  // Conflicting names resolved automatically
}

// Map automatically - no configuration needed!
input := grape.Input(data)
input.ToModel(&user)
```

### How It Works

- **Automatic Conversion**: Converts `snake_case` JSON keys to `PascalCase` struct fields
- **Nil Preservation**: `nil` values in input don't overwrite existing struct values
- **Type Safety**: Uses reflection to safely assign compatible types
- **Case Insensitive**: Matches fields regardless of naming convention differences

### GORM Integration

```go
// Complete example with GORM
func createUserHandler(c *gin.Context) error {
    // 1. Validate input
    input, err := userSchema.BindAndValidateReader(c.Request.Body, "create")
    if err != nil {
        return c.JSON(400, gin.H{"error": err.Error()})
    }

    // 2. Map to GORM model - simple one-liner!
    user := User{}
    input.ToModel(&user)

    // 3. Save to database
    if err := db.Create(&user).Error; err != nil {
        return c.JSON(500, gin.H{"error": "Failed to create user"})
    }

    // 4. Present response with detailed view
    result := grape.Present(user, userPresenter, grape.H{"view": "detailed"})
    return c.JSON(201, result)
}

func updateUserHandler(c *gin.Context) error {
    userID := c.Param("id")

    // Get existing user
    var user User
    if err := db.First(&user, userID).Error; err != nil {
        return c.JSON(404, gin.H{"error": "User not found"})
    }

    // Validate input
    input, err := userSchema.BindAndValidateReader(c.Request.Body, "update")
    if err != nil {
        return c.JSON(400, gin.H{"error": err.Error()})
    }

    // Map only provided fields (preserves existing data)
    input.ToModel(&user)

    // Save changes
    if err := db.Save(&user).Error; err != nil {
        return c.JSON(500, gin.H{"error": "Failed to update user"})
    }

    result := grape.Present(user, userPresenter, grape.H{"view": "detailed"})
    return c.JSON(200, result)
}
```

## Error Handling

The library provides detailed error messages:

```go
input, err := schema.BindAndValidate(data, "create")
if err != nil {
    // Error messages include field name and validation failure reason
    // "field 'email' validation failed: Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag"
    // "missing required field 'name' for create"
    // "field 'age' must be int"
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Related Projects

- [ruby-grape/grape](https://github.com/ruby-grape/grape) - Ruby on Rails Grape gem
- [go-playground/validator](https://github.com/go-playground/validator) - Validation library used internally
- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Popular HTTP framework
- [labstack/echo](https://github.com/labstack/echo) - High performance HTTP framework
- [go-chi/chi](https://github.com/go-chi/chi) - Lightweight HTTP router
- [gofiber/fiber](https://github.com/gofiber/fiber) - Fast HTTP engine
- [gorm.io/gorm](https://github.com/go-gorm/gorm) - ORM library