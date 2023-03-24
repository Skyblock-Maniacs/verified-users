package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"verified-users/mongo"
	"verified-users/requests"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type UserRequestBody struct {
	UUID 		string `json:"uuid"`
	DiscordId 	string `json:"discordId"`
}

type CloudflarePost struct {
	Cf string `json:"cf-turnstile-response"`
}
type CloudflareRes struct {
	Success 		bool `json:"success"`
	Errors 			[]string `json:"error-codes"`
	Challenge_ts 	string `json:"challenge_ts"`
	Hostname 		string `json:"hostname"`
	Action 			string `json:"action"`
	Cdata 			string `json:"cdata"`
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

	lookup := r.Group("/api/v1/lookup", cloudflareMiddleware())
	{
		lookup.POST("/ign/:ign", lookupIgn)
		lookup.POST("/discord/:discordId", lookupDiscord)
		lookup.POST("/uuid/:uuid", lookupUuid)
	}

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

func cloudflareMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body CloudflarePost
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cloudflare Error"})
			c.Abort()
			return
		}

		form := url.Values{}
		form.Add("secret", os.Getenv("CLOUDFLARE_TURNSILE_SECRET"))
		form.Add("response", body.Cf)
		form.Add("remoteip", c.ClientIP())

		CloudflareResp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Cloudflare Error"})
			c.Abort()
			return
		}
		var CloudflareRes CloudflareRes
		if err := json.NewDecoder(CloudflareResp.Body).Decode(&CloudflareRes); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			return
		}
		if !CloudflareRes.Success {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Cloudflare Error", "errors": CloudflareRes.Errors})
			c.Abort()
			return
		}

		c.Next()
	}
}

func lookupUuid(c *gin.Context) {
	mojangProfile, err := requests.MojangProfileRequest(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	dbUser, err := mongo.GetUserByUUID(mojangProfile.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User does not exist within our database."})
		return
	}
	discordUser, err := requests.DiscordRequest(dbUser.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error contacting the discord API. Please try again later."})
	}
	cape, _ := requests.DetermineCape(mojangProfile.Id, mojangProfile.Name)

	c.JSON(http.StatusOK, gin.H{
		"data": 
			gin.H{
				"ign": mojangProfile.Name,
				"uuid": mojangProfile.Id,
				"discordId": discordUser.Id,
				"disordUser": discordUser,
				"skin": mojangProfile.Properties[0].Value,
				"cape": cape,
			},
		})
}

func lookupIgn(c *gin.Context) {
	mojangUser, err := requests.MojangRequest(c.Param("ign"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	mojangProfile, err := requests.MojangProfileRequest(mojangUser.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	dbUser, err := mongo.GetUserByUUID(mojangUser.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User does not exist within our database."})
		return
	}
	discordUser, err := requests.DiscordRequest(dbUser.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error contacting the discord API. Please try again later."})
	}
	cape, _ := requests.DetermineCape(mojangProfile.Id, mojangProfile.Name)
	c.JSON(http.StatusOK, gin.H{
		"data": 
			gin.H{
				"ign": mojangUser.Name,
				"uuid": mojangUser.Id,
				"discordId": discordUser.Id,
				"disordUser": discordUser,
				"skin": mojangProfile.Properties[0].Value,
				"cape": cape,
			},
		})
}

func lookupDiscord(c *gin.Context) {
	dbUser, err := mongo.GetUserByDiscordID(c.Param("discordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User does not exist within our database."})
		return
	}
	mojangProfile, err := requests.MojangProfileRequest(dbUser.Uuid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	discordUser, err := requests.DiscordRequest(dbUser.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error contacting the discord API. Please try again later."})
		return
	}
	cape, _ := requests.DetermineCape(mojangProfile.Id, mojangProfile.Name)
	c.JSON(http.StatusOK, gin.H{
		"data": 
			gin.H{
				"ign": mojangProfile.Name,
				"uuid": mojangProfile.Id,
				"discordId": discordUser.Id,
				"disordUser": discordUser,
				"skin": mojangProfile.Properties[0].Value,
				"cape": cape,
			},
		})
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

	hypixelUser, err := requests.HypixelRequest(requestBody.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error contacting the hypixel API. Please try again later."})
		return
	}

	discordUser, err := requests.DiscordRequest(requestBody.DiscordId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "There was an error contacting the discord API. Please try again later."})
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