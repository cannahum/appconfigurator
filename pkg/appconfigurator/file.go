package appconfigurator

import (
	"encoding/json"
	"fmt"
	"os"
)

// File is a file loaded into memory. In config JSON it is written as a path
// string; the `resolve:"file"` resolver reads that path and fills Data.
//
//	type TLSConfig struct {
//	    ClientCert appconfigurator.File `json:"client_cert" resolve:"file"`
//	}
//
// After Resolve, ClientCert.Data holds the file contents (e.g. for
// tls.X509KeyPair). An untagged File is left unread (Path only).
type File struct {
	Path string
	Data []byte
}

// UnmarshalJSON accepts the JSON path string.
func (f *File) UnmarshalJSON(b []byte) error {
	var p string
	if err := json.Unmarshal(b, &p); err != nil {
		return fmt.Errorf("appconfigurator: File expects a path string: %w", err)
	}
	f.Path = p
	return nil
}

// MarshalJSON renders back to the path string.
func (f File) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Path)
}

// load reads the file at Path into Data. A blank Path is a no-op.
func (f *File) load() error {
	if f.Path == "" {
		return nil
	}
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return err
	}
	f.Data = data
	return nil
}
