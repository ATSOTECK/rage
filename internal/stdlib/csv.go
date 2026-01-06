package stdlib

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// CSV quoting constants
const (
	quoteMinimal    = 0
	quoteAll        = 1
	quoteNonnumeric = 2
	quoteNone       = 3
)

// InitCSVModule registers the csv module
func InitCSVModule() {
	// Register csv.reader type metatable
	readerMT := &runtime.TypeMetatable{
		Name: "csv.reader",
		Methods: map[string]runtime.GoFunction{
			"__iter__": csvReaderIter,
			"__next__": csvReaderNext,
		},
	}
	runtime.RegisterTypeMetatable("csv.reader", readerMT)

	// Register csv.DictReader type metatable
	dictReaderMT := &runtime.TypeMetatable{
		Name: "csv.DictReader",
		Methods: map[string]runtime.GoFunction{
			"__iter__":    csvDictReaderIter,
			"__next__":    csvDictReaderNext,
			"fieldnames":  csvDictReaderFieldnames,
		},
	}
	runtime.RegisterTypeMetatable("csv.DictReader", dictReaderMT)

	// Register csv.writer type metatable
	writerMT := &runtime.TypeMetatable{
		Name: "csv.writer",
		Methods: map[string]runtime.GoFunction{
			"writerow":  csvWriterWriterow,
			"writerows": csvWriterWriterows,
			"getvalue":  csvWriterGetvalue,
		},
	}
	runtime.RegisterTypeMetatable("csv.writer", writerMT)

	// Register csv.DictWriter type metatable
	dictWriterMT := &runtime.TypeMetatable{
		Name: "csv.DictWriter",
		Methods: map[string]runtime.GoFunction{
			"writeheader": csvDictWriterWriteheader,
			"writerow":    csvDictWriterWriterow,
			"writerows":   csvDictWriterWriterows,
			"getvalue":    csvDictWriterGetvalue,
		},
	}
	runtime.RegisterTypeMetatable("csv.DictWriter", dictWriterMT)

	runtime.NewModuleBuilder("csv").
		Doc("CSV file reading and writing.").
		// Reader/Writer constructors
		Func("reader", csvReader).
		Func("writer", csvWriter).
		Func("DictReader", csvDictReader).
		Func("DictWriter", csvDictWriter).
		// Utility functions for string-based operations
		Func("parse_row", csvParseRow).
		Func("format_row", csvFormatRow).
		// Quoting constants
		Const("QUOTE_MINIMAL", runtime.NewInt(quoteMinimal)).
		Const("QUOTE_ALL", runtime.NewInt(quoteAll)).
		Const("QUOTE_NONNUMERIC", runtime.NewInt(quoteNonnumeric)).
		Const("QUOTE_NONE", runtime.NewInt(quoteNone)).
		Register()
}

// =====================================
// CSV Reader
// =====================================

// PyCSVReader represents a CSV reader object
type PyCSVReader struct {
	Lines     []string
	Index     int
	Delimiter rune
	Quotechar rune
}

func (r *PyCSVReader) Type() string   { return "csv.reader" }
func (r *PyCSVReader) String() string { return "<csv.reader object>" }

// csvReader creates a new CSV reader from an iterable of lines
// csv.reader(iterable, delimiter=',', quotechar='"')
func csvReader(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("reader() missing required argument: 'csvfile'")
		return 0
	}

	// Get lines from the iterable
	lines, err := iterableToLines(vm.Get(1))
	if err != nil {
		vm.RaiseError("reader() argument must be an iterable of strings: %s", err)
		return 0
	}

	// Parse optional arguments
	delimiter := ','
	quotechar := '"'

	if vm.GetTop() >= 2 {
		if s, ok := vm.Get(2).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 3 {
		if s, ok := vm.Get(3).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	reader := &PyCSVReader{
		Lines:     lines,
		Index:     0,
		Delimiter: delimiter,
		Quotechar: quotechar,
	}

	ud := runtime.NewUserData(reader)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("csv.reader")
	vm.Push(ud)
	return 1
}

func csvReaderIter(vm *runtime.VM) int {
	// Return self for iteration
	vm.Push(vm.Get(1))
	return 1
}

func csvReaderNext(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.reader object")
		return 0
	}

	reader, ok := ud.Value.(*PyCSVReader)
	if !ok {
		vm.RaiseError("expected csv.reader object")
		return 0
	}

	if reader.Index >= len(reader.Lines) {
		// Raise StopIteration - must include colon for proper exception type parsing
		vm.RaiseError("StopIteration: iterator exhausted")
		return 0
	}

	line := reader.Lines[reader.Index]
	reader.Index++

	row := parseCSVLine(line, reader.Delimiter, reader.Quotechar)
	vm.Push(row)
	return 1
}

