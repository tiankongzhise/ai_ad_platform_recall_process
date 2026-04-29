package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/internal/repository"
)

var (
	ErrStateFormatError      = errors.New("state参数格式错误，应为Uid_Platform_UserTag格式")
	ErrMissingRequiredParam  = errors.New("缺少必填参数")
	ErrInvalidUid            = errors.New("uid格式错误，应为32位十六进制字符串")
	ErrInvalidPlatformNumber = errors.New("platform格式错误，应为不超过13位的纯数字")
	ErrInvalidUserTag        = errors.New("user_tag格式错误，应为不超过13位的纯数字")
)

// 32位十六进制字符串的正则表达式
var uidRegex = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
// 不超过13位的纯数字正则表达式
var numberRegex = regexp.MustCompile(`^[0-9]{1,13}$`)

type RecallService struct {
	recallRepo *repository.RecallRepository
	userRepo   *repository.UserRepository
}

func NewRecallService() *RecallService {
	return &RecallService{
		recallRepo: repository.NewRecallRepository(),
		userRepo:   repository.NewUserRepository(),
	}
}

type RecallParams struct {
	UID        string
	Platform   string
	UserTag    string
	UserName   string // 通过 UID 查询得到
	ExtraParams          map[string]string
}

type RecallResponse struct {
	Success bool `json:"success"`
}

func (s *RecallService) ProcessRecallWithParams(state string) (*RecallParams, []string, error) {
	// state 格式：Uid_Platform_UserTag
	// 不需要 URL 解码，直接解析
	parts := strings.Split(state, "_")
	if len(parts) != 3 {
		return nil, nil, ErrStateFormatError
	}

	uid := parts[0]
	platform := parts[1]
	userTag := parts[2]

	// 验证 uid 格式：32位十六进制字符串
	if !uidRegex.MatchString(uid) {
		return nil, nil, ErrInvalidUid
	}

	// 验证 platform：不超过13位的纯数字
	if !numberRegex.MatchString(platform) {
		return nil, nil, ErrInvalidPlatformNumber
	}

	// 验证 user_tag：不超过13位的纯数字
	if !numberRegex.MatchString(userTag) {
		return nil, nil, ErrInvalidUserTag
	}

	// 通过 uid 查找对应 user_name
	user, err := s.userRepo.FindByUID(uid)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	params := &RecallParams{
		UID:      uid,
		Platform: platform,
		UserTag:  userTag,
		UserName: user.UserName,
		ExtraParams:          make(map[string]string),
	}

	return params, nil, nil
}

func (s *RecallService) SaveRecall(params *RecallParams, extraParams map[string]string) (*RecallResponse, error) {
	record := &model.RecallRecord{
		UserName: params.UserName,
		UID:      params.UID,
		Platform: params.Platform,
		UserTag:  params.UserTag,
	}

	// 合并 state 中解析出的额外参数和 URL 中的额外参数
	allParams := make(map[string]string)
	// 添加 UID 和数字字段到额外参数中，方便追踪
	allParams["uid"] = params.UID
	allParams["platform"] = params.Platform
	allParams["user_tag"] = params.UserTag
	for k, v := range params.ExtraParams {
		allParams[k] = v
	}
	for k, v := range extraParams {
		allParams[k] = v
	}

	if len(allParams) > 0 {
		jsonBytes, _ := json.Marshal(allParams)
		record.Params = string(jsonBytes)
	}

	if err := s.recallRepo.Create(record); err != nil {
		return nil, err
	}

	return &RecallResponse{Success: true}, nil
}

func (s *RecallService) Query(params repository.QueryParams) (*repository.QueryResult, error) {
	return s.recallRepo.Query(params)
}

func (s *RecallService) QueryLatest(params repository.QueryParams) (*model.RecallRecord, error) {
	return s.recallRepo.QueryLatest(params)
}

func (s *RecallService) QueryAll(params repository.QueryParams) ([]model.RecallRecord, error) {
	return s.recallRepo.QueryAll(params)
}
