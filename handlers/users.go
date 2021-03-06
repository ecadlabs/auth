package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jq"
	"github.com/ecadlabs/auth/jsonpatch"
	"github.com/ecadlabs/auth/middleware"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultLimit = 20
)

type Users struct {
	Storage Storage
	Timeout time.Duration

	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod

	UsersPath       string
	RefreshPath     string
	ResetPath       string
	LogPath         string
	EmailUpdatePath string
	Namespace       string

	Notifier notification.Notifier

	Enforcer rbac.Enforcer

	AuxLogger *log.Logger
}

func (u *Users) UsersURL(c *middleware.DomainConfigData) string {
	return c.GetBaseURL() + u.UsersPath
}

func (u *Users) RefreshURL(c *middleware.DomainConfigData) string {
	return c.GetBaseURL() + u.RefreshPath
}

func (u *Users) ResetURL(c *middleware.DomainConfigData) string {
	return c.GetBaseURL() + u.ResetPath
}

func (u *Users) LogURL(c *middleware.DomainConfigData) string {
	return c.GetBaseURL() + u.LogPath
}

func (u *Users) EmailUpdateURL(c *middleware.DomainConfigData) string {
	return c.GetBaseURL() + u.EmailUpdatePath
}

func (u *Users) context(r *http.Request) (context.Context, context.CancelFunc) {
	if u.Timeout != 0 {
		return context.WithTimeout(r.Context(), u.Timeout)
	}
	return r.Context(), func() {}
}

func (u *Users) checkWritePermissions(role rbac.Role, typ string, self bool) (resTyp string, err error) {
	perm := []string{permissionFull, permissionWrite}
	if self {
		perm = append(perm, permissionWriteSelf)
	}

	regularGranted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return
	}

	perm = []string{permissionServiceFull, permissionServiceWrite}
	if self {
		perm = append(perm, permissionWriteSelf)
	}

	serviceGranted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return
	}

	switch {
	case regularGranted && !serviceGranted:
		resTyp = storage.AccountRegular
	case !regularGranted && serviceGranted:
		resTyp = storage.AccountService
	case !regularGranted && !serviceGranted:
		err = errors.ErrForbidden
		return
	}

	if typ != "" && resTyp != "" && typ != resTyp {
		err = errors.ErrForbidden
	}

	return
}

func (u *Users) checkReadPermissions(role rbac.Role, typ string, self bool) (resTyp string, err error) {
	perm := []string{permissionFull, permissionRead}
	if self {
		perm = append(perm, permissionReadSelf)
	}

	regularGranted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return
	}

	perm = []string{permissionServiceFull, permissionServiceRead}
	if self {
		perm = append(perm, permissionReadSelf)
	}

	serviceGranted, err := role.IsAnyGranted(perm...)
	if err != nil {
		return
	}

	switch {
	case regularGranted && !serviceGranted:
		resTyp = storage.AccountRegular
	case !regularGranted && serviceGranted:
		resTyp = storage.AccountService
	case !regularGranted && !serviceGranted:
		err = errors.ErrForbidden
		return
	}

	if typ != "" && resTyp != "" && typ != resTyp {
		err = errors.ErrForbidden
	}

	return
}

func (u *Users) getUserById(ctx context.Context, uid uuid.UUID, w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	user, err := u.Storage.GetUserByID(ctx, "", uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	_, err = u.checkReadPermissions(role, user.Type, self.ID == uid)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := u.context(r)
	defer cancel()

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	u.getUserById(ctx, uid, w, r)
}

func (u *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)
	r.ParseForm()
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	q, err := jq.FromValues(r.Form)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeQuerySyntax)
		return
	}

	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	typ, err := u.checkReadPermissions(role, "", false)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	userSlice, count, nextQuery, err := u.Storage.GetUsers(ctx, typ, q)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
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
		nextUrl, err := url.Parse(u.UsersURL(site))
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		nextUrl.RawQuery = nextQuery.Values().Encode()
		res.Next = nextUrl.String()
	}

	utils.JSONResponse(w, http.StatusOK, &res)
}

