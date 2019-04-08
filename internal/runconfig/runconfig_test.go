// Copyright 2019 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package runconfig

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sorintlab/agola/internal/config"
	rstypes "github.com/sorintlab/agola/internal/services/runservice/types"
	"github.com/sorintlab/agola/internal/services/types"
	"github.com/sorintlab/agola/internal/util"
)

var uuid = &util.TestUUIDGenerator{}

func TestGenTasksLevels(t *testing.T) {
	type task struct {
		ID      string
		Level   int
		Depends []*rstypes.RunConfigTaskDepend
	}
	tests := []struct {
		name string
		in   []task
		out  []task
		err  error
	}{
		{
			name: "test single task",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
			},
			out: []task{
				{
					ID:    "1",
					Level: 0,
				},
			},
		},
		{
			name: "test multiple root tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
				},
			},
			out: []task{
				{
					ID:    "1",
					Level: 0,
				},
				{
					ID:    "2",
					Level: 0,
				},
			},
		},
		{
			name: "test dependency between two tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			out: []task{
				{
					ID:    "1",
					Level: 0,
				},
				{
					ID:    "2",
					Level: 1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
		},
		{
			name: "Test circular dependency between two tasks: a -> b -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			err: fmt.Errorf("circular dependency detected"),
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			err: fmt.Errorf("circular dependency detected"),
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> b",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
			},
			err: fmt.Errorf("circular dependency detected"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRunConfig := &rstypes.RunConfig{Tasks: map[string]*rstypes.RunConfigTask{}}
			for _, t := range tt.in {
				inRunConfig.Tasks[t.ID] = &rstypes.RunConfigTask{
					ID:      t.ID,
					Level:   t.Level,
					Depends: t.Depends,
				}

			}
			outRunConfig := &rstypes.RunConfig{Tasks: map[string]*rstypes.RunConfigTask{}}
			for _, t := range tt.out {
				outRunConfig.Tasks[t.ID] = &rstypes.RunConfigTask{
					ID:      t.ID,
					Level:   t.Level,
					Depends: t.Depends,
				}

			}
			if err := GenTasksLevels(inRunConfig); err != nil {
				if err.Error() != tt.err.Error() {
					t.Fatalf("got error: %v, want error: %v", err, tt.err)
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("got nil error, want error: %v", tt.err)
			}
			if !reflect.DeepEqual(inRunConfig.Tasks, outRunConfig.Tasks) {
				t.Fatalf("got %s, expected %s", util.Dump(inRunConfig), util.Dump(outRunConfig))
			}
		})
	}
}

func TestGetAllParents(t *testing.T) {
	type task struct {
		ID      string
		Level   int
		Depends []*rstypes.RunConfigTaskDepend
	}
	tests := []struct {
		name string
		in   []task
		out  map[string][]string
	}{
		{
			name: "test single task",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
			},
			out: map[string][]string{
				"1": []string{},
			},
		},
		{
			name: "test multiple root tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
				},
			},
			out: map[string][]string{
				"1": []string{},
				"2": []string{},
			},
		},
		{
			name: "test dependency from a task to itself",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			out: map[string][]string{
				"1": []string{"1"},
			},
		},
		{
			name: "test dependency between two tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			out: map[string][]string{
				"1": []string{},
				"2": []string{"1"},
			},
		},
		{
			name: "Test dependency between 5 tasks: a -> (b, c) -> (d, e)",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "4",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "5",
						},
					},
				},
				{
					ID:    "4",
					Level: -1,
				},
				{
					ID:    "5",
					Level: -1,
				},
			},
			out: map[string][]string{
				"1": []string{"2", "3", "4", "5"},
				"2": []string{"4"},
				"3": []string{"5"},
				"4": []string{},
				"5": []string{},
			},
		},
		{
			name: "Test circular dependency between two tasks: a -> b -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			out: map[string][]string{
				"1": []string{"2", "1"},
				"2": []string{"1", "2"},
			},
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			out: map[string][]string{
				"1": []string{"2", "3", "1"},
				"2": []string{"3", "1", "2"},
				"3": []string{"1", "2", "3"},
			},
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> b",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
			},
			out: map[string][]string{
				"1": []string{"2", "3"},
				"2": []string{"3", "2"},
				"3": []string{"2", "3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRunConfig := &rstypes.RunConfig{Tasks: map[string]*rstypes.RunConfigTask{}}
			for _, t := range tt.in {
				inRunConfig.Tasks[t.ID] = &rstypes.RunConfigTask{
					ID:      t.ID,
					Level:   t.Level,
					Depends: t.Depends,
				}

			}

			for _, task := range inRunConfig.Tasks {
				allParents := GetAllParents(inRunConfig, task)

				allParentsList := []string{}
				for _, p := range allParents {
					allParentsList = append(allParentsList, p.ID)
				}
				if !util.CompareStringSliceNoOrder(tt.out[task.ID], allParentsList) {
					t.Fatalf("task: %s, got %s, expected %s", task.ID, util.Dump(allParentsList), util.Dump(tt.out[task.ID]))
				}
			}
		})
	}
}

