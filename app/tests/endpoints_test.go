package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"main/cmd/web/handlers"
	"main/internal/e"
	"main/internal/segment"
	"main/internal/user"
	"main/pkg"
	"main/pkg/utils"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func clearRedis(ctx context.Context, rdb *redis.Client) error {
	return rdb.FlushAll(ctx).Err()
}

func uniqueKey() string {
	return uuid.New().String()
}

func TestCreateSegmentsEndpoint(t *testing.T) {
	cfg := utils.LoadConfig("../config/app.yaml")
	ctx := context.Background()

	rdb, err := pkg.NewRedisClient(ctx, cfg)
	require.NoError(t, err)

	psqlClient, err := pkg.NewPsqlClient(ctx, cfg)
	require.NoError(t, err)

	segmentRepo := segment.NewRepo(psqlClient)

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

	key := uniqueKey()
	for _, tc := range testCases {
		segmentRepo.Delete(ctx, tc.segmentName)
		body := fmt.Sprintf(`{"slug": "%s"}`, tc.segmentName)
		req := httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Segments(segmentRepo, rdb)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
		require.NoError(t, segmentRepo.Delete(ctx, tc.segmentName))
		rdb.Del(ctx, key)
	}

	// Test wrong body
	segmentRepo.Delete(ctx, "AVITO_DISCOUNT_50_test")
	req := httptest.NewRequest(
		"POST",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)

	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Segments(segmentRepo, rdb)(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test idempotent key already exists
	req = httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(`{"slug": "AVITO"}`)))
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, rdb)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	rdb.Del(ctx, key)

	// Test segment already exists
	segmentRepo.Delete(ctx, "AVITO_DISCOUNT_50_test")
	req = httptest.NewRequest("POST", "/segment", bytes.NewBuffer([]byte(`{"slug": "AVITO_DISCOUNT_50_test"}`)))
	req.Header.Add("Idempotency-Key", key)
	handlers.Segments(segmentRepo, rdb)(httptest.NewRecorder(), req)
	rdb.Del(ctx, key)

	req = httptest.NewRequest(
		"POST",
		"/segment",
		bytes.NewBuffer([]byte(`{"slug": "AVITO_DISCOUNT_50_test"}`)), // AVITO_DISCOUNT_50_test already exists
	)
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, rdb)(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	segmentRepo.Delete(ctx, "AVITO_DISCOUNT_50_test")
	rdb.Del(ctx, key)
}

func TestDeleteSegmentsEndpoint(t *testing.T) {
	cfg := utils.LoadConfig("../config/app.yaml")
	ctx := context.Background()

	rdb, err := pkg.NewRedisClient(ctx, cfg)
	require.NoError(t, err)

	psqlClient, err := pkg.NewPsqlClient(ctx, cfg)
	require.NoError(t, err)

	segmentRepo := segment.NewRepo(psqlClient)

	key := uniqueKey()
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
		segmentRepo.Create(ctx, &segment.Segment{Slug: tc.segmentName})
		body := fmt.Sprintf(`{"slug": "%s"}`, tc.segmentName)
		req := httptest.NewRequest("DELETE", "/segment", bytes.NewBuffer([]byte(body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Segments(segmentRepo, rdb)(rr, req)
		rdb.Del(ctx, key)
		assert.Equal(t, tc.expectedStatus, rr.Code)
		require.NoError(t, segmentRepo.Delete(ctx, tc.segmentName))
	}

	// Test wrong body
	req := httptest.NewRequest(
		"DELETE",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)
	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Segments(segmentRepo, rdb)(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test Idempotency-Key already exists
	req = httptest.NewRequest(
		"DELETE",
		"/segment",
		bytes.NewBuffer([]byte(`{"slag": "AVITO_DISCOUNT_50_test"}`)), // slag - bad request
	)
	req.Header.Add("Idempotency-Key", key)
	rr = httptest.NewRecorder()
	handlers.Segments(segmentRepo, rdb)(rr, req)
	rdb.Del(ctx, key)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestGetUserActiveSegmentsEndpoint(t *testing.T) {
	cfg := utils.LoadConfig("../config/app.yaml")
	ctx := context.Background()

	rdb, err := pkg.NewRedisClient(ctx, cfg)
	require.NoError(t, err)

	psqlClient, err := pkg.NewPsqlClient(ctx, cfg)
	require.NoError(t, err)

	userRepo := user.NewRepo(psqlClient)
	segmentRepo := segment.NewRepo(psqlClient)

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
		{
			name:           "active_segments_not_found",
			expectedStatus: http.StatusNoContent,
			segmentsAdd:    []string{},
		},
	}

	for _, tc := range testCases {
		userId, err := userRepo.CreateUser(ctx)
		for _, slug := range tc.segmentsAdd {
			err = segmentRepo.Create(ctx, &segment.Segment{Slug: slug})
			var duplicate *e.DuplicateSegmentError
			if err != nil && !errors.As(err, &duplicate) {
				t.Errorf("error: %s", err)
			}
		}

		err = userRepo.AddDelSegments(ctx, &user.SegmentsAddDelDto{
			UserId:      userId,
			SegmentsAdd: tc.segmentsAdd,
		})
		if err != nil && !e.IsDuplicateError(err) {
			t.Errorf("error: %s", err)
		}

		req := httptest.NewRequest("GET", "/segment/user", nil)
		req.URL.RawQuery = fmt.Sprintf("id=%d", userId)
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, rdb)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
		require.NoError(t, userRepo.DelUser(ctx, userId))
		if tc.expectedStatus != 200 {
			continue
		}

		var s user.SegmentsDto
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &s))

		segmentsAddOut := make([]string, 0, len(s.Segments))
		for _, seg := range s.Segments {
			segmentsAddOut = append(segmentsAddOut, seg.Slug)
		}
		assert.Equal(t, tc.segmentsAdd, segmentsAddOut)
		assert.Equal(t, userId, s.UserId)

		require.NoError(t, userRepo.AddDelSegments(ctx, &user.SegmentsAddDelDto{
			UserId:      userId,
			SegmentsDel: tc.segmentsAdd,
		}))
		for _, seg := range segmentsAddOut {
			require.NoError(t, segmentRepo.Delete(ctx, seg))
		}
		require.NoError(t, userRepo.DelUser(ctx, userId))
	}

	maxId, err := userRepo.GetMaxId(ctx)
	require.NoError(t, err)

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
			name:           "user_id_negative",
			expectedStatus: http.StatusBadRequest,
			userId:         "-32",
		},
		{
			name:           "user_id_ascii",
			expectedStatus: http.StatusBadRequest,
			userId:         "hello",
		},
		{
			name:           "user_id_ascii",
			expectedStatus: http.StatusNoContent,
			userId:         strconv.Itoa(maxId + 1),
		},
	}

	for _, tc := range testCases2 {
		req := httptest.NewRequest("GET", "/segment/user", nil)
		req.URL.RawQuery = "id=" + tc.userId
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, rdb)(rr, req)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}
}

