package models

type FileData struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size,omitempty"`
}

type CopyFileInfoData struct {
	Files []FileData `json:"files"`
	Index int64      `json:"index,omitempty"`
	IP    string     `json:"ip"`
	Port  int        `json:"port"`
}