func TestCheckRunConfig(t *testing.T) {
	type task struct {
		ID      string
		Level   int
		Depends []*rstypes.RunConfigTaskDepend
	}
	tests := []struct {
		name string
		in   []task
		err  error
	}{
		{
			name: "test single task",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
			},
		},
		{
			name: "test multiple root tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
				},
			},
		},
		{
			name: "test dependency between two tasks",
			in: []task{
				{
					ID:    "1",
					Level: -1,
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
		},
		{
			name: "Test circular dependency between two tasks: a -> b -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			err: &util.Errors{
				Errs: []error{
					errors.Errorf("circular dependency between task %q and tasks %q", "task1", "task2"),
					errors.Errorf("circular dependency between task %q and tasks %q", "task2", "task1"),
				},
			},
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> a",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "1",
						},
					},
				},
			},
			err: &util.Errors{
				Errs: []error{
					errors.Errorf("circular dependency between task %q and tasks %q", "task1", "task3"),
					errors.Errorf("circular dependency between task %q and tasks %q", "task2", "task1"),
					errors.Errorf("circular dependency between task %q and tasks %q", "task3", "task2"),
				},
			},
		},
		{
			name: "Test circular dependency between 3 tasks: a -> b -> c -> b",
			in: []task{
				{
					ID:    "1",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
				{
					ID:    "2",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "3",
						},
					},
				},
				{
					ID:    "3",
					Level: -1,
					Depends: []*rstypes.RunConfigTaskDepend{
						&rstypes.RunConfigTaskDepend{
							TaskID: "2",
						},
					},
				},
			},
			err: &util.Errors{
				Errs: []error{
					errors.Errorf("circular dependency between task %q and tasks %q", "task2", "task3"),
					errors.Errorf("circular dependency between task %q and tasks %q", "task3", "task2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRunConfig := &rstypes.RunConfig{Tasks: map[string]*rstypes.RunConfigTask{}}
			for _, t := range tt.in {
				inRunConfig.Tasks[t.ID] = &rstypes.RunConfigTask{
					Name:    fmt.Sprintf("task%s", t.ID),
					ID:      t.ID,
					Level:   t.Level,
					Depends: t.Depends,
				}

			}

			if err := CheckRunConfig(inRunConfig); err != nil {
				if errs, ok := err.(*util.Errors); ok {
					if !errs.Equal(tt.err) {
						t.Fatalf("got error: %v, want error: %v", err, tt.err)
					}
				} else {
					if err.Error() != tt.err.Error() {
						t.Fatalf("got error: %v, want error: %v", err, tt.err)
					}
				}
				return
			}
			if tt.err != nil {
				t.Fatalf("got nil error, want error: %v", tt.err)
			}
		})
	}
}

