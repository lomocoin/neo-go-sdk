package response

type (
	// Error represents the JSON schema of a response from a NEO node, where the expected
	// an error message.
	Error struct {
		ID      int    `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
)
