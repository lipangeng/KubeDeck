package plugins

import (
	"fmt"

	"kubedeck/backend/pkg/sdk"
)

func ExecuteAction(
	providers []sdk.CapabilityProvider,
	request sdk.ActionExecutionRequest,
) (sdk.ActionExecutionResult, error) {
	for _, provider := range providers {
		executor, ok := provider.(sdk.ActionExecutor)
		if !ok {
			continue
		}
		result, err := executor.ExecuteAction(request)
		if err == nil || result.Accepted || result.Summary != "" {
			return result, err
		}
	}
	return sdk.ActionExecutionResult{}, fmt.Errorf(
		"no action executor registered for %s/%s",
		request.WorkflowDomainID,
		request.ActionID,
	)
}