func TestGenRunConfig(t *testing.T) {
	tests := []struct {
		name      string
		in        *config.Config
		env       map[string]string
		variables map[string]string
		out       *rstypes.RunConfig
	}{
		{
			name: "test runconfig generation",
			in: &config.Config{
				Runtimes: map[string]*config.Runtime{
					"runtime01": &config.Runtime{
						Name: "runtime01",
						Type: "pod",
						Arch: "",
						Containers: []*config.Container{
							&config.Container{
								Image: "image01",
								Environment: map[string]config.EnvVar{
									"ENV01":             config.EnvVar{Type: config.EnvVarTypeString, Value: "ENV01"},
									"ENVFROMVARIABLE01": config.EnvVar{Type: config.EnvVarTypeFromVariable, Value: "variable01"},
								},
								User: "",
							},
						},
					},
				},
				Tasks: map[string]*config.Task{
					"task01": &config.Task{
						Name:    "task01",
						Runtime: "runtime01",
						Environment: map[string]config.EnvVar{
							"ENV01":             config.EnvVar{Type: config.EnvVarTypeString, Value: "ENV01"},
							"ENVFROMVARIABLE01": config.EnvVar{Type: config.EnvVarTypeFromVariable, Value: "variable01"},
						},
						WorkingDir: "",
						Shell:      "",
						User:       "",
						Steps: []interface{}{
							&config.RunStep{
								Step: config.Step{
									Type: "run",
									Name: "command01",
								},
								Command: "command01",
							},
							&config.RunStep{
								Step: config.Step{
									Type: "run",
									Name: "name different than command",
								},
								Command: "command02",
							},
							&config.RunStep{
								Step: config.Step{
									Type: "run",
									Name: "command03",
								},
								Command: "command03",
								Environment: map[string]config.EnvVar{
									"ENV01":             config.EnvVar{Type: config.EnvVarTypeString, Value: "ENV01"},
									"ENVFROMVARIABLE01": config.EnvVar{Type: config.EnvVarTypeFromVariable, Value: "variable01"},
								},
							},
						},
					},
				},
				Pipelines: map[string]*config.Pipeline{
					"pipeline01": &config.Pipeline{
						Name: "pipeline01",
						Elements: map[string]*config.Element{
							"element01": &config.Element{
								Name:          "element01",
								Task:          "task01",
								Depends:       []*config.Depend{},
								IgnoreFailure: false,
								Approval:      false,
								When: &types.When{
									Branch: &types.WhenConditions{Include: []types.WhenCondition{{Match: "master"}}},
									Tag:    &types.WhenConditions{Include: []types.WhenCondition{{Match: "v1.x"}, {Match: "v2.x"}}},
									Ref: &types.WhenConditions{
										Include: []types.WhenCondition{{Match: "master"}},
										Exclude: []types.WhenCondition{{Match: "branch01", Type: types.WhenConditionTypeRegExp}, {Match: "branch02"}},
									},
								},
							},
						},
					},
				},
			},
			env: map[string]string{
				"ENV01": "ENVVALUE01",
			},
			variables: map[string]string{
				"variable01": "VARVALUE01",
			},
			out: &rstypes.RunConfig{
				Name: "pipeline01",
				Environment: map[string]string{
					"ENV01": "ENVVALUE01",
				},
				Tasks: map[string]*rstypes.RunConfigTask{
					uuid.New("element01").String(): &rstypes.RunConfigTask{
						ID:   uuid.New("element01").String(),
						Name: "element01", Depends: []*rstypes.RunConfigTaskDepend{},
						Runtime: &rstypes.Runtime{Type: rstypes.RuntimeType("pod"),
							Containers: []*rstypes.Container{
								{
									Image: "image01",
									Environment: map[string]string{
										"ENV01":             "ENV01",
										"ENVFROMVARIABLE01": "VARVALUE01",
									},
								},
							},
						},
						Environment: map[string]string{
							"ENV01":             "ENV01",
							"ENVFROMVARIABLE01": "VARVALUE01",
						},
						Steps: []interface{}{
							&rstypes.RunStep{Step: rstypes.Step{Type: "run", Name: "command01"}, Command: "command01", Environment: map[string]string{}},
							&rstypes.RunStep{Step: rstypes.Step{Type: "run", Name: "name different than command"}, Command: "command02", Environment: map[string]string{}},
							&rstypes.RunStep{Step: rstypes.Step{Type: "run", Name: "command03"}, Command: "command03", Environment: map[string]string{"ENV01": "ENV01", "ENVFROMVARIABLE01": "VARVALUE01"}},
						},
						Skip: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := GenRunConfig(uuid, tt.in, "pipeline01", tt.env, tt.variables, "", "", "")

			//if err != nil {
			//	t.Fatalf("unexpected error: %v", err)
			//}
			if diff := cmp.Diff(tt.out, out); diff != "" {
				t.Error(diff)
			}
		})
	}
}
