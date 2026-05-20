package service

import (
	"encoding/json"

	"github.com/mhsanaei/3x-ui/v3/database"
	"github.com/mhsanaei/3x-ui/v3/database/model"
	"github.com/mhsanaei/3x-ui/v3/util/crypto"
	"gorm.io/gorm"
)

func (s *UserService) GetAdminUser() (*model.User, error) {
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).
		Where("role = ? OR role = ? OR role IS NULL", "admin", "").
		First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAllSubAdmins() ([]*model.User, error) {
	db := database.GetDB()
	var users []*model.User
	err := db.Model(model.User{}).Where("role = ?", "sub-admin").Find(&users).Error
	for _, u := range users {
		u.Password = ""
	}
	return users, err
}

func (s *UserService) CreateSubAdmin(username, password string, allowedInbounds []int) (*model.User, error) {
	db := database.GetDB()
	hashed, err := crypto.HashPasswordAsBcrypt(password)
	if err != nil {
		return nil, err
	}
	allowedJSON, err := json.Marshal(allowedInbounds)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:        username,
		Password:        hashed,
		Role:            "sub-admin",
		AllowedInbounds: string(allowedJSON),
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateSubAdmin(id int, username, password string, allowedInbounds []int) error {
	db := database.GetDB()
	allowedJSON, err := json.Marshal(allowedInbounds)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"username":         username,
		"allowed_inbounds": string(allowedJSON),
	}
	if password != "" {
		hashed, err := crypto.HashPasswordAsBcrypt(password)
		if err != nil {
			return err
		}
		updates["password"] = hashed
		updates["login_epoch"] = gorm.Expr("login_epoch + 1")
	}
	return db.Model(model.User{}).Where("id = ? AND role = ?", id, "sub-admin").Updates(updates).Error
}

func (s *UserService) DeleteSubAdmin(id int) error {
	db := database.GetDB()
	return db.Delete(&model.User{}, "id = ? AND role = ?", id, "sub-admin").Error
}

func (s *UserService) GetAllowedInboundIDs(user *model.User) []int {
	if user.Role == "admin" || user.Role == "" {
		return nil
	}
	var ids []int
	if user.AllowedInbounds == "" || user.AllowedInbounds == "[]" {
		return ids
	}
	_ = json.Unmarshal([]byte(user.AllowedInbounds), &ids)
	return ids
}