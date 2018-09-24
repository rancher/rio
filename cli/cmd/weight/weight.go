package weight

import (
	"fmt"
	"strconv"
	"strings"

	service2 "github.com/rancher/rio/cli/pkg/service"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Weight struct {
}

func (w *Weight) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range app.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scaleStr = strings.TrimSuffix(scaleStr, "%")

		if scaleStr == "" {
			return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: mystack/myservice=10%%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		service, err := service2.Lookup(ctx, name)
		if err != nil {
			return err
		}

		data := map[string]interface{}{}
		values.PutValue(data, int64(scale),
			client.ServiceFieldWeight)

		_, err = ctx.Client.Service.Update(service, data)
		if err != nil {
			return err
		}

		waiter.Add(&service.Resource)
	}

	return waiter.Wait()
}
