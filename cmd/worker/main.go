package main

import (
	"context"

	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
	"github.com/a-tal/esi-isk/isk/worker"
)

func main() {
	worker.Run(cx.NewOptions(api.NewProvider(context.Background())))
}
