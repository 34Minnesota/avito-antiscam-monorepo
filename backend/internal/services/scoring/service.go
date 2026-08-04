package scoring

import (
	"math/big"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

// Это по идее должно попадать в движок попытки
// TODO(owner: training-engine): вызвать этот сервис при завершении попытки
type Result struct {
	Score  domain.Score
	Passed bool
}

type Service struct{}

func New() Service {
	return Service{}
}

func (Service) Evaluate(points, maxPoints, passPercent int) (Result, error) {
	if err := validatePassPercent(passPercent); err != nil {
		return Result{}, err
	}

	score, err := domain.NewScore(points, maxPoints)
	if err != nil {
		return Result{}, err
	}

	return result(score, passPercent), nil
}

func (Service) Apply(current domain.Score, delta, passPercent int) (Result, error) {
	if err := validatePassPercent(passPercent); err != nil {
		return Result{}, err
	}

	score, err := current.Apply(delta)
	if err != nil {
		return Result{}, err
	}

	return result(score, passPercent), nil
}

func result(score domain.Score, passPercent int) Result {
	points := new(big.Int).Mul(big.NewInt(int64(score.Points())), big.NewInt(100))
	threshold := new(big.Int).Mul(big.NewInt(int64(score.MaxPoints())), big.NewInt(int64(passPercent)))
	passed := points.Cmp(threshold) >= 0

	return Result{Score: score, Passed: passed}
}

func validatePassPercent(passPercent int) error {
	if passPercent < 1 || passPercent > 100 {
		return core_errors.ErrInvalidPassRate
	}

	return nil
}
