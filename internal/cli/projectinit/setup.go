package projectinit

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// CheckUninitialized reads only module identity, before questions or environment
// access. Init is adoption, never an existing application's maintenance command.
func CheckUninitialized(root string) error {
	data, _, err := readRegular(filepath.Join(root, "go.mod"))
	if err != nil {
		return errors.New("initial setup module unavailable")
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "module" {
			count++
			if len(fields) != 2 || fields[1] != OriginalModule() {
				return errors.New("init cannot rerun on an adopted application; use maintenance commands")
			}
		}
	}
	if count != 1 {
		return errors.New("initial setup module unavailable")
	}
	return nil
}

// EffectiveEnvironment parses the exact planned generated snapshot, or the
// explicitly kept regular dotenv. It never consults the process environment.
// Each call returns a fresh map, and diagnostics never include dotenv contents.
func (p *Plan) EffectiveEnvironment() (map[string]string, error) {
	selected, err := p.SelectEnvironment()
	if err != nil {
		return nil, err
	}
	return selected.Values, nil
}

// SelectedEnvironment binds effective settings to the kept file's exact bytes,
// mode and identity. Generated settings are already bound by Plan's transaction.
// This is a quiescence check, not a lock against concurrent application edits.
type SelectedEnvironment struct {
	Values   map[string]string
	path     string
	data     []byte
	mode     fs.FileMode
	identity os.FileInfo
}

func (p *Plan) SelectEnvironment() (*SelectedEnvironment, error) {
	s := &SelectedEnvironment{}
	data := p.environment
	if data == nil {
		s.path = filepath.Join(p.root, ".env")
		var err error
		data, s.mode, err = readRegular(s.path)
		if err != nil {
			return nil, errors.New("kept environment unavailable")
		}
		s.identity, err = os.Lstat(s.path)
		if err != nil {
			return nil, errors.New("kept environment unavailable")
		}
		s.data = data
	}
	values, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return nil, errors.New("invalid effective environment")
	}
	s.Values = values
	return s, nil
}

func (s *SelectedEnvironment) Check() error {
	if s.path == "" {
		return nil
	}
	data, mode, err := readRegular(s.path)
	if err != nil {
		return errors.New("kept environment changed since preview")
	}
	info, err := os.Lstat(s.path)
	if err != nil || !os.SameFile(info, s.identity) || mode != s.mode || !bytes.Equal(data, s.data) {
		return errors.New("kept environment changed since preview")
	}
	return nil
}
