package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

type FileServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewFileServiceClient(clients *clients.Clients) *FileServiceClient {
	if clients == nil {
		panic("NewFileServiceClient: clients is nil")
	}

	return &FileServiceClient{clients: clients}
}

func (c FileServiceClient) UpLoadFile(ctx context.Context, ticketFile string, body string) error {
	resp, err := c.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, ticketFile, body)
	if err != nil {
		return fmt.Errorf("failed to print ticket: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return nil
	case http.StatusCreated:
		// receipt was created
		return nil
	case http.StatusConflict:
		log.FromContext(ctx).With("file", ticketFile).Info("file already exists")
		return nil
	default:
		return fmt.Errorf("unexpected status code for POST File Service: %d", resp.StatusCode())
	}
}
