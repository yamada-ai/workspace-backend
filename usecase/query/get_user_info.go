package query

import (
	"context"
	"time"

	"github.com/yamada-ai/workspace-backend/domain"
	"github.com/yamada-ai/workspace-backend/domain/repository"
)

// GetUserInfoInput represents the input for GetUserInfo query
type GetUserInfoInput struct {
	UserName string
}

// GetUserInfoOutput represents the output of GetUserInfo query
type GetUserInfoOutput struct {
	UserID               int64 `json:"user_id"`
	RemainingMinutes     int   `json:"remaining_minutes"`
	TodayTotalMinutes    int   `json:"today_total_minutes"`
	LifetimeTotalMinutes int   `json:"lifetime_total_minutes"`
}

// GetUserInfoUseCase handles retrieving user session information
type GetUserInfoUseCase struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	now               func() time.Time
}

// NewGetUserInfoUseCase creates a new use case instance
func NewGetUserInfoUseCase(
	userRepository repository.UserRepository,
	sessionRepository repository.SessionRepository,
) *GetUserInfoUseCase {
	return &GetUserInfoUseCase{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		now:               func() time.Time { return time.Now().UTC() },
	}
}

// Execute retrieves user session information
func (uc *GetUserInfoUseCase) Execute(ctx context.Context, input GetUserInfoInput) (*GetUserInfoOutput, error) {
	// 1. Find user
	user, err := uc.userRepository.FindByName(ctx, input.UserName)
	if err != nil {
		return nil, err
	}

	// 2. Find active session (必須: アクティブセッションがなければエラー)
	activeSession, err := uc.sessionRepository.FindActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 3. Calculate remaining minutes until planned_end
	currentTime := uc.now().UTC()
	// Ensure session times are in UTC for consistent calculations
	activeSessionPlannedEndUTC := activeSession.PlannedEnd.UTC()
	remainingDuration := activeSessionPlannedEndUTC.Sub(currentTime)
	remainingMinutes := int(remainingDuration.Minutes())
	if remainingMinutes < 0 {
		remainingMinutes = 0 // 既に終了時刻を過ぎている場合は0分
	}

	// 4. Get today's sessions (00:00 to 23:59)
	todayStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := todayStart.Add(24 * time.Hour)

	todaySessions, err := uc.sessionRepository.FindByUserIDAndDateRange(ctx, user.ID, todayStart, todayEnd)
	if err != nil {
		return nil, err
	}

	// 5. Calculate today's total minutes
	todayTotalMinutes := calculateTotalMinutes(todaySessions, currentTime)

	// 6. Get all user sessions for lifetime calculation
	// LIMIT を大きめに設定して全件取得（実際の運用では適切な値に調整）
	allSessions, err := uc.sessionRepository.ListByUserID(ctx, user.ID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// 7. Calculate lifetime total minutes
	lifetimeTotalMinutes := calculateTotalMinutes(allSessions, currentTime)

	return &GetUserInfoOutput{
		UserID:               user.ID,
		RemainingMinutes:     remainingMinutes,
		TodayTotalMinutes:    todayTotalMinutes,
		LifetimeTotalMinutes: lifetimeTotalMinutes,
	}, nil
}

// calculateTotalMinutes calculates the total work minutes from sessions
func calculateTotalMinutes(sessions []*domain.Session, currentTime time.Time) int {
	totalMinutes := 0
	for _, session := range sessions {
		var duration time.Duration
		// Ensure all times are in UTC for consistent calculations
		startTimeUTC := session.StartTime.UTC()
		if session.ActualEnd != nil {
			// 完了したセッション: actual_end - start_time
			actualEndUTC := session.ActualEnd.UTC()
			duration = actualEndUTC.Sub(startTimeUTC)
		} else if session.IsActive() {
			// 現在アクティブなセッション: 現在時刻 - start_time
			duration = currentTime.Sub(startTimeUTC)
		}
		// actual_end がnilでかつ非アクティブなセッションは無視（想定外だが安全のため）
		totalMinutes += int(duration.Minutes())
	}
	return totalMinutes
}
