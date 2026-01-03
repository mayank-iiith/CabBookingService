package models

type Account struct {
	BaseModel
	Username string `gorm:"size:255;not null;unique"`
	Password string `gorm:"not null"`

	// Many-to-Many relationship with Role
	Roles []Role `gorm:"many2many:account_roles"`
}

func (*Account) TableName() string {
	return "accounts"
}
