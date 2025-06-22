//go:build wasm || wasip1

package providers

func StartServer() error {
	CreateRequestHandler(
		NewProtocolHandlers(),
		NewWorkspaceHandlers(),
		NewTreeHandlers(),
		NewConfigurationHandlers(),
	)

	return nil
}
