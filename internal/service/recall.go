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
	ErrStateFormatError      = errors.New("state参数格式错误，应为RecallServiceUserUid_PlatformNumber_UserNumber格式")
	ErrMissingRequiredParam  = errors.New("缺少必填参数")
	ErrInvalidUid           = errors.New("RecallServiceUserUid格式错误，应为32位十六进制字符串")
	ErrInvalidPlatformNumber = errors.New("PlatformNumber格式错误，应为不超过13位的纯数字")
	ErrInvalidUserNumber    = errors.New("UserNumber格式错误，应为不超过13位的纯数字")
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
	RecallServiceUserUid string
	PlatformNumber       string
	UserNumber           string
	RecallServiceName    string // 通过 UID 查询得到
	ExtraParams          map[string]string
}

type RecallResponse struct {
	RecordID uint64 `json:"record_id"`
}

func (s *RecallService) ProcessRecallWithParams(state string) (*RecallParams, []string, error) {
	// 新格式：RecallServiceUserUid_PlatformNumber_User_Number
	// 不需要 URL 解码，直接解析
	parts := strings.Split(state, "_")
	if len(parts) != 3 {
		return nil, nil, ErrStateFormatError
	}

	uid := parts[0]
	platformNumber := parts[1]
	userNumber := parts[2]

	// 验证 RecallServiceUserUid 格式：32位十六进制字符串
	if !uidRegex.MatchString(uid) {
		return nil, nil, ErrInvalidUid
	}

	// 验证 PlatformNumber：不超过13位的纯数字
	if !numberRegex.MatchString(platformNumber) {
		return nil, nil, ErrInvalidPlatformNumber
	}

	// 验证 User_Number：不超过13位的纯数字
	if !numberRegex.MatchString(userNumber) {
		return nil, nil, ErrInvalidUserNumber
	}

	// 通过 RecallServiceUserUid 查找对应的 RecallServiceName
	user, err := s.userRepo.FindByRecallServiceUserUid(uid)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	params := &RecallParams{
		RecallServiceUserUid: uid,
		PlatformNumber:       platformNumber,
		UserNumber:           userNumber,
		RecallServiceName:    user.RecallServiceName,
		ExtraParams:          make(map[string]string),
	}

	return params, nil, nil
}

func (s *RecallService) SaveRecall(params *RecallParams, extraParams map[string]string) (*RecallResponse, error) {
	record := &model.RecallRecord{
		RecallServiceName: params.RecallServiceName,
		Platform:          params.PlatformNumber,  // 使用 PlatformNumber
		UserName:          params.UserNumber,      // 使用 UserNumber
	}

	// 合并 state 中解析出的额外参数和 URL 中的额外参数
	allParams := make(map[string]string)
	// 添加 UID 和数字字段到额外参数中，方便追踪
	allParams["recall_service_user_uid"] = params.RecallServiceUserUid
	allParams["platform_number"] = params.PlatformNumber
	allParams["user_number"] = params.UserNumber
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

	return &RecallResponse{RecordID: record.ID}, nil
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
