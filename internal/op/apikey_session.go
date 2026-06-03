package op

import (
	"encoding/json"
	"strings"
	"time"
	"urlAPI/internal/database"

	"github.com/pkg/errors"
)

func fetchAPIKeys(info *Session) error {
	keys, err := database.APIKeyStore.List()
	if err != nil {
		return errors.WithStack(err)
	}

	body, err := json.Marshal(map[string]any{"keys": keys})
	if err != nil {
		return errors.WithStack(err)
	}
	info.SettingBody = body
	return nil
}

func createAPIKey(info *Session) error {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Role        string   `json:"role"`
		QuotaDay    int      `json:"quota_day"`
		QuotaMonth  int      `json:"quota_month"`
		AllowedIPs  []string `json:"allowed_ips"`
		ExpiresAt   string   `json:"expires_at"`
	}

	if len(info.SettingBody) == 0 {
		return errors.New("missing request body")
	}

	if err := json.Unmarshal(info.SettingBody, &req); err != nil {
		return errors.WithStack(err)
	}

	var expiresAt time.Time
	if req.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil && !strings.Contains(req.ExpiresAt, "T") {
			expiresAt, err = time.Parse("2006-01-02", req.ExpiresAt)
		}
		if err != nil {
			expiresAt, err = time.Parse("2006-01-02T15:04", req.ExpiresAt)
		}
		if err != nil {
			return errors.WithStack(err)
		}
	}

	apiKey, err := database.APIKeyStore.Create(req.Name, req.Description, req.Role, req.QuotaDay, req.QuotaMonth, req.AllowedIPs, expiresAt)
	if err != nil {
		return errors.WithStack(err)
	}

	body, err := json.Marshal(map[string]string{"api_key": apiKey})
	if err != nil {
		return errors.WithStack(err)
	}
	info.SettingBody = body
	return nil
}

func deleteAPIKey(info *Session) error {
	var req struct {
		ID uint `json:"api_key_id"`
	}

	if err := json.Unmarshal(info.SettingBody, &req); err != nil {
		return errors.WithStack(err)
	}

	if err := database.APIKeyStore.Delete(req.ID); err != nil {
		return errors.WithStack(err)
	}

	info.SettingBody = []byte(`{"message":"deleted"}`)
	return nil
}

func updateAPIKey(info *Session) error {
	var req struct {
		ID   uint           `json:"api_key_id"`
		Data map[string]any `json:"api_key_data"`
	}

	if err := json.Unmarshal(info.SettingBody, &req); err != nil {
		return errors.WithStack(err)
	}

	if err := database.APIKeyStore.Update(req.ID, req.Data); err != nil {
		return errors.WithStack(err)
	}

	info.SettingBody = []byte(`{"message":"updated"}`)
	return nil
}
