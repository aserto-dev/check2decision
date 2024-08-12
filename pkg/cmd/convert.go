package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/check2decision/api"
	"github.com/aserto-dev/check2decision/pkg/resource"
	"github.com/aserto-dev/check2decision/pkg/version"
	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	aza2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConvertCmd struct {
	Input        string `flag:"" short:"i" help:"assertions file path" xor:"input,stdin"`
	Output       string `flag:"" short:"o" help:"decisions file path"`
	PolicyName   string `flag:"" default:"policy-rebac" help:"policy name"`
	PolicyPath   string `flag:"" default:"rebac.check" help:"policy package path"`
	PolicyRule   string `flag:"" default:"allowed" help:"policy rule name"`
	IdentityType string `flag:"" default:"sub" help:"identity type (sub|jwt|manual|none)" enum:"sub,jwt,manual,none"`
	StdIn        bool   `flag:"" name:"stdin" help:"read input from StdIn" xor:"input,stdin"`
	Version      bool   `flag:"" help:"version info"`
}

func (cmd *ConvertCmd) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if cmd.Version {
		fmt.Fprintln(os.Stdout, version.GetInfo().String())
		return nil
	}

	checkAssertions, err := cmd.load(ctx)
	if err != nil {
		return err
	}

	decisionAssertions := cmd.transform(ctx, checkAssertions)

	if err := cmd.persist(ctx, decisionAssertions); err != nil {
		return err
	}

	return nil
}

type CheckAssertions struct {
	Assertions []*api.CheckAssertion `json:"assertions"`
}

func (cmd *ConvertCmd) load(_ context.Context) (*CheckAssertions, error) {
	a := CheckAssertions{}

	var r *os.File

	if cmd.StdIn {
		r = os.Stdin
	}
	if cmd.Input != "" {
		fi, err := os.Stat(cmd.Input)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			return nil, status.Errorf(codes.NotFound, cmd.Input)
		}
		r, err = os.Open(cmd.Input)
		if err != nil {
			return nil, err
		}
	}
	defer r.Close()

	dec := json.NewDecoder(r)
	if err := dec.Decode(&a); err != nil {
		return nil, err
	}

	return &a, nil
}

type DecisionAssertions struct {
	Assertions []*api.DecisionAssertion `json:"assertions"`
}

func (cmd *ConvertCmd) transform(_ context.Context, a *CheckAssertions) *DecisionAssertions {
	d := DecisionAssertions{}

	identityType := aza2.IdentityType(aza2.IdentityType_value["IDENTITY_TYPE_"+strings.ToUpper(cmd.IdentityType)])

	for i := 0; i < len(a.Assertions); i++ {
		decision := api.DecisionAssertion{
			CheckDecision: &az2.IsRequest{
				IdentityContext: &aza2.IdentityContext{
					Type:     identityType,
					Identity: a.Assertions[i].Check.SubjectId,
				},
				ResourceContext: resource.Context{
					ObjectType: a.Assertions[i].Check.ObjectType,
					ObjectID:   a.Assertions[i].Check.ObjectId,
					Relation:   a.Assertions[i].Check.Relation,
				}.Struct(),
				PolicyContext: &aza2.PolicyContext{
					Path:      cmd.PolicyPath,
					Decisions: []string{cmd.PolicyRule},
				},
				PolicyInstance: &aza2.PolicyInstance{
					Name: cmd.PolicyName,
				},
			},
			Expected: a.Assertions[i].Expected,
		}

		d.Assertions = append(d.Assertions, &decision)
	}
	return &d
}

func (cmd *ConvertCmd) persist(_ context.Context, d *DecisionAssertions) error {
	var w *os.File

	w = os.Stdout
	if cmd.Output != "" {
		var err error
		w, err = os.Create(cmd.Output)
		if err != nil {
			return err
		}
	}
	defer w.Close()

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		return err
	}

	return nil
}
