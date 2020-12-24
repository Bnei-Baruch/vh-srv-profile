package models

import (
	"log"
	"time"

	"gitlab.bbdev.team/vh/vh-srv-profile/app"
)

//User
type User struct {
	tableName struct{}  `sql:"users"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Active    bool      `json:"active" `
	Roles     []int     `json:"roles" pg:",array"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Country   int       `json:"country"`
	Language  int       `json:"language"`
	BirthDate time.Time `json:"birthdate"`
	Gender    int       `json:"gender"`
	Token     string    `json:"token"`
}

//PasswordData password data for update
type PasswordData struct {
	CurrentPassword      string `json:"current_password" form:"current_password" `
	Password             string `json:"password" form:"password" `
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation"  `
}

//InsertUser insert new user
func InsertUser(data *User) error {

	data.Created = time.Now()
	data.Updated = data.Created

	_, err := DB.Model(data).Insert()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return err
}

func UpdateUser(data *User) error {

	data.Updated = time.Now()

	_, err := DB.Model(data).
		Where("id = ?", data.ID).
		Update()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return err
}

//ValidateUser validate user data
func ValidateUser(data *User, checkPassword bool) (errors []string, result bool) {

	result = true

	if !app.IsEmailValid(data.Email) {
		result = false
		errors = append(errors, app.GetSystemMessage("email_not_valid"))
	}

	//check email
	count, err := DB.Model(data).Where("email = ?", data.Email).Count()

	if err != nil {
		result = false
		errors = append(errors, err.Error())
		log.Println("Error: " + err.Error())
	}

	if count > 0 {
		result = false
		errors = append(errors, app.GetSystemMessage("email_exists"))

	}

	//check phone
	count, err = DB.Model(data).Where("phone = ?", data.Phone).Count()
	if err != nil {
		result = false
		errors = append(errors, err.Error())
		log.Println("Error: " + err.Error())
	}
	if count > 0 {
		result = false
		errors = append(errors, app.GetSystemMessage("phone_exists"))

	}
	if checkPassword {
		if len(data.Password) < 6 {
			result = false
			errors = append(errors, app.GetSystemMessage("password_min_length_error"))
		}

		if len(data.Password) > 32 {
			result = false
			errors = append(errors, app.GetSystemMessage("password_max_length_error"))
		}
	}

	return
}

func ValidateUpdateUser(data *User, passwordData PasswordData) (errors []string, result bool) {

	result = true

	if !app.IsEmailValid(data.Email) {
		result = false
		errors = append(errors, app.GetSystemMessage("email_not_valid"))
	}

	//check email
	count, err := DB.Model(data).Where("email = ?", data.Email).Where("id != ?", data.ID).Count()

	if err != nil {
		result = false
		errors = append(errors, err.Error())
		log.Println("Error: " + err.Error())
	}

	if count > 0 {
		result = false
		errors = append(errors, app.GetSystemMessage("email_exists"))

	}

	//check phone
	count, err = DB.Model(data).Where("phone = ?", data.Phone).Where("id != ?", data.ID).Count()
	if err != nil {
		result = false
		errors = append(errors, err.Error())
		log.Println("Error: " + err.Error())
	}
	if count > 0 {
		result = false
		errors = append(errors, app.GetSystemMessage("phone_exists"))

	}

	if passwordData.Password != "" {

		//verify current password

		user, _ := FindUserByID(data.ID)

		if !app.VerifyPassword(user.Password, passwordData.CurrentPassword) {
			result = false
			errors = append(errors, app.GetSystemMessage("wrong_current_password"))
		}

		if len(passwordData.Password) < 6 {
			result = false
			errors = append(errors, app.GetSystemMessage("password_min_length_error"))
		}

		if len(passwordData.Password) > 32 {
			result = false
			errors = append(errors, app.GetSystemMessage("password_max_length_error"))
		}

		if passwordData.Password != passwordData.PasswordConfirmation {
			result = false
			errors = append(errors, app.GetSystemMessage("wrong_password_confirmation"))
		}
	}

	return
}

func FindUserByEmail(email string) (user *User, err error) {

	user = new(User)

	err = DB.Model(user).WhereOr("email = ?", email).First()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return
}

func FindUserByID(id uint64) (user *User, err error) {

	user = new(User)

	err = DB.Model(user).Where("id = ?", id).Select()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return
}

func FindUserByToken(token string) (user *User, err error) {

	user = new(User)

	err = DB.Model(user).Where("token = ?", token).Select()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return
}

func SetPassword(password string, userID uint64) error {

	newPassword, _ := app.HashPassword(password)

	_, err := DB.Model((*User)(nil)).Where("id = ?", userID).Set("password = ?", newPassword).Update()

	return err

}

func SetToken(token string, userID uint64) error {

	_, err := DB.Model((*User)(nil)).Where("id = ?", userID).Set("token = ?", token).Update()

	return err

}
