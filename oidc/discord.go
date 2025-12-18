package oidc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const discordBaseURL = "https://discord.com/api"

var ErrEmailNotVerified = fmt.Errorf("user does not have a verified email address")

type discordUserInfo struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	AvatarHash string `json:"avatar"`

	Email         string `json:"email"`
	EmailVerified bool   `json:"verified"`
}

func GetDiscordUserInfo(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", discordBaseURL+"/users/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		slog.Error("failed to get discord user info", "body", string(bytes))
		return nil, ErrRequestFailed
	}

	var data discordUserInfo
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	if !data.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	avatar := data.AvatarHash
	if avatar != "" {
		ext := "png"
		if strings.HasPrefix(avatar, "a_") {
			ext = "gif"
		}

		avatar = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", data.ID, avatar, ext)
	}

	userInfo := UserInfo{
		ID:             data.ID,
		Username:       data.Username,
		FirstName:      "",
		LastName:       "",
		ProfilePicture: avatar,
		Email:          data.Email,
	}

	return &userInfo, nil
}
