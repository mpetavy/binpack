package main

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type pathlist []string

var dirs pathlist
var packageName *string
var variablePrefix *string
var fileName *string

func init() {
	common.Init("1.0.0", "2018", "Tool to embedd resource files into go source files", "mpetavy", common.APACHE, false, nil, nil, run, 0)
	flag.Var(&dirs, "i", "include directory or file")
	packageName = flag.String("p", "main", "package name")
	variablePrefix = flag.String("v", "binpack", "package name")
	fileName = flag.String("f", "binpack.go", "file name for output")
}

func (i *pathlist) String() string {
	return "my string representation"
}

func (i *pathlist) Find(path string) int {
	for i, item := range *i {
		if item == path {
			return i
		}
	}

	return -1
}

func (i *pathlist) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func run() error {
	goFile, err := os.Create(*fileName)
	if common.Error(err) {
		return err
	}

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("package %s\n", *packageName)))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("import (\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"archive/zip\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"bytes\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"encoding/hex\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"github.com/mpetavy/common\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"io\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"io/ioutil\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"os\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"path/filepath\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"strings\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    )\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("type %sFile struct {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    Dir string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    File string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    Content string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("var %sDirs []string\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("var %sFiles []%sFile\n", strings.Title(*variablePrefix), strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, "\n"))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func init() {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sDirs = make([]string,0)\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sFiles = make([]%sFile,0)\n", strings.Title(*variablePrefix), strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, "\n"))

	for _, dir := range dirs {
		common.Info("binpack dir: %s", dir)

		files := make(pathlist, 0)

		notRecursive := strings.HasSuffix(dir, string(filepath.Separator))
		if notRecursive {
			dir = dir[:len(dir)-1]
		}

		cleanDir := common.CleanPath(dir)
		if common.ContainsWildcard(cleanDir) {
			cleanDir = filepath.Dir(cleanDir)
		}

		err := common.WalkFilepath(common.CleanPath(dir), !notRecursive, func(file string) error {
			files = append(files, file)

			return nil
		})

		if common.ContainsWildcard(dir) {
			dir = filepath.Dir(dir)
		}

		if common.Error(err) {
			return err
		}

		if len(files) > 0 {
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sDirs = append(%sDirs,\"%s\")\n", strings.Title(*variablePrefix), strings.Title(*variablePrefix), filepath.ToSlash(dir))))

			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sFiles = append(%sFiles,[]%sFile{\n", strings.Title(*variablePrefix), strings.Title(*variablePrefix), strings.Title(*variablePrefix))))
			for _, file := range files {
				common.Info("binpack file: %s", file)

				ba, err := ioutil.ReadFile(file)
				if common.Error(err) {
					return err
				}

				buf := new(bytes.Buffer)

				w := zip.NewWriter(buf)
				f, err := w.Create(file)
				if common.Error(err) {
					return err
				}

				_, err = f.Write(ba)
				if common.Error(err) {
					return err
				}

				err = w.Close()
				if common.Error(err) {
					return err
				}

				cleanFile := common.CleanPath(file)
				cleanFile = cleanFile[len(cleanDir)+1:]

				var content string

				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        %sFile{\n", strings.Title(*variablePrefix))))
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            Dir: \"%s\",\n", filepath.ToSlash(dir))))
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            File: \"%s\",\n", filepath.ToSlash(cleanFile))))
				if len(ba) > buf.Len() {
					common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            Content: \"zip:\"+\n")))
					content = hex.EncodeToString(buf.Bytes())
				} else {
					common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            Content:\n")))
					content = hex.EncodeToString(ba)
				}

				for index := 0; index < len(content); index += 1000 {
					end := common.Min(index+1000, len(content))
					ch := "+"

					if end == len(content) {
						ch = ","
					}
					common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                \"%s\"%s\n", content[index:end], ch)))
				}
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            },\n")))
			}
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        }...,)\n")))
		}
	}
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func (this *%sFile) Unpack(path string) error {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	fn := filepath.ToSlash(filepath.Join(path, filepath.Base(this.File)))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	file, err := os.Create(fn)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	defer func() {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		common.DebugError(file.Close())\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if strings.HasPrefix(this.Content, \"zip:\") {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		ba, err := hex.DecodeString(this.Content[4:])\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		br := bytes.NewReader(ba)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		zr, err := zip.NewReader(br, int64(len(ba)))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		for _, f := range zr.File {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			i, err := f.Open()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			defer func() {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				common.DebugError(i.Close())\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			_, err = io.Copy(file, i)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	ba, err := hex.DecodeString(this.Content)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	err = ioutil.WriteFile(fn, ba, common.FileMode(true, true, false))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func UnpackDir(src string,dest string) error {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    for _, file := range %sFiles {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        if file.Dir == src {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            fn := filepath.ToSlash(filepath.Join(dest,file.Dir,file.File))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            path := filepath.Dir(fn)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            b, err := common.FileExists(path)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            if !b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                err = os.MkdirAll(path, common.FileMode(true, true, true))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                    return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            err = file.Unpack(path)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

	err = goFile.Close()
	if common.Error(err) {
		return err
	}

	cmd := exec.Command("gofmt", "-w", *fileName)
	err = cmd.Run()
	if common.Error(err) {
		return err
	}

	return nil
}

func main() {
	//UnpackDir("static","d:/temp")
	//os.Exit(0)

	defer common.Done()

	common.Run(nil)
}
