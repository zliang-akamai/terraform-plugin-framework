package proto6server

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfprotov6.ProviderServer = &Server{}

// Provider server implementation.
type Server struct {
	FrameworkServer fwserver.Server

	contextCancels   []context.CancelFunc
	contextCancelsMu sync.Mutex
}

func (s *Server) registerContext(in context.Context) context.Context {
	ctx, cancel := context.WithCancel(in)
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	s.contextCancels = append(s.contextCancels, cancel)
	return ctx
}

func (s *Server) cancelRegisteredContexts(_ context.Context) {
	s.contextCancelsMu.Lock()
	defer s.contextCancelsMu.Unlock()
	for _, cancel := range s.contextCancels {
		cancel()
	}
	s.contextCancels = nil
}

func (s *Server) GetProviderSchema(ctx context.Context, proto6Req *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwReq := fromproto6.GetProviderSchemaRequest(ctx, proto6Req)
	fwResp := &fwserver.GetProviderSchemaResponse{}

	s.FrameworkServer.GetProviderSchema(ctx, fwReq, fwResp)

	return toproto6.GetProviderSchemaResponse(ctx, fwResp), nil
}

func (s *Server) ValidateProviderConfig(ctx context.Context, proto6Req *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateProviderConfigResponse{}

	providerSchema, diags := s.FrameworkServer.ProviderSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateProviderConfigRequest(ctx, proto6Req, providerSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateProviderConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateProviderConfigResponse(ctx, fwResp), nil
}

func (s *Server) ConfigureProvider(ctx context.Context, proto6Req *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &tfsdk.ConfigureProviderResponse{}

	providerSchema, diags := s.FrameworkServer.ProviderSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ConfigureProviderRequest(ctx, proto6Req, providerSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ConfigureProvider(ctx, fwReq, fwResp)

	return toproto6.ConfigureProviderResponse(ctx, fwResp), nil
}

func (s *Server) StopProvider(ctx context.Context, _ *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	s.cancelRegisteredContexts(ctx)

	return &tfprotov6.StopProviderResponse{}, nil
}

func (s *Server) ValidateResourceConfig(ctx context.Context, proto6Req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateResourceConfigResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateResourceConfigRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateResourceConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
}

func (s *Server) UpgradeResourceState(ctx context.Context, proto6Req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.UpgradeResourceStateResponse{}

	if proto6Req == nil {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.UpgradeResourceStateRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.UpgradeResourceState(ctx, fwReq, fwResp)

	return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
}

func (s *Server) ReadResource(ctx context.Context, proto6Req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ReadResourceResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ReadResourceRequest(ctx, proto6Req, resourceType, resourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadResourceResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ReadResource(ctx, fwReq, fwResp)

	return toproto6.ReadResourceResponse(ctx, fwResp), nil
}

func markComputedNilsAsUnknown(ctx context.Context, config tftypes.Value, resourceSchema tfsdk.Schema) func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error) {
	return func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		ctx = logging.FrameworkWithAttributePath(ctx, path.String())

		// we are only modifying attributes, not the entire resource
		if len(path.Steps()) < 1 {
			return val, nil
		}
		configVal, _, err := tftypes.WalkAttributePath(config, path)
		if err != tftypes.ErrInvalidStep && err != nil {
			logging.FrameworkError(ctx, "error walking attribute path")
			return val, err
		} else if err != tftypes.ErrInvalidStep && !configVal.(tftypes.Value).IsNull() {
			logging.FrameworkTrace(ctx, "attribute not null in config, not marking unknown")
			return val, nil
		}
		attribute, err := resourceSchema.AttributeAtPath(path)
		if err != nil {
			if errors.Is(err, tfsdk.ErrPathInsideAtomicAttribute) {
				// ignore attributes/elements inside schema.Attributes, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is a non-schema attribute, not marking unknown")
				return val, nil
			}
			logging.FrameworkError(ctx, "couldn't find attribute in resource schema")
			return tftypes.Value{}, fmt.Errorf("couldn't find attribute in resource schema: %w", err)
		}
		if !attribute.Computed {
			logging.FrameworkTrace(ctx, "attribute is not computed in schema, not marking unknown")
			return val, nil
		}
		logging.FrameworkDebug(ctx, "marking computed attribute that is null in the config as unknown")
		return tftypes.NewValue(val.Type(), tftypes.UnknownValue), nil
	}
}

// planResourceChangeResponse is a thin abstraction to allow native Diagnostics usage
type planResourceChangeResponse struct {
	PlannedState    *tfprotov6.DynamicValue
	Diagnostics     diag.Diagnostics
	RequiresReplace []*tftypes.AttributePath
	PlannedPrivate  []byte
}

func (r planResourceChangeResponse) toTfprotov6() *tfprotov6.PlanResourceChangeResponse {
	return &tfprotov6.PlanResourceChangeResponse{
		PlannedState:    r.PlannedState,
		Diagnostics:     toproto6.Diagnostics(r.Diagnostics),
		RequiresReplace: r.RequiresReplace,
		PlannedPrivate:  r.PlannedPrivate,
	}
}

func (s *Server) PlanResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)
	resp := &planResourceChangeResponse{}

	s.planResourceChange(ctx, req, resp)

	return resp.toTfprotov6(), nil
}

