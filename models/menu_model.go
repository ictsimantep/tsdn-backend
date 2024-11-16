package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Menu struct {
	gorm.Model
	Slug              string     `json:"slug"`
	MenuTitle         string     `json:"menuTitle"`
	MenuType          string     `json:"menuType"`
	HasMegamenu       bool       `json:"hasMegamenu"`
	HasMenuChild      bool       `json:"hasMenuChild"`
	HasPositionStatic bool       `json:"hasPositionStatic"`
	HasSideBanner     bool       `json:"hasSideBanner"`
	BannerImage       string     `json:"bannerImage"`
	MenuItems         []MenuItem `json:"menuItems" gorm:"foreignKey:MenuID"`
	UUID              uuid.UUID  `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

type MenuItem struct {
	gorm.Model
	MenuID       uint       `json:"menuID"`
	Title        string     `json:"title"`
	Link         string     `json:"link"`
	Slug         string     `json:"slug"`
	HasMenuChild bool       `json:"hasSubmenu"`
	ParentID     *uint      `json:"parentID"`
	SubMenuItems []MenuItem `json:"submenuItems,omitempty" gorm:"-"`
	UUID         uuid.UUID  `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

type MenuAdmin struct {
	gorm.Model
	MenuTitle         string     `json:"menu_title"`
	MenuType          string     `json:"menu_type"`
	HasMegamenu       bool       `json:"has_megamenu"`
	HasMenuChild      bool       `json:"has_menu_child"`
	HasPositionStatic bool       `json:"has_position_static"`
	HasSideBanner     bool       `json:"has_side_banner"`
	BannerImage       string     `json:"banner_imange"`
	MenuItems         []MenuItem `json:"menu_items" gorm:"foreignKey:MenuID"`
	UUID              uuid.UUID  `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

func (MenuAdmin) TableName() string {
	return "menus"
}

type MenuItemAdmin struct {
	gorm.Model
	MenuID       uint      `json:"menu_id"`
	Title        string    `json:"title"`
	Link         string    `json:"link"`
	Slug         string    `json:"slug"`
	HasMenuChild bool      `json:"has_sub_menu"`
	ParentID     *uint     `json:"parent_id"`
	UUID         uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid()"`
}

func (MenuItemAdmin) TableName() string {
	return "menu_items"
}
