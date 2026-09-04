// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package main

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name directus

import (
	"context"
	"flag"
	"log"

	"github.com/chop-sticks/terraform-provider-directus/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/chop-sticks/directus",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