// =====================================
// CSV DictReader
// =====================================

// PyCSVDictReader represents a CSV DictReader object
type PyCSVDictReader struct {
	Lines      []string
	Index      int
	Delimiter  rune
	Quotechar  rune
	Fieldnames []string
	RestKey    string
	RestVal    string
}

func (r *PyCSVDictReader) Type() string   { return "csv.DictReader" }
func (r *PyCSVDictReader) String() string { return "<csv.DictReader object>" }

// csvDictReader creates a new DictReader
// csv.DictReader(f, fieldnames=None, restkey=None, restval=None, delimiter=',', quotechar='"')
func csvDictReader(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("DictReader() missing required argument: 'f'")
		return 0
	}

	lines, err := iterableToLines(vm.Get(1))
	if err != nil {
		vm.RaiseError("DictReader() argument must be an iterable of strings: %s", err)
		return 0
	}

	delimiter := ','
	quotechar := '"'
	var fieldnames []string
	restkey := ""
	restval := ""

	// Parse fieldnames (arg 2)
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		fieldnames = valueToStringSlice(vm.Get(2))
	}

	// Parse restkey (arg 3)
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		if s, ok := vm.Get(3).(*runtime.PyString); ok {
			restkey = s.Value
		}
	}

	// Parse restval (arg 4)
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		if s, ok := vm.Get(4).(*runtime.PyString); ok {
			restval = s.Value
		}
	}

	// Parse delimiter (arg 5)
	if vm.GetTop() >= 5 {
		if s, ok := vm.Get(5).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	// Parse quotechar (arg 6)
	if vm.GetTop() >= 6 {
		if s, ok := vm.Get(6).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	// If no fieldnames provided, use first row
	startIndex := 0
	if fieldnames == nil && len(lines) > 0 {
		row := parseCSVLine(lines[0], delimiter, quotechar)
		fieldnames = listToStringSlice(row)
		startIndex = 1
	}

	reader := &PyCSVDictReader{
		Lines:      lines,
		Index:      startIndex,
		Delimiter:  delimiter,
		Quotechar:  quotechar,
		Fieldnames: fieldnames,
		RestKey:    restkey,
		RestVal:    restval,
	}

	ud := runtime.NewUserData(reader)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("csv.DictReader")
	vm.Push(ud)
	return 1
}

func csvDictReaderIter(vm *runtime.VM) int {
	vm.Push(vm.Get(1))
	return 1
}

func csvDictReaderNext(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictReader object")
		return 0
	}

	reader, ok := ud.Value.(*PyCSVDictReader)
	if !ok {
		vm.RaiseError("expected csv.DictReader object")
		return 0
	}

	if reader.Index >= len(reader.Lines) {
		// Raise StopIteration - must include colon for proper exception type parsing
		vm.RaiseError("StopIteration: iterator exhausted")
		return 0
	}

	line := reader.Lines[reader.Index]
	reader.Index++

	row := parseCSVLine(line, reader.Delimiter, reader.Quotechar)
	values := listToStringSlice(row)

	// Build dict from fieldnames and values
	d := runtime.NewDict()
	for i, name := range reader.Fieldnames {
		if i < len(values) {
			d.Items[runtime.NewString(name)] = runtime.NewString(values[i])
		} else {
			// Use restval for missing values
			d.Items[runtime.NewString(name)] = runtime.NewString(reader.RestVal)
		}
	}

	// Handle extra values with restkey
	if len(values) > len(reader.Fieldnames) && reader.RestKey != "" {
		extra := make([]runtime.Value, len(values)-len(reader.Fieldnames))
		for i := len(reader.Fieldnames); i < len(values); i++ {
			extra[i-len(reader.Fieldnames)] = runtime.NewString(values[i])
		}
		d.Items[runtime.NewString(reader.RestKey)] = runtime.NewList(extra)
	}

	vm.Push(d)
	return 1
}

