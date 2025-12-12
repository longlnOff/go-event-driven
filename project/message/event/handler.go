package event

type Handler struct {
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
}

func NewEventHandler(
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
) *Handler {
	if spreadsheetsAPI == nil {
		panic("missing spreadsheetsAPI")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	return &Handler{
		spreadsheetsAPI: spreadsheetsAPI,
		receiptsService: receiptsService,
	}
}
