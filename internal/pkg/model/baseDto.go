package model

type RespondedBody struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

func Ok() RespondedBody {
	return RespondedBody{ErrorCode: 0, Message: "ok"}
}
