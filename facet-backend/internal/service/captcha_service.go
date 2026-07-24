package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	Client     *redis.Client
	Prefix     string
	Expiration time.Duration
}

func (s *RedisStore) Set(id string, value string) error {
	return s.Client.Set(context.Background(), s.Prefix+id, value, s.Expiration).Err()
}

func (s *RedisStore) Get(id string, clear bool) string {
	key := s.Prefix + id
	val, err := s.Client.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	if clear {
		if err := s.Client.Del(context.Background(), key).Err(); err != nil {
			slog.Warn("failed to delete captcha key", "error", err)
		}
	}
	return val
}

func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val == answer
}

const (
	CaptchaTimeout = 3 * time.Minute
)

type CaptchaResponse struct {
	CaptchaID   string `json:"captchaId"`
	Base64Image string `json:"base64Image"`
}

type CaptchaService interface {
	Generate() (*CaptchaResponse, error)
	Verify(id, answer string) bool
}

type captchaService struct {
	captcha *base64Captcha.Captcha
}

func NewCaptchaService(redisClient *redis.Client) CaptchaService {
	store := &RedisStore{
		Client:     redisClient,
		Expiration: CaptchaTimeout,
		Prefix:     "captcha:",
	}

	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)

	return &captchaService{
		captcha: base64Captcha.NewCaptcha(driver, store),
	}
}

func (s *captchaService) Generate() (*CaptchaResponse, error) {
	id, content, answer, err := s.captcha.Generate()
	slog.Info("Generate Captcha, id is: {}, and answer is: {}", id, answer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate captcha: %w", err)
	}

	return &CaptchaResponse{
		CaptchaID:   id,
		Base64Image: content,
	}, nil
}

func (s *captchaService) Verify(id, answer string) bool {
	return s.captcha.Verify(id, answer, true)
}
