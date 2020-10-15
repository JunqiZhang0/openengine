package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/nsf/jsondiff"
)

type TestData string

func (t TestData) Reader() ([]byte, error) {
	tJSON, err := ioutil.ReadFile(filepath.Join("testdata", string(t)+".json"))
	if err != nil {
		return nil, fmt.Errorf("failed reading: %s", err)
	}
	return tJSON, nil
}

type TestSolution struct {
	Resource       Resource         `json:"resource"`
	System         System           `json:"system"`
	Provider       Provider         `json:"provider"`
	Provisioner    Provisioner      `json:"provisioner"`
	Resolved       bool             `json:"resolved"`
	ResolutionTree map[string]Param `json:"resolution_tree"`
	Parent         *Solution        `json:"parent"`
	Size           int              `json:"size"`
	Action         string           `json:"action"`
	Output         string           `json:"output"`
	Debug          bool             `json:"debug"`
}

func (s *Solution) UnmarshalJSON(b []byte) error {
	temp := &TestSolution{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	s.resolved = temp.Resolved
	s.resolutionTree = temp.ResolutionTree
	s.size = temp.Size
	s.action = temp.Action
	s.debug = temp.Debug
	s.parent = temp.Parent
	s.System = temp.System
	s.Resource = temp.Resource
	s.Provisioner = temp.Provisioner
	s.Provider = temp.Provider
	s.Output = temp.Output

	return nil
}

func TestSolution_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		test    TestData
		want    TestData
		wantErr bool
	}{
		{"empty", "solution.empty", "solution.empty", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJSON, _ := tt.test.Reader()
			wantJSON, _ := tt.want.Reader()
			testSolution := &Solution{}
			json.Unmarshal(testJSON, &testSolution)
			opt := jsondiff.DefaultConsoleOptions()
			got, err := testSolution.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result, diff := jsondiff.Compare(wantJSON, got, &opt); result.String() != "FullMatch" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestSolution_Run(t *testing.T) {
	tests := []struct {
		name    string
		test    TestData
		args    map[string]interface{}
		want    string
		wantErr bool
	}{
		{"empty", "solution.empty", map[string]interface{}{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJSON, _ := tt.test.Reader()
			testSolution := &Solution{}
			json.Unmarshal(testJSON, &testSolution)
			got, err := testSolution.Run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Fatalf("Run() got %v but, wanted %v", got, tt.want)
			}
		})
	}
}

func TestSolution_ToJson(t *testing.T) {
	tests := []struct {
		name string
		in   TestData
		out  TestData
	}{
		{"empty", "solution.empty", "solution.empty.ToJSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inJSON, err := tt.in.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			outJSON, err := tt.out.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			inSolution := &Solution{}
			json.Unmarshal(inJSON, inSolution)
			opt := jsondiff.DefaultConsoleOptions()
			if result, diff := jsondiff.Compare([]byte(inSolution.ToJSON()), outJSON, &opt); result.String() != "FullMatch" {
				t.Errorf("%v", diff)
			}
		})
	}
}

func TestSolution_decouple(t *testing.T) {
	tests := []struct {
		name string
		test TestData
		want []TestData
	}{
		{"empty", "solution.empty", []TestData{"solution.empty"}},
		{"implicit", "solution.implicit", []TestData{"solution.implicit.decouple"}},
	}
	opt := jsondiff.DefaultConsoleOptions()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var want = []Solution{}
			for _, file := range tt.want {
				wantJSON, _ := file.Reader()
				wantSolution := &Solution{}
				json.Unmarshal(wantJSON, &wantSolution)
				want = append(want, *wantSolution)
			}
			testJSON, _ := tt.test.Reader()
			testSolution := &Solution{}
			json.Unmarshal(testJSON, &testSolution)
			got := testSolution.decouple()
			gJSON, _ := json.Marshal(got)
			wJSON, _ := json.Marshal(want)
			if result, diff := jsondiff.Compare(gJSON, wJSON, &opt); result.String() != "FullMatch" {
				t.Errorf("%v", diff)
			}
		})
	}
}

func TestSolution_equals(t *testing.T) {
	tests := []struct {
		name     string
		testData TestData
		want     TestData
	}{
		{"emtpy", "solution.empty", "solution.empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inJSON, err := tt.testData.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			outJSON, err := tt.want.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			inSolution := &Solution{}
			json.Unmarshal(inJSON, inSolution)
			outSolution := &Solution{}
			json.Unmarshal(outJSON, outSolution)
			if !inSolution.equals(*outSolution) {
				t.Errorf("Solution not equal")
			}
		})
	}
}

