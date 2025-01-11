# jsonmask
[![Build Status](https://github.com/axkit/bitset/actions/workflows/go.yml/badge.svg)](https://github.com/axkit/jsonmask/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/axkit/jsonmask)](https://goreportcard.com/report/github.com/axkit/jsonmask)
[![GoDoc](https://pkg.go.dev/badge/github.com/axkit/jsonmask)](https://pkg.go.dev/github.com/axkit/jsonmask)
[![Coverage Status](https://coveralls.io/repos/github/axkit/jsonmask/badge.svg?branch=main)](https://coveralls.io/github/axkit/jsonmask?branch=main)

The `jsonmask` package is designed to mask sensitive data in JSON payloads. It allows developers to define masking rules on struct fields using tags, or programmatically, to anonymize or transform data as needed.


## Features

- **Customizable Masking Functions**: Transform values with predefined or custom masking functions.
- **Field-Level Control**: Use struct tags to specify masking rules for individual fields.
- **Nested Data Support**: Apply masking to nested objects, arrays, and slices.
- **Built-in Masking Functions**:
  - Convert strings to uppercase or lowercase.
  - Truncate strings or replace them with empty values.
  - Mask email addresses.
  - Set numeric values to zero or nullify fields.

## Installation

```bash
go get github.com/axkit/jsonmask
```

## Usage

### 1. Define Your Struct

Annotate your struct fields with the `mask` tag to specify masking rules. Use `-` to exclude fields from output entirely.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/axkit/jsonmask"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName" mask:"initialChar"`
	LastName  string `json:"lastName" mask:"first4"`
	Email     string `json:"email" mask:"email"`
	Password  string `json:"password" mask:"-"`
}

func main() {
	user := User{
		ID:        1,
		FirstName: "Robert",
		LastName:  "Egorov",
		Email:     "robert.egorov@example.com",
		Password:  "supersecret",
	}

	jm := jsonmask.New()
	rules := jm.ParseStruct(user)
	jsonData, _ := json.Marshal(user)
	maskedData, _ := jm.Mask(jsonData, rules)

	fmt.Println(string(maskedData))
}
```

**Output:**
```json
{
	"id": 1,
	"firstName": "R",
	"lastName": "Egor...",
	"email": "r***********v@e********.com"
}
```

### 2. Add Custom Masking Functions

Extend `jsonmask` with your own masking logic by registering custom functions.

```go
jm.AddFunc("customMask", func(s string) []byte {
	return []byte(`"custom_value"`)
})
```

Then, use `customMask` in your struct tags or rules.

### 3. Use with Arrays and Nested Structures

`jsonmask` supports arrays, slices, and nested structures.

```go
type Nested struct {
	Key string `json:"key" mask:"upper"`
}

type Parent struct {
	Items []Nested `json:"items"`
}

parent := Parent{
	Items: []Nested{
		{Key: "value1"},
		{Key: "value2"},
	},
}

jsonData, _ := json.Marshal(parent)
rules := jm.ParseStruct(parent)
maskedData, _ := jm.Mask(jsonData, rules)
fmt.Println(string(maskedData))
```

**Output:**
```json
{
	"items": [
		{"key": "VALUE1"},
		{"key": "VALUE2"}
	]
}
```


### 4. Practical Example: Masking Sensitive Data in Logs

Suppose you have a web server that responds to client requests with JSON payloads. For debugging purposes, each response is logged. To prevent sensitive data from leaking into the logs, you can use the `jsonmask` package to mask values in sensitive fields before logging.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/axkit/jsonmask"
	"net/http"
)

type ApiResponse struct {
	UserID    int    `json:"userId"`
	UserName  string `json:"userName" mask:"initialChar"`
	Email     string `json:"email" mask:"email"`
	SecretKey string `json:"secretKey" mask:"-"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	response := ApiResponse{
		UserID:    42,
		UserName:  "JohnDoe",
		Email:     "johndoe@example.com",
		SecretKey: "supersecretkey",
	}

	jm := jsonmask.New()
	rules := jm.ParseStruct(response)
	jsonData, _ := json.Marshal(response)
	maskedData, _ := jm.Mask(jsonData, rules)

	// Log the masked response
	fmt.Println("Masked Response for Logs:", string(maskedData))

	// Send the original response to the client
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	http.HandleFunc("/api", handler)
	http.ListenAndServe(":8080", nil)
}
```

**Output in Logs:**
```json
{
	"userId": 42,
	"userName": "J",
	"email": "j***e@e********.com"
}
```

## Predefined Masking Functions

- **`upper`**: Converts strings to uppercase.
- **`lower`**: Converts strings to lowercase.
- **`initialChar`**: Extracts the first character in uppercase.
- **`truncate`**: Replaces non-null strings with an empty string.
- **`null`**: Sets the field to `null`.
- **`email`**: Masks email addresses by anonymizing the local and domain parts.
- **`zero`**: Sets numeric fields to `0`.

## Testing

Run the provided tests to ensure the package works as expected.

```bash
go test ./...
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

