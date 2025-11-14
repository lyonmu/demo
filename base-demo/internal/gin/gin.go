package gin

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/lyonmu/demo/base-demo/internal/metrics"
	"github.com/soheilhy/cmux"
)

func NewGin(m cmux.CMux) error {

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(cors.Default())
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	if err := metrics.RegisterMetrics(r); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register metrics: %v", err)
		return err
	}

	if gin.Mode() != gin.ReleaseMode {
		pprof.Register(r)
	}

	go (&http.Server{Handler: r}).Serve(m.Match(cmux.HTTP1Fast()))

	return nil
}
