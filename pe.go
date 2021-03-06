// Package pe provides access to Portable Executable (PE) files.
//
// ref: https://docs.microsoft.com/en-us/windows/desktop/debug/pe-format
package pe

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/mewmew/pe/enum"
	"github.com/mewmew/pe/internal/pe"
	"github.com/pkg/errors"
)

// ParseFile parses the given PE file.
func ParseFile(path string) (*File, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ParseBytes(buf)
}

// Parse parses the given PE file, reading from r.
func Parse(r io.Reader) (*File, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ParseBytes(buf)
}

// ParseBytes parses the given PE file, reading from content.
func ParseBytes(content []byte) (*File, error) {
	return parse(content)
}

// reader is the interface that groups the basic Read, ReadAt and Seek methods.
type reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// PE signature.
var signature = []byte("PE\x00\x00")

// parse parses the given PE file, reading from content.
func parse(content []byte) (*File, error) {
	file := &File{
		Content: content,
	}
	r := bytes.NewReader(content)
	// Parse COFF file header.
	fileHdr, err := parseFileHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.FileHdr = fileHdr
	// Parse optional header.
	optHdr, err := parseOptHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.OptHdr = optHdr
	// Parse data directories.
	dataDirs, err := file.parseDataDirs(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.DataDirs = dataDirs
	// Parse section headers.
	//
	// After parsing the section headers, we may read data using relative
	// addresses (relative to image base).
	sectHdrs, err := file.parseSectionHdrs(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	file.SectHdrs = sectHdrs
	// Parse contents of data directories.
	if err := file.parseDataDirsContent(r); err != nil {
		return nil, errors.WithStack(err)
	}
	return file, nil
}

// parseFileHeader parses the COFF file header of the given PE file.
func parseFileHeader(r reader) (*FileHeader, error) {
	// Get offset of COFF file header.
	var offset uint32
	sr := io.NewSectionReader(r, 0x3C, 4)
	if err := binary.Read(sr, binary.LittleEndian, &offset); err != nil {
		return nil, errors.WithStack(err)
	}
	// Parse PE signature.
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, errors.WithStack(err)
	}
	sig := make([]byte, 4)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, errors.WithStack(err)
	}
	if !bytes.Equal(signature, sig) {
		return nil, errors.Errorf("invalid PE signature; expected %q, got %q", signature, sig)
	}
	// Parse COFF file header.
	raw := &pe.RawFileHeader{}
	if err := binary.Read(r, binary.LittleEndian, raw); err != nil {
		return nil, errors.WithStack(err)
	}
	return goFileHeader(raw), nil
}

const (
	// Magic value of optional header for PE32 (32-bit).
	magic32 = 0x010B
	// Magic value of optional header for PE32+ (64-bit).
	magic64 = 0x020B
)

// parseOptHeader parses the optional header of the given PE file.
func parseOptHeader(r reader) (*OptHeader, error) {
	// Get magic number to determine type of optional header (PE32 vs. PE32+).
	var magic uint16
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, errors.WithStack(err)
	}
	switch magic {
	case magic32:
		// PE32 (32-bit).
		raw := &pe.RawOptHeader32{}
		if err := binary.Read(r, binary.LittleEndian, raw); err != nil {
			return nil, errors.WithStack(err)
		}
		return goOptHeader32(raw, magic), nil
	case magic64:
		// PE32+ (64-bit).
		raw := &pe.RawOptHeader64{}
		if err := binary.Read(r, binary.LittleEndian, raw); err != nil {
			return nil, errors.WithStack(err)
		}
		return goOptHeader64(raw, magic), nil
	default:
		return nil, errors.Errorf("invalid optional header magic number; expected 0x%04X or 0x%04X, got 0x%04X", magic32, magic64, magic)
	}
}

// parseDataDirs parses the data directories of the given PE file.
func (file *File) parseDataDirs(r reader) ([]DataDirectory, error) {
	dataDirs := make([]DataDirectory, file.OptHdr.NDataDirs)
	for idx := range dataDirs {
		if err := binary.Read(r, binary.LittleEndian, &dataDirs[idx]); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return dataDirs, nil
}

// parseSectionHdrs parses the section headers of the given PE file.
func (file *File) parseSectionHdrs(r reader) ([]SectionHeader, error) {
	var sectHdrs []SectionHeader
	for i := 0; i < int(file.FileHdr.NSections); i++ {
		var raw pe.RawSectionHeader
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			return nil, errors.WithStack(err)
		}
		sectHdrs = append(sectHdrs, goSectionHeader(raw))
	}
	return sectHdrs, nil
}

