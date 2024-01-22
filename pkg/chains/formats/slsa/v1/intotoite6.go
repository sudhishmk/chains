/*
Copyright 2021 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"

	"github.com/tektoncd/chains/pkg/chains/formats"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/internal/slsaconfig"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/v1/pipelinerun"
	"github.com/tektoncd/chains/pkg/chains/formats/slsa/v1/taskrun"
	"github.com/tektoncd/chains/pkg/chains/objects"
	"github.com/tektoncd/chains/pkg/config"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const (
	PayloadTypeInTotoIte6 = formats.PayloadTypeInTotoIte6
	PayloadTypeSlsav1     = formats.PayloadTypeSlsav1
)

func init() {
	formats.RegisterPayloader(PayloadTypeInTotoIte6, NewFormatter)
	formats.RegisterPayloader(PayloadTypeSlsav1, NewFormatter)
}

type InTotoIte6 struct {
	slsaConfig *slsaconfig.SlsaConfig
}

func NewFormatter(cfg config.Config) (formats.Payloader, error) {
	return &InTotoIte6{
		slsaConfig: &slsaconfig.SlsaConfig{
			BuilderID:             cfg.Builder.ID,
			DeepInspectionEnabled: cfg.Artifacts.PipelineRuns.DeepInspectionEnabled,
		},
	}, nil
}

func (i *InTotoIte6) Wrap() bool {
	return true
}

func (i *InTotoIte6) CreatePayload(ctx context.Context, obj interface{}) (interface{}, error) {
	switch v := obj.(type) {
	case *objects.TaskRunObjectV1:
		tro := obj.(*objects.TaskRunObjectV1)
		trV1Beta1 := &v1beta1.TaskRun{} //nolint:staticcheck
		if err := trV1Beta1.ConvertFrom(ctx, tro.GetObject().(*v1.TaskRun)); err != nil {
			return nil, fmt.Errorf("error converting Tekton TaskRun from version v1 to v1beta1: %s", err)
		}
		return taskrun.GenerateAttestation(ctx, objects.NewTaskRunObjectV1Beta1(trV1Beta1), i.slsaConfig)
	case *objects.PipelineRunObjectV1:
		pro := obj.(*objects.PipelineRunObjectV1)
		prV1Beta1 := &v1beta1.PipelineRun{} //nolint:staticcheck
		if err := prV1Beta1.ConvertFrom(ctx, pro.GetObject().(*v1.PipelineRun)); err != nil {
			return nil, fmt.Errorf("error converting Tekton PipelineRun from version v1 to v1beta1: %s", err)
		}
		proV1Beta1 := objects.NewPipelineRunObjectV1Beta1(prV1Beta1)
		trs := pro.GetTaskRuns()
		for _, tr := range trs {
			trV1Beta1 := &v1beta1.TaskRun{} //nolint:staticcheck
			if err := trV1Beta1.ConvertFrom(ctx, tr); err != nil {
				return nil, fmt.Errorf("error converting Tekton TaskRun from version v1 to v1beta1: %s", err)
			}
			proV1Beta1.AppendTaskRun(trV1Beta1)
		}
		return pipelinerun.GenerateAttestation(ctx, proV1Beta1, i.slsaConfig)
	case *objects.TaskRunObjectV1Beta1:
		return taskrun.GenerateAttestation(ctx, v, i.slsaConfig)
	case *objects.PipelineRunObjectV1Beta1:
		return pipelinerun.GenerateAttestation(ctx, v, i.slsaConfig)
	default:
		return nil, fmt.Errorf("intoto does not support type: %s", v)
	}
}

func (i *InTotoIte6) Type() config.PayloadType {
	return formats.PayloadTypeSlsav1
}
