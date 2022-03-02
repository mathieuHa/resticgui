package Models

type PathStruct struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	AbsPath string `json:"absPath"`
	RelPath string `json:"relPath"`
}
