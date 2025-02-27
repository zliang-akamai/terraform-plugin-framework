package fwschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// AttributeMissingAttributeTypesDiag returns an error diagnostic to provider
// developers about missing the AttributeTypes field on an Attribute
// implementation. This can cause unexpected errors or panics.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/699
func AttributeMissingAttributeTypesDiag(attributePath path.Path) diag.Diagnostic {
	// The diagnostic path is intentionally omitted as it is invalid in this
	// context. Diagnostic paths are intended to be mapped to actual data,
	// while this path information must be synthesized.
	return diag.NewErrorDiagnostic(
		"Invalid Attribute Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is missing the AttributeTypes or CustomType field on an object Attribute. ", attributePath)+
			"One of these fields is required to prevent other unexpected errors or panics.",
	)
}

// AttributeMissingElementTypeDiag returns an error diagnostic to provider
// developers about missing the ElementType field on an Attribute
// implementation. This can cause unexpected errors or panics.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/699
func AttributeMissingElementTypeDiag(attributePath path.Path) diag.Diagnostic {
	// The diagnostic path is intentionally omitted as it is invalid in this
	// context. Diagnostic paths are intended to be mapped to actual data,
	// while this path information must be synthesized.
	return diag.NewErrorDiagnostic(
		"Invalid Attribute Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is missing the CustomType or ElementType field on a collection Attribute. ", attributePath)+
			"One of these fields is required to prevent other unexpected errors or panics.",
	)
}
