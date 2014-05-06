package extract_strings

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"

	"encoding/json"
	"io/ioutil"

	common "gi18n/common"
)

type ExtractStrings struct {
	Options          common.Options
	Filename         string
	I18nFilename     string
	ExtractedStrings map[string]common.StringInfo
	FilteredStrings  map[string]string
	TotalStringsDir  int
	TotalStrings     int
	TotalFiles       int
}

func NewExtractStrings(options common.Options) ExtractStrings {
	return ExtractStrings{Options: options,
		Filename:         "extracted_strings.json",
		ExtractedStrings: nil,
		FilteredStrings:  nil,
		TotalStringsDir:  0,
		TotalStrings:     0,
		TotalFiles:       0}
}

func (es *ExtractStrings) Println(a ...interface{}) (int, error) {
	if es.Options.VerboseFlag {
		return fmt.Println(a...)
	}

	return 0, nil
}

func (es *ExtractStrings) Printf(msg string, a ...interface{}) (int, error) {
	if es.Options.VerboseFlag {
		return fmt.Printf(msg, a...)
	}

	return 0, nil
}

func (es *ExtractStrings) InspectFile(filename string) error {
	es.Println("gi18n: extracting strings from file:", filename)

	es.ExtractedStrings = make(map[string]common.StringInfo)
	es.FilteredStrings = make(map[string]string)

	es.setFilename(filename)
	es.setI18nFilename(filename)

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		es.Println(err)
		return err
	}

	err = es.loadExcludedStrings()
	if err != nil {
		es.Println(err)
		return err
	}

	es.excludeImports(astFile)

	es.extractString(astFile, fset)
	es.TotalStringsDir += len(es.ExtractedStrings)
	es.TotalStrings += len(es.ExtractedStrings)
	es.TotalFiles += 1

	es.Printf("Extracted %d strings from file: %s\n", len(es.ExtractedStrings), filename)

	err = es.saveExtractedStrings()
	if err != nil {
		es.Println(err)
		return err
	}

	err = es.saveI18nStrings()
	if err != nil {
		es.Println(err)
		return err
	}

	if es.Options.PoFlag {
		err = es.saveI18nStringsInPo()
		if err != nil {
			es.Println(err)
			return err
		}
	}

	return nil
}

func (es *ExtractStrings) InspectDir(dirName string, recursive bool) error {
	es.Printf("gi18n: inspecting dir %s, recursive: %t\n", dirName, recursive)
	es.Println()

	fset := token.NewFileSet()
	es.TotalStringsDir = 0

	packages, err := parser.ParseDir(fset, dirName, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		es.Println(err)
		return err
	}

	for k, pkg := range packages {
		es.Println("Extracting strings in package:", k)
		for fileName, _ := range pkg.Files {
			if !strings.HasPrefix(fileName, ".") && strings.HasSuffix(fileName, ".go") {
				err = es.InspectFile(fileName)
				if err != nil {
					es.Println(err)
				}
			}
		}
	}
	es.Printf("Extracted total of %d strings\n\n", es.TotalStringsDir)

	if recursive {
		fileInfos, _ := ioutil.ReadDir(dirName)
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() && !strings.HasPrefix(fileInfo.Name(), ".") {
				err = es.InspectDir(dirName + "/" + fileInfo.Name(), recursive)
				if err != nil {
					es.Println(err)
				}
			}
		}
	}

	return nil
}

func (es *ExtractStrings) saveExtractedStrings() error {
	es.Println("Saving extracted strings to file:", es.Filename)

	stringInfos := make([]common.StringInfo, 0)
	for _, stringInfo := range es.ExtractedStrings {
		stringInfos = append(stringInfos, stringInfo)
	}

	jsonData, err := json.Marshal(stringInfos)
	if err != nil {
		es.Println(err)
		return err
	}

	file, err := os.Create(es.Filename)
	if err != nil {
		es.Println(err)
		return err
	}

	file.Write(jsonData)
	defer file.Close()

	return nil
}

