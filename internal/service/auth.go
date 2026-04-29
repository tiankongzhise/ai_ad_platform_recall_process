package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"ai_ad_platform_recall_process/internal/config"
	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/internal/repository"
	"ai_ad_platform_recall_process/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrUserExists          = errors.New("用户名已存在")
	ErrInvalidCredentials  = errors.New("用户名或密码错误")
	ErrUserNotFound        = errors.New("用户不存在")
	ErrInvalidToken        = errors.New("Token无效或已过期")
	ErrInvalidCode         = errors.New("验证码错误")
	ErrPhoneNotFound       = errors.New("手机号未注册")
	ErrPhoneAlreadyExists  = errors.New("手机号已被注册")
	ErrInvalidApiToken     = errors.New("ApiToken无效")
	ErrInvalidRefreshToken = errors.New("RefreshToken无效或已过期")
	ErrUserNotActive       = errors.New("账户已注销或未生效")
	ErrUsernamePhoneMismatch = errors.New("用户名与手机号不匹配")
)

type AuthService struct {
	userRepo         *repository.UserRepository
	tokenRepo        *repository.TokenRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	smsService       *SMSService
	cfg              *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:         repository.NewUserRepository(),
		tokenRepo:        repository.NewTokenRepository(),
		refreshTokenRepo: repository.NewRefreshTokenRepository(),
		smsService:       NewSMSService(),
		cfg:              cfg,
	}
}

// generateApiToken 生成一个随机的API Token
func generateApiToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateUID 生成用户唯一UID
// 格式：32位十六进制字符串
func generateUID() (string, error) {
	bytes := make([]byte, 16) // 16字节 = 32个十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone" binding:"required,len=11"`
	Code     string `json:"code" binding:"required"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	UID      string `json:"uid"` // 注册时自动生成的用户唯一标识
	ApiToken string `json:"api_token"` // 注册时自动生成长期有效的ApiToken
}

func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error) {
	// 检查是否存在活跃用户（logout_at = -1）
	_, err := s.userRepo.FindActiveByUsername(req.Username)
	if err == nil {
		// 存在同名活跃用户，不允许注册
		return nil, ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.Phone != "" && req.Code != "" {
		if !s.smsService.ValidateRegisterCode(req.Phone, req.Code) {
			return nil, ErrInvalidCode
		}

		_, err := s.userRepo.FindActiveByPhone(req.Phone)
		if err == nil {
			return nil, ErrPhoneAlreadyExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 注册时自动生成ApiToken
	apiToken, err := generateApiToken()
	if err != nil {
		return nil, err
	}

	// 注册时自动生成UID
	uid, err := generateUID()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		UserName: req.Username,
		UID:      uid,
		Phone:    req.Phone,
		Password: hashedPassword,
		ApiToken: apiToken,
		Status:   1,    // 显式设置活跃状态
		LogoutAt: -1,  // 显式设置活跃用户标记
	}

	// 调试日志：确认 LogoutAt 值
	log.Printf("[DEBUG Register] 创建用户 user_name=%s, logout_at=%d", user.UserName, user.LogoutAt)

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		Username: user.UserName,
		UID:      uid,
		ApiToken: apiToken,
	}, nil
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Username  string    `json:"username"`
	UID       string    `json:"uid"` // 用户唯一标识，显示在用户名旁边
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// 使用 FindActiveByUsername 查询活跃用户（logout_at = -1）
	user, err := s.userRepo.FindActiveByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 双重检查：确保用户状态正常（防御性编程）
	if user.Status == 0 || user.LogoutAt != -1 {
		return nil, ErrInvalidCredentials
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	tokenStr, expiresAt, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		Username:  user.UserName,
		UID:       user.UID,
	}, nil
}

// CreateJWTTokenRequest 用ApiToken换取JWT和RefreshToken
type CreateJWTTokenRequest struct {
	ApiToken string `json:"api_token" binding:"required"`
}

type CreateJWTTokenResponse struct {
	JWTToken      string    `json:"jwt_token"`
	RefreshToken  string    `json:"refresh_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	RefreshExpiry time.Time `json:"refresh_expires_at"`
}

