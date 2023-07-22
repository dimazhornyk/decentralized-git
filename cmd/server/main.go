package main

import (
	"fmt"
	"git-test/internal/common"
	"git-test/internal/connectors"
	"git-test/internal/logic"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			common.NewConfig,
			connectors.NewStorage,
			connectors.NewRepository,
			logic.NewTokenManager,
			logic.NewService,
			logic.NewGin,
		),
		fx.Invoke(func(e *gin.Engine, config *common.Config) {
			if err := e.Run(fmt.Sprintf(":%d", config.Port)); err != nil {
				panic(err)
			}
		}),
	).Run()

}
