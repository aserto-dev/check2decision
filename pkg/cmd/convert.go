package cmd

import (
	"bytes"
	"context"
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
	"google.golang.org/protobuf/encoding/protojson"
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

func (cmd *ConvertCmd) load(_ context.Context) (*api.CheckAssertions, error) {
	a := &api.CheckAssertions{}

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

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}

	uOpts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}

	if err := uOpts.Unmarshal(buf.Bytes(), a); err != nil {
		return nil, err
	}

	return a, nil
}

type DecisionAssertions struct {
	Assertions []*api.DecisionAssertion `json:"assertions"`
}

func (cmd *ConvertCmd) transform(_ context.Context, a *api.CheckAssertions) *api.DecisionAssertions {
	d := &api.DecisionAssertions{}

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
					Name:          cmd.PolicyName,
					InstanceLabel: "",
				},
			},
			Expected: a.Assertions[i].Expected,
		}

		d.Assertions = append(d.Assertions, &decision)
	}
	return d
}

func (cmd *ConvertCmd) persist(_ context.Context, d *api.DecisionAssertions) error {
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

	buf, err := protojson.MarshalOptions{
		Multiline:         true,
		Indent:            "  ",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   false,
		EmitDefaultValues: false,
	}.Marshal(d)
	if err != nil {
		return err
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}
