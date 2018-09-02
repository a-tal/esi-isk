package main

import (
	"context"

	"github.com/a-tal/esi-isk/isk"
	"github.com/a-tal/esi-isk/isk/api"
	"github.com/a-tal/esi-isk/isk/cx"
)

func main() {
	isk.RunServer(cx.NewOptions(api.NewProvider(context.Background())))
}
