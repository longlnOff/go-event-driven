package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lithammer/shortuuid/v3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tickets/adapters"
	dbAdapters "tickets/db"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/message"
	"tickets/service"
)

func TestComponent(t *testing.T) {
	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	redisClient := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadsheetsAPI := &adapters.SpreadsheetsAPIStub{}
	receiptsService := &adapters.ReceiptsServiceStub{IssuedReceipts: map[string]entities.IssueReceiptRequest{}}
	filesAPI := &adapters.FilesApiStub{}

	go func() {
		svc := service.New(
			db,
			spreadsheetsAPI,
			receiptsService,
			filesAPI,
			redisClient,
		)
		assert.NoError(t, svc.Run(ctx))
	}()

	waitForHttpServer(t)

	ticket := ticketsHttp.TicketStatusRequest{
		TicketID: uuid.NewString(),
		Status:   "confirmed",
		Price: entities.Money{
			Amount:   "50.30",
			Currency: "GBP",
		},
		CustomerEmail: "email@example.com",
	}

	idempotencyKey := uuid.NewString()

	// check idempotency
	for i := 0; i < 3; i++ {
		sendTicketsStatus(
			t,
			ticketsHttp.TicketsStatusRequest{
				Tickets: []ticketsHttp.TicketStatusRequest{ticket},
			},
			idempotencyKey,
		)
	}

	assertReceiptForTicketIssued(t, receiptsService, ticket)
	assertTicketPrinted(t, filesAPI, ticket)
	assertRowToSheetAdded(
		t,
		spreadsheetsAPI,
		ticket,
		"tickets-to-print",
	)
	assertTicketStoredInRepository(t, db, ticket)

	ticket.Status = "canceled"
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{ticket},
	}, uuid.NewString())

	assertRowToSheetAdded(
		t,
		spreadsheetsAPI,
		ticket,
		"tickets-to-refund",
	)
}

func assertTicketStoredInRepository(t *testing.T, db *sqlx.DB, ticket ticketsHttp.TicketStatusRequest) {
	ticketsRepo := dbAdapters.NewTicketsRepository(db)

	assert.Eventually(
		t,
		func() bool {
			tickets, err := ticketsRepo.FindAll(context.Background())
			if err != nil {
				return false
			}

			for _, t := range tickets {
				if t.TicketID == ticket.TicketID {
					return true
				}
			}

			return false
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertRowToSheetAdded(t *testing.T, spreadsheetsAPI *adapters.SpreadsheetsAPIStub, ticket ticketsHttp.TicketStatusRequest, sheetName string) bool {
	t.Helper()

	return assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			rows, ok := spreadsheetsAPI.Rows[sheetName]
			if !assert.True(t, ok, "sheet %s not found", sheetName) {
				return
			}

			var ticketRow []string

			for _, row := range rows {
				for _, col := range row {
					if col == ticket.TicketID {
						ticketRow = row
						break
					}
				}
			}
			if !assert.NotEmpty(t, ticketRow, "ticket row not found in sheet %s", sheetName) {
				return
			}

			expectedRow := []string{
				ticket.TicketID,
				ticket.CustomerEmail,
				ticket.Price.Amount,
				ticket.Price.Currency,
			}

			assert.Equal(
				t,
				expectedRow,
				ticketRow,
			)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertTicketPrinted(t *testing.T, filesAPI *adapters.FilesApiStub, ticket ticketsHttp.TicketStatusRequest) bool {
	return assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			content, err := filesAPI.DownloadFile(context.Background(), ticket.TicketID+"-ticket.html")
			if !assert.NoError(t, err) {
				return
			}

			if assert.NotEmpty(t, content) {
				return
			}

			assert.Contains(t, content, ticket.TicketID)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *adapters.ReceiptsServiceStub, ticket ticketsHttp.TicketStatusRequest) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceipts)

			assert.Equal(t, 1, issuedReceipts, "receipt for ticket %s not found", ticket.TicketID)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	receipt, ok := lo.Find(lo.Values(receiptsService.IssuedReceipts), func(r entities.IssueReceiptRequest) bool {
		return r.TicketID == ticket.TicketID
	})
	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest, idempotencyKey string) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}