func notInArr(search string, arr []string) bool {
	for _, el := range arr {
		if el == search {
			return false
		}
	}
	return true
}

func TestAddDelSegmentsEndpoint(t *testing.T) {
	cfg := utils.LoadConfig("../config/app.yaml")
	ctx := context.Background()

	rdb, err := pkg.NewRedisClient(ctx, cfg)
	require.NoError(t, err)

	psqlClient, err := pkg.NewPsqlClient(ctx, cfg)
	require.NoError(t, err)

	userRepo := user.NewRepo(psqlClient)
	segmentRepo := segment.NewRepo(psqlClient)

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

	key := uniqueKey()
	for _, tc := range testCasesErr {
		req := httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer([]byte(tc.body)))
		req.Header.Add("Idempotency-Key", key)
		rr := httptest.NewRecorder()
		handlers.Users(userRepo, rdb)(rr, req)
		rdb.Del(ctx, key)
		assert.Equal(t, tc.expectedStatus, rr.Code)
	}

	// Test idempotent key already processed
	_, err = rdb.Set(ctx, key, "", 0).Result()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("Idempotency-Key", key)
	rr := httptest.NewRecorder()
	handlers.Users(userRepo, rdb)(rr, req)
	rdb.Del(ctx, key)
	assert.Equal(t, http.StatusConflict, rr.Code)

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
		userId, err := userRepo.CreateUser(ctx)
		require.NoError(t, err)

		for _, slug := range tc.add {
			err = segmentRepo.Create(ctx, &segment.Segment{Slug: slug})
			if e.IsDuplicateError(err) {
				continue
			}
		}

		segmentsAddDelDto := user.SegmentsAddDelDto{
			UserId:      userId,
			SegmentsAdd: tc.add,
			SegmentsDel: tc.del,
		}
		body, err := json.Marshal(segmentsAddDelDto)
		require.NoError(t, err)

		req = httptest.NewRequest("POST", "/segment/user", bytes.NewBuffer(body))
		req.Header.Add("Idempotency-Key", key)
		rr = httptest.NewRecorder()
		handlers.Users(userRepo, rdb)(rr, req)
		rdb.Del(ctx, key)
		assert.Equal(t, tc.expectedStatus, rr.Code)

		segments, err := userRepo.FindByUserId(ctx, userId)
		require.NoError(t, err)
		for _, seg := range segments.Segments {
			if notInArr(seg.Slug, tc.add) {
				t.Errorf("segment not found when reqiered")
			}
		}

		require.NoError(t, userRepo.AddDelSegments(ctx, &user.SegmentsAddDelDto{
			UserId:      userId,
			SegmentsDel: tc.add,
		}))

		for _, slug := range tc.add {
			require.NoError(t, segmentRepo.Delete(ctx, slug))
		}
		require.NoError(t, userRepo.DelUser(ctx, userId))
	}
}
