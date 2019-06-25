package customErrors

import (
	"errors"
	"fmt"
)

var (
	ErrUsersForReviewNotFound = errors.New("users for review not found")
	ErrUserNorRegistered      = errors.New("user not registered")
	ErrCreateUser             = errors.New("user not created")
	ErrCreateReview           = errors.New("review not created")
	ErrChangeUserActivity     = errors.New("change is active user error")
	ErrChangeReviewApprove    = errors.New("change is approved review error")
	ErrInvalidVariableType    = errors.New("invalid variable type")
)

func Wrap(err error, msg string) error {
	if msg == "" {
		return err
	}
	return errors.New(fmt.Sprintf("%s: %v", msg, err))
}