// parseDataDirsContent parses the contents of the data directories.
func (file *File) parseDataDirsContent(r reader) error {
	for idx, dataDir := range file.DataDirs {
		zero := DataDirectory{}
		if dataDir == zero {
			continue
		}
		switch idx {
		case 0:
			// Export Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 1:
			// Import Table
			imps, err := file.parseImports(dataDir)
			if err != nil {
				return errors.WithStack(err)
			}
			file.Imps = imps
		case 2:
			// Resource Table
			// TODO: parse resource table.
			//panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 3:
			// Exception Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 4:
			// Certificate Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 5:
			// Base Relocation Table
			baseRelocBlocks, err := file.parseBaseRelocBlocks(dataDir)
			if err != nil {
				return errors.WithStack(err)
			}
			file.BaseRelocBlocks = baseRelocBlocks
		case 6:
			// Debug data
			dbgData, err := file.parseDebugData(dataDir)
			if err != nil {
				return errors.WithStack(err)
			}
			file.DbgData = dbgData
		case 7:
			// Architecture
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 8:
			// Global Pointer Register
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 9:
			// TLS Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 10:
			// Load Config Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 11:
			// Bound Import Table
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 12:
			// Import Address Table
			// already handled when parsing import table.
		case 13:
			// Delay Import Descriptor
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 14:
			// CLR Header
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		case 15:
			// Reserved
			panic(fmt.Errorf("support for data directory index %d not yet implemented", idx))
		default:
			panic(fmt.Errorf("invalid data directory index; expected < 16, got %d", idx))
		}
	}
	return nil
}

// --- [ 1 - Import Table ] ----------------------------------------------------

// parseImports parses the import table of the given data directory.
func (file *File) parseImports(dataDir DataDirectory) ([]ImportEntry, error) {
	impDirs, err := file.parseImportDirs(dataDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var imps []ImportEntry
	for _, impDir := range impDirs {
		imp, err := file.parseImportEntry(impDir)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		imps = append(imps, imp)
	}
	return imps, nil
}

// parseImportDirs parses the import data directories.
func (file *File) parseImportDirs(dataDir DataDirectory) ([]ImportDirectory, error) {
	addr := file.OptHdr.ImageBase + uint64(dataDir.RelAddr)
	buf := file.ReadData(addr, int64(dataDir.Size))
	r := bytes.NewReader(buf)
	var impDirs []ImportDirectory
	for {
		var raw pe.RawImportDirectory
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return nil, errors.WithStack(err)
		}
		zero := pe.RawImportDirectory{}
		if raw == zero {
			// Last entry of table is zero.
			break
		}
		impDir := file.goImportDirectory(raw)
		impDirs = append(impDirs, impDir)
	}
	return impDirs, nil
}

// parseImportEntry parses the import entry based on the given import data
// directories.
func (file *File) parseImportEntry(impDir ImportDirectory) (ImportEntry, error) {
	// Parse import name table.
	imp := ImportEntry{
		ImpDir: impDir,
	}
	if impDir.INTRelAddr != 0 {
		ints, err := file.parseINTs(impDir.INTRelAddr)
		if err != nil {
			return ImportEntry{}, errors.WithStack(err)
		}
		imp.INTs = ints
	}
	// Parse import address table (IAT is identical in structure to INT).
	iats, err := file.parseINTs(impDir.IATRelAddr)
	if err != nil {
		return ImportEntry{}, errors.WithStack(err)
	}
	imp.IATs = iats
	return imp, nil
}

// parseINTs parses the import name table located at the given
// relative address.
func (file *File) parseINTs(intRelAddr uint32) ([]INTEntry, error) {
	var ints []INTEntry
	addr := file.OptHdr.ImageBase + uint64(intRelAddr)
loop:
	for {
		switch file.OptHdr.Magic {
		case magic32:
			// PE32 (32-bit).
			const rawSize = 4
			buf := file.ReadData(addr, rawSize)
			r := bytes.NewReader(buf)
			var raw pe.RawINTEntry32
			if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
				return nil, errors.WithStack(err)
			}
			if raw == 0 {
				// Last entry of table is zero.
				break loop
			}
			addr += rawSize
			intEntry := file.goINTEntry32(raw)
			ints = append(ints, intEntry)
		case magic64:
			// PE32+ (64-bit).
			const rawSize = 8
			buf := file.ReadData(addr, rawSize)
			r := bytes.NewReader(buf)
			var raw pe.RawINTEntry64
			if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
				return nil, errors.WithStack(err)
			}
			if raw == 0 {
				// Last entry of table is zero.
				break loop
			}
			addr += rawSize
			intEntry := file.goINTEntry64(raw)
			ints = append(ints, intEntry)
		default:
			return nil, errors.Errorf("invalid optional header magic number; expected 0x%04X or 0x%04X, got 0x%04X", magic32, magic64, file.OptHdr.Magic)
		}
	}
	return ints, nil
}

// --- [ 5 - Base Relocation Table ] -------------------------------------------

