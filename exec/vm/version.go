package vm

import (
	"fmt"

	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

// Version is a RICE module version. The preamble is always magic, major, minor.
type Version struct {
	Major uint16
	Minor uint16
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func (v Version) Less(o Version) bool {
	return v.Major < o.Major || (v.Major == o.Major && v.Minor < o.Minor)
}

func (v Version) Max(o Version) Version {
	if v.Less(o) {
		return o
	}
	return v
}

// Supports reports whether a VM at version v can load a module of version mod.
// A module X.y loads when MinMajor <= X <= v.Major and, if X == v.Major, y <= v.Minor.
func (v Version) Supports(mod Version) bool {
	if mod.Major < MinMajor || mod.Major > v.Major {
		return false
	}
	if mod.Major < v.Major {
		return true
	}
	return mod.Minor <= v.Minor
}

var (
	// CurrentVersion is the newest module this VM produces and accepts.
	CurrentVersion = Version{Major: 1, Minor: 0}
	// MinMajor is the oldest major this VM still decodes. Raised only to drop a major.
	MinMajor uint16 = 1
)

var v1_0 = Version{Major: 1, Minor: 0}

func versionLoadError(mod Version) error {
	if mod.Major < MinMajor {
		return fmt.Errorf("module is RICE %s, this VM requires major >= %d", mod, MinMajor)
	}
	if !CurrentVersion.Supports(mod) {
		return fmt.Errorf("module needs RICE %s, this VM supports up to %s", mod, CurrentVersion)
	}
	return nil
}

// RequiredVersion is the lowest version that covers every opcode, constant tag,
// and flag this module uses. Empty modules require 1.0.
func (m *Module) RequiredVersion() Version {
	req := v1_0
	for _, c := range m.Constants {
		if s, ok := constTagSince[constTag(c)]; ok && req.Less(s) {
			req = s
		}
	}
	for _, h := range m.Hotspots {
		if s, err := flagSince(h.Flags, hotspotFlagSince); err == nil && req.Less(s) {
			req = s
		}
	}
	for _, fn := range m.Functions {
		if fn == nil {
			continue
		}
		flags := byte(0)
		if fn.Variadic {
			flags |= FuncVariadic
		}
		if s, err := flagSince(flags, funcFlagSince); err == nil && req.Less(s) {
			req = s
		}
		for _, op := range fn.Bytecode {
			if s, ok := opcodeSince(op); ok && req.Less(s) {
				req = s
			}
		}
	}
	return req
}

func verifyVersion(m *Module) error {
	for i, c := range m.Constants {
		tag := constTag(c)
		s, ok := constTagSince[tag]
		if !ok {
			return fmt.Errorf("constant %d: unknown tag %d", i, tag)
		}
		if m.Version.Less(s) {
			return fmt.Errorf("constant %d: tag %d requires RICE %s, module is %s", i, tag, s, m.Version)
		}
	}
	for i, h := range m.Hotspots {
		s, err := flagSince(h.Flags, hotspotFlagSince)
		if err != nil {
			return fmt.Errorf("hotspot %d: %w", i, err)
		}
		if m.Version.Less(s) {
			return fmt.Errorf("hotspot %d: flags require RICE %s, module is %s", i, s, m.Version)
		}
	}
	for fi, fn := range m.Functions {
		if fn == nil {
			continue
		}
		flags := byte(0)
		if fn.Variadic {
			flags |= FuncVariadic
		}
		s, err := flagSince(flags, funcFlagSince)
		if err != nil {
			return fmt.Errorf("function %d: %w", fi, err)
		}
		if m.Version.Less(s) {
			return fmt.Errorf("function %d: flags require RICE %s, module is %s", fi, s, m.Version)
		}
		for ip, op := range fn.Bytecode {
			os, ok := opcodeSince(op)
			if !ok {
				return fmt.Errorf("function %d ip %d: unknown opcode %d", fi, ip, op)
			}
			if m.Version.Less(os) {
				return fmt.Errorf("function %d ip %d: opcode %s requires RICE %s, module is %s", fi, ip, op, os, m.Version)
			}
		}
	}
	return nil
}

func constTag(c types.Value) byte {
	if c == nil {
		return ConstNull
	}
	switch c.(type) {
	case values.Bool:
		return ConstBool
	case values.Int:
		return ConstInt
	case values.Float:
		return ConstFloat
	case values.String:
		return ConstString
	default:
		return 0
	}
}

func flagSince(flags byte, table []flagMeta) (Version, error) {
	req := Version{}
	used := flags
	for _, f := range table {
		if flags&f.bit != 0 {
			if req.Less(f.since) {
				req = f.since
			}
			used &^= f.bit
		}
	}
	if used != 0 {
		return req, fmt.Errorf("unknown flag bits 0x%x", used)
	}
	return req, nil
}
