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
	files := make(pathlist, 0)

	for _, path := range dirs {
		notRecursive := strings.HasSuffix(path, string(filepath.Separator))
		if notRecursive {
			path = path[:len(path)-1]
		}

		err := common.WalkFilepath(path, !notRecursive, func(file string) error {
			files = append(files, file)

			return nil
		})

		if common.Error(err) {
			return err
		}
	}

	if len(files) > 0 {
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
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("type %sFile struct {\n", *variablePrefix)))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    file string\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    content string\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("var %sDirs []string\n", strings.Title(*variablePrefix))))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("var %sFiles []%sFile\n", strings.Title(*variablePrefix), *variablePrefix)))
		common.Ignore(fmt.Fprintf(goFile, "\n"))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func init() {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sDirs = []string{\n", strings.Title(*variablePrefix))))

		for _, dir := range dirs {
			if common.IsWindowsOS() {
				dir = filepath.ToSlash(dir)
			}
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    	\"%s\",\n", dir)))
		}
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    	}\n")))

		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sFiles = []%sFile{\n", strings.Title(*variablePrefix), *variablePrefix)))
		for _, file := range files {
			ba, err := ioutil.ReadFile(file)
			buf := new(bytes.Buffer)

			w := zip.NewWriter(buf)
			f, err := w.Create(file)
			if common.Error(err) {
				return err
			}

			f.Write(ba)

			err = w.Close()
			if common.Error(err) {
				return err
			}

			if common.Error(err) {
				return err
			}
			file = filepath.ToSlash(file)

			var content string

			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        {\n")))
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            file: \"%s\",\n", file)))
			if len(ba) > buf.Len() {
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            content: \"zip:\"+\n")))
				content = hex.EncodeToString(buf.Bytes())
			} else {
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            content:\n")))
				content = hex.EncodeToString(ba)
			}

			for index := 0; index < len(content); index += 1000 {
				end := common.Min(index+1000, len(content))
				ch := "+"

				if end == len(content) {
					ch = ","
				}
				common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("                       \"%s\"%s\n",content[index:end], ch)))
			}
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        },\n")))
		}
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    }\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func (this *binpackFile) Extract(path string) error {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	fn := filepath.ToSlash(filepath.Join(path, filepath.Base(this.file)))\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	file, err := os.Create(fn)\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	defer func() {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		common.DebugError(file.Close())\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}()\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if strings.HasPrefix(this.content, \"zip:\") {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		ba, err := hex.DecodeString(this.content[4:])\n")))
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
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	ba, err := hex.DecodeString(this.content)\n")))
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
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func ExtractDir(src string,dest string) error {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    prefix := filepath.Dir(src)\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    for _, file := range BinpackFiles {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        b, err := common.EqualWildcards(file.file, filepath.ToSlash(src))\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        if common.Error(err) {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            return err\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        }\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        if b {\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            fn := filepath.Join(dest,file.file[len(prefix):])\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            path := filepath.Dir(fn)\n")))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            b, err = common.FileExists(path)\n")))
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
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("            err = file.Extract(path)\n")))
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
	}

	return nil
}

func main() {
	//ExtractBinpackDir("d:\\go\\src\\*.bat","d:\\temp")
	//BinpackFiles[0].Extract("d:/temp")
	//os.Exit(0)

	defer common.Done()

	common.Run(nil)
}

