package models

import (
	"net/url"
	"strings"
)

type User struct {
	Id        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
func (rb *User) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if strings.TrimSpace(rb.FirstName) == "" {
		errors = append(errors, map[string]interface{}{"Field": "FirstName", "Message": "Please provide a first name"})
	}

	if strings.TrimSpace(rb.LastName) == "" {
		errors = append(errors, map[string]interface{}{"Field": "LastName", "Message": "Please provide your last name"})
	}

	if strings.TrimSpace(rb.Password) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Password", "Message": "Password is a necessary field. Please provide it."})
	}

	if strings.TrimSpace(rb.Email) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Email", "Message": "Email is a necessary field. Please provide it."})
	}

	return errors
}

func (rb *User) Save() {

}

func (rb *User) Update() {

}

func (rb *User) Delete() {

}
