// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"fmt"
)

var GitCommit string

const Name = "vault-ssh-helper"
const Version = "0.2.1"
const VersionPrerelease = ""

// formattedVersion returns a formatted version string which includes the git
// commit and development information.
func formattedVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "%s v%s", Name, Version)

	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", VersionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}
	return versionString.String()
}