func (s *AuthService) CreateJWTToken(req CreateJWTTokenRequest) (*CreateJWTTokenResponse, error) {
	user, err := s.userRepo.FindByApiToken(req.ApiToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidApiToken
		}
		return nil, err
	}

	// 生成JWT和RefreshToken
	jwtToken, jwtExpiresAt, err := s.generateJWTToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &CreateJWTTokenResponse{
		JWTToken:      jwtToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     jwtExpiresAt,
		RefreshExpiry: refreshExpiresAt,
	}, nil
}

// RefreshJWTByRefreshTokenRequest 使用RefreshToken刷新JWT
type RefreshJWTByRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshJWTResponse struct {
	JWTToken      string    `json:"jwt_token"`
	RefreshToken  string    `json:"refresh_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	RefreshExpiry time.Time `json:"refresh_expires_at"`
}

func (s *AuthService) RefreshJWTByRefreshToken(req RefreshJWTByRefreshTokenRequest) (*RefreshJWTResponse, error) {
	refreshToken, err := s.refreshTokenRepo.FindByToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(refreshToken.ExpiresAt) {
		_ = s.refreshTokenRepo.DeleteByToken(req.RefreshToken)
		return nil, ErrInvalidRefreshToken
	}

	userID := refreshToken.UserID

	// 删除旧的refresh_token
	_ = s.refreshTokenRepo.DeleteByToken(req.RefreshToken)

	// 删除旧的JWT记录
	_ = s.tokenRepo.DeleteByUserID(userID)

	// 生成新的JWT和RefreshToken
	jwtToken, jwtExpiresAt, err := s.generateJWTToken(userID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshExpiresAt, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &RefreshJWTResponse{
		JWTToken:      jwtToken,
		RefreshToken:  newRefreshToken,
		ExpiresAt:     jwtExpiresAt,
		RefreshExpiry: refreshExpiresAt,
	}, nil
}

// RefreshJWTByApiTokenRequest 使用ApiToken刷新JWT
type RefreshJWTByApiTokenRequest struct {
	ApiToken string `json:"api_token" binding:"required"`
}

func (s *AuthService) RefreshJWTByApiToken(req RefreshJWTByApiTokenRequest) (*RefreshJWTResponse, error) {
	user, err := s.userRepo.FindByApiToken(req.ApiToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidApiToken
		}
		return nil, err
	}

	// 删除旧的refresh_token和JWT记录
	_ = s.refreshTokenRepo.DeleteByUserID(user.ID)
	_ = s.tokenRepo.DeleteByUserID(user.ID)

	// 生成新的JWT和RefreshToken
	jwtToken, jwtExpiresAt, err := s.generateJWTToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &RefreshJWTResponse{
		JWTToken:      jwtToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     jwtExpiresAt,
		RefreshExpiry: refreshExpiresAt,
	}, nil
}

// GetJWTTokenInfo 获取当前JWT和RefreshToken信息
type GetJWTTokenInfoResponse struct {
	JWTToken      string    `json:"jwt_token"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	RefreshExpiry time.Time `json:"refresh_expires_at"`
	JWTTTLHours   int       `json:"jwt_ttl_hours"`
	RefreshTTLHours int     `json:"refresh_ttl_hours"`
}

