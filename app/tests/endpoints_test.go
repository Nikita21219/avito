package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"main/cmd/web/handlers"
	redisRepoMock "main/internal/cache/mocks"
	"main/internal/e"
	historyRepoMock "main/internal/history/mocks"
	"main/internal/segment"
	segmentRepoMock "main/internal/segment/mocks"
	"main/internal/user"
	userRepoMock "main/internal/user/mocks"
	"main/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

//func UniqueKey() string {
//	return uuid.New().String()
//}

func TestCreateSegmentsEndpoint(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	segmentRepo := segmentRepoMock.NewMockRepository(ctl)
	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	ctx := context.Background()

	testCases := []struct {
		name           string
		expectedStatus int
		segmentName    string
	}{
		{
			name:           "create_1",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_VOICE_MESSAGES_test",
		},
		{
			name:           "create_2",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_PERFORMANCE_VAS_test",
		},
		{
			name:           "create_3",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_DISCOUNT_30_test",
		},
		{
			name:           "create_4",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_DISCOUNT_50_test",
		},
		{
			name:           "segment_name_empty",
			expectedStatus: http.StatusBadRequest,
			segmentName:    "",
		},
	}

	key := handlers.UniqueKey()
	for _, tc := range testCases {
		if tc.expectedStatus == http.StatusOK {
			segmentRepo.EXPECT().Create(ctx, &segment.Segment{Slug: tc.segmentName})
		}
		cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)

		body := fmt.Sprintf(`{"slug": "%s"}`, tc.segmentName)
		req := httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Segments(segmentRepo, cacheRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	// Test wrong body
	req := httptest.NewRequest(
		"POST",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
	cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)
	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test idempotent key already exists
	req = httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(`{"slug": "AVITO"}`)))
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(1), nil)
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)

	// Test idempotent key empty
	req = httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(`{"slug": "AVITO"}`)))
	req.Header.Add("Idempotency-Key", "")
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test segment already exists
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
	cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)
	segmentRepo.EXPECT().Create(
		ctx,
		&segment.Segment{Slug: "AVITO_DISCOUNT_50_test"},
	).Return(
		&e.DuplicateSegmentError{SegmentName: "AVITO_DISCOUNT_50_test"},
	)

	req = httptest.NewRequest(
		"POST",
		"/segment",
		bytes.NewBuffer([]byte(`{"slug": "AVITO_DISCOUNT_50_test"}`)), // AVITO_DISCOUNT_50_test already exists
	)
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestDeleteSegmentsEndpoint(t *testing.T) {
	ctx := context.Background()

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	segmentRepo := segmentRepoMock.NewMockRepository(ctl)
	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	key := handlers.UniqueKey()
	testCases := []struct {
		name           string
		expectedStatus int
		segmentName    string
	}{
		{
			name:           "delete1",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_VOICE_MESSAGES_test",
		},
		{
			name:           "delete2",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_PERFORMANCE_VAS_test",
		},
		{
			name:           "delete3",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_DISCOUNT_30_test",
		},
		{
			name:           "delete4",
			expectedStatus: http.StatusOK,
			segmentName:    "AVITO_DISCOUNT_50_test",
		},
		{
			name:           "segment_name_empty",
			expectedStatus: http.StatusBadRequest,
			segmentName:    "",
		},
	}

	for _, tc := range testCases {
		cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)
		if tc.expectedStatus == http.StatusOK {
			segmentRepo.EXPECT().Delete(ctx, tc.segmentName).Return(nil)
		}

		body := fmt.Sprintf(`{"slug": "%s"}`, tc.segmentName)
		req := httptest.NewRequest("DELETE", "/segment", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Segments(segmentRepo, cacheRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	// Test wrong body
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
	cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)

	req := httptest.NewRequest(
		"DELETE",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)
	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test Idempotency-Key already exists
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(1), nil)

	req = httptest.NewRequest(
		"DELETE",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, cacheRepo)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestGetUserActiveSegmentsEndpoint(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	userRepo := userRepoMock.NewMockRepository(ctl)
	historyRepo := historyRepoMock.NewMockRepository(ctl)
	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	testCases := []struct {
		name           string
		expectedStatus int
		segmentsAdd    []string
	}{
		{
			name:           "active_segments_1",
			expectedStatus: http.StatusOK,
			segmentsAdd:    []string{"AVITO_VOICE_MESSAGES_TEST", "AVITO_PERFORMANCE_VAS_TEST", "AVITO_DISCOUNT_30_TEST"},
		},
		{
			name:           "active_segments_2",
			expectedStatus: http.StatusOK,
			segmentsAdd:    []string{"AVITO_VOICE_MESSAGES_TEST"},
		},
	}

	userId := 1
	for _, tc := range testCases {
		cacheRepo.EXPECT().Get(
			ctx,
			fmt.Sprintf("avito_user_%d", userId),
			gomock.Any(),
		).DoAndReturn(func(ctx context.Context, key string, result interface{}) error {
			segments := user.Segments{UserId: userId, Segments: []*segment.Segment{
				{Id: 1, Slug: "test"},
			}}
			*result.(*user.Segments) = segments
			return nil
		})

		req := httptest.NewRequest("GET", "/segment/user", nil)
		req.URL.RawQuery = fmt.Sprintf("id=%d", userId)
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	testCases2 := []struct {
		name           string
		expectedStatus int
		userId         string
	}{
		{
			name:           "user_id_zero",
			expectedStatus: http.StatusBadRequest,
			userId:         "0",
		},
		{
			name:           "user_id_several",
			expectedStatus: http.StatusBadRequest,
			userId:         "50;100",
		},
		{
			name:           "user_id_negative",
			expectedStatus: http.StatusBadRequest,
			userId:         "-32",
		},
		{
			name:           "user_id_ascii",
			expectedStatus: http.StatusBadRequest,
			userId:         "hello",
		},
	}

	for _, tc := range testCases2 {
		req := httptest.NewRequest("GET", "/segment/user", nil)
		req.URL.RawQuery = "id=" + tc.userId
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	// Test user id not found redis
	cacheRepo.EXPECT().Get(ctx, fmt.Sprintf("avito_user_%d", userId), gomock.Any()).Return(errors.New("redis: nil"))
	userRepo.EXPECT().FindByUserId(ctx, userId).Return(nil, &e.UserNotFoundError{UserId: userId})

	req := httptest.NewRequest("GET", "/segment/user", nil)
	req.URL.RawQuery = "id=1"
	rr := httptest.NewRecorder()
	handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Test user id not found DB
	cacheRepo.EXPECT().Get(ctx, fmt.Sprintf("avito_user_%d", userId), gomock.Any()).Return(errors.New("redis: nil"))
	userRepo.EXPECT().FindByUserId(ctx, userId).Return(&user.Segments{UserId: 0, Segments: make([]*segment.Segment, 0)}, nil)

	req = httptest.NewRequest("GET", "/segment/user", nil)
	req.URL.RawQuery = "id=1"
	rr = httptest.NewRecorder()
	handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAddDelSegmentsEndpoint(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	userRepo := userRepoMock.NewMockRepository(ctl)
	cacheRepo := redisRepoMock.NewMockRepository(ctl)
	historyRepo := historyRepoMock.NewMockRepository(ctl)

	testCasesErr := []struct {
		name           string
		expectedStatus int
		body           string
	}{
		{
			name:           "wrong_body_1",
			expectedStatus: http.StatusBadRequest,
			body:           `{"WRONG_user_id": 4, "add": ["AVITO_VOICE_MESSAGES_TEST"], "del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
		{
			name:           "wrong_body_2",
			expectedStatus: http.StatusBadRequest,
			body:           `{"WRONG_user_id": 4, "add": [""], "del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
		{
			name:           "wrong_body_3",
			expectedStatus: http.StatusBadRequest,
			body:           `{"WRONG_user_id": 4, "add": ["AVITO_VOICE_MESSAGES_TEST"], "del": [""]}`,
		},
		{
			name:           "wrong_body_4",
			expectedStatus: http.StatusBadRequest,
			body:           `{"user_id": 4, "WRONG_add": ["AVITO_VOICE_MESSAGES_TEST"], "del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
		{
			name:           "wrong_body_5",
			expectedStatus: http.StatusBadRequest,
			body:           `{"user_id": 4, "add": ["AVITO_VOICE_MESSAGES_TEST"], "WRONG_del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
		{
			name:           "wrong_body_6",
			expectedStatus: http.StatusBadRequest,
			body:           `{"user_id": 4}`,
		},
		{
			name:           "wrong_user_id_zero",
			expectedStatus: http.StatusBadRequest,
			body:           `{"user_id": 0, "add": ["AVITO_VOICE_MESSAGES_TEST"], "del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
		{
			name:           "wrong_user_id_negative",
			expectedStatus: http.StatusBadRequest,
			body:           `{"user_id": -1, "add": ["AVITO_VOICE_MESSAGES_TEST"], "del": ["AVITO_DISCOUNT_50_TEST"]}`,
		},
	}

	key := handlers.UniqueKey()
	for _, tc := range testCasesErr {
		cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)

		req := httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer([]byte(tc.body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	testCasesOk := []struct {
		name           string
		expectedStatus int
		add            []string
		del            []string
	}{
		{
			name:           "add_1",
			expectedStatus: http.StatusOK,
			add:            []string{"AVITO_VOICE_MESSAGES_TEST"},
			del:            []string{},
		},
		{
			name:           "add_2",
			expectedStatus: http.StatusOK,
			add:            []string{"AVITO_VOICE_MESSAGES_TEST", "AVITO_DISCOUNT_50_TEST", "AVITO_DISCOUNT_30_TEST"},
			del:            []string{},
		},
	}

	for _, tc := range testCasesOk {
		userId := 1
		cacheRepo.EXPECT().Exists(ctx, key).Return(int64(0), nil)
		cacheRepo.EXPECT().Set(ctx, key, true, 60*time.Minute)
		s := user.SegmentsAddDelDto{
			UserId:      userId,
			SegmentsAdd: tc.add,
			SegmentsDel: tc.del,
		}

		userRepo.EXPECT().AddDelSegments(ctx, &s, historyRepo)

		body, err := json.Marshal(s)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer(body))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	// Test idempotent key already processed
	cacheRepo.EXPECT().Exists(ctx, key).Return(int64(1), nil)
	req := httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Users(userRepo, cacheRepo, historyRepo)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRateLimiter(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	limiterMiddleware := handlers.RateLimiter(handler)

	for i := 0; i < 10; i++ {
		limiterMiddleware.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		rr = httptest.NewRecorder()
		time.Sleep(time.Millisecond)
	}

	limiterMiddleware.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestReports(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	historyRepo := historyRepoMock.NewMockRepository(ctl)
	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	cfg := utils.LoadConfig("../config/app.yaml")

	testCases := []struct {
		name           string
		expectedStatus int
		date           string
	}{
		{
			name:           "report_1",
			expectedStatus: http.StatusBadRequest,
			date:           "hello world",
		},
		{
			name:           "report_2",
			expectedStatus: http.StatusBadRequest,
			date:           "2023-08-30 28",
		},
		{
			name:           "report_3",
			expectedStatus: http.StatusBadRequest,
			date:           "2023-08-30",
		},
		{
			name:           "report_4",
			expectedStatus: http.StatusBadRequest,
			date:           "2023.08.30 10:28",
		},
		{
			name:           "report_5",
			expectedStatus: http.StatusBadRequest,
			date:           "10:28",
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/report", nil)
		req.URL.RawQuery = fmt.Sprintf("date=%s", tc.date)
		rr := httptest.NewRecorder()
		handlers.Reports(historyRepo, cacheRepo, cfg)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}
}

func TestReportCheckBadRequest(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	testCases := []struct {
		name           string
		expectedStatus int
		rawQuery       string
	}{
		{
			name:           "report_check_1",
			expectedStatus: http.StatusBadRequest,
			rawQuery:       "task_id=32dsafasdffdsa;asdf",
		},
		{
			name:           "report_check_2",
			expectedStatus: http.StatusBadRequest,
			rawQuery:       "32dsafasdffdsa;asdf",
		},
		{
			name:           "report_check_3",
			expectedStatus: http.StatusBadRequest,
			rawQuery:       "32dsafasdffdsa",
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/report_check", nil)
		req.URL.RawQuery = tc.rawQuery
		rr := httptest.NewRecorder()
		handlers.ReportCheck(cacheRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}
}

func TestReportCheck(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cacheRepo := redisRepoMock.NewMockRepository(ctl)

	testCases := []struct {
		name           string
		expectedStatus int
		taskId         string
	}{
		{
			name:           "report_check_ok_1",
			expectedStatus: http.StatusOK,
			taskId:         "some_task_id",
		},
		{
			name:           "report_check_ko_2",
			expectedStatus: http.StatusInternalServerError,
			taskId:         "some_task_id",
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		if tc.expectedStatus == http.StatusInternalServerError {
			var report handlers.Report
			cacheRepo.EXPECT().Get(ctx, handlers.TaskPrefix+tc.taskId, &report).Return(fmt.Errorf("some error"))
		} else {
			var report handlers.Report
			cacheRepo.EXPECT().Get(ctx, handlers.TaskPrefix+tc.taskId, &report)
		}

		req := httptest.NewRequest("GET", "/report_check", nil)
		req.URL.RawQuery = "task_id=" + tc.taskId
		rr := httptest.NewRecorder()
		handlers.ReportCheck(cacheRepo)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}
}

func TestDownloadFile(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus int
		id             string
	}{
		{
			name:           "download_1",
			expectedStatus: http.StatusBadRequest,
			id:             "../..",
		},
		{
			name:           "download_2",
			expectedStatus: http.StatusBadRequest,
			id:             "..",
		},
		{
			name:           "download_3",
			expectedStatus: http.StatusBadRequest,
			id:             "~/",
		},
		{
			name:           "download_4",
			expectedStatus: http.StatusBadRequest,
			id:             "./././././././..",
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/report", nil)
		req.URL.RawQuery = fmt.Sprintf("id=%s", tc.id)
		rr := httptest.NewRecorder()
		handlers.DownloadFile()(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}
}
