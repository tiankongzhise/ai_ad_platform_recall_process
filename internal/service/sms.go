package service

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	MockRegisterCode = "930108"
	MockResetCode     = "931216"
)

var (
	registerCodes = make(map[string]string)
	resetCodes    = make(map[string]string)
	mu            sync.RWMutex
)

type SMSService struct{}

func NewSMSService() *SMSService {
	return &SMSService{}
}

type SMSProvider interface {
	SendCode(phone, code string) error
}

type MockSMSProvider struct{}

func (p *MockSMSProvider) SendCode(phone, code string) error {
	log.Printf("[SMS Mock] 发送验证码 %s 到 %s", code, phone)
	return nil
}

func (s *SMSService) SendRegisterCode(phone string) (string, error) {
	code := MockRegisterCode

	provider := &MockSMSProvider{}
	if err := provider.SendCode(phone, code); err != nil {
		return "", err
	}

	mu.Lock()
	registerCodes[phone] = code
	mu.Unlock()

	go func() {
		time.Sleep(5 * time.Minute)
		mu.Lock()
		delete(registerCodes, phone)
		mu.Unlock()
	}()

	return code, nil
}

func (s *SMSService) SendResetCode(phone string) (string, error) {
	code := MockResetCode

	provider := &MockSMSProvider{}
	if err := provider.SendCode(phone, code); err != nil {
		return "", err
	}

	mu.Lock()
	resetCodes[phone] = code
	mu.Unlock()

	go func() {
		time.Sleep(5 * time.Minute)
		mu.Lock()
		delete(resetCodes, phone)
		mu.Unlock()
	}()

	return code, nil
}

func (s *SMSService) ValidateRegisterCode(phone, code string) bool {
	mu.RLock()
	defer mu.RUnlock()
	expected, ok := registerCodes[phone]
	return ok && expected == code
}

func (s *SMSService) ValidateResetCode(phone, code string) bool {
	mu.RLock()
	defer mu.RUnlock()
	expected, ok := resetCodes[phone]
	return ok && expected == code
}

func (s *SMSService) GenerateRandomCode(length int) string {
	const digits = "0123456789"
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}