func (s *Server) planResourceChange(ctx context.Context, req *tfprotov6.PlanResourceChangeRequest, resp *planResourceChangeResponse) {
	// get the type of resource, so we can get its schema and create an
	// instance
	resourceType, diags := s.FrameworkServer.ResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get the schema from the resource type, so we can embed it in the
	// config and plan
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema")
	resourceSchema, diags := resourceType.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := req.Config.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing configuration",
			"An unexpected error was encountered trying to parse the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	plan, err := req.ProposedNewState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing plan",
			"There was an unexpected error parsing the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	state, err := req.PriorState.Unmarshal(resourceSchema.TerraformType(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing prior state",
			"An unexpected error was encountered trying to parse the prior state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	resp.PlannedState = req.ProposedNewState

	// create the resource instance, so we can call its methods and handle
	// the request
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.FrameworkServer.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute any AttributePlanModifiers.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers are run again after that marking to allow setting values
	// and preventing extraneous plan differences.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// Execute any resource-level ModifyPlan method.
	//
	// This pass is before any Computed-only attributes are marked as unknown
	// to ensure any plan changes will trigger that behavior. These plan
	// modifiers be run again after that marking to allow setting values and
	// preventing extraneous plan differences.
	//
	// TODO: Enabling this pass will generate the following test error:
	//
	//     --- FAIL: TestServerPlanResourceChange/two_modifyplan_add_list_elem (0.00s)
	// serve_test.go:3303: An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:
	//
	// ElementKeyInt(1).AttributeName("name") still remains in the path: step cannot be applied to this value
	//
	// To fix this, (Config).GetAttribute() should return nil instead of the error.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/150
	// See also: https://github.com/hashicorp/terraform-plugin-framework/pull/167

	// After ensuring there are proposed changes, mark any computed attributes
	// that are null in the config as unknown in the plan, so providers have
	// the choice to update them.
	//
	// Later attribute and resource plan modifier passes can override the
	// unknown with a known value using any plan modifiers.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !plan.IsNull() && !plan.Equal(state) {
		logging.FrameworkTrace(ctx, "marking computed null values as unknown")
		modifiedPlan, err := tftypes.Transform(plan, markComputedNilsAsUnknown(ctx, config, resourceSchema))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error modifying plan",
				"There was an unexpected error updating the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return
		}
		if !plan.Equal(modifiedPlan) {
			logging.FrameworkTrace(ctx, "at least one value was changed to unknown")
		}
		plan = modifiedPlan
	}

	// Execute any AttributePlanModifiers again. This allows overwriting
	// any unknown values.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	if !plan.IsNull() {
		modifySchemaPlanReq := ModifySchemaPlanRequest{
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}

		providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if providerMetaSchema != nil {
			modifySchemaPlanReq.ProviderMeta = tfsdk.Config{
				Schema: *providerMetaSchema,
				Raw:    tftypes.NewValue(providerMetaSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(providerMetaSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				modifySchemaPlanReq.ProviderMeta.Raw = pmValue
			}
		}

		modifySchemaPlanResp := ModifySchemaPlanResponse{
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			Diagnostics: resp.Diagnostics,
		}

		SchemaModifyPlan(ctx, resourceSchema, modifySchemaPlanReq, &modifySchemaPlanResp)
		resp.RequiresReplace = append(resp.RequiresReplace, modifySchemaPlanResp.RequiresReplace...)
		plan = modifySchemaPlanResp.Plan.Raw
		resp.Diagnostics = modifySchemaPlanResp.Diagnostics
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Execute any resource-level ModifyPlan method again. This allows
	// overwriting any unknown values.
	//
	// We do this regardless of whether the plan is null or not, because we
	// want resources to be able to return diagnostics when planning to
	// delete resources, e.g. to inform practitioners that the resource
	// _can't_ be deleted in the API and will just be removed from
	// Terraform's state
	var modifyPlanResp tfsdk.ModifyResourcePlanResponse
	if resource, ok := resource.(tfsdk.ResourceWithModifyPlan); ok {
		logging.FrameworkTrace(ctx, "Resource implements ResourceWithModifyPlan")
		modifyPlanReq := tfsdk.ModifyResourcePlanRequest{
			Config: tfsdk.Config{
				Schema: resourceSchema,
				Raw:    config,
			},
			State: tfsdk.State{
				Schema: resourceSchema,
				Raw:    state,
			},
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
		}

		providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if providerMetaSchema != nil {
			modifyPlanReq.ProviderMeta = tfsdk.Config{
				Schema: *providerMetaSchema,
				Raw:    tftypes.NewValue(providerMetaSchema.TerraformType(ctx), nil),
			}

			if req.ProviderMeta != nil {
				pmValue, err := req.ProviderMeta.Unmarshal(providerMetaSchema.TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"Error parsing provider_meta",
						"There was an error parsing the provider_meta block. Please report this to the provider developer:\n\n"+err.Error(),
					)
					return
				}
				modifyPlanReq.ProviderMeta.Raw = pmValue
			}
		}

		modifyPlanResp = tfsdk.ModifyResourcePlanResponse{
			Plan: tfsdk.Plan{
				Schema: resourceSchema,
				Raw:    plan,
			},
			RequiresReplace: []*tftypes.AttributePath{},
			Diagnostics:     resp.Diagnostics,
		}
		logging.FrameworkDebug(ctx, "Calling provider defined Resource ModifyPlan")
		resource.ModifyPlan(ctx, modifyPlanReq, &modifyPlanResp)
		logging.FrameworkDebug(ctx, "Called provider defined Resource ModifyPlan")
		resp.Diagnostics = modifyPlanResp.Diagnostics
		plan = modifyPlanResp.Plan.Raw
	}

	plannedState, err := tfprotov6.NewDynamicValue(plan.Type(), plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting response",
			"There was an unexpected error converting the state in the response to a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}
	resp.PlannedState = &plannedState
	resp.RequiresReplace = append(resp.RequiresReplace, modifyPlanResp.RequiresReplace...)

	// ensure deterministic RequiresReplace by sorting and deduplicating
	resp.RequiresReplace = normaliseRequiresReplace(ctx, resp.RequiresReplace)
}

// normaliseRequiresReplace sorts and deduplicates the slice of AttributePaths
// used in the RequiresReplace response field.
// Sorting is lexical based on the string representation of each AttributePath.
func normaliseRequiresReplace(ctx context.Context, rs []*tftypes.AttributePath) []*tftypes.AttributePath {
	if len(rs) < 2 {
		return rs
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].String() < rs[j].String()
	})

	ret := make([]*tftypes.AttributePath, len(rs))
	ret[0] = rs[0]

	// deduplicate
	j := 1
	for i := 1; i < len(rs); i++ {
		if rs[i].Equal(ret[j-1]) {
			logging.FrameworkDebug(ctx, "attribute found multiple times in RequiresReplace, removing duplicate", map[string]interface{}{logging.KeyAttributePath: rs[i]})
			continue
		}
		ret[j] = rs[i]
		j++
	}
	return ret[:j]
}

