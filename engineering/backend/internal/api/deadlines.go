package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// liftWriteDeadline removes the server's write timeout for one route.
//
// The server gives every response a fixed time to be written, and Go starts
// that clock when the request arrives, before the handler runs. That suits
// routes that answer quickly. A route that does slow work first, such as
// batch-save, which waits for ingestion to finish before it responds, would
// find the deadline already expired when it writes: the work has committed, but
// the client never receives the result.
//
// Use it after authentication, so unauthenticated requests keep the deadline.
// It only lifts the write deadline: the header, read and idle timeouts still
// apply. If the connection cannot change its deadline the server's own value
// stays in force, which is logged.
func liftWriteDeadline() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{}); err != nil {
			log.Printf("could not lift the write deadline for %s %s: %v", c.Request.Method, c.FullPath(), err)
		}
		c.Next()
	}
}
