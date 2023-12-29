package htmx

const (
	HeaderTriggerAfterSettle = "HX-Trigger-After-Settle"
	HeaderReswap             = "HX-Reswap"
)

const (
	EventCreateSuccess = "createSuccess"
	EventUpdateSuccess = "updateSuccess"
	EventSaveSuccess   = "saveSuccess"
	EventDeleteSuccess = "deleteSuccess"
)

type EventSaveError struct {
	ErrorMessage string `json:"saveError"`
}

const (
	SwapNone = "none"
)
