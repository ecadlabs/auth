package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.ecadlabs.com/ecad/auth/jsonpatch"
	"git.ecadlabs.com/ecad/auth/notification"
	"git.ecadlabs.com/ecad/auth/query"
	"git.ecadlabs.com/ecad/auth/storage"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserContextKey = "user"
	DefaultLimit   = 20
)

type Users struct {
	Storage *storage.Storage
	Timeout time.Duration

	SessionMaxAge    time.Duration
	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod

	BaseURL         func() string
	UsersPath       string
	RefreshPath     string
	ResetPath       string
	LogPath         string
	EmailUpdatePath string
	Namespace       string

	Notifier               notification.Notifier
	ResetTokenMaxAge       time.Duration
	EmailUpdateTokenMaxAge time.Duration

	AuxLogger *log.Logger
}

func (u *Users) UsersURL() string {
	return u.BaseURL() + u.UsersPath
}

func (u *Users) RefreshURL() string {
	return u.BaseURL() + u.RefreshPath
}

func (u *Users) ResetURL() string {
	return u.BaseURL() + u.ResetPath
}

func (u *Users) LogURL() string {
	return u.BaseURL() + u.LogPath
}

func (u *Users) EmailUpdateURL() string {
	return u.BaseURL() + u.EmailUpdatePath
}

func (u *Users) context(r *http.Request) context.Context {
	if u.Timeout != 0 {
		ctx, _ := context.WithTimeout(r.Context(), u.Timeout)
		return ctx
	}
	return r.Context()
}

func errorHTTPStatus(err error) int {
	if e, ok := err.(*storage.Error); ok {
		return e.HTTPStatus
	}

	return http.StatusInternalServerError
}

func (u *Users) GetUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = self.Roles.Get().IsGranted(permissionGet, map[string]interface{}{
		"self": self.ID,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := u.Storage.GetUserByID(u.context(r), uid)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	self := r.Context().Value(UserContextKey).(*storage.User)

	if err := self.Roles.Get().IsGranted(permissionList, nil); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	q, err := query.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	userSlice, count, nextQuery, err := u.Storage.GetUsers(u.context(r), q)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	if len(userSlice) == 0 && !q.TotalCount {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pagination
	res := utils.Paginated{
		Value: userSlice,
	}

	if q.TotalCount {
		res.TotalCount = &count
	}

	if nextQuery != nil {
		nextUrl, err := url.Parse(u.UsersURL())
		if err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextUrl.RawQuery = nextQuery.Values().Encode()
		res.Next = nextUrl.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}

func (u *Users) resetToken(user *storage.User) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.ResetTokenMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.ResetURL(),
		utils.NSClaim(u.Namespace, "gen"): user.PasswordGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		return "", err
	}

	return token.SignedString(secret)
}

// Lazy email syntax verification
func validEmail(s string) bool {
	i := strings.IndexByte(s, '@')
	return i >= 1 && i < len(s)-1 && i == strings.LastIndexByte(s, '@')
}

func (u *Users) NewUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	var user storage.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validEmail(user.Email) {
		utils.JSONError(w, "Invalid email syntax", http.StatusBadRequest)
		return
	}

	user.EmailVerified = false
	user.PasswordHash = nil

	if !user.Roles.HasPrefix(RolePrefix) {
		user.Roles.Add(RoleRegular)
	}

	if err := self.Roles.Get().IsGranted(permissionCreate, map[string]interface{}{"user": &user}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	ret, err := u.Storage.NewUser(u.context(r), &user)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	// Create reset token
	token, err := u.resetToken(ret)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationInvite, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		CurrentUser: self,
		TargetUser:  ret,
		Token:       token,
		TokenMaxAge: u.ResetTokenMaxAge,
	}); err != nil {
		log.Error(err)
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvCreate, self.ID, ret.ID, r)).WithFields(log.Fields{
			"email":          ret.Email,
			"name":           ret.Name,
			"added":          ret.Added,
			"email_verified": ret.EmailVerified,
			"roles":          ret.Roles,
		}).Printf("User %v created account %v", self.ID, ret.ID)
	}

	w.Header().Set("Location", u.UsersURL()+ret.ID.String())
	utils.JSONResponse(w, http.StatusCreated, ret)
}

