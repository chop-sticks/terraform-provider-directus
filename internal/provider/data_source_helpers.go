// Copyright IBM Corp. 2026

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Computed-attribute constructors for data source schemas. Data sources reuse
// the resource model structs (matching tfsdk tags), so every attribute is
// Computed except the lookup key, which is Required.

func dsComputedString(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func dsComputedBool(desc string) dschema.BoolAttribute {
	return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func dsComputedInt64(desc string) dschema.Int64Attribute {
	return dschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
}

func dsComputedJSON(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, CustomType: jsontypes.NormalizedType{}, Computed: true}
}

func dsComputedStringList(desc string) dschema.ListAttribute {
	return dschema.ListAttribute{MarkdownDescription: desc, ElementType: types.StringType, Computed: true}
}

func dsRequiredString(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Required: true}
}
