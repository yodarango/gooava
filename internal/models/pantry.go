package models

type Pantry struct {
	Id        uint   `json:id`
	UserId    uint   `json:user_id`
	CreatedAt string `json:created_at`
	Name      string `json:name`
}

/**
* I validate that all the required fields are set
 */

func (ri *Pantry) Validate() []FieldValidationError {

	errors := make([]FieldValidationError, 0, 3)

	if !(ri.UserId > 0) {
		errors = append(errors, FieldValidationError{Field: "PantryId", Message: "Pantry items need to be associated with a user"})
	}

	return errors
}
