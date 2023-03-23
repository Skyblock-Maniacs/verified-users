package requests

import (
	"encoding/json"
	"net/http"
	"errors"
	"os"
	"strconv"

	"github.com/disgoorg/snowflake/v2"
)

type DiscordRes struct {
	Id 				string `json:"id"`
	Username 		string `json:"username"`
	Discriminator   string `json:"discriminator"`
	Avatar 			string `json:"avatar"`
	Bot 			bool   `json:"bot"`
	System 			bool   `json:"system"`
	PublicFlags 	int    `json:"public_flags"`
	Banner 			string `json:"banner"`
	BannerColor 	string `json:"banner_color"`
	AccentColor 	int    `json:"accent_color"`
	CreatedAt 		int    `json:"created_at"`
}

type HypixelRes struct {
	Success 				bool   `json:"success"`
	Player struct {
		SocialMedia struct {
			Links struct {
				DISCORD 	string `json:"DISCORD"`
			} 					   `json:"links"`
		} 						   `json:"socialMedia"`
	} 							   `json:"player"`
}

type MojangRes struct {
	Id   	string `json:"id"`
	Name 	string `json:"name"`
}

type MojangProfileRes struct {
	Id		string `json:"id"`
	Name 	string `json:"name"`
	Properties []struct {
		Name	string `json:"name"`
		Value	string `json:"value"`
	} `json:"properties"`
}

func DiscordRequest(id string) (*DiscordRes, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/users/" + id, nil)
	req.Header.Set("Authorization", "Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	DiscordResp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var discordUser *DiscordRes
	if err := json.NewDecoder(DiscordResp.Body).Decode(&discordUser); err != nil {
		return nil, err
	}
	idInt, err := strconv.ParseInt(discordUser.Id, 10, 0)
	if err != nil {
		return nil, err
	}
	discordUser.CreatedAt = int(snowflake.ID(idInt).Time().Unix())
	return discordUser, nil
}

func HypixelRequest(uuid string) (*HypixelRes, error) {
	HypixelResp, err := http.Get("https://api.hypixel.net/player?key=" + os.Getenv("HYPIXEL_API_KEY") + "&uuid=" + uuid)
	if err != nil {
		return nil, err
	}
	var hypixelUser *HypixelRes
	if err := json.NewDecoder(HypixelResp.Body).Decode(&hypixelUser); err != nil {
		return nil, err
	}
	return hypixelUser, nil
}

func MojangRequest(ign string) (*MojangRes, error) {
	MojangResp, err := http.Get("https://api.mojang.com/users/profiles/minecraft/" + ign)
	if MojangResp.StatusCode != 200 {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	var mojangUser *MojangRes
	if err := json.NewDecoder(MojangResp.Body).Decode(&mojangUser); err != nil {
		return nil, err
	}
	return mojangUser, nil
}

func MojangProfileRequest(uuid string) (*MojangProfileRes, error) {
	MojangProfileResp, err := http.Get("https://sessionserver.mojang.com/session/minecraft/profile/" + uuid)
	if MojangProfileResp.StatusCode != 200 {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	var mojangProfileUser *MojangProfileRes
	if err := json.NewDecoder(MojangProfileResp.Body).Decode(&mojangProfileUser); err != nil {
		return nil, err
	}
	return mojangProfileUser, nil
}