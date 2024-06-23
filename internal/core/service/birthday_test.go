package service

import (
	"birthdayapp/internal/config"
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port/mock"
	"bytes"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSendInviteForUsers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepo(ctrl)
	mockTelegram := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := &config.Config{
		BirthdayGroupID: 12345,
		GroupOwnerID:    67890,
	}

	bs := &BirthdayService{
		tg:  mockTelegram,
		ur:  mockUserRepo,
		cfg: cfg,
		log: log,
	}

	usersForSendInvite := &[]domain.User{
		{TelegramID: 123},
		{TelegramID: 456},
	}

	birthdayUsers := "@user1, @user2"

	inviteLink := "http://invite.com"
	mockTelegram.EXPECT().GetInviteLink(gomock.Any(), birthdayUsers).Return(inviteLink, nil).Times(1)

	mockTelegram.EXPECT().UnBanUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockTelegram.EXPECT().SendMessage(int64(123), gomock.Any()).Times(1)
	mockTelegram.EXPECT().SendMessage(int64(456), gomock.Any()).Times(1)

	bs.sendInviteForUsers(usersForSendInvite, birthdayUsers)

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}

	assert.Equal(t, len(logSlice), 0)
}

func TestSendInviteForUsers_ErrGetInviteLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepo(ctrl)
	mockTelegram := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := &config.Config{
		BirthdayGroupID: 12345,
		GroupOwnerID:    67890,
	}

	bs := &BirthdayService{
		tg:  mockTelegram,
		ur:  mockUserRepo,
		cfg: cfg,
		log: log,
	}

	usersForSendInvite := &[]domain.User{
		{TelegramID: 123},
		{TelegramID: 456},
	}

	birthdayUsers := "@user1, @user2"

	inviteLink := "http://invite.com"
	mockTelegram.EXPECT().GetInviteLink(gomock.Any(), birthdayUsers).Return(inviteLink, errors.New("test")).Times(1)

	mockTelegram.EXPECT().UnBanUser(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockTelegram.EXPECT().SendMessage(int64(123), gomock.Any()).Times(1)
	mockTelegram.EXPECT().SendMessage(int64(456), gomock.Any()).Times(1)

	bs.sendInviteForUsers(usersForSendInvite, birthdayUsers)

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}

	assert.Equal(t, len(logSlice), 1)
	assert.Contains(t, logSlice[0], "error generate invite link")
}

func TestKickUsers_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		TimeToKick:      1 * time.Second,
		BirthdayGroupID: 12345,
		GroupOwnerID:    11111,
	}

	bs := &BirthdayService{
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	usersToKick := []domain.User{
		{TelegramID: 22222},
		{TelegramID: 33333},
		{TelegramID: 11111}, //group owner
	}

	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(33333)).Return(nil).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(22222)).Return(nil).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(11111)).Return(nil).Times(0)

	go bs.kickUsers(ctx, &usersToKick)

	time.Sleep(3 * cfg.TimeToKick)

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 0, len(logSlice))
}

func TestKickUsers_CtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		TimeToKick:      10 * time.Second,
		BirthdayGroupID: 12345,
		GroupOwnerID:    11111,
	}

	bs := &BirthdayService{
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	usersToKick := []domain.User{
		{TelegramID: 22222},
		{TelegramID: 33333},
		{TelegramID: 11111}, //group owner
	}

	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(33333)).Return(nil).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(22222)).Return(nil).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(11111)).Return(nil).Times(0)

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.kickUsers(ctx, &usersToKick)
	}()
	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 0, len(logSlice))
}

func TestKickUsers_KickErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		TimeToKick:      1 * time.Second,
		BirthdayGroupID: 12345,
		GroupOwnerID:    11111,
	}

	bs := &BirthdayService{
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	usersToKick := []domain.User{
		{TelegramID: 22222},
		{TelegramID: 33333},
		{TelegramID: 11111}, //group owner
	}

	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(33333)).Return(errors.New("test")).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(22222)).Return(errors.New("test")).Times(1)
	mockTg.EXPECT().KickUser(cfg.BirthdayGroupID, int64(11111)).Return(nil).Times(0)

	mockTg.EXPECT().SendMessage(int64(22222), "please, leave from group. We'll wait for next birthday")
	mockTg.EXPECT().SendMessage(int64(33333), "please, leave from group. We'll wait for next birthday")

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.kickUsers(ctx, &usersToKick)
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 2, len(logSlice))
}

