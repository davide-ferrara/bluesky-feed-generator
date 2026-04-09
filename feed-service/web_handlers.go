package main

import (
	"bsky-schwartz/internal/bsky"
	"bsky-schwartz/internal/models"
	"bsky-schwartz/views"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	dbpkg "bsky-schwartz/db"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func (user *User) HasPerm(session sessions.Session, c *gin.Context) error {
	u, err := GetUser(session)
	if err != nil {
		err := fmt.Errorf("%s has no permission to access this page", u.Handle)
		logger.Error(err)
		c.Redirect(http.StatusFound, "/login")
		return err
	}
	return nil
}

func RootHandler(c *gin.Context) {
	session := sessions.Default(c)

	user, err := GetUser(session)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get userDID from session
	userDID := GetSessionString(session, "userDID")

	// Get username from session
	username := GetSessionString(session, "username")

	// Load weights from SQLite
	var weights map[string]float64
	if userDID != "" {
		dbWeights, err := dbpkg.GetWeightsByDID(userDID)
		if err == nil && len(dbWeights) > 0 {
			weights = dbWeights
		}
	}

	// Fallback: try session (for backwards compatibility)
	if weights == nil {
		weights = make(map[string]float64)
		if w := session.Get("weights"); w != nil {
			if data, ok := w.([]byte); ok {
				if err := json.Unmarshal(data, &weights); err != nil {
					fmt.Printf("Could not Marshal: %v", err)
				}
			}
		}
	}

	// Only fetch from Bluesky if we have no weights at all
	if len(weights) == 0 {
		client, err := bsky.NewClient(user.Handle, user.AppPassword)
		if err != nil {
			fmt.Println(err)
			c.Redirect(http.StatusFound, "/login")
			return
		}

		ctx := context.Background()
		username, err = bsky.GetProfileInfo(ctx, &client, user.Handle)
		if err != nil {
			fmt.Println(err)
		} else {
			session.Set("username", username)
		}

		// Get user DID from profile
		userDID, err = bsky.ResolveDID(ctx, &client, user.Handle)
		if err != nil {
			fmt.Println("Could not resolve handle, using handle as key")
			userDID = user.Handle
		}
		session.Set("userDID", userDID)

		// Read weights from PDS
		weights, err = bsky.GetWeights(ctx, &client, user.Handle)
		if err != nil {
			fmt.Println("ℹ️ INFO - Nessun weights nel PDS, uso default:", err)
			weights = make(map[string]float64)
			for _, v := range models.SwartzValues {
				weights[v.ID] = 0.0
			}
		} else {
			fmt.Println("WEIGHTS caricati dal PDS")
		}

		// Save to SQLite
		SaveUser(user.Handle, userDID)
		SaveWeights(userDID, weights)
	}

	// Get flash message
	var message string
	var messageType string
	if msg := session.Get("message"); msg != nil {
		message = msg.(string)
		session.Delete("message")
	}
	if mt := session.Get("messageType"); mt != nil {
		messageType = mt.(string)
		session.Delete("messageType")
	}
	if err := session.Save(); err != nil {
		fmt.Printf("Could not save session, trying clearing it: %v", err)
		session.Clear()
		return
	}

	component := views.SlidersPage(username, messageType, message, weights)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		panic(err)
	}
}

func LoginGetHandler(c *gin.Context) {
	session := sessions.Default(c)

	var message string
	var messageType string
	if msg := session.Get("message"); msg != nil {
		message = msg.(string)
		session.Delete("message")
	}
	if mt := session.Get("messageType"); mt != nil {
		messageType = mt.(string)
		session.Delete("messageType")
	}

	if err := session.Save(); err != nil {
		fmt.Printf("Could not save session, trying clearing it: %v", err)
		session.Clear()
		return
	}

	component := views.LoginPage(messageType, message)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		fmt.Println(err)
	}
}

