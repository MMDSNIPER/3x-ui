package service

import (
	"errors"

	"github.com/mhsanaei/3x-ui/v3/database"
	"github.com/mhsanaei/3x-ui/v3/database/model"
	"github.com/mhsanaei/3x-ui/v3/logger"
	"github.com/mhsanaei/3x-ui/v3/util/crypto"
	ldaputil "github.com/mhsanaei/3x-ui/v3/util/ldap"
	"github.com/xlzd/gotp"
	"gorm.io/gorm"
)

import (
	"encoding/json"
 
	"github.com/mhsanaei/3x-ui/v3/database"
	"github.com/mhsanaei/3x-ui/v3/database/model"
	"github.com/mhsanaei/3x-ui/v3/util/crypto"
	"gorm.io/gorm"
)

// UserService provides business logic for user management and authentication.
// It handles user creation, login, password management, and 2FA operations.
type UserService struct {
	settingService SettingService
}

// GetFirstUser retrieves the first user from the database.
// This is typically used for initial setup or when there's only one admin user.
func (s *UserService) GetFirstUser() (*model.User, error) {
	db := database.GetDB()

	user := &model.User{}
	err := db.Model(model.User{}).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) CheckUser(username string, password string, twoFactorCode string) (*model.User, error) {
	db := database.GetDB()

	user := &model.User{}

	err := db.Model(model.User{}).
		Where("username = ?", username).
		First(user).
		Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("invalid credentials")
	} else if err != nil {
		logger.Warning("check user err:", err)
		return nil, err
	}

	if !crypto.CheckPasswordHash(user.Password, password) {
		ldapEnabled, _ := s.settingService.GetLdapEnable()
		if !ldapEnabled {
			return nil, errors.New("invalid credentials")
		}

		host, _ := s.settingService.GetLdapHost()
		port, _ := s.settingService.GetLdapPort()
		useTLS, _ := s.settingService.GetLdapUseTLS()
		bindDN, _ := s.settingService.GetLdapBindDN()
		ldapPass, _ := s.settingService.GetLdapPassword()
		baseDN, _ := s.settingService.GetLdapBaseDN()
		userFilter, _ := s.settingService.GetLdapUserFilter()
		userAttr, _ := s.settingService.GetLdapUserAttr()

		cfg := ldaputil.Config{
			Host:       host,
			Port:       port,
			UseTLS:     useTLS,
			BindDN:     bindDN,
			Password:   ldapPass,
			BaseDN:     baseDN,
			UserFilter: userFilter,
			UserAttr:   userAttr,
		}
		ok, err := ldaputil.AuthenticateUser(cfg, username, password)
		if err != nil || !ok {
			return nil, errors.New("invalid credentials")
		}
	}

	twoFactorEnable, err := s.settingService.GetTwoFactorEnable()
	if err != nil {
		logger.Warning("check two factor err:", err)
		return nil, err
	}

	if twoFactorEnable {
		twoFactorToken, err := s.settingService.GetTwoFactorToken()

		if err != nil {
			logger.Warning("check two factor token err:", err)
			return nil, err
		}

		if gotp.NewDefaultTOTP(twoFactorToken).Now() != twoFactorCode {
			return nil, errors.New("invalid 2fa code")
		}
	}

	return user, nil
}

func (s *UserService) BumpLoginEpoch() error {
	db := database.GetDB()
	return db.Model(model.User{}).
		Where("1 = 1").
		Update("login_epoch", gorm.Expr("login_epoch + 1")).
		Error
}

func (s *UserService) UpdateUser(id int, username string, password string) error {
	db := database.GetDB()
	hashedPassword, err := crypto.HashPasswordAsBcrypt(password)

	if err != nil {
		return err
	}

	twoFactorEnable, err := s.settingService.GetTwoFactorEnable()
	if err != nil {
		return err
	}

	if twoFactorEnable {
		s.settingService.SetTwoFactorEnable(false)
		s.settingService.SetTwoFactorToken("")
	}

	return db.Model(model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"username":    username,
			"password":    hashedPassword,
			"login_epoch": gorm.Expr("login_epoch + 1"),
		}).
		Error
}

func (s *UserService) UpdateFirstUser(username string, password string) error {
	if username == "" {
		return errors.New("username can not be empty")
	} else if password == "" {
		return errors.New("password can not be empty")
	}
	hashedPassword, er := crypto.HashPasswordAsBcrypt(password)

	if er != nil {
		return er
	}

	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).First(user).Error
	if database.IsNotFound(err) {
		user.Username = username
		user.Password = hashedPassword
		return db.Model(model.User{}).Create(user).Error
	} else if err != nil {
		return err
	}
	user.Username = username
	user.Password = hashedPassword
	user.LoginEpoch++
	return db.Save(user).Error
}

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
 
// GetAllSubAdmins lists every user whose role is "sub-admin".
func (s *UserService) GetAllSubAdmins() ([]*model.User, error) {
	db := database.GetDB()
	var users []*model.User
	err := db.Model(model.User{}).Where("role = ?", "sub-admin").Find(&users).Error
	for _, u := range users {
		u.Password = ""
	}
	return users, err
}
 
// CreateSubAdmin inserts a new sub-admin with bcrypt-hashed password and the
// given allowed inbound IDs (stored as a JSON array string).
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
 
// UpdateSubAdmin updates a sub-admin record.  Password is re-hashed only when
// the caller supplies a non-empty string; omitting it leaves it unchanged.
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
 
// DeleteSubAdmin removes a sub-admin by ID. Refuses to delete a non-sub-admin row.
func (s *UserService) DeleteSubAdmin(id int) error {
	db := database.GetDB()
	return db.Delete(&model.User{}, "id = ? AND role = ?", id, "sub-admin").Error
}
 
// GetAllowedInboundIDs parses the JSON AllowedInbounds column and returns the
// slice of permitted inbound IDs. Returns nil for admin users (no restriction).
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
 
