package main

import cobragenskill "github.com/theburrowhub/go-cobra-gen-skill"

func init() {
	cobragenskill.RegisterCommand(rootCmd,
		cobragenskill.WithVersion(version),
		cobragenskill.WithLicense("MIT"),
		cobragenskill.WithMetadata("author", "theburrowhub"),
	)
}
