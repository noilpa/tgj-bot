package customErrors

import (
	"errors"
	"fmt"
	"log"
)

var (
	ErrUsersForReviewNotFound = errors.New("users for review not found")
	ErrUserNorRegistered      = errors.New("user not registered")
	ErrCreateUser             = errors.New("user not created")
	ErrCreateReview           = errors.New("review not created")
	ErrChangeUserActivity     = errors.New("change is active user error")
	ErrChangeReviewApprove    = errors.New("change is approved review error")
	ErrChangeReviewComment    = errors.New("change is commented review error")
	ErrCloseMRs               = errors.New("close merge request error")
	ErrInvalidVariableType    = errors.New("invalid variable type")
	ErrGetUsersWithPayload    = errors.New("users with payload not found")
)

func Wrap(err error, msg string) error {
	if msg == "" {
		return err
	}
	return errors.New(fmt.Sprintf("%s: %v", msg, err))
}

func WrapWithLog(err error, msg string) error {
	err = Wrap(err, msg)
	log.Println(err)
	return err
}