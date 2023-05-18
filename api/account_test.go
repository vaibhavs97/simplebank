package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/vaibhavs97/simplebank/db/mock"
	db "github.com/vaibhavs97/simplebank/db/sqlc"
	"github.com/vaibhavs97/simplebank/util"
)

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAcc db.Account
	err = json.Unmarshal(data, &gotAcc)
	require.NoError(t, err)
	require.Equal(t, account, gotAcc)
}

func Test_getAccount(t *testing.T) {
	account := randomAccount()

	tests := []struct {
		name      string
		accountID int64
		mockCalls func(mockStore *mockdb.MockStore)
		checkResp func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Success",
			accountID: account.ID,
			mockCalls: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Return(account, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "Not found",
			accountID: account.ID,
			mockCalls: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, db.Account{})
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			mockCalls: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, db.Account{})
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			mockCalls: func(mockStore *mockdb.MockStore) {},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, db.Account{})
			},
		},
	}

	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			tt.mockCalls(mockStore)

			// Start test server and send request
			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tt.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			// check response
			tt.checkResp(t, recorder)
		})
	}
}