// parseBaseRelocBlocks parses the base relocation table of the given data
// directory.
func (file *File) parseBaseRelocBlocks(dataDir DataDirectory) ([]BaseRelocBlock, error) {
	addr := file.OptHdr.ImageBase + uint64(dataDir.RelAddr)
	buf := file.ReadData(addr, int64(dataDir.Size))
	r := bytes.NewReader(buf)
	var blocks []BaseRelocBlock
	for {
		block, err := file.parseBaseRelocBlock(r)
		if err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return nil, errors.WithStack(err)
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

// parseBaseRelocBlock parses a base relocation block, reading from r.
func (file *File) parseBaseRelocBlock(r io.Reader) (BaseRelocBlock, error) {
	var raw pe.RawBaseRelocBlock
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		return BaseRelocBlock{}, errors.WithStack(err)
	}
	block := BaseRelocBlock{
		PageRelAddr: raw.PageRelAddr,
		BlockSize:   raw.BlockSize,
	}
	lr := &io.LimitedReader{
		R: r,
		N: int64(raw.BlockSize) - 8, // exclude size of RawBaseRelocBlock.
	}
	for {
		var rawEntry pe.RawBaseRelocEntry
		if err := binary.Read(lr, binary.LittleEndian, &rawEntry); err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return BaseRelocBlock{}, errors.WithStack(err)
		}
		block.Entries = append(block.Entries, goBaseRelocEntry(rawEntry))
	}
	return block, nil
}

// --- [ 6 - Debug ] -----------------------------------------------------------

// parseDebugData parses the debug data of the given data directory.
func (file *File) parseDebugData(dataDir DataDirectory) ([]DebugData, error) {
	dbgDirs, err := file.parseDebugDirs(dataDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var dbgData []DebugData
	for _, dbgDir := range dbgDirs {
		buf := file.readDebugData(dbgDir)
		switch dbgDir.Type {
		case enum.DebugTypeCodeView:
			dbgCodeView, err := parseDebugCodeViewInfo(dbgDir, buf)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			dbgData = append(dbgData, dbgCodeView)
		case enum.DebugTypeFPO:
			dbgFPO, err := parseDebugFPO(dbgDir, buf)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			dbgData = append(dbgData, dbgFPO)
		case enum.DebugTypeMisc:
			// Miscellaneous debug data format is application specific; store raw
			// content.
			dbgMisc := &DebugMisc{
				DbgDir:  dbgDir,
				Content: buf,
			}
			dbgData = append(dbgData, dbgMisc)
		default:
			panic(fmt.Errorf("support for debug data type %q not yet implemented", dbgDir.Type))
		}
	}
	return dbgData, nil
}

// parseDebugDirs parses the debug data directories.
func (file *File) parseDebugDirs(dataDir DataDirectory) ([]DebugDirectory, error) {
	addr := file.OptHdr.ImageBase + uint64(dataDir.RelAddr)
	buf := file.ReadData(addr, int64(dataDir.Size))
	r := bytes.NewReader(buf)
	var dbgDirs []DebugDirectory
	for {
		var raw pe.RawDebugDirectory
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return nil, errors.WithStack(err)
		}
		dbgDirs = append(dbgDirs, goDebugDirectory(raw))
	}
	return dbgDirs, nil
}

// readDebugData reads the debug data of the given debug data directory.
func (file *File) readDebugData(dbgDir DebugDirectory) []byte {
	if dbgDir.RelAddr != 0 {
		addr := file.OptHdr.ImageBase + uint64(dbgDir.RelAddr)
		return file.ReadData(addr, int64(dbgDir.Size))
	}
	start := dbgDir.Offset
	end := start + dbgDir.Size
	return file.Content[start:end]
}

// ~~~ [ CodeView ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseDebugCodeViewInfo parses the CodeView debug data of the given debug data
// directory contents.
func parseDebugCodeViewInfo(dbgDir DebugDirectory, buf []byte) (*DebugCodeView, error) {
	var raw pe.RawCodeViewInfo
	r := bytes.NewReader(buf)
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		return nil, errors.WithStack(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pdbPath := parseCString(b)
	dbgCodeView := &DebugCodeView{
		DbgDir:       dbgDir,
		CodeViewInfo: goCodeViewInfo(raw, pdbPath),
	}
	return dbgCodeView, nil
}

// ~~~ [ FPO ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// parseDebugFPO parses the FPO debug data of the given debug data directory
// contents.
func parseDebugFPO(dbgDir DebugDirectory, buf []byte) (*DebugFPO, error) {
	r := bytes.NewReader(buf)
	var fpoData []FPOData
	for {
		var raw pe.RawFPOData
		if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return nil, errors.WithStack(err)
		}
		fpoData = append(fpoData, goFPOData(raw))
	}
	dbgFPO := &DebugFPO{
		DbgDir:  dbgDir,
		FPOData: fpoData,
	}
	return dbgFPO, nil
}