func (u *Users) resetToken(user *storage.User, conf *middleware.DomainConfigData) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":                             user.ID,
		"exp":                             now.Add(conf.ResetTokenMaxAge).Unix(),
		"iat":                             now.Unix(),
		"iss":                             conf.GetBaseURL(),
		"aud":                             u.ResetURL(conf),
		utils.NSClaim(u.Namespace, "gen"): user.PasswordGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		return "", err
	}

	return token.SignedString(secret)
}

func (u *Users) NewUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)

	var user storage.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error(err)
		utils.JSONError(w, "Invalid user data", errors.CodeBadRequest)
		return
	}

	if user.Type == "" {
		user.Type = storage.AccountRegular
	} else if user.Type != storage.AccountRegular && user.Type != storage.AccountService {
		utils.JSONError(w, "Invalid account type", errors.CodeBadRequest)
	}

	service := user.Type == storage.AccountService

	if !service && !utils.ValidEmail(user.Email) {
		utils.JSONErrorResponse(w, errors.ErrEmailFmt)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Check write access
	var fullGranted bool
	if service {
		fullGranted, err = role.IsAnyGranted(permissionServiceFull)
	} else {
		fullGranted, err = role.IsAnyGranted(permissionFull)
	}
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	var writeGranted bool
	if service {
		writeGranted, err = role.IsAnyGranted(permissionServiceWrite)
	} else {
		writeGranted, err = role.IsAnyGranted(permissionWrite)
	}
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !fullGranted && !writeGranted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	// Check delegating access
	if !fullGranted {
		granted, err := member.CanDelegate(role, user.Roles, permissionDelegatePrefix)

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
		if !granted {
			utils.JSONErrorResponse(w, errors.ErrForbidden)
			return
		}
	}

	user.EmailVerified = false
	user.PasswordHash = nil

	ret, err := u.Storage.NewUser(ctx, &user)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Create reset token
	token, err := u.resetToken(ret, site)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if user.Type == storage.AccountRegular {
		if err = u.Notifier.Notify(ctx, notification.NotificationInvite, &notification.NotificationData{
			Addr:        utils.GetRemoteAddr(r),
			CurrentUser: self,
			TargetUser:  ret,
			Token:       token,
			TokenMaxAge: site.ResetTokenMaxAge,
			Misc:        &site.TemplateData,
		}); err != nil {
			log.Error(err)
		}
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvCreate, member.ID, ret.ID, r)).WithFields(log.Fields{
			"email":          ret.Email,
			"name":           ret.Name,
			"added":          ret.Added,
			"email_verified": ret.EmailVerified,
		}).Printf("User %v created account %v from tenant %v", self.ID, ret.ID, member.TenantID)
	}

	w.Header().Set("Location", u.UsersURL(site)+ret.ID.String())
	utils.JSONResponse(w, http.StatusCreated, ret)
}

func (u *Users) PatchUser(w http.ResponseWriter, r *http.Request) {
	// TODO Email verification
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	var p jsonpatch.Patch
	if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ops, err := storage.OpsFromPatch(p)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	user, err := u.Storage.GetUserByID(ctx, "", uid)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	typ, err := u.checkWritePermissions(role, user.Type, self.ID == uid)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	delete(ops.Update, "password_hash")

	if typ == storage.AccountRegular && (ops.Add["address_whitelist"] != nil || ops.Remove["address_whitelist"] != nil) {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	if v, ok := ops.Update["password"]; ok {
		delete(ops.Update, "password")

		if p, ok := v.(string); ok {
			if p == "" {
				utils.JSONError(w, "Password is empty", errors.CodeBadRequest)
				return
			}

			var hash []byte
			if hash, err = bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost); err != nil {
				log.Error(err)
				utils.JSONErrorResponse(w, err)
				return
			}

			ops.Update["password_hash"] = hash
		}
	}

	user, err = u.Storage.UpdateUser(ctx, typ, uid, ops)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		if len(ops.Update) != 0 {
			u.AuxLogger.WithFields(logFields(EvUpdate, member.ID, uid, r)).WithFields(log.Fields(ops.Update)).Printf("User %v updated account %v from tenant %v", self.ID, uid, member.TenantID)

		}
	}

	utils.JSONResponse(w, http.StatusOK, user)
}