func (s *Server) ApplyResourceChange(ctx context.Context, proto6Req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ApplyResourceChangeResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ApplyResourceChangeRequest(ctx, proto6Req, resourceType, resourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ApplyResourceChange(ctx, fwReq, fwResp)

	return toproto6.ApplyResourceChangeResponse(ctx, fwResp), nil
}

func (s *Server) ValidateDataResourceConfig(ctx context.Context, proto6Req *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateDataSourceConfigResponse{}

	dataSourceType, diags := s.FrameworkServer.DataSourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	dataSourceSchema, diags := s.FrameworkServer.DataSourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateDataSourceConfigRequest(ctx, proto6Req, dataSourceType, dataSourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateDataSourceConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateDataSourceConfigResponse(ctx, fwResp), nil
}

func (s *Server) ReadDataSource(ctx context.Context, proto6Req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ReadDataSourceResponse{}

	dataSourceType, diags := s.FrameworkServer.DataSourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	dataSourceSchema, diags := s.FrameworkServer.DataSourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	providerMetaSchema, diags := s.FrameworkServer.ProviderMetaSchema(ctx)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ReadDataSourceRequest(ctx, proto6Req, dataSourceType, dataSourceSchema, providerMetaSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ReadDataSource(ctx, fwReq, fwResp)

	return toproto6.ReadDataSourceResponse(ctx, fwResp), nil
}
