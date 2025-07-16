package main

import (
	"fmt"
)

type AppError struct {
	Op   string
	Code string
	Err  error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Code, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

const (
	ErrCodeConfig     = "CONFIG_ERROR"
	ErrCodeDiscord    = "DISCORD_ERROR"
	ErrCodeOpenAI     = "OPENAI_ERROR"
	ErrCodeHTTP       = "HTTP_ERROR"
	ErrCodeValidation = "VALIDATION_ERROR"
)

func NewAppError(op, code string, err error) *AppError {
	return &AppError{
		Op:   op,
		Code: code,
		Err:  err,
	}
}