func (u *Users) DeleteUser(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)

	uid, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	user, err := u.Storage.GetUserByID(ctx, "", uid)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	typ, err := u.checkWritePermissions(role, user.Type, self.ID == uid)
	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	tenants, err := u.Storage.GetTenantsSoleMember(ctx, uid)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// TODO: archive user
	if err := u.Storage.DeleteUser(ctx, typ, uid); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvDelete, member.ID, uid, r)).Printf("User %v deleted account %v from tenant %v", self.ID, uid, member.TenantID)
	}

	individualTenants := []*storage.TenantModel{}

	for _, tenant := range tenants {
		if tenant.TenantType == storage.TenantTypeIndividual {
			individualTenants = append(individualTenants, tenant)
		}
	}

	// Archive any individual tenant made orphan by this deletion
	if len(individualTenants) > 0 {
		for _, tenant := range individualTenants {
			deleteErr := u.Storage.DeleteTenant(ctx, tenant.ID)
			// Log
			if deleteErr == nil && u.AuxLogger != nil {
				u.AuxLogger.WithFields(logFields(EvArchiveTenant, member.ID, tenant.ID, r)).Printf("User %v from tenant %v archived tenant %v", self.ID, member.TenantID, tenant.ID)
			}

			if deleteErr != nil {
				err = deleteErr
			}
		}

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	// Archive any tenant made orphan by this deletion
	if len(tenants) > 0 {
		for _, tenant := range tenants {
			deleteErr := u.Storage.DeleteTenant(ctx, tenant.ID)
			// Log
			if deleteErr == nil && u.AuxLogger != nil {
				u.AuxLogger.WithFields(logFields(EvArchiveTenant, member.ID, tenant.ID, r)).Printf("User %v archived tenant %v", self.ID, tenant.ID)
			}

			if deleteErr != nil {
				err = deleteErr
			}
		}

		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	// Allow authorized requests too
	if token, err := jwtmiddleware.FromAuthHeader(r); err != nil {
		utils.JSONError(w, err.Error(), errors.CodeUnauthorized)
		return
	} else if token != "" {
		request.Token = token
	}

	if request.Token == "" {
		utils.JSONErrorResponse(w, errors.ErrTokenEmpty)
		return
	}

	if request.Password == "" {
		utils.JSONErrorResponse(w, errors.ErrPasswordEmpty)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeTokenFmt)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.ResetURL(site), true) {
		utils.JSONErrorResponse(w, errors.ErrAudience)
		return
	}

	// Get password generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	err = u.Storage.UpdatePasswordWithGen(ctx, id, hash, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvReset, id, id, r)).Printf("Password for account %v reset", id)
	}
}

