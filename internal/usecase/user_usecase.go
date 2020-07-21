package usecase

import (
	"context"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/wicaker/user/internal/domain"
)

type userUsecase struct {
	userRepo       domain.UserRepository
	contextTimeout time.Duration
}

// NewUserUsecase will create new an userUsecase object representation of domain.UserUsecase interface
func NewUserUsecase(timeout time.Duration, userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{
		contextTimeout: timeout,
		userRepo:       userRepo,
	}
}

/**
 * Used to register a new user. Pseudocode:
 * - set context.WithTimeout
 * - check user input in database
 * - if not exist, do hashing password
 * - do sync data before persist to db
 * - save a new user or update if existing user isActive=false
 * - create token as a key for user activation
 */
func (u *userUsecase) Register(ctx context.Context, user *domain.User) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// check user
	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"email": user.Email,
	}, nil)
	if err != nil {
		return "", err
	}
	if checkUser != nil && checkUser.IsActive == true {
		return "", domain.ErrUserAlreadyExist
	}

	// hash password
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "Password Encryption failed")
	}

	// Save new user or update data
	if checkUser == nil {
		user.Password = string(password)
		user, err = u.userRepo.Store(ctx, user)
		if err != nil {
			return "", errors.Wrap(err, "Store user data")
		}
	} else {
		checkUser.Password = string(password)
		user, err = u.userRepo.Update(ctx, checkUser)
		if err != nil {
			return "", errors.Wrap(err, "Update user data")
		}
	}

	// create token
	expiresAt := time.Now().Add(time.Hour * 24 * 30).Unix()
	tk := &domain.JWToken{
		UUID:  user.UUID,
		Email: user.Email,
		Salt:  user.Salt,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, nil
}

/**
 * Used to login. Pseudocode:
 * - set context.WithTimeout
 * - check user input in database
 * - if exist do compare password
 * - if match do create token
 */
func (u *userUsecase) Login(ctx context.Context, user *domain.User) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// check user
	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"email":     user.Email,
		"is_active": true,
	}, nil)
	if err != nil {
		return "", err
	}
	if checkUser == nil {
		return "", domain.ErrUserNotFound
	}

	// check password
	err = bcrypt.CompareHashAndPassword([]byte(checkUser.Password), []byte(user.Password))
	if err != nil {
		return "", domain.ErrWrongPassword
	}

	// create token
	expiresAt := time.Now().Add(time.Minute * 100000).Unix()
	tk := &domain.JWToken{
		UUID:  checkUser.UUID,
		Email: checkUser.Email,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err
}

/**
 * Used to change email address. Pseudocode:
 * - set context.WithTimeout
 * - check token user id and email in db
 * - if exist, do compare password
 * - if match, do sync data
 * - update
 */
func (u *userUsecase) ChangeEmail(ctx context.Context, user *domain.User, parsedToken domain.JWToken) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"uuid":      parsedToken.UUID,
		"email":     parsedToken.Email,
		"is_active": true,
	}, nil)
	if err != nil {
		return err
	}
	if checkUser == nil {
		return domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(checkUser.Password), []byte(user.Password))
	if err != nil {
		return domain.ErrWrongPassword
	}

	checkUser.Email = user.Email

	_, err = u.userRepo.Update(ctx, checkUser)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Used to change password. Pseudocode:
 * - set context.WithTimeout
 * - check token user id, email and is_active=true in db
 * - if exist, do compare password
 * - if match, create hash new password
 * - sync data
 * - update
 * - create token as a key for change password confirmation
 */
func (u *userUsecase) ChangePassword(ctx context.Context, user *domain.User, parsedToken domain.JWToken) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"uuid":      parsedToken.UUID,
		"email":     parsedToken.Email,
		"is_active": true,
	}, nil)
	if err != nil {
		return "", err
	}
	if checkUser == nil {
		return "", domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(checkUser.Password), []byte(user.Password))
	if err != nil {
		return "", domain.ErrWrongPassword
	}

	// hash new password
	newPassword, err := bcrypt.GenerateFromPassword([]byte(*user.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "Password Encryption failed")
	}

	newPass := string(newPassword)
	checkUser.NewPassword = &newPass

	user, err = u.userRepo.Update(ctx, checkUser)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Minute * 60).Unix()
	tk := &domain.JWToken{
		UUID:  user.UUID,
		Email: user.Email,
		Salt:  user.Salt,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	tok := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := tok.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, nil
}

/**
 * Used to activate user after register for first time. Pseudocode:
 * - set context.WithTimeout
 * - check token user uuid, email, salt, is_active=false in db
 * - if match, do sync data
 * - update
 */
func (u *userUsecase) Activation(ctx context.Context, parsedToken domain.JWToken) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"uuid":      parsedToken.UUID,
		"email":     parsedToken.Email,
		"salt":      parsedToken.Salt,
		"is_active": false,
	}, nil)
	if err != nil {
		return err
	}
	if checkUser == nil {
		return domain.ErrUserNotFound
	}

	checkUser.IsActive = true

	_, err = u.userRepo.Update(ctx, checkUser)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Used to confirm new user password. Pseudocode:
 * - set context.WithTimeout
 * - check token user uuid, email, salt, is_active=true in db
 * - if match, do sync data (password= new_password)
 * - update
 */
func (u *userUsecase) PasswordConfirm(ctx context.Context, parsedToken domain.JWToken) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"uuid":      parsedToken.UUID,
		"email":     parsedToken.Email,
		"salt":      parsedToken.Salt,
		"is_active": true,
	}, nil)
	if err != nil {
		return err
	}
	if checkUser == nil {
		return domain.ErrUserNotFound
	}

	checkUser.Password = *checkUser.NewPassword

	_, err = u.userRepo.Update(ctx, checkUser)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Used when user forgot their password. Pseudocode:
 * - set context.WithTimeout
 * - check email in db
 * - if match, return token
 */
func (u *userUsecase) ForgotPasswordRequest(ctx context.Context, email string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"email":     email,
		"is_active": true,
	}, nil)
	if err != nil {
		return "", err
	}
	if checkUser == nil {
		return "", domain.ErrUserNotFound
	}

	expiresAt := time.Now().Add(time.Minute * 60).Unix()
	tk := &domain.JWToken{
		UUID:  checkUser.UUID,
		Email: checkUser.Email,
		Salt:  checkUser.Salt,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	tok := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := tok.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, nil
}

/**
 * Used when user confirm their forgot password via email. Pseudocode:
 * - set context.WithTimeout
 * - check token user uuid, email, salt, is_active=true in db
 * - if match, sync data
 * - update new data or password
 */
func (u *userUsecase) ForgotPasswordConfirm(ctx context.Context, user *domain.User, parsedToken domain.JWToken) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	checkUser, err := u.userRepo.FindOneBy(ctx, map[string]interface{}{
		"uuid":      parsedToken.UUID,
		"email":     parsedToken.Email,
		"salt":      parsedToken.Salt,
		"is_active": true,
	}, nil)
	if err != nil {
		return err
	}
	if checkUser == nil {
		return domain.ErrUserNotFound
	}

	checkUser.Password = *user.NewPassword

	_, err = u.userRepo.Update(ctx, checkUser)
	if err != nil {
		return err
	}

	return nil
}
