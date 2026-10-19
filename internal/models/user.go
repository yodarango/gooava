package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gooava/internal/lib"
	"gooava/internal/utils"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
)
type User struct {
	VerificationCode string `json:"verification_code"` // has the user verified their account 
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Status Status `json:"status"`
	TwoFAEnabled bool `json:"two_fa_enabled"`
	Avatar string `json:"avatar"`
	Email string `json:"email"`
	Id int `json:"id"`
}

// Status represents the status of a user as an enum type
type Status string

const (
	StatusPending Status = "pending"
	StatusActive  Status = "active"
	StatusDeleted Status = "deleted"
)


func (u *User) RequestToStruct(body io.ReadCloser) (error){

	bodyBuffer, err := io.ReadAll(body)

	if err != nil {
		return fmt.Errorf("error decoding bsingle word ody data %w", err)
	}

	body.Close()

	err = json.Unmarshal(bodyBuffer, &u)

	if err != nil {
		return fmt.Errorf("error unmarshalling bsingle word ody data %w", err)
	}

	return nil
}

/**************************************************************************************
* Creates a new user and initially it will set it up as pending until their email is 
* verified.
* 
* status: ✅
****************************************************************************************/
func (u *User) Create() (error){
	var newStudySession = `
		INSERT INTO users(
		first_name,
		last_name,
		username,
		email
		) 
		values(?,?,?,?)
	`
	result, err := ModelsRepo.DB.Conn.Exec(newStudySession, 
		u.FirstName, 
		u.LastName,
		u.Username,
		u.Email, 
	)

	if err != nil {
		return fmt.Errorf("could not create new user. %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return fmt.Errorf("could not get last user id. %w", err)
	}

	u.Id = int(id)

	return nil
}


/**************************************************************************************
* Updates the user's profile: first name, last name, username, email and avatar.
* Username and email must remain unique across OTHER accounts.
*
* status: ✅
****************************************************************************************/
func (u *User) Update(userId uint) (error){
	// basic validation — the table has no CHECK constraints for these
	if strings.TrimSpace(u.FirstName) == "" || strings.TrimSpace(u.LastName) == "" {
		return fmt.Errorf("first name and last name are required")
	}

	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if strings.TrimSpace(u.Email) == "" || !strings.Contains(u.Email, "@") {
		return fmt.Errorf("a valid email address is required")
	}

	// username and email must be unique among other users
	queryConflict := `
		SELECT id
		FROM users
		WHERE (username = ? OR email = ?) AND id != ?
	`

	rows, err := ModelsRepo.DB.Conn.Query(queryConflict, u.Username, u.Email, userId)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("could not check for conflicting accounts \n %w", err)
	}

	conflictingIds := make([]int, 0)

	for rows.Next(){
		var id int

		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("error scanning conflict rows \n %w", err)
		}

		conflictingIds = append(conflictingIds, id)
	}
	rows.Close()

	if len(conflictingIds) > 0 {
		return fmt.Errorf("this username or email is already linked to another account")
	}

	query := `
		UPDATE users SET
		first_name = ?,
		last_name = ?,
		username = ?,
		email = ?,
		avatar = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(query,
		u.FirstName,
		u.LastName,
		u.Username,
		u.Email,
		u.Avatar,
		userId,
	)

	if err != nil {
		return fmt.Errorf("could not update this user \n %w", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("could not get rows affected: \n %w", err)
	}

	return nil
}

/**************************************************************************************
* Updates the password of a user, and in order to change it, they will need to enter
* the current password. This function is used for both forgotten passwords and 
* changing passwords. If a user forget thier password an email will be sent to them
* 
* status: ✅
****************************************************************************************/
func (u *User) UpdatePassword() (error) {
	query := `
		UPDATE users SET password = ? WHERE id = ?
	`

	userById := `
		SELECT password
		FROM users 
		WHERE id = ?
	`

	row := ModelsRepo.DB.Conn.QueryRow(userById, u.Id)

	var password string

	err := row.Scan(&password)

	if err != nil {
		return fmt.Errorf("Erro scaning user row \n %w", err)
	}

	// check if the current password matches the new one 

	result, err := ModelsRepo.DB.Conn.Exec(query, 
		u.Password,
		u.Id,
	)

	if err != nil {
		return fmt.Errorf("could not update password, please try again \n %w", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("could not get rows affected: \n %w", err)
	}

	return nil
}

/**************************************************************************************
* Checks if there is a user with the same username or email before creating a new one.
* Initially, the user will be set with status = pending. Then sends an email with 
* a unique 6 digit verification code.
* 
* status: ✅
****************************************************************************************/
func (u *User) Signup() (*string, error){
	
	var authUser AuthUser

	query := `SELECT id FROM users WHERE email = ? OR username = ?;`

	rows, err := ModelsRepo.DB.Conn.Query(query, &u.Email, &u.Username)


	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("Erro creting account please try again %w", err)
	}

	existsingUserIds := make([]int, 0)

	for rows.Next(){
		var userId int
		rows.Scan(&userId)

		existsingUserIds = append(existsingUserIds, userId)
	}

	if len(existsingUserIds) > 0 {
		return nil, fmt.Errorf("there is a user with this credentials already, please login instead.")
	}

	hashPasswrd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("could not hash password \n %w", err)
	}

	// Generate verification code for new user
	verificationCode := utils.GenerateRandomString(6)

	queryCreateUser := `
		INSERT INTO users(first_name, last_name, username, email, avatar, password, verification_code, status)
		VALUES(?,?,?,?,?,?,?,?)
	`

	result, err := ModelsRepo.DB.Conn.Exec(queryCreateUser, &u.FirstName, &u.LastName, &u.Username, &u.Email, &u.Avatar, &hashPasswrd, verificationCode, StatusPending)

	if err != nil {
		return nil, fmt.Errorf("could insert result. \n %w", err)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("could not get the record for the last saved user.")
	}

	u.Id = int(lastInsertedId)

	// Send verification email
	emailService := lib.NewEmailService()
	err = emailService.SendVerificationEmail(u.Email, u.FirstName, verificationCode)
	if err != nil {
		return nil, fmt.Errorf("could not send verification email: %w", err)
	}

	authUser.Status = StatusPending
	authUser.FirstName = u.FirstName
	authUser.LastName	=u.LastName
	authUser.Username = u.Username
	authUser.Avatar = u.Avatar
	authUser.Email = u.Email
	authUser.Id = uint(u.Id)

	// set the user status too
	u.Status = StatusPending

	token, err := authUser.GenerateJWT()

	return &token, nil
}


/**************************************************************************************
* Very simple function that just checks the credentials of the user before logging them 
* in.
* 
* status: ✅
****************************************************************************************/
func (u *User) Login() (*string, bool, error){

	var authUser AuthUser
	var storedPassword string

	query := `
		SELECT id, first_name, last_name, username, email, avatar, password, status, verification_code, two_fa_enabled
		FROM users
		WHERE username = ?
	`

	row := ModelsRepo.DB.Conn.QueryRow(query, u.Username)

	err := row.Scan(&u.Id, &u.FirstName, &u.LastName, &u.Username, &u.Email, &u.Avatar, &storedPassword, &u.Status, &u.VerificationCode, &u.TwoFAEnabled)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, fmt.Errorf("no user found with these credentials, please check your email/username and try again")
		}
		return nil, false, fmt.Errorf("error querying for user with these credentials \n %w", err)
	}

	// verify the password
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(u.Password))

	if err != nil {
		return nil, false, fmt.Errorf("invalid password, please try again")
	}

	// 2FA: the password alone is not enough — email a 6-digit code (stored in
	// the same verification_code column as the signup verification) and hold
	// the token until the code is confirmed via the verify-2fa endpoint
	if u.TwoFAEnabled {
		code := utils.GenerateRandomString(6)

		_, err := ModelsRepo.DB.Conn.Exec("UPDATE users SET verification_code = ? WHERE id = ?", code, u.Id)
		if err != nil {
			return nil, false, fmt.Errorf("could not save your sign-in code \n %w", err)
		}

		emailService := lib.NewEmailService()
		if err := emailService.SendTwoFACodeEmail(u.Email, u.FirstName, code); err != nil {
			return nil, false, fmt.Errorf("could not email your sign-in code: %w", err)
		}

		return nil, true, nil
	}

	// create auth user for JWT generation
	authUser.FirstName = u.FirstName
	authUser.LastName = u.LastName
	authUser.Username = u.Username
	authUser.Status = u.Status
	authUser.Avatar = u.Avatar
	authUser.Email = u.Email
	authUser.TwoFAEnabled = u.TwoFAEnabled
	authUser.Id = uint(u.Id)

	token, err := authUser.GenerateJWT()

	if err != nil {
		return nil, false, fmt.Errorf("could not generate authentication token \n %w", err)
	}

	return &token, false, nil
}

/**************************************************************************************
* Enables or disables two-factor authentication for the user.
*
* status: ✅
****************************************************************************************/
func (u *User) SetTwoFA(userId uint, enabled bool) error {
	query := `
		UPDATE users
		SET two_fa_enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(query, enabled, userId)
	if err != nil {
		return fmt.Errorf("could not update the two-factor setting \n %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify the two-factor update: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found to update")
	}

	return nil
}

