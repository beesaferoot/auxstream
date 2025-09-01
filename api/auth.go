package api

import (
	"auxstream/db"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// cookie based auth session
// credits: https://github.com/Depado/gin-auth-example
const userKey = "user"

func CookieAuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get(userKey)
	if userId == nil {
		// Abort the request with the appropriate error code
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

type AuthForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func Signup(c *gin.Context, r db.UserRepo) {
	var form AuthForm

	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pHash, err := hashPassword(form.Password)
	if err != nil {
		log.Println("password hash failure: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup user"})
		return
	}
	_, err = r.CreateUser(c, form.Username, pHash)
	if err != nil {
		log.Println("CreateUser: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to signup user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully signed up user"})
}

func Login(c *gin.Context, r db.UserRepo) {
	session := sessions.Default(c)

	var form AuthForm

	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := r.GetUserByUsername(c, form.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	if !cmpHashString(user.PasswordHash, form.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	// Save the userid in the session
	session.Set(userKey, user.ID)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully authenticated user"})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userKey)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session token"})
		return
	}
	session.Delete("user")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

func hashPassword(password string) (hash string, err error) {
	// Generate a bcrypt hash of the password
	hashByte, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error generating hash:", err)
		return "", err
	}
	hash = string(hashByte)
	return
}

func cmpHashString(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}
	return false
}
