package mcpserver

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/service"
)

// toolError turns an error from the service or repository layer into one that
// is safe to show the model and the user. It is the MCP counterpart of the REST
// layer's respondRepoError:
//   - a missing row becomes "<tool>: not found";
//   - a rejected business rule (service.ErrValidation) keeps its own message,
//     since that message is written for the caller;
//   - anything else is logged here and replaced by "<tool>: internal error",
//     so database details (table and column names, driver messages) never
//     reach the client.
//
// publicMessage names what failed, normally the tool name. Call it only with
// a non-nil err.
func toolError(publicMessage string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: not found", publicMessage)
	}
	if errors.Is(err, service.ErrValidation) {
		return err
	}
	log.Printf("%s: %v", publicMessage, err)
	return fmt.Errorf("%s: internal error", publicMessage)
}