func LoginPostHandler(c *gin.Context) {
	handle := c.PostForm("handle")
	appPassword := c.PostForm("appPassword")

	session := sessions.Default(c)

	if handle == "" || appPassword == "" {
		session.Set("message", "Handle e App Password sono obbligatori")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	client, err := bsky.NewClient(handle, appPassword)
	if err != nil {
		fmt.Println("ERRORE LOGIN - bsky.NewClient:", err)
		session.Set("message", "Handle o App Password non validi")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	fmt.Println("LOGIN OK - Handle:", handle)

	token, err := GenerateToken(handle, appPassword)
	if err != nil {
		fmt.Println("ERRORE - Generazione token:", err)
		session.Set("message", "Errore interno durante il login")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	session.Set("authToken", token)
	session.Save()

	ctx := context.Background()
	username, err := bsky.GetProfileInfo(ctx, &client, handle)
	if err != nil {
		fmt.Println("WARNING - GetProfileInfo:", err)
	} else {
		fmt.Println("USERNAME:", username)
		session.Set("username", username)
	}

	// Get user DID from profile (for weights storage)
	// For now we use handle as DID
	userDID, err := bsky.ResolveDID(ctx, &client, handle)
	if err != nil {
		fmt.Println("Could not resolve handle, using handle as key")
		userDID = handle
	}

	session.Set("userDID", userDID)

	// Try to load weights from SQLite first
	var weights map[string]float64
	dbDID, dbWeights, err := GetUserFromDB(handle)
	if err == nil && len(dbWeights) > 0 {
		fmt.Println("WEIGHTS caricati da SQLite")
		weights = dbWeights
		userDID = dbDID
	} else {
		// Read weights from PDS
		weights, err = bsky.GetWeights(ctx, &client, handle)
		if err != nil {
			fmt.Println("ℹ️ INFO - Nessun weights nel PDS, uso default:", err)
			weights = make(map[string]float64)
			for _, v := range models.SwartzValues {
				weights[v.ID] = 0.0
			}
		} else {
			fmt.Println("WEIGHTS caricati dal PDS")
		}
	}

	// Also save in session for backwards compatibility
	if data, err := json.Marshal(weights); err == nil {
		session.Set("weights", data)
	}

	if err := session.Save(); err != nil {
		fmt.Println("ERRORE salvataggio sessione:", err)
	}

	c.Redirect(http.StatusFound, "/")
}

func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("authToken")
	session.Delete("username")
	session.Delete("userDID")
	session.Delete("weights")

	if err := session.Save(); err != nil {
		fmt.Println(err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func ValuesHandler(c *gin.Context) {
	session := sessions.Default(c)
	username := GetSessionString(session, "username")
	component := views.ValuesPage(username)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		panic(err)
	}
}

func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)

	user, err := GetUser(session)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	username := GetSessionString(session, "username")

	var message string
	var messageType string
	if msg := session.Get("message"); msg != nil {
		message = msg.(string)
		session.Delete("message")
	}
	if mt := session.Get("messageType"); mt != nil {
		messageType = mt.(string)
		session.Delete("messageType")
	}
	if err := session.Save(); err != nil {
		fmt.Printf("Could not save session, trying clearing it: %v", err)
		session.Clear()
		return
	}

	component := views.ProfilePage(username, user.Handle, messageType, message)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		panic(err)
	}
}

// Weights webpage
func PreferencesHandler(c *gin.Context) {
	session := sessions.Default(c)

	user, err := GetUser(session)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse form data into map
	weights := make(map[string]float64)
	for _, value := range models.SwartzValues {
		if val := c.PostForm(value.ID); val != "" {
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				weights[value.ID] = f
			}
		}
	}

	// Get user DID
	userDID := GetSessionString(session, "userDID")

	// Save in SQLite (database)
	if err := SaveUser(user.Handle, userDID); err != nil {
		fmt.Println("ERRORE - Salvataggio utente in SQLite:", err)
	}
	if err := SaveWeights(userDID, weights); err != nil {
		fmt.Println("ERRORE - Salvataggio pesi in SQLite:", err)
	}
	fmt.Println("WEIGHTS salvati in SQLite")

	// Create client for PDS
	client, err := bsky.NewClient(user.Handle, user.AppPassword)
	if err != nil {
		fmt.Println("ERRORE - Creazione client fallita:", err)
		session.Set("message", "Errore di connessione a Bluesky")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Save weights to PDS (Bluesky protocol) - async, non-blocking
	go func() {
		if err := bsky.SaveWeights(context.Background(), &client, userDID, weights); err != nil {
			fmt.Printf("ERRORE async - Salvataggio PDS fallito per %s: %v\n", userDID, err)
		} else {
			fmt.Printf("WEIGHTS salvati nel PDS (async) per %s\n", userDID)
		}
	}()

	session.Set("message", "Preferenze salvate con successo!")
	session.Set("messageType", "success")

	// Also save in session for backwards compatibility
	if data, err := json.Marshal(weights); err == nil {
		session.Set("weights", data)
	}
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func DeleteWeightsHanlder(c *gin.Context) {
	session := sessions.Default(c)

	user, err := GetUser(session)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	client, err := bsky.NewClient(user.Handle, user.AppPassword)
	if err != nil {
		fmt.Println("ERRORE LOGIN - bsky.NewClient:", err)
		session.Set("message", "Handle o App Password non validi")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	ctx := context.Background()
	err = bsky.DeleteWeights(ctx, &client, user.Handle)
	if err != nil {
		session.Set("message", "Errore nell'eliminare il record")
		session.Set("messageType", "error")
		session.Save()
		c.Redirect(http.StatusPermanentRedirect, "/profile")
		return
	}

	c.Redirect(http.StatusPermanentRedirect, "/preferences")
	return
}

func GetSessionString(session sessions.Session, name string) string {
	if val := session.Get(name); val != nil {
		return val.(string)
	}
	return ""
}

func GetUser(session sessions.Session) (User, error) {
	var user User

	tokenStr := GetSessionString(session, "authToken")
	if tokenStr == "" {
		return user, fmt.Errorf("no auth token found")
	}

	handle, appPassword, err := ValidateToken(tokenStr)
	if err != nil {
		return user, fmt.Errorf("invalid token: %w", err)
	}

	user.Handle = handle
	user.AppPassword = appPassword
	return user, nil
}
