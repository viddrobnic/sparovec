package routes

const (
	HtmxHeaderTriggerAfterSettle = "HX-Trigger-After-Settle"
	HtmxHeaderReswap             = "HX-Reswap"
)

const (
	HtmxEventCreateSuccess = "createSuccess"
	HtmxEventUpdateSuccess = "updateSuccess"
	HtmxEventSaveSuccess   = "saveSuccess"
	HtmxEventDeleteSuccess = "deleteSuccess"
)

type HtmxEventSaveError struct {
	ErrorMessage string `json:"saveError"`
}

const (
	HtmxSwapNone = "none"
)