func (u *Users) SendResetRequest(w http.ResponseWriter, r *http.Request) {
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)

	var request struct {
		Email string `json:"email"`
	}

	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), errors.CodeBadRequest)
			return
		}
	} else {
		// Allow GET requests for testing purposes
		request.Email = r.FormValue("email")
	}

	if request.Email == "" {
		utils.JSONErrorResponse(w, errors.ErrEmailEmpty)
		return
	}

	// Don't return anything
	defer w.WriteHeader(http.StatusNoContent)

	ctx, cancel := u.context(r)
	defer cancel()

	user, err := u.Storage.GetUserByEmail(ctx, storage.AccountRegular, request.Email)
	if err != nil {
		log.Error(err)
		return
	}

	if user.Type == storage.AccountService {
		utils.JSONErrorResponse(w, errors.ErrService)
		return
	}

	// Create reset token
	token, err := u.resetToken(user, site)
	if err != nil {
		log.Error(err)
		return
	}

	if err = u.Notifier.Notify(r.Context(), notification.NotificationReset, &notification.NotificationData{
		Addr:        utils.GetRemoteAddr(r),
		CurrentUser: user,
		TargetUser:  user,
		Token:       token,
		TokenMaxAge: site.ResetTokenMaxAge,
		Misc:        &site.TemplateData,
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
	self := r.Context().Value(middleware.UserContextKey).(*storage.User)
	member := r.Context().Value(middleware.MembershipContextKey).(*storage.Membership)
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)

	var request struct {
		Email string    `json:"email"`
		ID    uuid.UUID `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if request.Email == "" {
		utils.JSONErrorResponse(w, errors.ErrEmailEmpty)
		return
	}

	if !utils.ValidEmail(request.Email) {
		utils.JSONErrorResponse(w, errors.ErrEmailFmt)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	user, err := u.Storage.GetUserByID(ctx, storage.AccountRegular, request.ID)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	role, err := u.Enforcer.GetRole(ctx, member.Roles.Get()...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	perm := []string{permissionFull, permissionWrite}
	if self.ID == request.ID {
		perm = append(perm, permissionWriteSelf)
	}

	granted, err := role.IsAnyGranted(perm...)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if !granted {
		utils.JSONErrorResponse(w, errors.ErrForbidden)
		return
	}

	// Create update token
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":                               user.ID,
		"exp":                               now.Add(site.EmailUpdateTokenMaxAge).Unix(),
		"iat":                               now.Unix(),
		"iss":                               site.GetBaseURL,
		"aud":                               u.EmailUpdateURL(site),
		utils.NSClaim(u.Namespace, "email"): request.Email,
		utils.NSClaim(u.Namespace, "gen"):   user.EmailGen,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)

	secret, err := u.JWTSecretGetter()
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	tokStr, err := token.SignedString(secret)
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Notifier.Notify(ctx, notification.NotificationEmailUpdateRequest, &notification.NotificationData{
		Addr:        utils.GetRemoteAddr(r),
		To:          []string{request.Email},
		Email:       request.Email,
		CurrentUser: user,
		TargetUser:  user,
		Token:       tokStr,
		TokenMaxAge: site.EmailUpdateTokenMaxAge,
		Misc:        &site.TemplateData,
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
	site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)

	var request struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	if request.Token == "" {
		utils.JSONErrorResponse(w, errors.ErrTokenEmpty)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errors.CodeTokenFmt)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(u.EmailUpdateURL(site), true) {
		utils.JSONErrorResponse(w, errors.ErrAudience)
		return
	}

	// Get email
	email, ok := claims[utils.NSClaim(u.Namespace, "email")].(string)
	if !ok || !utils.ValidEmail(email) {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get email generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONErrorResponse(w, errors.ErrInvalidToken)
		return
	}

	ctx, cancel := u.context(r)
	defer cancel()

	user, prevEmail, err := u.Storage.UpdateEmailWithGen(ctx, id, email, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	if err = u.Notifier.Notify(ctx, notification.NotificationEmailUpdate, &notification.NotificationData{
		Addr:        utils.GetRemoteAddr(r),
		Email:       prevEmail,
		To:          []string{prevEmail, user.Email},
		CurrentUser: user,
		TargetUser:  user,
		Misc:        &site.TemplateData,
	}); err != nil {
		log.Error(err)
		utils.JSONErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvEmailUpdate, id, id, r)).WithField("email", email).Printf("Email for account %v updated", id)
	}
}

func (u *Users) FindUserByMembershipID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := u.context(r)
	defer cancel()

	membershipID, err := uuid.FromString(mux.Vars(r)["memberId"])
	if err != nil {
		utils.JSONError(w, err.Error(), errors.CodeBadRequest)
		return
	}

	userID, err := u.Storage.GetUserIDByMembershipID(ctx, "", membershipID)

	if err != nil {
		utils.JSONErrorResponse(w, err)
		return
	}

	u.getUserById(ctx, userID, w, r)
}
