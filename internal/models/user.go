package models

import (
	"net/url"
	"strings"
)

type User struct {
	Id        uint   `json:id`
	FirstName string `json:first_name`
	LastName  string `json:last_name`
	Email     string `json:email`
	Password  string `json:password`
	CreatedAt string `json:created_at`
	UpdatedAt string `json:created_at`
}

/**
* I map the form fields to the struct fields
 */
func (u *User) MapFormToStruct(form url.Values) {
	u.FirstName = form.Get("first_name")
	u.LastName = form.Get("last_name")
	u.Password = form.Get("password")
	u.Email = form.Get("email")
}

/**
* I validate that all the required fields are set
 */
func (rb *User) Validate() []FieldValidationError {

	errors := make([]FieldValidationError, 0, 5)

	if strings.TrimSpace(rb.FirstName) == "" {
		errors = append(errors, FieldValidationError{Field: "FirstName", Message: "Please provide a first name"})
	}

	if strings.TrimSpace(rb.LastName) == "" {
		errors = append(errors, FieldValidationError{Field: "LastName", Message: "Please provide your last name"})
	}

	if strings.TrimSpace(rb.Password) == "" {
		errors = append(errors, FieldValidationError{Field: "Password", Message: "Password is a necessary field. Please provide it."})
	}

	if strings.TrimSpace(rb.Email) == "" {
		errors = append(errors, FieldValidationError{Field: "Email", Message: "The prompt used to created this batch must be specified"})
	}

	return errors
}

func (rb *User) Save() {

}

func (rb *User) Update() {

}

func (rb *User) Delete() {

}
