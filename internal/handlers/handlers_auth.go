package handlers

import (
	"context"
	"diplom_part1/internal/cookie"
	"diplom_part1/internal/encryption"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (cfg *ConfigHndl) LoginUse(ctx context.Context, login string) (bool, error) {
	use, err := cfg.DB.LoginUse(ctx, login)
	return use, err
}

func (cfg *ConfigHndl) NewUser(ctx context.Context, key string, login string, pass string) (string, error) {
	// create hash
	msg := login + pass
	hash := encryption.Encrypt(msg, key)

	// write in db login/hash
	userID, err := cfg.DB.WriteNewUser(ctx, login, hash)
	if err != nil {
		return "", err
	}
	// return userID
	return userID, nil
}

func (cfg *ConfigHndl) AuthorizeUser(ctx context.Context, key string, login string, pass string) (string, error) {
	// create hash
	msg := login + pass
	hash := encryption.Encrypt(msg, key)

	// read in db login/hash
	userID, err := cfg.DB.ReadUser(ctx, login, hash)
	if err != nil {
		return "", err
	}
	// return userID
	return userID, nil
}

func (cfg *ConfigHndl) userRegister(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type in struct {
		Login string `json:"login"`
		Pass  string `json:"password"`
	}

	valueIn := in{}

	if err := json.Unmarshal(body, &valueIn); err != nil || valueIn.Login == "" || valueIn.Pass == "" {
		http.Error(w, "register unmarshal error", http.StatusBadRequest)
		return
	}

	use, err := cfg.LoginUse(r.Context(), valueIn.Login)
	if err != nil {
		http.Error(w, "data base err", http.StatusInternalServerError)
		return
	}
	if use {
		http.Error(w, "login already in use", http.StatusConflict)
		return
	}

	userID, err := cfg.NewUser(r.Context(), cfg.Key, valueIn.Login, valueIn.Pass)
	if err != nil {
		http.Error(w, "data base err", http.StatusInternalServerError)
		return
	}

	cookie.AddCookie("userID", userID, w, r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(""))
	fmt.Fprint(w)
}

func (cfg *ConfigHndl) userLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type in struct {
		Login string `json:"login"`
		Pass  string `json:"password"`
	}

	valueIn := in{}

	if err := json.Unmarshal(body, &valueIn); err != nil || valueIn.Login == "" || valueIn.Pass == "" {
		http.Error(w, "login unmarshal error", http.StatusBadRequest)
		return
	}

	userID, err := cfg.AuthorizeUser(r.Context(), cfg.Key, valueIn.Login, valueIn.Pass)
	if err != nil {
		http.Error(w, "data base err", http.StatusInternalServerError)
		return
	}
	if userID == "" {
		http.Error(w, "invalid username/password pair", http.StatusUnauthorized)
		return
	}

	cookie.AddCookie("userID", userID, w, r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(""))
	fmt.Fprint(w)
}

func (cfg *ConfigHndl) CheckAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// получим куки для идентификации пользователя
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		userID := cookie.GetCookie(r, cfg.Key, "userID")
		if userID == "" {
			// no cookie
			http.Error(w, "CheckAuth/ userID no cookie", http.StatusUnauthorized)
			return
		}

		exist, err := cfg.DB.ExistsUserID(r.Context(), userID)
		if err != nil {
			// error server
			http.Error(w, "CheckAuth/ data base err", http.StatusInternalServerError)
			return
		}
		if !exist {
			// no in data base
			http.Error(w, "CheckAuth/ user not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