func (u *Users) PatchUser(w http.ResponseWriter, r *http.Request) {
	// TODO Email verification
	self := r.Context().Value(UserContextKey).(*storage.User)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var p jsonpatch.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ops, err := storage.OpsFromPatch(p)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	if _, ok := ops.Update["password_hash"]; ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if v, ok := ops.Update["password"]; ok {
		delete(ops.Update, "password")

		if p, ok := v.(string); ok {
			if p == "" {
				utils.JSONError(w, "Password is empty", http.StatusBadRequest)
				return
			}

			hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			if err != nil {
				log.Error(err)
				utils.JSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ops.Update["password_hash"] = hash
		}
	}

	userRoles := self.Roles.Get()

	if err = userRoles.IsGranted(permissionModify, map[string]interface{}{
		"self": self.ID,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	for _, r := range ops.AddRoles {
		if err := userRoles.IsGranted(permissionAddRole, map[string]interface{}{"role": r}); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	for _, r := range ops.RemoveRoles {
		if err := userRoles.IsGranted(permissionDeleteRole, map[string]interface{}{"role": r}); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	user, err := u.Storage.UpdateUser(u.context(r), uid, ops)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	// Log
	if u.AuxLogger != nil {
		if len(ops.Update) != 0 {
			u.AuxLogger.WithFields(logFields(EvUpdate, self.ID, uid, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated account %v", self.ID, uid)
		}

		for _, role := range ops.AddRoles {
			u.AuxLogger.WithFields(logFields(EvAddRole, self.ID, uid, r)).WithField("role", role).Printf("User %v added role `%s' to account %v", self.ID, role, uid)
		}

		for _, role := range ops.RemoveRoles {
			u.AuxLogger.WithFields(logFields(EvRemoveRole, self.ID, uid, r)).WithField("role", role).Printf("User %v removed role `%s' from account %v", self.ID, role, uid)
		}
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) DeleteUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = self.Roles.Get().IsGranted(permissionDelete, map[string]interface{}{
		"self": self.ID,
		"id":   uid,
	}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := u.Storage.DeleteUser(u.context(r), uid); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvDelete, self.ID, uid, r)).Printf("User %v deleted account %v", self.ID, uid)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Allow authorized requests too
	if token, err := jwtmiddleware.FromAuthHeader(r); err != nil {
		utils.JSONError(w, err.Error(), http.StatusUnauthorized)
		return
	} else if token != "" {
		request.Token = token
	}

	if request.Token == "" {
		utils.JSONError(w, "Token must not be empty", http.StatusBadRequest)
		return
	}

	if request.Password == "" {
		utils.JSONError(w, "Password must not be empty", http.StatusBadRequest)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.ResetURL(), true) {
		utils.JSONError(w, "Not a reset token", http.StatusUnauthorized)
		return
	}

	// Get password generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = u.Storage.UpdatePasswordWithGen(u.context(r), id, hash, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvReset, id, id, r)).Printf("Password for account %v reset", id)
	}
}

func (u *Users) SendResetRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Allow GET requests for testing purposes
		request.Email = r.FormValue("email")
	}

	if request.Email == "" {
		utils.JSONError(w, "Email must not be empty", http.StatusBadRequest)
		return
	}

	// Don't return anything
	defer w.WriteHeader(http.StatusNoContent)

	user, err := u.Storage.GetUserByEmail(u.context(r), request.Email)
	if err != nil {
		log.Error(err)
		return
	}

	// Create reset token
	token, err := u.resetToken(user)
	if err != nil {
		log.Error(err)
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationReset, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		CurrentUser: user,
		TargetUser:  user,
		Token:       token,
		TokenMaxAge: u.ResetTokenMaxAge,
	}); err != nil {
		log.Error(err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvResetRequest, user.ID, user.ID, r)).WithField("email", user.Email).Printf("User %v requested password reset", user.ID)
	}
}

func (u *Users) SendUpdateEmailRequest(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	var request struct {
		Email string    `json:"email"`
		ID    uuid.UUID `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Email == "" {
		utils.JSONError(w, "Email must not be empty", http.StatusBadRequest)
		return
	}

	if !validEmail(request.Email) {
		utils.JSONError(w, "Invalid email syntax", http.StatusBadRequest)
		return
	}

	userRoles := self.Roles.Get()
	if err := userRoles.IsGranted(permissionModify, map[string]interface{}{
		"self": self.ID,
		"id":   request.ID,
	}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusForbidden)
		return
	}

	user, err := u.Storage.GetUserByID(u.context(r), request.ID)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	// Create update token
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.EmailUpdateTokenMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.EmailUpdateURL(),
		utils.NSClaim(u.Namespace, "email"): request.Email,
		utils.NSClaim(u.Namespace, "gen"):   user.EmailGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokStr, err := token.SignedString(secret)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationEmailUpdateRequest, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		To:          []string{request.Email},
		Email:       request.Email,
		CurrentUser: user,
		TargetUser:  user,
		Token:       tokStr,
		TokenMaxAge: u.EmailUpdateTokenMaxAge,
	}); err != nil {
		log.Error(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvEmailUpdateRequest, self.ID, user.ID, r)).WithField("email", request.Email).Printf("User %v requested email update for user %v", self.ID, user.ID)
	}
}

func (u *Users) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Token == "" {
		utils.JSONError(w, "Token must not be empty", http.StatusBadRequest)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.EmailUpdateURL(), true) {
		utils.JSONError(w, "Not a email update token", http.StatusUnauthorized)
		return
	}

	// Get email
	email, ok := claims[utils.NSClaim(u.Namespace, "email")].(string)
	if !ok || !validEmail(email) {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get email generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, prevEmail, err := u.Storage.UpdateEmailWithGen(u.context(r), id, email, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationEmailUpdate, &notification.NotificationData{
		Addr:        getRemoteAddr(r),
		Email:       prevEmail,
		To:          []string{prevEmail, user.Email},
		CurrentUser: user,
		TargetUser:  user,
	}); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvEmailUpdate, id, id, r)).WithField("email", email).Printf("Email for account %v updated", id)
	}
}
