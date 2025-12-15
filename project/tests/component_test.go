package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"testing"
	ticketsAdapter "tickets/adapters"
	ticketsEntity "tickets/entities"
	ticketsHttp "tickets/http"
	ticketsMessage "tickets/message"
	ticketsService "tickets/service"
	"time"

	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	// place for your tests!
	redisClient := ticketsMessage.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spreadsheetsAPI := ticketsAdapter.SpreadsheetsAPIStub{}
	receiptsService := ticketsAdapter.ReceiptsServiceStub{}

	go func() {
		svc := ticketsService.New(
			&spreadsheetsAPI,
			&receiptsService,
			redisClient,
		)
		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()
	waitForHttpServer(t)

	// Test confirmed ticket
	tk1 := ticketsHttp.TicketStatusRequest{
		TicketID: shortuuid.New(),
		Status:   "confirmed",
		Price: ticketsEntity.Money{
			Currency: "USD",
			Amount:   "12.0",
		},
		CustomerEmail: "longln@gmail.com",
	}

	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{tk1},
	})
	assertReceiptForTicketIssued(t, &receiptsService, tk1)
	assertTicketAppendedToTracker(t, &spreadsheetsAPI, tk1, "tickets-to-print")

	// Test canceled ticket
	tk1.Status = "canceled"
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{tk1},
	})
	assertTicketAppendedToTracker(t, &spreadsheetsAPI, tk1, "tickets-to-refund")
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

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest) {
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
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(
	t *testing.T,
	receiptsService *ticketsAdapter.ReceiptsServiceStub,
	ticket ticketsHttp.TicketStatusRequest,
) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceipts)

			assert.Greater(t, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt ticketsEntity.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertTicketAppendedToTracker(
	t *testing.T,
	spreadsheetsAPI *ticketsAdapter.SpreadsheetsAPIStub,
	ticket ticketsHttp.TicketStatusRequest,
	sheetName string,
) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			AppendedRows := len(spreadsheetsAPI.Rows)
			parentT.Log("appended rows", AppendedRows)

			assert.Greater(t, AppendedRows, 0, "no ticket appended")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var ok bool = false
	var standardRow []string = nil
	for _, row := range spreadsheetsAPI.Rows[sheetName] {
		standardRow = []string{ticket.TicketID, ticket.CustomerEmail, ticket.Price.Amount, ticket.Price.Currency}
		if slices.Equal(standardRow, row) {
			ok = true
			break
		}
	}

	require.Truef(t, ok, "tracker for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, standardRow[0])
	assert.Equal(t, ticket.CustomerEmail, standardRow[1])
	assert.Equal(t, ticket.Price.Amount, standardRow[2])
	assert.Equal(t, ticket.Price.Currency, standardRow[3])
}