func (es *ExtractStrings) saveI18nStrings() error {
	es.Println("Saving extracted i18n strings to file:", es.I18nFilename)

	i18nStringInfos := make([]common.I18nStringInfo, len(es.ExtractedStrings))
	i := 0
	for _, stringInfo := range es.ExtractedStrings {
		i18nStringInfos[i] = common.I18nStringInfo{ID: stringInfo.Value, Translation: stringInfo.Value}
		i++
	}

	jsonData, err := json.Marshal(i18nStringInfos)
	if err != nil {
		es.Println(err)
		return err
	}

	file, err := os.Create(es.I18nFilename)
	if err != nil {
		es.Println(err)
		return err
	}

	file.Write(jsonData)
	defer file.Close()

	return nil
}

func (es *ExtractStrings) saveI18nStringsInPo() error {
	poFilename := es.I18nFilename[:len(es.I18nFilename)-len(".json")] + ".po"
	es.Println("Creating and saving i18n strings to .po file:", poFilename)

	file, err := os.Create(poFilename)
	if err != nil {
		es.Println(err)
		return err
	}

	for _, stringInfo := range es.ExtractedStrings {
		file.Write([]byte("# filename: " + stringInfo.Filename +
			", offset: " + strconv.Itoa(stringInfo.Offset) +
			", line: " + strconv.Itoa(stringInfo.Line) +
			", column: " + strconv.Itoa(stringInfo.Column) + "\n"))
		file.Write([]byte("msgid " + strconv.Quote(stringInfo.Value) + "\n"))
		file.Write([]byte("msgstr " + strconv.Quote(stringInfo.Value) + "\n"))
		file.Write([]byte("\n"))
	}

	defer file.Close()

	return nil
}

func (es *ExtractStrings) setFilename(filename string) {
	es.Filename = filename + ".extracted.json"
}

func (es *ExtractStrings) setI18nFilename(filename string) {
	es.I18nFilename = filename + ".en.json"
}

func (es *ExtractStrings) loadExcludedStrings() error {
	es.Println("Excluding strings in file:", es.Options.ExcludedFilenameFlag)

	content, err := ioutil.ReadFile(es.Options.ExcludedFilenameFlag)
	if err != nil {
		fmt.Print(err)
		return err
	}

	var excludedStrings common.ExcludedStrings
	err = json.Unmarshal(content, &excludedStrings)
	if err != nil {
		fmt.Print(err)
		return err
	}

	for i := range excludedStrings.ExcludedStrings {
		es.FilteredStrings[excludedStrings.ExcludedStrings[i]] = excludedStrings.ExcludedStrings[i]
	}

	return nil
}

func (es *ExtractStrings) extractString(f *ast.File, fset *token.FileSet) error {
	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s, _ = strconv.Unquote(x.Value)
			if len(s) > 0 && x.Kind == token.STRING && s != "\t" && s != "\n" && s != " " && !es.filter(s) { //TODO: fix to remove these: s != "\\t" && s != "\\n" && s != " "
				position := fset.Position(n.Pos())
				stringInfo := common.StringInfo{Value: s,
					Filename: position.Filename,
					Offset:   position.Offset,
					Line:     position.Line,
					Column:   position.Column}
				es.ExtractedStrings[s] = stringInfo
			}
		}
		return true
	})

	return nil
}

func (es *ExtractStrings) excludeImports(astFile *ast.File) {
	for i := range astFile.Imports {
		importString, _ := strconv.Unquote(astFile.Imports[i].Path.Value)
		es.FilteredStrings[importString] = importString
	}

}

func (es *ExtractStrings) filter(aString string) bool {
	for i := range common.BLANKS {
		if aString == common.BLANKS[i] {
			return true
		}
	}

	if es.FilteredStrings[aString] != "" {
		return true
	}
	return false
}