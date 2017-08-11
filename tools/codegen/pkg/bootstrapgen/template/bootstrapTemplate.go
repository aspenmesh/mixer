// Copyright 2016 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package template

// InterfaceTemplate defines the template used to generate the adapter
// interfaces for Mixer for a given aspect.
var InterfaceTemplate = `// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// THIS FILE IS AUTOMATICALLY GENERATED.

package {{.PkgName}}

import (
	"github.com/golang/protobuf/proto"
	"fmt"
	"istio.io/mixer/pkg/attribute"
	rpc "github.com/googleapis/googleapis/google/rpc"
	"github.com/hashicorp/go-multierror"
	"istio.io/mixer/pkg/expr"
	"github.com/golang/glog"
	"istio.io/mixer/pkg/status"
	"istio.io/mixer/pkg/adapter"
	"istio.io/api/mixer/v1/config/descriptor"
	"istio.io/mixer/pkg/template"
	adptTmpl "istio.io/mixer/pkg/adapter/template"
	{{range .TemplateModels}}
		"{{.PackageImportPath}}"
	{{end}}
)

var (
	SupportedTmplInfo = map[string]template.Info {
	{{range .TemplateModels}}
		{{.GoPackageName}}.TemplateName: {
			CtrCfg:  &{{.GoPackageName}}.InstanceParam{},
			Variety:   adptTmpl.{{.VarietyName}},
			BldrName:  "{{.PackageImportPath}}.{{.Name}}HandlerBuilder",
			HndlrName: "{{.PackageImportPath}}.{{.Name}}Handler",
			SupportsTemplate: func(hndlrBuilder adapter.HandlerBuilder) bool {
				_, ok := hndlrBuilder.({{.GoPackageName}}.{{.Name}}HandlerBuilder)
				return ok
			},
			HandlerSupportsTemplate: func(hndlr adapter.Handler) bool {
				_, ok := hndlr.({{.GoPackageName}}.{{.Name}}Handler)
				return ok
			},
			InferType: func(cp proto.Message, tEvalFn template.TypeEvalFn) (proto.Message, error) {
				var err error = nil
				cpb := cp.(*{{.GoPackageName}}.InstanceParam)
				infrdType := &{{.GoPackageName}}.Type{}

				{{range .TemplateMessage.Fields}}
					{{if containsValueType .GoType}}
						{{if .GoType.IsMap}}
							infrdType.{{.GoName}} = make(map[{{.GoType.MapKey.Name}}]istio_mixer_v1_config_descriptor.ValueType, len(cpb.{{.GoName}}))
							for k, v := range cpb.{{.GoName}} {
								if infrdType.{{.GoName}}[k], err = tEvalFn(v); err != nil {
									return nil, err
								}
							}
						{{else}}
							if cpb.{{.GoName}} == "" {
								return nil, fmt.Errorf("expression for field {{.GoName}} cannot be empty")
							}
							if infrdType.{{.GoName}}, err = tEvalFn(cpb.{{.GoName}}); err != nil {
								return nil, err
							}
						{{end}}
					{{else}}
						{{if .GoType.IsMap}}
							for _, v := range cpb.{{.GoName}} {
								if t, e := tEvalFn(v); e != nil || t != {{getValueType .GoType.MapValue}} {
									if e != nil {
										return nil, fmt.Errorf("failed to evaluate expression for field {{.GoName}}: %v", e)
									}
									return nil, fmt.Errorf("error type checking for field {{.GoName}}: Evaluated expression type %v want %v", t, {{getValueType .GoType.MapValue}})
								}
							}
						{{else}}
							if cpb.{{.GoName}} == "" {
								return nil, fmt.Errorf("expression for field {{.GoName}} cannot be empty")
							}
							if t, e := tEvalFn(cpb.{{.GoName}}); e != nil || t != {{getValueType .GoType}} {
								if e != nil {
									return nil, fmt.Errorf("failed to evaluate expression for field {{.GoName}}: %v", e)
								}
								return nil, fmt.Errorf("error type checking for field {{.GoName}}: Evaluated expression type %v want %v", t, {{getValueType .GoType}})
							}
						{{end}}
					{{end}}
				{{end}}
				_ = cpb
				return infrdType, err
			},
			ConfigureType: func(types map[string]proto.Message, builder *adapter.HandlerBuilder) error {
				// Mixer framework should have ensured the type safety.
				castedBuilder := (*builder).({{.GoPackageName}}.{{.Name}}HandlerBuilder)
				castedTypes := make(map[string]*{{.GoPackageName}}.Type, len(types))
				for k, v := range types {
					// Mixer framework should have ensured the type safety.
					v1 := v.(*{{.GoPackageName}}.Type)
					castedTypes[k] = v1
				}
				return castedBuilder.Configure{{.Name}}Handler(castedTypes)
			},
			{{if eq .VarietyName "TEMPLATE_VARIETY_REPORT"}}
				ProcessReport: func(insts map[string]proto.Message, attrs attribute.Bag, mapper expr.Evaluator, handler adapter.Handler) rpc.Status {
					result := &multierror.Error{}
					var instances []*{{.GoPackageName}}.Instance

					castedInsts := make(map[string]*{{.GoPackageName}}.InstanceParam, len(insts))
					for k, v := range insts {
						v1 := v.(*{{.GoPackageName}}.InstanceParam)
						castedInsts[k] = v1
					}
					for name, md := range castedInsts {
						{{range .TemplateMessage.Fields}}
							{{if .GoType.IsMap}}
								{{.GoName}}, err := template.EvalAll(md.{{.GoName}}, attrs, mapper)
							{{else}}
								{{.GoName}}, err := mapper.Eval(md.{{.GoName}}, attrs)
							{{end}}
								if err != nil {
									result = multierror.Append(result, fmt.Errorf("failed to eval {{.GoName}} for instance '%s': %v", name, err))
									continue
								}
						{{end}}

						instances = append(instances, &{{.GoPackageName}}.Instance{
							Name:       name,
							{{range .TemplateMessage.Fields}}
								{{if containsValueType .GoType}}
									{{.GoName}}: {{.GoName}},
								{{else}}
									{{if .GoType.IsMap}}
										{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
											res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
											for k, v := range m {
												res[k] = v.({{.GoType.MapValue.Name}})
											}
											return res
										}({{.GoName}}),
									{{else}}
										{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),
									{{end}}
								{{end}}
							{{end}}
						})
						_ = md
					}

					if err := handler.({{.GoPackageName}}.{{.Name}}Handler).Handle{{.Name}}(instances); err != nil {
						result = multierror.Append(result, fmt.Errorf("failed to report all values: %v", err))
					}

					err := result.ErrorOrNil()
					if err != nil {
						return status.WithError(err)
					}

					return status.OK
				},
				ProcessCheck: nil,
				ProcessQuota: nil,
			{{else if eq .VarietyName "TEMPLATE_VARIETY_CHECK"}}
				ProcessCheck: func(instName string, inst proto.Message, attrs attribute.Bag, mapper expr.Evaluator,
				handler adapter.Handler) (rpc.Status, adapter.CacheabilityInfo) {
					var found bool
					var err error

					castedInst := inst.(*{{.GoPackageName}}.InstanceParam)
					var instances []*{{.GoPackageName}}.Instance
					{{range .TemplateMessage.Fields}}
						{{if .GoType.IsMap}}
							{{.GoName}}, err := template.EvalAll(castedInst.{{.GoName}}, attrs, mapper)
						{{else}}
							{{.GoName}}, err := mapper.Eval(castedInst.{{.GoName}}, attrs)
						{{end}}
							if err != nil {
								return status.WithError(err), adapter.CacheabilityInfo{}
							}
					{{end}}

					instance := &{{.GoPackageName}}.Instance{
						Name:	instName,
						{{range .TemplateMessage.Fields}}
							{{if containsValueType .GoType}}
								{{.GoName}}: {{.GoName}},
							{{else}}
								{{if .GoType.IsMap}}
									{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
										res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
										for k, v := range m {
											res[k] = v.({{.GoType.MapValue.Name}})
										}
										return res
									}({{.GoName}}),
								{{else}}
									{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),
								{{end}}
							{{end}}
						{{end}}
					}
					_ = castedInst

					var cacheInfo adapter.CacheabilityInfo
					if found, cacheInfo, err = handler.({{.GoPackageName}}.{{.Name}}Handler).Handle{{.Name}}(instance); err != nil {
						return status.WithError(err), adapter.CacheabilityInfo{}
					}

					if found {
						return status.OK, cacheInfo
					}

					return status.WithPermissionDenied(fmt.Sprintf("%s rejected", instances)), adapter.CacheabilityInfo{}
				},
				ProcessReport: nil,
				ProcessQuota: nil,
			{{else}}
				ProcessQuota: func(quotaName string, inst proto.Message, attrs attribute.Bag, mapper expr.Evaluator, handler adapter.Handler,
				qma adapter.QuotaRequestArgs) (rpc.Status, adapter.CacheabilityInfo, adapter.QuotaResult) {
					castedInst := inst.(*{{.GoPackageName}}.InstanceParam)
					{{range .TemplateMessage.Fields}}
						{{if .GoType.IsMap}}
							{{.GoName}}, err := template.EvalAll(castedInst.{{.GoName}}, attrs, mapper)
						{{else}}
							{{.GoName}}, err := mapper.Eval(castedInst.{{.GoName}}, attrs)
						{{end}}
							if err != nil {
								msg := fmt.Sprintf("failed to eval {{.GoName}} for instance '%s': %v", quotaName, err)
								glog.Error(msg)
								return status.WithInvalidArgument(msg), adapter.CacheabilityInfo{}, adapter.QuotaResult{}
							}
					{{end}}

					instance := &{{.GoPackageName}}.Instance{
						Name:       quotaName,
						{{range .TemplateMessage.Fields}}
							{{if containsValueType .GoType}}
								{{.GoName}}: {{.GoName}},
							{{else}}
								{{if .GoType.IsMap}}
									{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
										res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
										for k, v := range m {
											res[k] = v.({{.GoType.MapValue.Name}})
										}
										return res
									}({{.GoName}}),
								{{else}}
									{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),
								{{end}}
							{{end}}
						{{end}}
					}

					var qr adapter.QuotaResult
					var cacheInfo adapter.CacheabilityInfo
					if qr, cacheInfo, err = handler.({{.GoPackageName}}.{{.Name}}Handler).Handle{{.Name}}(instance, qma); err != nil {
						glog.Errorf("Quota allocation failed: %v", err)
						return status.WithError(err), adapter.CacheabilityInfo{}, adapter.QuotaResult{}
					}
					if qr.Amount == 0 {
						msg := fmt.Sprintf("Unable to allocate %v units from quota %s", qma.QuotaAmount, quotaName)
						glog.Warning(msg)
						return status.WithResourceExhausted(msg), adapter.CacheabilityInfo{}, adapter.QuotaResult{}
					}
					if glog.V(2) {
						glog.Infof("Allocated %v units from quota %s", qma.QuotaAmount, quotaName)
					}
					return status.OK, cacheInfo, qr
				},
				ProcessReport: nil,
				ProcessCheck: nil,
			{{end}}
		},
	{{end}}
	}
)

`