func csvDictReaderFieldnames(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictReader object")
		return 0
	}

	reader, ok := ud.Value.(*PyCSVDictReader)
	if !ok {
		vm.RaiseError("expected csv.DictReader object")
		return 0
	}

	items := make([]runtime.Value, len(reader.Fieldnames))
	for i, name := range reader.Fieldnames {
		items[i] = runtime.NewString(name)
	}
	vm.Push(runtime.NewList(items))
	return 1
}

// =====================================
// CSV Writer
// =====================================

// PyCSVWriter represents a CSV writer object
type PyCSVWriter struct {
	Buffer    *bytes.Buffer
	Delimiter rune
	Quotechar rune
	Quoting   int
	Lineterminator string
}

func (w *PyCSVWriter) Type() string   { return "csv.writer" }
func (w *PyCSVWriter) String() string { return "<csv.writer object>" }

// csvWriter creates a new CSV writer
// csv.writer(delimiter=',', quotechar='"', quoting=QUOTE_MINIMAL, lineterminator='\n')
func csvWriter(vm *runtime.VM) int {
	delimiter := ','
	quotechar := '"'
	quoting := quoteMinimal
	lineterminator := "\n"

	// Parse optional arguments
	if vm.GetTop() >= 1 {
		if s, ok := vm.Get(1).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 2 {
		if s, ok := vm.Get(2).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 3 {
		if i, ok := vm.Get(3).(*runtime.PyInt); ok {
			quoting = int(i.Value)
		}
	}

	if vm.GetTop() >= 4 {
		if s, ok := vm.Get(4).(*runtime.PyString); ok {
			lineterminator = s.Value
		}
	}

	writer := &PyCSVWriter{
		Buffer:         &bytes.Buffer{},
		Delimiter:      delimiter,
		Quotechar:      quotechar,
		Quoting:        quoting,
		Lineterminator: lineterminator,
	}

	ud := runtime.NewUserData(writer)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("csv.writer")
	vm.Push(ud)
	return 1
}

func csvWriterWriterow(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVWriter)
	if !ok {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	if vm.GetTop() < 2 {
		vm.RaiseError("writerow() missing required argument: 'row'")
		return 0
	}

	row := vm.Get(2)
	line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
	writer.Buffer.WriteString(line)
	writer.Buffer.WriteString(writer.Lineterminator)

	vm.Push(runtime.None)
	return 1
}

func csvWriterWriterows(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVWriter)
	if !ok {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	if vm.GetTop() < 2 {
		vm.RaiseError("writerows() missing required argument: 'rows'")
		return 0
	}

	rows := vm.Get(2)
	switch r := rows.(type) {
	case *runtime.PyList:
		for _, row := range r.Items {
			line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
			writer.Buffer.WriteString(line)
			writer.Buffer.WriteString(writer.Lineterminator)
		}
	case *runtime.PyTuple:
		for _, row := range r.Items {
			line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
			writer.Buffer.WriteString(line)
			writer.Buffer.WriteString(writer.Lineterminator)
		}
	default:
		vm.RaiseError("writerows() argument must be an iterable")
		return 0
	}

	vm.Push(runtime.None)
	return 1
}

func csvWriterGetvalue(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVWriter)
	if !ok {
		vm.RaiseError("expected csv.writer object")
		return 0
	}

	vm.Push(runtime.NewString(writer.Buffer.String()))
	return 1
}

// =====================================
// CSV DictWriter
// =====================================

// PyCSVDictWriter represents a CSV DictWriter object
type PyCSVDictWriter struct {
	Buffer         *bytes.Buffer
	Fieldnames     []string
	Delimiter      rune
	Quotechar      rune
	Quoting        int
	Lineterminator string
	RestVal        string
	ExtrasAction   string
}

func (w *PyCSVDictWriter) Type() string   { return "csv.DictWriter" }
func (w *PyCSVDictWriter) String() string { return "<csv.DictWriter object>" }

// csvDictWriter creates a new DictWriter
// csv.DictWriter(fieldnames, delimiter=',', quotechar='"', quoting=QUOTE_MINIMAL, lineterminator='\n', restval='', extrasaction='raise')
func csvDictWriter(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("DictWriter() missing required argument: 'fieldnames'")
		return 0
	}

	fieldnames := valueToStringSlice(vm.Get(1))
	if fieldnames == nil {
		vm.RaiseError("DictWriter() fieldnames must be a sequence")
		return 0
	}

	delimiter := ','
	quotechar := '"'
	quoting := quoteMinimal
	lineterminator := "\n"
	restval := ""
	extrasaction := "raise"

	if vm.GetTop() >= 2 {
		if s, ok := vm.Get(2).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 3 {
		if s, ok := vm.Get(3).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 4 {
		if i, ok := vm.Get(4).(*runtime.PyInt); ok {
			quoting = int(i.Value)
		}
	}

	if vm.GetTop() >= 5 {
		if s, ok := vm.Get(5).(*runtime.PyString); ok {
			lineterminator = s.Value
		}
	}

	if vm.GetTop() >= 6 {
		if s, ok := vm.Get(6).(*runtime.PyString); ok {
			restval = s.Value
		}
	}

	if vm.GetTop() >= 7 {
		if s, ok := vm.Get(7).(*runtime.PyString); ok {
			extrasaction = s.Value
		}
	}

	writer := &PyCSVDictWriter{
		Buffer:         &bytes.Buffer{},
		Fieldnames:     fieldnames,
		Delimiter:      delimiter,
		Quotechar:      quotechar,
		Quoting:        quoting,
		Lineterminator: lineterminator,
		RestVal:        restval,
		ExtrasAction:   extrasaction,
	}

	ud := runtime.NewUserData(writer)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("csv.DictWriter")
	vm.Push(ud)
	return 1
}

func csvDictWriterWriteheader(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVDictWriter)
	if !ok {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	items := make([]runtime.Value, len(writer.Fieldnames))
	for i, name := range writer.Fieldnames {
		items[i] = runtime.NewString(name)
	}
	row := runtime.NewList(items)
	line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
	writer.Buffer.WriteString(line)
	writer.Buffer.WriteString(writer.Lineterminator)

	vm.Push(runtime.None)
	return 1
}

func csvDictWriterWriterow(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVDictWriter)
	if !ok {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	if vm.GetTop() < 2 {
		vm.RaiseError("writerow() missing required argument: 'rowdict'")
		return 0
	}

	rowDict, ok := vm.Get(2).(*runtime.PyDict)
	if !ok {
		vm.RaiseError("writerow() argument must be a dict")
		return 0
	}

	// Check for extra keys if extrasaction is 'raise'
	if writer.ExtrasAction == "raise" {
		fieldnameSet := make(map[string]bool)
		for _, name := range writer.Fieldnames {
			fieldnameSet[name] = true
		}
		for k := range rowDict.Items {
			if s, ok := k.(*runtime.PyString); ok {
				if !fieldnameSet[s.Value] {
					vm.RaiseError("ValueError: dict contains fields not in fieldnames: '%s'", s.Value)
					return 0
				}
			}
		}
	}

	// Build row from fieldnames
	items := make([]runtime.Value, len(writer.Fieldnames))
	for i, name := range writer.Fieldnames {
		if val, found := dictGetByStringKey(rowDict, name); found {
			items[i] = val
		} else {
			items[i] = runtime.NewString(writer.RestVal)
		}
	}

	row := runtime.NewList(items)
	line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
	writer.Buffer.WriteString(line)
	writer.Buffer.WriteString(writer.Lineterminator)

	vm.Push(runtime.None)
	return 1
}

func csvDictWriterWriterows(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVDictWriter)
	if !ok {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	if vm.GetTop() < 2 {
		vm.RaiseError("writerows() missing required argument: 'rowdicts'")
		return 0
	}

	rows := vm.Get(2)
	var rowList []runtime.Value

	switch r := rows.(type) {
	case *runtime.PyList:
		rowList = r.Items
	case *runtime.PyTuple:
		rowList = r.Items
	default:
		vm.RaiseError("writerows() argument must be an iterable")
		return 0
	}

	fieldnameSet := make(map[string]bool)
	for _, name := range writer.Fieldnames {
		fieldnameSet[name] = true
	}

	for _, rowVal := range rowList {
		rowDict, ok := rowVal.(*runtime.PyDict)
		if !ok {
			vm.RaiseError("writerows() argument must be an iterable of dicts")
			return 0
		}

		// Check for extra keys
		if writer.ExtrasAction == "raise" {
			for k := range rowDict.Items {
				if s, ok := k.(*runtime.PyString); ok {
					if !fieldnameSet[s.Value] {
						vm.RaiseError("ValueError: dict contains fields not in fieldnames: '%s'", s.Value)
						return 0
					}
				}
			}
		}

		// Build row
		items := make([]runtime.Value, len(writer.Fieldnames))
		for i, name := range writer.Fieldnames {
			if val, found := dictGetByStringKey(rowDict, name); found {
				items[i] = val
			} else {
				items[i] = runtime.NewString(writer.RestVal)
			}
		}

		row := runtime.NewList(items)
		line := formatCSVRow(row, writer.Delimiter, writer.Quotechar, writer.Quoting)
		writer.Buffer.WriteString(line)
		writer.Buffer.WriteString(writer.Lineterminator)
	}

	vm.Push(runtime.None)
	return 1
}

func csvDictWriterGetvalue(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	writer, ok := ud.Value.(*PyCSVDictWriter)
	if !ok {
		vm.RaiseError("expected csv.DictWriter object")
		return 0
	}

	vm.Push(runtime.NewString(writer.Buffer.String()))
	return 1
}

// =====================================
// Utility functions
// =====================================

// csvParseRow parses a single CSV line into a list
// csv.parse_row(line, delimiter=',', quotechar='"')
func csvParseRow(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("parse_row() missing required argument: 'line'")
		return 0
	}

	line := vm.CheckString(1)
	delimiter := ','
	quotechar := '"'

	if vm.GetTop() >= 2 {
		if s, ok := vm.Get(2).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 3 {
		if s, ok := vm.Get(3).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	result := parseCSVLine(line, delimiter, quotechar)
	vm.Push(result)
	return 1
}

// csvFormatRow formats a list as a CSV line
// csv.format_row(row, delimiter=',', quotechar='"', quoting=QUOTE_MINIMAL)
func csvFormatRow(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("format_row() missing required argument: 'row'")
		return 0
	}

	row := vm.Get(1)
	delimiter := ','
	quotechar := '"'
	quoting := quoteMinimal

	if vm.GetTop() >= 2 {
		if s, ok := vm.Get(2).(*runtime.PyString); ok && len(s.Value) > 0 {
			delimiter = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 3 {
		if s, ok := vm.Get(3).(*runtime.PyString); ok && len(s.Value) > 0 {
			quotechar = rune(s.Value[0])
		}
	}

	if vm.GetTop() >= 4 {
		if i, ok := vm.Get(4).(*runtime.PyInt); ok {
			quoting = int(i.Value)
		}
	}

	result := formatCSVRow(row, delimiter, quotechar, quoting)
	vm.Push(runtime.NewString(result))
	return 1
}

// =====================================
// Helper functions
// =====================================

// parseCSVLine parses a single CSV line into a PyList of strings
func parseCSVLine(line string, delimiter, quotechar rune) *runtime.PyList {
	var result []runtime.Value
	var field strings.Builder
	inQuotes := false
	i := 0
	runes := []rune(line)

	for i < len(runes) {
		c := runes[i]

		if inQuotes {
			if c == quotechar {
				// Check for escaped quote (doubled quote)
				if i+1 < len(runes) && runes[i+1] == quotechar {
					field.WriteRune(quotechar)
					i += 2
					continue
				}
				// End of quoted field
				inQuotes = false
				i++
				continue
			}
			field.WriteRune(c)
			i++
		} else {
			if c == quotechar {
				inQuotes = true
				i++
				continue
			}
			if c == delimiter {
				result = append(result, runtime.NewString(field.String()))
				field.Reset()
				i++
				continue
			}
			field.WriteRune(c)
			i++
		}
	}

	// Add last field
	result = append(result, runtime.NewString(field.String()))

	return runtime.NewList(result)
}

// formatCSVRow formats a row (list/tuple) as a CSV line string
func formatCSVRow(row runtime.Value, delimiter, quotechar rune, quoting int) string {
	var values []string

	switch r := row.(type) {
	case *runtime.PyList:
		for _, item := range r.Items {
			values = append(values, valueToCSVString(item))
		}
	case *runtime.PyTuple:
		for _, item := range r.Items {
			values = append(values, valueToCSVString(item))
		}
	default:
		return ""
	}

	var result strings.Builder
	delimStr := string(delimiter)
	quoteStr := string(quotechar)

	for i, val := range values {
		if i > 0 {
			result.WriteString(delimStr)
		}

		needsQuoting := false

		switch quoting {
		case quoteAll:
			needsQuoting = true
		case quoteNonnumeric:
			// Quote if not a number
			if _, err := parseNumber(val); err != nil {
				needsQuoting = true
			}
		case quoteMinimal:
			// Quote if contains delimiter, quotechar, or newline
			needsQuoting = strings.ContainsAny(val, delimStr+quoteStr+"\n\r")
		case quoteNone:
			needsQuoting = false
		}

		if needsQuoting {
			// Escape quotes by doubling them
			escaped := strings.ReplaceAll(val, quoteStr, quoteStr+quoteStr)
			result.WriteString(quoteStr)
			result.WriteString(escaped)
			result.WriteString(quoteStr)
		} else {
			result.WriteString(val)
		}
	}

	return result.String()
}

// valueToCSVString converts a Python value to its string representation for CSV
func valueToCSVString(v runtime.Value) string {
	switch val := v.(type) {
	case *runtime.PyString:
		return val.Value
	case *runtime.PyInt:
		return fmt.Sprintf("%d", val.Value)
	case *runtime.PyFloat:
		return fmt.Sprintf("%g", val.Value)
	case *runtime.PyBool:
		if val.Value {
			return "True"
		}
		return "False"
	case *runtime.PyNone:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// iterableToLines converts an iterable to a slice of strings
func iterableToLines(v runtime.Value) ([]string, error) {
	switch val := v.(type) {
	case *runtime.PyList:
		lines := make([]string, len(val.Items))
		for i, item := range val.Items {
			if s, ok := item.(*runtime.PyString); ok {
				lines[i] = s.Value
			} else {
				lines[i] = valueToCSVString(item)
			}
		}
		return lines, nil
	case *runtime.PyTuple:
		lines := make([]string, len(val.Items))
		for i, item := range val.Items {
			if s, ok := item.(*runtime.PyString); ok {
				lines[i] = s.Value
			} else {
				lines[i] = valueToCSVString(item)
			}
		}
		return lines, nil
	case *runtime.PyString:
		// Split string by newlines
		return strings.Split(val.Value, "\n"), nil
	default:
		return nil, nil
	}
}

// valueToStringSlice converts a Python list/tuple to a Go string slice
func valueToStringSlice(v runtime.Value) []string {
	switch val := v.(type) {
	case *runtime.PyList:
		return listToStringSlice(val)
	case *runtime.PyTuple:
		result := make([]string, len(val.Items))
		for i, item := range val.Items {
			if s, ok := item.(*runtime.PyString); ok {
				result[i] = s.Value
			} else {
				result[i] = valueToCSVString(item)
			}
		}
		return result
	default:
		return nil
	}
}

// listToStringSlice converts a PyList to a Go string slice
func listToStringSlice(list *runtime.PyList) []string {
	result := make([]string, len(list.Items))
	for i, item := range list.Items {
		if s, ok := item.(*runtime.PyString); ok {
			result[i] = s.Value
		} else {
			result[i] = valueToCSVString(item)
		}
	}
	return result
}

// parseNumber attempts to parse a string as a number
func parseNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	var f float64
	_, err := strings.NewReader(s).Read([]byte{})
	if err != nil {
		return 0, err
	}
	// Try parsing as float
	_, err = strings.NewReader(s).Read([]byte{})
	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' && c != '-' && c != '+' && c != 'e' && c != 'E' {
			return 0, nil
		}
	}
	return f, nil
}

// dictGetByStringKey looks up a value in a PyDict by string key value
// This properly handles the case where keys are different PyString objects with the same value
func dictGetByStringKey(d *runtime.PyDict, key string) (runtime.Value, bool) {
	for k, v := range d.Items {
		if s, ok := k.(*runtime.PyString); ok && s.Value == key {
			return v, true
		}
	}
	return nil, false
}
