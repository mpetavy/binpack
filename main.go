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

type fileitem struct {
	fullname string
	relname  string
}

var paths pathlist
var packageName *string
var variablePrefix *string
var fileName *string

func init() {
	common.Init("1.0.0", "2018", "Tool to embedd resource files into go source files", "mpetavy", common.APACHE, false, nil, nil, run, 0)
	flag.Var(&paths, "i", "include directory or file")
	packageName = flag.String("p", "main", "package name")
	variablePrefix = flag.String("v", "binpack", "variable prefix")
	fileName = flag.String("f", "binpack.go", "go file name")
}

func (i *pathlist) String() string {
	if i == nil {
		return ""
	}
	return strings.Join(paths, ",")
}

func (i *pathlist) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func filenameToVar(path string) string {
	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = path[1:]
	}

	path = strings.ReplaceAll(path, "-", " ")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "\\", "_")
	path = strings.ReplaceAll(path, ".", " ")

	path = strings.Title(path)

	path = strings.ReplaceAll(path, " ", "")

	return strings.Title(*variablePrefix) + "_" + path
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
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"flag\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"fmt\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"github.com/mpetavy/common\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"io\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"io/ioutil\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"os\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"path/filepath\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"strings\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	\"time\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    )\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("type %sFile struct {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    File string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    FileDate string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    MimeType string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    Content string\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("var (\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	%sReadFiles *bool\n", *variablePrefix)))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sFiles = make(map[string]*%sFile)\n", strings.Title(*variablePrefix), strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

	files := make([]fileitem, 0)

	for _, dir := range paths {
		recursive := !strings.HasSuffix(dir, string(filepath.Separator))
		if !recursive {
			dir = dir[:len(dir)-1]
		}

		cleanDir := common.CleanPath(dir)

		err := common.WalkFilepath(cleanDir, recursive, func(file string) error {
			files = append(files, fileitem{
				fullname: file,
				relname:  file[len(filepath.Dir(cleanDir))+1:],
			})

			return nil
		})

		if common.Error(err) {
			return err
		}
	}

	for _, fileItem := range files {
		varname := filenameToVar(fileItem.relname)

		common.Info("binpack file: %s", fileItem.relname)

		ba, err := ioutil.ReadFile(fileItem.fullname)
		if common.Error(err) {
			return err
		}

		mt := common.DetectMimeType(fileItem.fullname, ba)

		fd, err := common.FileDate(fileItem.fullname)
		if common.Error(err) {
			return err
		}

		fileDate, err := fd.MarshalText()
		if common.Error(err) {
			return err
		}

		buf := new(bytes.Buffer)

		zipWriter := zip.NewWriter(buf)
		zipFile, err := zipWriter.Create(fileItem.relname)
		if common.Error(err) {
			return err
		}

		_, err = zipFile.Write(ba)
		if common.Error(err) {
			return err
		}

		err = zipWriter.Close()
		if common.Error(err) {
			return err
		}

		var content string

		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %s = &%sFile{\n", varname, strings.Title(*variablePrefix))))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        File: \"%s\",\n", filepath.ToSlash(fileItem.relname))))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        FileDate: \"%s\",\n", fileDate)))
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        MimeType: \"%s\",\n", mt.MimeType)))

		if len(ba) > buf.Len() {
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        Content: \"zip:\"+\n")))
			content = hex.EncodeToString(buf.Bytes())
		} else {
			common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        Content:\n")))
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
		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    }\n")))
	}

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf(")\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func init() {")))

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	%sReadFiles = flag.Bool(\"%s.readfiles\", !common.IsRunningAsExecutable(), \"Read resource from filesystem\")\n", *variablePrefix, *variablePrefix)))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))

	for _, fileItem := range files {
		varname := filenameToVar(fileItem.relname)

		common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    %sFiles[\"%s\"] = %s\n", strings.Title(*variablePrefix), filepath.ToSlash(fileItem.relname), varname)))
	}

	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	common.RegisterResourceLoader(func(name string) []byte {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		l := make([]string, 0)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		for k := range %sFiles {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if filepath.Base(k) == filepath.Base(name) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				l = append(l, k)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if len(l) == 0 {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if len(l) > 1 {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			common.WarnError(fmt.Errorf(\"multiple resources with name %s found: %s\", name, l))\n", "%%s", "%%v")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		b := %sFiles[l[0]]\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		ba, err := b.Unpack()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return ba\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	})\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func (this *%sFile) Unpack() ([]byte, error) {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("  if *%sReadFiles {\n", *variablePrefix)))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    filename := \"\"\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    b, _ := common.FileExists(filepath.Base(this.File))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    if b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      filename = filepath.Base(this.File)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    } else {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      b, _ = common.FileExists(this.File)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      if b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        filename = this.File\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    if b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      common.Warn(\"Read resource from filesystem: %s\",filename)\n", "%%s")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      ba, err := ioutil.ReadFile(filename)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      if err == nil {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("        return ba, nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("      }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("    }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("  }\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if strings.HasPrefix(this.Content, \"zip:\") {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		ba, err := hex.DecodeString(this.Content[4:])\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return ba, err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		br := bytes.NewReader(ba)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		zr, err := zip.NewReader(br, int64(len(ba)))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return ba, err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		buf := bytes.Buffer{}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		for _, f := range zr.File {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			i, err := f.Open()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				return ba, err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			defer func() {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				common.DebugError(i.Close())\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			_, err = io.Copy(&buf, i)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				return buf.Bytes(), err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return buf.Bytes(), nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	} else {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		ba, err := hex.DecodeString(this.Content)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			return ba, err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return ba, nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func (this *%sFile) UnpackFile(path string) error {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	fileName := filepath.ToSlash(filepath.Join(path, filepath.Base(this.File)))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	var fileDate time.Time\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	err := fileDate.UnmarshalText([]byte(this.FileDate))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	fs, _ := os.Stat(fileName)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	b := fs == nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if !b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		b = fileDate != fs.ModTime()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if !b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		common.DebugFunc(\"Unpack file %s --> %s [skip]\", this.File, fileName)\n", "%%s", "%%s")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	common.DebugFunc(\"Unpack file %s --> %s\", this.File, fileName)\n", "%%s", "%%s")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	ba, err := this.Unpack()\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	err = ioutil.WriteFile(fileName, ba, common.DefaultFileMode)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	err = os.Chtimes(fileName, fileDate, fileDate)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	return nil\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("func UnpackDir(src string, dest string) error {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	common.DebugFunc(\"Unpack dir %s --> %s\", src, dest)\n", "%%s", "%%s")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	for _, file := range %sFiles {\n", strings.Title(*variablePrefix))))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		if strings.HasPrefix(file.File,src) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			fn := filepath.ToSlash(filepath.Clean(filepath.Join(dest, file.File)))\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			path := filepath.Dir(fn)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			b, err := common.FileExists(path)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			if !b {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				err = os.MkdirAll(path, common.DefaultDirMode)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				if common.Error(err) {\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("					return err\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("				}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("			err = file.UnpackFile(path)\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("		}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	}\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("\n")))
	common.Ignore(fmt.Fprintf(goFile, fmt.Sprintf("	return nil\n")))
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
	//UnpackDir("/go/src", "d:/temp")
	//os.Exit(0)

	defer common.Done()

	common.Run([]string{"i"})
}