func (s *AuthService) GetJWTTokenInfo(userID uint64) (*GetJWTTokenInfoResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取最新的JWT记录
	tokens, err := s.tokenRepo.FindByUserID(userID)
	if err != nil || len(tokens) == 0 {
		return &GetJWTTokenInfoResponse{
			JWTToken:       "",
			RefreshToken:   "",
			ExpiresAt:      time.Time{},
			RefreshExpiry:  time.Time{},
			JWTTTLHours:    s.cfg.Token.JWTTTLHours,
			RefreshTTLHours: s.cfg.Token.RefreshTTLHours,
		}, nil
	}

	// 获取最新的JWT
	var latestJWT *model.Token
	for i := range tokens {
		if latestJWT == nil || tokens[i].CreatedAt.After(latestJWT.CreatedAt) {
			latestJWT = &tokens[i]
		}
	}

	// 获取最新的RefreshToken
	refreshTokens, err := s.refreshTokenRepo.FindByUserID(userID)
	if err != nil {
		refreshTokens = nil
	}

	var latestRefresh *model.RefreshToken
	for i := range refreshTokens {
		if latestRefresh == nil || refreshTokens[i].CreatedAt.After(latestRefresh.CreatedAt) {
			latestRefresh = &refreshTokens[i]
		}
	}

	response := &GetJWTTokenInfoResponse{
		JWTToken:        latestJWT.Token,
		ExpiresAt:       latestJWT.ExpiresAt,
		JWTTTLHours:     s.cfg.Token.JWTTTLHours,
		RefreshTTLHours: s.cfg.Token.RefreshTTLHours,
	}

	if latestRefresh != nil {
		response.RefreshToken = latestRefresh.Token
		response.RefreshExpiry = latestRefresh.ExpiresAt
	}

	_ = user // 避免未使用警告

	return response, nil
}

// UpdateApiTokenRequest 更换ApiToken
type UpdateApiTokenRequest struct {
	OldApiToken string `json:"old_api_token" binding:"required"`
}

type UpdateApiTokenResponse struct {
	ApiToken string `json:"api_token"`
}

func (s *AuthService) UpdateApiToken(userID uint64, req UpdateApiTokenRequest) (*UpdateApiTokenResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user.ApiToken != req.OldApiToken {
		return nil, ErrInvalidApiToken
	}

	// 生成新的ApiToken
	newApiToken, err := generateApiToken()
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateApiToken(userID, newApiToken); err != nil {
		return nil, err
	}

	// 清除用户的所有JWT和RefreshToken，强制重新登录
	_ = s.tokenRepo.DeleteByUserID(userID)
	_ = s.refreshTokenRepo.DeleteByUserID(userID)

	return &UpdateApiTokenResponse{
		ApiToken: newApiToken,
	}, nil
}

// GetApiToken 获取用户的ApiToken
type GetApiTokenResponse struct {
	ApiToken string `json:"api_token"`
}

func (s *AuthService) GetApiToken(userID uint64) (*GetApiTokenResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return &GetApiTokenResponse{
		ApiToken: user.ApiToken,
	}, nil
}

// GetAccountInfo 获取用户账户信息
type AccountInfoResponse struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
}

func (s *AuthService) GetAccountInfo(userID uint64) (*AccountInfoResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return &AccountInfoResponse{
		Username: user.UserName,
		UID:      user.UID,
	}, nil
}

type RefreshResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *AuthService) RefreshToken(userID uint64) (*RefreshResponse, error) {
	_ = s.tokenRepo.DeleteByUserID(userID)

	tokenStr, expiresAt, err := s.generateToken(userID)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	return s.tokenRepo.DeleteByToken(tokenStr)
}

func (s *AuthService) ValidateToken(tokenStr string) (*model.User, error) {
	// 首先尝试解析JWT，验证签名和过期时间
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Token.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// 验证用户是否存在且未注销
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 检查用户是否已注销
	if user.Status == 0 || user.LogoutAt != -1 {
		return nil, ErrInvalidToken
	}

	return user, nil
}

type Claims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// generateToken 生成旧的Token（兼容旧接口）
func (s *AuthService) generateToken(userID uint64) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.Token.ExpiryHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.Token.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	tokenModel := &model.Token{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.Create(tokenModel); err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

// generateJWTToken 生成新的JWT Token
func (s *AuthService) generateJWTToken(userID uint64) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.Token.JWTTTLHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.Token.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	tokenModel := &model.Token{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.Create(tokenModel); err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

// generateRefreshToken 生成RefreshToken
func (s *AuthService) generateRefreshToken(userID uint64) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.Token.RefreshTTLHours) * time.Hour)

	tokenStr, err := generateApiToken()
	if err != nil {
		return "", time.Time{}, err
	}

	refreshTokenModel := &model.RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiresAt, nil
}

type SendCodeRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Username string `json:"username"` // 可选，忘记密码时需要
}

type SendCodeResponse struct {
	Message string `json:"message"`
}

func (s *AuthService) SendRegisterCode(req SendCodeRequest) (*SendCodeResponse, error) {
	_, err := s.smsService.SendRegisterCode(req.Phone)
	if err != nil {
		return nil, err
	}
	return &SendCodeResponse{Message: "验证码已发送"}, nil
}

func (s *AuthService) SendResetCode(req SendCodeRequest) (*SendCodeResponse, error) {
	// 使用 FindActiveByPhone 查找活跃用户（status=1 且 logout_at=-1）
	user, err := s.userRepo.FindActiveByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPhoneNotFound
		}
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrPhoneNotFound
	}

	// 如果请求中包含用户名，验证用户名与手机号是否匹配
	if req.Username != "" {
		if user.UserName != req.Username {
			return nil, ErrUsernamePhoneMismatch
		}
	}

	_, err = s.smsService.SendResetCode(req.Phone)
	if err != nil {
		return nil, err
	}
	return &SendCodeResponse{Message: "验证码已发送"}, nil
}

type ResetPasswordRequest struct {
	Username    string `json:"username" binding:"required,min=3"`
	Phone       string `json:"phone" binding:"required,len=11"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

func (s *AuthService) ResetPassword(req ResetPasswordRequest) (*ResetPasswordResponse, error) {
	if !s.smsService.ValidateResetCode(req.Phone, req.Code) {
		return nil, ErrInvalidCode
	}

	// 使用 FindActiveByPhone 查找活跃用户（status=1 且 logout_at=-1）
	user, err := s.userRepo.FindActiveByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPhoneNotFound
		}
		return nil, err
	}

	// 验证用户名与手机号是否匹配
	if user.UserName != req.Username {
		return nil, ErrUsernamePhoneMismatch
	}

	hashedPwd, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdatePassword(user.ID, hashedPwd); err != nil {
		return nil, err
	}

	return &ResetPasswordResponse{Message: "密码重置成功"}, nil
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

func (s *AuthService) ChangePassword(userID uint64, req ChangePasswordRequest) (*ChangePasswordResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return nil, ErrInvalidCredentials
	}

	hashedPwd, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdatePassword(userID, hashedPwd); err != nil {
		return nil, err
	}

	return &ChangePasswordResponse{Message: "密码修改成功"}, nil
}

type DeleteAccountRequest struct {
	ConfirmUsername string `json:"confirm_username" binding:"required"`
}

type DeleteAccountResponse struct {
	Message string `json:"message"`
}

func (s *AuthService) DeleteAccount(userID uint64, req DeleteAccountRequest) (*DeleteAccountResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 检查用户是否已注销
	if user.Status == 0 || user.LogoutAt != -1 {
		return nil, errors.New("账户已注销")
	}

	if user.UserName != req.ConfirmUsername {
		return nil, errors.New("账户名称不正确")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		return nil, err
	}

	_ = s.tokenRepo.DeleteByUserID(userID)
	_ = s.refreshTokenRepo.DeleteByUserID(userID)

	return &DeleteAccountResponse{Message: "账户已注销"}, nil
}

// GetUidByUsername 通过用户名查询 UID（包含已注销用户）
type GetUidByUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type GetUidByUsernameResponse struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
}

func (s *AuthService) GetUidByUsername(req GetUidByUsernameRequest) (*GetUidByUsernameResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &GetUidByUsernameResponse{
		Username: user.UserName,
		UID:      user.UID,
	}, nil
}

// GetActivateUidByUsername 通过用户名查询活跃用户的 UID（仅查询 logout_at = -1 的用户）
type GetActivateUidByUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type GetActivateUidByUsernameResponse struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
}

func (s *AuthService) GetActivateUidByUsername(req GetActivateUidByUsernameRequest) (*GetActivateUidByUsernameResponse, error) {
	user, err := s.userRepo.FindActiveByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &GetActivateUidByUsernameResponse{
		Username: user.UserName,
		UID:      user.UID,
	}, nil
}
