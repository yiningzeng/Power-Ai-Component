package main

type PowerAiAsset struct {
	Asset Asset `json:"asset"`
	Regions []string `json:"regions"`
	Version string `json:"version"`
}

type Asset struct {
	Format string `json:"format"`
	Id     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   Size   `json:"size"`
	State  int    `json:"state"`
	Type   int    `json:"type"`
	Sort   string `json:"sort"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