/**************************************************************************************
* Confirms the emailed 6-digit 2FA code for a login attempt. The code lives in the
* same verification_code column used by signup verification, scoped here by username
* so a code alone can never be guessed against every account. Codes are single-use:
* cleared on success and overwritten on every new login attempt.
*
* status: ✅
****************************************************************************************/
func (u *User) VerifyTwoFA(username, code string) (*string, error) {
	query := `
		SELECT id, first_name, last_name, username, email, avatar, status, two_fa_enabled
		FROM users
		WHERE username = ? AND verification_code = ? AND verification_code != ""
	`

	row := ModelsRepo.DB.Conn.QueryRow(query, username, code)

	err := row.Scan(&u.Id, &u.FirstName, &u.LastName, &u.Username, &u.Email, &u.Avatar, &u.Status, &u.TwoFAEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired code, please try again")
		}
		return nil, fmt.Errorf("error verifying code: %w", err)
	}

	if !u.TwoFAEnabled {
		return nil, fmt.Errorf("two-factor authentication is not enabled for this account")
	}

	// burn the code — each one is single-use
	_, err = ModelsRepo.DB.Conn.Exec(`UPDATE users SET verification_code = "" WHERE id = ?`, u.Id)
	if err != nil {
		return nil, fmt.Errorf("could not clear the used code: %w", err)
	}

	authUser := AuthUser{
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     u.Username,
		Email:        u.Email,
		Avatar:       u.Avatar,
		Status:       u.Status,
		TwoFAEnabled: u.TwoFAEnabled,
		Id:           uint(u.Id),
	}

	token, err := authUser.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("could not generate authentication token \n %w", err)
	}

	return &token, nil
}

