package main

import "embed"

//go:embed static/*
var staticAssets embed.FS
