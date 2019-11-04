package scale

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/merr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Scale struct {
}

func scaleParam(scaleStr string) (*int32, *int32, bool, error) {
	min, max := kv.Split(scaleStr, "-")
	minScale, err := strconv.Atoi(min)
	if err != nil {
		return nil, nil, false, errors.Wrapf(err, "failed to parse %s", scaleStr)
	}

	if max == "" {
		return util.ToInt32(minScale), new(int32), false, nil
	}

	maxScale, err := strconv.Atoi(max)
	if err != nil {
		return nil, nil, false, errors.Wrapf(err, "failed to parse %s", scaleStr)
	}

	return util.ToInt32(minScale), util.ToInt32(maxScale), true, nil
}

func (s *Scale) Run(ctx *clicontext.CLIContext) error {
	var errors []error

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		if strings.TrimSpace(scaleStr) == "" {
			return fmt.Errorf("missing scale, format SERVICE=SCALE")
		}
		err := ctx.Update(name, func(obj runtime.Object) error {
			min, max, autoscale, err := scaleParam(scaleStr)
			if err != nil {
				return err
			}

			switch v := obj.(type) {
			case *v1.Service:
				if autoscale {
					if v.Spec.Autoscale == nil {
						v.Spec.Autoscale = &v1.AutoscaleConfig{
							Concurrency: 10,
						}
					}
					v.Spec.Autoscale.MinReplicas = min
					v.Spec.Autoscale.MaxReplicas = max
				} else {
					v.Spec.Autoscale = nil
					v.Spec.Replicas = util.ToInt(min)
				}
			case *appsv1.Deployment:
				v.Spec.Replicas = min
			default:
				return fmt.Errorf("non-scalable type %s", obj.GetObjectKind().GroupVersionKind().String())
			}

			return nil
		})
		errors = append(errors, err)
	}

	return merr.NewErrors(errors...)
}
