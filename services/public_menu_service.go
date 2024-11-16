package services

import (
	"backend-school/config"
	"backend-school/models"
)

type MenuService struct {
}

func NewMenuService() *MenuService {
	return &MenuService{}
}

func (s *MenuService) GetAllMenus() ([]models.Menu, error) {
	var menus []models.Menu
	if err := config.DB.Preload("MenuItems").Find(&menus).Error; err != nil {
		return nil, err
	}

	// Iterate over each menu to filter and structure menu items
	for i := range menus {
		menus[i].MenuItems = s.buildMenuHierarchy(menus[i].MenuItems)
	}

	return menus, nil
}

// buildMenuHierarchy organizes MenuItems into a hierarchy based on ParentID.
func (s *MenuService) buildMenuHierarchy(menuItems []models.MenuItem) []models.MenuItem {
	var rootItems []models.MenuItem
	subMenuMap := make(map[uint][]models.MenuItem)

	// Separate items into those with children and those without
	for _, item := range menuItems {
		if item.ParentID != nil {
			// Store child items in a map keyed by ParentID
			subMenuMap[*item.ParentID] = append(subMenuMap[*item.ParentID], item)
		} else {
			// Directly add items without a parent reference as root items
			rootItems = append(rootItems, item)
		}
	}

	// Attach child items to their respective parent items
	for i := range rootItems {
		rootItems[i].SubMenuItems = s.populateSubMenus(rootItems[i], subMenuMap)
	}

	return rootItems
}

// populateSubMenus recursively attaches child items to their parent.
func (s *MenuService) populateSubMenus(parent models.MenuItem, subMenuMap map[uint][]models.MenuItem) []models.MenuItem {
	if subItems, exists := subMenuMap[parent.ID]; exists {
		for i := range subItems {
			subItems[i].SubMenuItems = s.populateSubMenus(subItems[i], subMenuMap)
		}
		return subItems
	}
	return nil
}

func (s *MenuService) GetMenuByID(id uint) (models.Menu, error) {
	var menu models.Menu
	if err := config.DB.Preload("MenuItems").First(&menu, id).Error; err != nil {
		return menu, err
	}
	return menu, nil
}
