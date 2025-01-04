package models

type Pantry struct {
	*ModelConfiguration
	Id        uint   `json:"id"`
	UserId    uint   `json:"user_id"`
	CreatedAt string `json:"created_at"`
	Name      string `json:"name"`
}

/**
* I validate that all the required fields are set
 */
func (ri *Pantry) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if !(ri.UserId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "UserId", "Message": "Pantry items need to be associated with a user"})
	}

	return errors
}
