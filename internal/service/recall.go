package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/internal/repository"
)

var (
	ErrStateFormatError     = errors.New("state参数格式错误，请使用URL编码")
	ErrMissingRequiredParam = errors.New("缺少必填参数")
)

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
	RecallServiceName string
	Platform          string
	UserName          string
	ExtraParams       map[string]string
}

type RecallResponse struct {
	RecordID uint64 `json:"record_id"`
}

func (s *RecallService) ProcessRecallWithParams(state string) (*RecallParams, []string, error) {
	decodedState, err := url.QueryUnescape(state)
	if err != nil {
		return nil, nil, ErrStateFormatError
	}

	params, missingParams := s.parseParams(decodedState)
	if len(missingParams) > 0 {
		return nil, missingParams, ErrMissingRequiredParam
	}

	return params, nil, nil
}

func (s *RecallService) SaveRecall(params *RecallParams, extraParams map[string]string) (*RecallResponse, error) {
	record := &model.RecallRecord{
		RecallServiceName: params.RecallServiceName,
		Platform:          params.Platform,
		UserName:          params.UserName,
	}

	// 合并 state 中解析出的额外参数和 URL 中的额外参数
	allParams := make(map[string]string)
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

func (s *RecallService) ProcessRecall(state string, extraParams map[string]string) (*RecallResponse, []string, error) {
	params, missingParams, err := s.ProcessRecallWithParams(state)
	if err != nil {
		return nil, missingParams, err
	}
	resp, err := s.SaveRecall(params, extraParams)
	return resp, nil, err
}

func (s *RecallService) parseParams(decodedState string) (*RecallParams, []string) {
	params := &RecallParams{
		ExtraParams: make(map[string]string),
	}
	var missing []string

	pairs := strings.Split(decodedState, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "recall_service_name":
			params.RecallServiceName = value
		case "platform":
			params.Platform = value
		case "user_name":
			params.UserName = value
		default:
			params.ExtraParams[key] = value
		}
	}

	if params.RecallServiceName == "" {
		missing = append(missing, "recall_service_name")
	}
	if params.Platform == "" {
		missing = append(missing, "platform")
	}
	if params.UserName == "" {
		missing = append(missing, "user_name")
	}

	return params, missing
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
