package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"ai_ad_platform_recall_process/internal/config"
	"ai_ad_platform_recall_process/internal/repository"
)

type NotifyService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
	client   *http.Client
}

func NewNotifyService(cfg *config.Config) *NotifyService {
	return &NotifyService{
		userRepo: repository.NewUserRepository(),
		cfg:      cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Notify.TimeoutSeconds) * time.Second,
		},
	}
}

type NotifyData struct {
	RecallServiceName string `json:"recall_service_name"`
	Platform          string `json:"platform"`
	UserName          string `json:"user_name"`
}

func (s *NotifyService) TriggerNotify(recallServiceName, platform, userName string) {
	go s.asyncNotify(recallServiceName, platform, userName)
}

func (s *NotifyService) asyncNotify(recallServiceName, platform, userName string) {
	user, err := s.userRepo.FindByUsername(recallServiceName)
	if err != nil {
		log.Printf("[Notify] User not found: %s, error: %v", recallServiceName, err)
		return
	}

	if user.NotifyURL == "" {
		log.Printf("[Notify] No notify URL set for user: %s", recallServiceName)
		return
	}

	data := NotifyData{
		RecallServiceName: recallServiceName,
		Platform:          platform,
		UserName:          userName,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[Notify] JSON marshal error: %v", err)
		return
	}

	var lastErr error
	for i := 0; i < s.cfg.Notify.RetryTimes; i++ {
		if i > 0 {
			time.Sleep(time.Duration(s.cfg.Notify.RetryIntervalSeconds) * time.Second)
		}

		err := s.sendNotify(user.NotifyURL, jsonData)
		if err == nil {
			log.Printf("[Notify] Successfully sent to %s for user %s", user.NotifyURL, recallServiceName)
			return
		}
		lastErr = err
		log.Printf("[Notify] Retry %d failed for %s: %v", i+1, user.NotifyURL, err)
	}

	log.Printf("[Notify] All retries failed for %s, last error: %v", user.NotifyURL, lastErr)
}

func (s *NotifyService) sendNotify(notifyURL string, jsonData []byte) error {
	req, err := http.NewRequest("POST", notifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *NotifyService) SetNotifyURL(userID uint64, notifyURL string) error {
	parsedURL, err := fmt_URL(notifyURL)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateNotifyURL(userID, parsedURL)
}

func (s *NotifyService) GetNotifyURL(userID uint64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	return user.NotifyURL, nil
}

func fmt_URL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", nil
	}
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}
	return rawURL, nil
}
