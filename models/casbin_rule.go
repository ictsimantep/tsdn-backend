package models

type CasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100;not null"`
	V0    string `gorm:"size:100;not null"`
	V1    string `gorm:"size:100;not null"`
	V2    string `gorm:"size:100"`
	V3    string `gorm:"size:100;default:'none'"`
	V4    string `gorm:"size:100;default:'none'"`
	V5    string `gorm:"size:100;default:'none'"`
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}
