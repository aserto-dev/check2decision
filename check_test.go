package check2decision_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aserto-dev/check2decision/api"
	"github.com/aserto-dev/check2decision/pkg/resource"
	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	aza2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCheck1(t *testing.T) {
	c := &api.CheckAssertion{
		Check: &reader.CheckRequest{
			ObjectType:  "obj_type",
			ObjectId:    "obj_id",
			Relation:    "rel",
			SubjectType: "sub_type",
			SubjectId:   "sub_id",
		},
		Expected: true,
	}

	buf, err := protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   false,
		EmitDefaultValues: false,
	}.Marshal(c)
	protojson.Format(c)

	assert.NoError(t, err)
	t.Log(string(buf))
}

func TestCheck2(t *testing.T) {
	a := struct {
		Assertions []*api.CheckAssertion `json:"assertions"`
	}{}

	r, err := os.Open("./checks.json")
	assert.NoError(t, err)

	dec := json.NewDecoder(r)
	if err := dec.Decode(&a); err != nil {
		assert.NoError(t, err)
	}

	t.Logf("length %d", len(a.Assertions))
	for i := 0; i < len(a.Assertions); i++ {
		t.Logf("%-4d %s:%s#%s@%s:%s - %t",
			i,
			a.Assertions[i].Check.ObjectType,
			a.Assertions[i].Check.ObjectId,
			a.Assertions[i].Check.Relation,
			a.Assertions[i].Check.SubjectType,
			a.Assertions[i].Check.SubjectId,
			a.Assertions[i].Expected,
		)
	}
}

func TestCheck3(t *testing.T) {
	a := struct {
		Assertions []*api.CheckAssertion `json:"assertions"`
	}{}

	r, err := os.Open("./checks.json")
	assert.NoError(t, err)

	dec := json.NewDecoder(r)
	if err := dec.Decode(&a); err != nil {
		assert.NoError(t, err)
	}

	var decisions []*api.DecisionAssertion

	t.Logf("length a: %d", len(a.Assertions))

	for i := 0; i < len(a.Assertions); i++ {
		decision := api.DecisionAssertion{
			CheckDecision: &az2.IsRequest{
				IdentityContext: &aza2.IdentityContext{
					Type:     aza2.IdentityType_IDENTITY_TYPE_SUB,
					Identity: a.Assertions[i].Check.SubjectId,
				},
				ResourceContext: resource.Context{
					ObjectType: a.Assertions[i].Check.ObjectType,
					ObjectID:   a.Assertions[i].Check.ObjectId,
					Relation:   a.Assertions[i].Check.Relation,
				}.Struct(),
				PolicyContext: &aza2.PolicyContext{
					Path:      "rebac.check",
					Decisions: []string{"allowed"},
				},
				PolicyInstance: &aza2.PolicyInstance{
					Name: "policy-rebac",
				},
			},
			Expected: a.Assertions[i].Expected,
		}
		decisions = append(decisions, &decision)
	}

	b := struct {
		Assertions []*api.DecisionAssertion `json:"assertions"`
	}{
		Assertions: decisions,
	}

	t.Logf("length b: %d", len(b.Assertions))

	for i := 0; i < len(b.Assertions); i++ {
		t.Logf("%-4d %s:%s#%s@%s:%s - %t",
			i,
			a.Assertions[i].Check.ObjectType,
			a.Assertions[i].Check.ObjectId,
			a.Assertions[i].Check.Relation,
			a.Assertions[i].Check.SubjectType,
			a.Assertions[i].Check.SubjectId,
			a.Assertions[i].Expected,
		)
	}
}