/**************************************************************************************
* It verifies the user's email if the code sent is correct. If the emil is verified,
* the user will be set to status = active and a new auth token will be generated.
* 
*
* status: ✅
***************************************************************************************/
func (u *User) VerifyEmail(code string) (*string, error) {
	// Find user by verification code
	query := `
		SELECT id, first_name, last_name, avatar,username, email, status, verification_code
		FROM users
		WHERE verification_code = ?
	`

	row := ModelsRepo.DB.Conn.QueryRow(query, code)

	err := row.Scan(&u.Id, &u.FirstName, &u.LastName, &u.Avatar, &u.Username, &u.Email, &u.Status, &u.VerificationCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired verification code")
		}
		return nil, fmt.Errorf("error verifying code: %w", err)
	}

	// Update user status to active and clear verification code
	updateQuery := `
		UPDATE users
		SET status = ?, verification_code = ""
		WHERE id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(updateQuery, StatusActive, u.Id)
	if err != nil {
		return nil, fmt.Errorf("could not activate user account: %w", err)
	}

	rowsUpdated, err := result.RowsAffected()

	if err != nil {
		return nil, fmt.Errorf("could not update this user %w", err)
	}

	if !(rowsUpdated > 0) {
		return nil, fmt.Errorf("could not update this user")
	}

	// Send welcome email
	emailService := lib.NewEmailService()
	err = emailService.SendWelcomeEmail(u.Email, u.FirstName)
	if err != nil {
		fmt.Printf("Warning: Could not send welcome email: %v\n", err)
	}

	var authUser AuthUser

	// create auth user for JWT generation
	authUser.FirstName = u.FirstName
	authUser.LastName = u.LastName
	authUser.Username = u.Username
	authUser.Status = StatusActive
	authUser.Avatar = u.Avatar
	authUser.Email = u.Email
	authUser.Id = uint(u.Id)

	token, err := authUser.GenerateJWT()
	if err != nil {
		fmt.Printf("I could not generate the auth token for this user: %v\n", err)
	}
	// Update user status
	u.Status = StatusActive
	u.VerificationCode = ""

	return &token, nil
}

/************************************************************************
* Handles forgot password functionality by sending a temporary password
* to the user's email address. The user can then use this temporary 
* password to log in and change their password.
*
* status: ✅
************************************************************************/

func (u *User) ForgotPassword(body io.ReadCloser) error {

	type ForgotPasswordEmail struct {
		Email string `json:"email"`
	}

	var requestBody ForgotPasswordEmail

	err := json.NewDecoder(body).Decode(&requestBody)
	if err != nil {
		return fmt.Errorf("invalid request format: %w", err)
	}

	fmt.Println(requestBody.Email)
	if len(requestBody.Email) == 0 {
		return fmt.Errorf("email is required")
	}

	// Find user by email
	query := `
		SELECT id, first_name
		FROM users
		WHERE email = ?
	`

	var userId int
	var firstName string
	row := ModelsRepo.DB.Conn.QueryRow(query, requestBody.Email)
	err = row.Scan(&userId, &firstName)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no user found with this email address")
		}
		return fmt.Errorf("error querying for user: %w", err)
	}

	// Generate temporary password
	temporaryPassword := utils.GenerateRandomString(8)

	// Hash the temporary password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not hash temporary password: %w", err)
	}

	// Update user password in database
	updateQuery := `
		UPDATE users
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(updateQuery, hashedPassword, userId)
	if err != nil {
		return fmt.Errorf("could not update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify password update: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password was not updated")
	}

	// Send email with temporary password
	emailService := lib.NewEmailService()
	err = emailService.SendPasswordResetEmail(requestBody.Email, firstName, temporaryPassword)
	if err != nil {
		return fmt.Errorf("could not send password reset email: %w", err)
	}

	return nil
}

