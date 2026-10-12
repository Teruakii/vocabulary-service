package constants

import "errors"

const (
	FAILED_TO_LOGIN          = "FAILED_TO_LOGIN"
	INVALID_CREDENTIALS      = "INVALID_CREDENTIALS"
	FAILED_TO_GENERATE_TOKEN = "FAILED_TO_GENERATE_TOKEN"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

// ListLimit is the page size for a requested limit: the default when none is
// given, and never more than MaxListLimit so one request cannot pull a whole
// table.
func ListLimit(requested int) int {
	if requested <= 0 {
		return DefaultListLimit
	}
	return min(requested, MaxListLimit)
}

const MaxWordsPerCategory = 20

var (
	ErrNotFound = errors.New("resource not found")
	ErrNotOwner = errors.New("you do not own this resource")
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailEmpty               = errors.New("email must not be empty")
	ErrEmailDuplicate           = errors.New("email already registered")
	ErrPasswordTooShort         = errors.New("password is too short")
	ErrPasswordTooLong          = errors.New("password must be at most 72 bytes")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrIdentityInvalid          = errors.New("invalid identity")
	ErrIdentityDuplicate        = errors.New("identity already linked")
	ErrIdentityNotFound         = errors.New("identity not found")
	ErrCannotUnlinkLastIdentity = errors.New("cannot unlink your only sign-in method")
	ErrPasswordResetUnavailable = errors.New("password reset is not available")
	ErrResetTokenInvalid        = errors.New("invalid or expired reset link")
)

var (
	ErrWordNotFound  = errors.New("word not found")
	ErrWordDuplicate = errors.New("word already exists")
	ErrWordInUse     = errors.New("word is still in use and cannot be deleted")
	ErrWordEmpty     = errors.New("word must not be empty")
	ErrWordForbidden = errors.New("cannot modify this word")
	ErrMeaningEmpty  = errors.New("meaning must not be empty")
)

var (
	ErrUserWordNotFound  = errors.New("user word not found")
	ErrUserWordDuplicate = errors.New("this word is already in your list")
	ErrInvalidQuality    = errors.New("invalid quality score")
)

var (
	ErrQuizNotFound            = errors.New("quiz not found")
	ErrQuestionNotFound        = errors.New("question not found")
	ErrQuizCompleted           = errors.New("quiz already completed")
	ErrQuestionAlreadyAnswered = errors.New("question already answered")
	ErrAnswerEmpty             = errors.New("answer must not be empty")
	ErrInvalidQuizSize         = errors.New("totalQuestions must be greater than 0")
	ErrInvalidQuizSource       = errors.New(`source must be "due" or "all"`)
	ErrNoWordsAvailable        = errors.New("no words available to quiz on")
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryNameEmpty = errors.New("category name must not be empty")
	ErrCategoryDuplicate = errors.New("category already exists")
	ErrCategoryInUse     = errors.New("category is still in use and cannot be deleted")
	ErrCategoryIsGlobal  = errors.New("global categories cannot be deleted")
	ErrTagDuplicate      = errors.New("word is already tagged with this category")
	ErrTagNotFound       = errors.New("tag not found")
	ErrCategoryForbidden = errors.New("cannot modify this category")
	ErrCategoryFull      = errors.New("category already has the maximum number of words")
	ErrVocabNotFound     = errors.New("vocab entry not found")
)

var (
	ErrTokenExpired        = errors.New("token expired")
	ErrOAuthExchangeFailed = errors.New("google oauth exchange failed")
	ErrEmailNotVerified    = errors.New("google email not verified")
	ErrInvalidToken        = errors.New("invalid token")
)
