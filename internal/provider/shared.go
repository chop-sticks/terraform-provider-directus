// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapToNormalized marshals a free-form Directus object into a jsontypes
// Normalized value for storage in state. An empty/nil map becomes null so
// unset blobs don't show spurious "{}" diffs. Normalized compares JSON
// semantically, so key ordering and whitespace never cause drift.
func mapToNormalized(m map[string]any) jsontypes.Normalized {
	if len(m) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(m)
	if err != nil {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

// normalizedToMap decodes a Normalized JSON attribute into a map for sending to
// the client. Null/unknown yields a nil map (omitted from the request).
func normalizedToMap(v jsontypes.Normalized, diags *diag.Diagnostics) map[string]any {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out map[string]any
	diags.Append(v.Unmarshal(&out)...)
	return out
}

// anyToNormalized marshals an arbitrary Directus scalar/object (e.g. a field
// default_value, which may be a string, number, bool, object, or null) into a
// Normalized JSON value. nil becomes null.
func anyToNormalized(v any) jsontypes.Normalized {
	if v == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

// normalizedToAny decodes a Normalized JSON attribute into an arbitrary Go
// value for sending to the client. Null/unknown yields nil.
func normalizedToAny(v jsontypes.Normalized, diags *diag.Diagnostics) any {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out any
	diags.Append(v.Unmarshal(&out)...)
	return out
}

// importInt64ID parses a string import identifier into the integer "id"
// attribute used by Directus resources with numeric primary keys (presets,
// permissions, ...). The default passthrough importer only handles string ids.
func importInt64ID(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(strings.TrimSpace(request.ID), 10, 64)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected a numeric id, got %q: %s", request.ID, err),
		)
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// isNotFound reports whether err is a Directus 404 response. The client returns
// errors formatted as "status: <code>, body <...>" with no typed error value,
// so this matches on that shape. Read handlers use it to detect drift (a
// resource deleted out-of-band) and remove the resource from state instead of
// erroring. Revisit if the client gains typed errors.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "status: 404")
}

// anyToStringID normalizes a Directus relational field into a string id.
// Directus returns either a bare id string or an expanded object (with an "id"
// key) depending on the request's fields; unset relations come back nil. Used
// by resources whose model exposes a m2o relation as a plain id string.
func anyToStringID(v any) types.String {
	switch t := v.(type) {
	case string:
		if t == "" {
			return types.StringNull()
		}
		return types.StringValue(t)
	case map[string]any:
		if id, ok := t["id"].(string); ok && id != "" {
			return types.StringValue(id)
		}
	}
	return types.StringNull()
}

// configureClient extracts the shared *directus.Client from a resource or data
// source ConfigureRequest's ProviderData.
//
// It returns nil in two cases the caller must handle by returning early:
//   - providerData is nil: the framework calls Configure once before the
//     provider itself is configured; this is expected and adds no diagnostic.
//   - providerData is the wrong type: adds an error diagnostic and returns nil.
//
// Every resource and data source funnels its Configure through this helper so
// the client type assertion lives in exactly one place.
func configureClient(providerData any, diags *diag.Diagnostics) *directus.Client {
	if providerData == nil {
		return nil
	}

	client, ok := providerData.(*directus.Client)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *directus.Client, got %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}

	return client
}