func TestBirthdayNotify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := mock.NewMockUserRepo(ctrl)
	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		BirthdayGroupID: 12345,
		TimeToKick:      1 * time.Second,
	}

	bs := &BirthdayService{
		ur:  mockUR,
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	birthdayUsers := []domain.User{
		{Username: "user1", TelegramID: 22222},
	}

	subscribers := []domain.User{
		{Username: "sub1", TelegramID: 33333},
	}

	mockUR.EXPECT().GetUsersWithBirthdayToday().Return(&birthdayUsers, nil)
	mockUR.EXPECT().GetUsersSubscribedToUsers(&birthdayUsers).Return(&subscribers, nil)
	mockTg.EXPECT().SendMessage(cfg.BirthdayGroupID, "happy birthday @user1")

	mockTg.EXPECT().SendMessage(gomock.Any(), gomock.Any()).AnyTimes().Times(2)
	mockTg.EXPECT().GetInviteLink(gomock.Any(), gomock.Any()).AnyTimes().Times(1)
	mockTg.EXPECT().UnBanUser(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockTg.EXPECT().KickUser(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.BirthdayNotify(ctx, wg)
		wg.Wait()
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 0, len(logSlice))
}

func TestBirthdayNotify_WithoutBirthday(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := mock.NewMockUserRepo(ctrl)
	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		BirthdayGroupID: 12345,
		TimeToKick:      1 * time.Second,
	}

	bs := &BirthdayService{
		ur:  mockUR,
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	birthdayUsers := []domain.User{}

	mockUR.EXPECT().GetUsersWithBirthdayToday().Return(&birthdayUsers, nil)

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.BirthdayNotify(ctx, wg)
		wg.Wait()
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 0, len(logSlice))
}

func TestBirthdayNotify_WithoutSubscribers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := mock.NewMockUserRepo(ctrl)
	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		BirthdayGroupID: 12345,
		TimeToKick:      1 * time.Second,
	}

	bs := &BirthdayService{
		ur:  mockUR,
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	birthdayUsers := []domain.User{
		{Username: "user1", TelegramID: 22222},
	}

	subscribers := []domain.User{}

	mockUR.EXPECT().GetUsersWithBirthdayToday().Return(&birthdayUsers, nil)
	mockUR.EXPECT().GetUsersSubscribedToUsers(&birthdayUsers).Return(&subscribers, nil)

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.BirthdayNotify(ctx, wg)
		wg.Wait()
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 0, len(logSlice))
}

func TestBirthdayNotify_ErrGetUsersWithBirthdayToday(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := mock.NewMockUserRepo(ctrl)
	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		BirthdayGroupID: 12345,
		TimeToKick:      1 * time.Second,
	}

	bs := &BirthdayService{
		ur:  mockUR,
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	birthdayUsers := []domain.User{
		{Username: "user1", TelegramID: 22222},
	}

	mockUR.EXPECT().GetUsersWithBirthdayToday().Return(&birthdayUsers, errors.New("test"))

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.BirthdayNotify(ctx, wg)
		wg.Wait()
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 1, len(logSlice))
}

func TestBirthdayNotify_ErrGetSubscribers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUR := mock.NewMockUserRepo(ctrl)
	mockTg := mock.NewMockTelegram(ctrl)

	var logBuf bytes.Buffer
	log := slog.New(
		slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Config{
		BirthdayGroupID: 12345,
		TimeToKick:      1 * time.Second,
	}

	bs := &BirthdayService{
		ur:  mockUR,
		tg:  mockTg,
		log: log,
		cfg: &cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	birthdayUsers := []domain.User{
		{Username: "user1", TelegramID: 22222},
	}

	subscribers := []domain.User{
		{Username: "sub1", TelegramID: 33333},
	}

	mockUR.EXPECT().GetUsersWithBirthdayToday().Return(&birthdayUsers, nil)
	mockUR.EXPECT().GetUsersSubscribedToUsers(&birthdayUsers).Return(&subscribers, errors.New("test"))

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		bs.BirthdayNotify(ctx, wg)
		wg.Wait()
	}()

	select {
	case <-done:
		//test completed within the timeout
	case <-timeout.C:
		t.Fatal("test timed out")
	}

	logSlice := strings.Split(logBuf.String(), "\n")
	if len(logSlice) > 0 {
		logSlice = logSlice[:len(logSlice)-1]
	}
	assert.Equal(t, 1, len(logSlice))
}
