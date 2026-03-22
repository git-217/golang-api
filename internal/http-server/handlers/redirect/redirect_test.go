package redirect

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"psql_crud/internal/storage"
	db_req "psql_crud/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockURLGetter struct {
	url  *db_req.CustomURL
	err  error
	call int
}

func (m *MockURLGetter) GetURL(ctx context.Context, alias string) (*db_req.CustomURL, error) {
	m.call++
	if m.err != nil {
		return nil, m.err
	}
	return m.url, nil
}

func createTestRequest(alias string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/"+alias, nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("alias", alias)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		alias          string
		mockURL        *db_req.CustomURL
		mockErr        error
		expectedStatus int
		expectedURL    string
		expectedError  string
	}{
		{
			name:           "success redirect",
			alias:          "testalias",
			mockURL:        &db_req.CustomURL{URL: "https://example.com", Alias: "testalias"},
			mockErr:        nil,
			expectedStatus: http.StatusFound,
			expectedURL:    "https://example.com",
		},
		{
			name:           "empty alias",
			alias:          "",
			mockURL:        nil,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedError:  "Invalid request",
		},
		{
			name:           "url not found",
			alias:          "nonexistent",
			mockURL:        nil,
			mockErr:        storage.ErrURLNotFound,
			expectedStatus: http.StatusOK,
			expectedError:  "URL not found",
		},
		{
			name:           "internal error",
			alias:          "erroralias",
			mockURL:        nil,
			mockErr:        errors.New("database error"),
			expectedStatus: http.StatusOK,
			expectedError:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := &MockURLGetter{
				url: tt.mockURL,
				err: tt.mockErr,
			}
			logger := newTestLogger()
			handler := New(logger, mockGetter)

			req := createTestRequest(tt.alias)
			w := httptest.NewRecorder()

			handler(w, req)

			if tt.expectedError != "" {
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.Equal(t, tt.expectedURL, w.Header().Get("Location"))
			}
		})
	}
}

func TestNew_RedirectStatusCode(t *testing.T) {
	mockGetter := &MockURLGetter{
		url: &db_req.CustomURL{URL: "https://example.com", Alias: "testalias"},
	}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest("testalias")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}

func TestNew_EmptyAlias_ReturnsError(t *testing.T) {

	mockGetter := &MockURLGetter{}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest("")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
	assert.Equal(t, 0, mockGetter.call, "GetURL should not be called with empty alias")
}

func TestNew_URLNotFound_ReturnsError(t *testing.T) {

	mockGetter := &MockURLGetter{
		err: storage.ErrURLNotFound,
	}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest("nonexistent")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "URL not found")
	assert.Equal(t, 1, mockGetter.call)
}

func TestNew_StorageError_ReturnsInternalError(t *testing.T) {

	mockGetter := &MockURLGetter{
		err: errors.New("database connection failed"),
	}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest("testalias")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "internal error")
	assert.Equal(t, 1, mockGetter.call)
}

func TestNew_WithLongAlias(t *testing.T) {

	longAlias := "verylongaliasthatwontbeusedanywhere"
	mockGetter := &MockURLGetter{
		url: &db_req.CustomURL{URL: "https://long-example.com", Alias: longAlias},
	}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest(longAlias)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://long-example.com", w.Header().Get("Location"))
}

func BenchmarkNew(b *testing.B) {
	mockGetter := &MockURLGetter{
		url: &db_req.CustomURL{URL: "https://example.com", Alias: "testalias"},
	}
	logger := newTestLogger()
	handler := New(logger, mockGetter)

	req := createTestRequest("testalias")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler(w, req)
	}
}

func requireEqual(t *testing.T, expected, actual interface{}) {
	require.Equal(t, expected, actual)
}
