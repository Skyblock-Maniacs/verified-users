package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"encoding/json"

	"verified-users/mongo"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

type UserRequestBody struct {
	UUID string `json:"uuid"`
	DiscordId string `json:"discordId"`
}
type HypixelRes struct {
	Success bool `json:"success"`
	Player struct {
		SocialMedia struct {
			Links struct {
				DISCORD string `json:"DISCORD"`
			} `json:"links"`
		} `json:"socialMedia"`
	} `json:"player"`
}
type DiscordRes struct {
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
}

func Init() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})
	r.GET("/favicon.ico", func (c *gin.Context) { c.Status(http.StatusAccepted) })
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Endpoint Not Found"})
	})

	v1 := r.Group("/api/v1", authMiddleware())
	{
		v1.GET("/user", getUser)
		v1.POST("/user", postUser)
		v1.DELETE("/user", deleteUser)
	}

	log.Println("Server started on " + os.Getenv("PORT"))
	r.Run()
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryToken, headerToken := c.GetHeader("Authorization"), c.Query("key")
		var token string;
		if headerToken == "" && queryToken == ""{
			c.JSON(http.StatusBadRequest, gin.H{"message": "API Key Not Found"})
			c.Abort()
			return
		}
		if headerToken == "" && queryToken != "" {
			token = queryToken
		} else if headerToken != "" && queryToken == "" {
			token = headerToken
		} else if headerToken != "" && queryToken != "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "API Key Found in both header and query"})
			c.Abort()
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "API Key Not Found"})
			c.Abort()
			return
		}
		data, keyErr := mongo.GetApiKeyData(token)
		if keyErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "API Key Invalid"})
			c.Abort()
			return
		}
		if c.Request.Method == "GET" && !findPerm(data.ApiKey.Permissions, "usersGet") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Insufficient Permissions"})
			c.Abort()
			return
		}
		if c.Request.Method == "POST" && !findPerm(data.ApiKey.Permissions, "usersPost") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Insufficient Permissions"})
			c.Abort()
			return
		}
		if c.Request.Method == "DELETE" && !findPerm(data.ApiKey.Permissions, "usersDelete") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Insufficient Permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getUser(c *gin.Context) {
	uuid, discordId := c.Query("uuid"), c.Query("discordId")
	if uuid == "" && discordId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neither UUID nor Discord ID not found in your request."})
		return
	}
	if uuid != "" && discordId != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Both UUID and Discord ID found in your request. Please request one at a time."})
		return
	}
	if uuid != "" {
		data, err := mongo.GetUserByUUID(strings.Replace(uuid, "-", "", -1))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "UUID not found within our database."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"uuid": data.Uuid, "discordId": data.Id})
		return
	} else if discordId != "" {
		data, err := mongo.GetUserByDiscordID(discordId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Discord ID not found our database."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"uuid": data.Uuid, "discordId": data.Id})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "Neither UUID nor Discord ID not found our database."})
}

func postUser(c *gin.Context) {
	var requestBody UserRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON"})
		return
	}
	if requestBody.UUID == "" || requestBody.DiscordId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing UUID or Discord ID"})
		return
	}

	HypixelResp, err := http.Get("https://api.hypixel.net/player?key=" + os.Getenv("HYPIXEL_API_KEY") + "&uuid=" + requestBody.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}
	var hypixelUser HypixelRes
	if err := json.NewDecoder(HypixelResp.Body).Decode(&hypixelUser); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/users/" + requestBody.DiscordId, nil)
	req.Header.Set("Authorization", "Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	DiscordResp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}
	var discordUser DiscordRes
	if err := json.NewDecoder(DiscordResp.Body).Decode(&discordUser); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	if hypixelUser.Player.SocialMedia.Links.DISCORD != discordUser.Username + "#" + discordUser.Discriminator {
		c.JSON(http.StatusAccepted, gin.H{"message": "Discord ID and UUID do not match"})
		return
	}
	
	if err := mongo.InsertUser(requestBody.UUID, requestBody.DiscordId); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully added"})
}

func deleteUser(c *gin.Context) {
	uuid, discordId := c.Query("uuid"), c.Query("discordId")
	if uuid == "" && discordId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Neither UUID nor Discord ID not found in your request."})
		return
	}
	if uuid != "" && discordId != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Both UUID and Discord ID found in your request. Please request one at a time."})
		return
	}
	if uuid != "" {
		if err := mongo.DeleteUserViaUUID(strings.Replace(uuid, "-", "", -1)); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error deleting that user."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
		return
	}
	if discordId != "" {
		if err := mongo.DeleteUserViaDiscordID(discordId); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error deleting that user."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
}

func findPerm(permissions []string, perm string) bool {
	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}