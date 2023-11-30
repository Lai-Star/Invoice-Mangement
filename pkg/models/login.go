package models

type Login struct {
	tableName string `pg:"logins"`

	LoginId         uint64       `json:"loginId" pg:"login_id,notnull,pk,type:'bigserial'"`
	Email           string       `json:"email" pg:"email,notnull,unique"`
	PasswordHash    string       `json:"-" pg:"password_hash,notnull"`
	PhoneNumber     *PhoneNumber `json:"-" pg:"phone_number,type:'text'"`
	IsEmailVerified bool         `json:"isEmailVerified" pg:"is_email_verified,notnull,use_zero"`
	IsPhoneVerified bool         `json:"isPhoneVerified" pg:"is_phone_verified,notnull,use_zero"`

	Users              []User              `json:"users,omitempty" pg:"rel:has-many"`
	EmailVerifications []EmailVerification `json:"emailVerifications,omitempty" pg:"rel:has-many"`
}
