package models

type Role struct {
	BaseModel
	Name        string `gorm:"size:100;not null;unique"`
	Description string `gorm:"size:255;"`
}

func (*Role) TableName() string {
	return "roles"
}
