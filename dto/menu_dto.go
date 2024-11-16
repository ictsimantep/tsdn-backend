package dto

type MenuItemDTO struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Slug  string `json:"slug,omitempty"`
}

type MenuDTO struct {
	MenuTitle         string        `json:"menu_title"`
	MenuType          string        `json:"menu_type"`
	HasMegamenu       bool          `json:"has_megamenu"`
	HasMenuChild      bool          `json:"has_menu_child"`
	HasPositionStatic bool          `json:"has_position_static"`
	HasSideBanner     bool          `json:"has_side_banner,omitempty"`
	BannerImage       string        `json:"banner_image,omitempty"`
	MenuItems         []MenuItemDTO `json:"menu_items,omitempty"`
}
