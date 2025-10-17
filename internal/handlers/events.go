package handlers

// websocketEvent is the custom type to represent all events that can be sent over the websocket connection.
type websocketEvent string

const dotCreated websocketEvent = "DotCreated"

// websocketMessage is the schema all websocket messages follow.
type websocketMessage struct {
	Event websocketEvent `json:"event"`
	Data  any            `json:"data"`
}