/**************************************************************************************
* Changes the user's password. Requires the current password for verification.
* Used when a user wants to change their password from the settings.
* 
* status: ✅
****************************************************************************************/
func (u *User) ChangePassword(body io.ReadCloser, userId uint) error {
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	var requestBody ChangePasswordRequest

	err := json.NewDecoder(body).Decode(&requestBody)
	if err != nil {
		return fmt.Errorf("invalid request format: %w", err)
	}

	if len(requestBody.CurrentPassword) == 0 {
		return fmt.Errorf("current password is required")
	}

	if len(requestBody.NewPassword) == 0 {
		return fmt.Errorf("new password is required")
	}

	if len(requestBody.NewPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters long")
	}

	// Get current password from database
	query := `
		SELECT password
		FROM users 
		WHERE id = ?
	`

	var storedPassword string
	row := ModelsRepo.DB.Conn.QueryRow(query, userId)
	err = row.Scan(&storedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("error retrieving user password: %w", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(requestBody.CurrentPassword))
	if err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Check if new password is different from current
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(requestBody.NewPassword))
	if err == nil {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not hash new password: %w", err)
	}

	// Update password in database
	updateQuery := `
		UPDATE users
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(updateQuery, hashedNewPassword, userId)
	if err != nil {
		return fmt.Errorf("could not update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify password update: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password was not updated")
	}

	return nil
}
