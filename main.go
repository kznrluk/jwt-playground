package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"
	"net/http"
	"time"
)

type User struct {
	UserID   string
	Password string
}

func GetUsers() []User {
	return []User{
		{
			UserID:   "Admin",
			Password: "p@ssw0rd",
		},
		{
			UserID:   "Guest",
			Password: "password",
		},
	}
}

func main() {
	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		userID := c.PostForm("userID")
		password := c.PostForm("password")

		var user User
		for _, us := range GetUsers() {
			if userID == us.UserID {
				if password != us.Password {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				user = us
			}
		}
		if user.UserID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["userID"] = userID
		claims["data"] = "my_first_jwt_token"
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		tokenString, err := token.SignedString([]byte("test_sign_key"))
		if err != nil {
			fmt.Println(err)
		}

		c.SetCookie("auth_token", tokenString, 0, "/", "", false, false)
		c.Redirect(302, "/yourUserName")
	})

	r.GET("/yourUserName", func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != "HS256" {
				return nil, xerrors.Errorf("invalid alg") // alg none attack countermeasures
			}
			return []byte("test_sign_key"), nil
		})
		if err != nil {
			spew.Dump(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		userName, ok := claims["userID"].(string)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.String(200, "あなたは"+userName+"!")
	})

	r.Static("/login/", "static")
	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}