func TestSolution_inLoop(t *testing.T) {
	tests := []struct {
		name     string
		testCase TestData
		arg      TestData
		want     bool
	}{
		{"empty", "solution.empty", "solution.empty", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJSON, err := tt.testCase.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			argJSON, err := tt.arg.Reader()
			if err != nil {
				t.Fatalf("%s", err)
			}
			testSolution := &Solution{}
			json.Unmarshal(testJSON, testSolution)
			argSolution := &Solution{}
			json.Unmarshal(argJSON, argSolution)
			if testSolution.inLoop(*argSolution) != tt.want {
				result := "'nt"
				if tt.want {
					result = ""
				}
				t.Errorf("Solution should%v be in loop", result)
			}
		})
	}
}

func TestSolution_resolveExplicit(t *testing.T) {
	tests := []struct {
		name string
		test TestData
		new  TestData
		want []string
	}{
		{"empty", "solution.empty", "solution.empty", nil},
		{"implicit", "solution.implicit", "solution.implicit", []string{"flavorRef", "name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJSON, _ := tt.test.Reader()
			newJSON, _ := tt.new.Reader()
			testSolution := &Solution{}
			err := json.Unmarshal(testJSON, &testSolution)
			if err != nil {
				t.Fatalf("%v", err)
			}
			opt := jsondiff.DefaultConsoleOptions()
			got := testSolution.resolveExplicit()
			sort.Strings(got)
			sort.Strings(tt.want)
			newnew, _ := json.Marshal(testSolution)
			if result, diff := jsondiff.Compare(newJSON, newnew, &opt); result.String() != "FullMatch" {
				t.Fatalf(diff)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("resolveExplicit() got %v but wanted %v", got, tt.want)
			}
		})
	}
}

func Test_combIntSlices(t *testing.T) {
	type args struct {
		seq [][]int
	}
	tests := []struct {
		name    string
		args    args
		wantOut [][]int
	}{
		{
			name: "zero",
			args: args{
				seq: [][]int{},
			},
			wantOut: nil,
		},
		{
			name: "[[1]]",
			args: args{
				seq: [][]int{{1}},
			},
			wantOut: nil,
		},
		{
			name: "[[1][2]]",
			args: args{
				seq: [][]int{{1}, {2}},
			},
			wantOut: [][]int{{1, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := combIntSlices(tt.args.seq); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("combIntSlices() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func Test_intersect(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{
				a: []string{},
				b: []string{},
			},
			want: []string{},
		},
		{
			name: "one",
			args: args{
				a: []string{"one"},
				b: []string{"one"},
			},
			want: []string{"one"},
		},
		{
			name: "none",
			args: args{
				a: []string{"one"},
				b: []string{"two"},
			},
			want: []string{},
		},
		{
			name: "empty a",
			args: args{
				a: []string{},
				b: []string{"two"},
			},
			want: []string{},
		},
		{
			name: "empty b",
			args: args{
				a: []string{"one"},
				b: []string{},
			},
			want: []string{},
		},
		{
			name: "intersect",
			args: args{
				a: []string{"one", "two"},
				b: []string{"two", "three"},
			},
			want: []string{"two"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intersect(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeRange(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "zero",
			args: args{
				min: 0,
				max: 0,
			},
			want: []int{0},
		},
		{
			name: "min=0",
			args: args{
				min: 0,
				max: 1,
			},
			want: []int{0, 1},
		},
		{
			name: "min>0",
			args: args{
				min: 1,
				max: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "min<0",
			args: args{
				min: -1,
				max: 1,
			},
			want: []int{-1, 0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeRange(tt.args.min, tt.args.max); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_solutionList_Len(t *testing.T) {
	tests := []struct {
		name string
		s    solutionList
		want int
	}{
		{
			name: "empty",
			s:    solutionList{},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_solutionList_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		s    solutionList
		args args
		want bool
	}{
		{
			name: "equal",
			s:    solutionList{Solution{size: 1}, Solution{size: 1}},
			args: args{
				i: 0,
				j: 1,
			},
			want: false,
		},
		{
			name: "bigger",
			s:    solutionList{Solution{size: 2}, Solution{size: 1}},
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_solutionList_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	s1 := Solution{size: 1}
	s2 := Solution{size: 2}
	tests := []struct {
		name string
		s    solutionList
		args args
		want solutionList
	}{
		{
			name: "simple",
			s:    solutionList{s1, s2},
			args: args{
				i: 0,
				j: 1,
			},
			want: solutionList{s2, s1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Swap(tt.args.i, tt.args.j)
			if tt.s[0].size != 2 {
				t.Errorf("Less() = %v, want %v", tt.s, tt.want)
			}
		})
	}
}
