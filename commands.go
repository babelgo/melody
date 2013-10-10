package melody

import (
	"flag"
	"fmt"
	//"github.com/emicklei/hopwatch"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

const (
	MELODY_IMPORT_PATH = "github.com/babelgo/melody"
)

func (cli *MelodyCli) getMethod(name string) (reflect.Method, bool) {
	methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	return reflect.TypeOf(cli).MethodByName(methodName)
}

func ParseCommands(args ...string) error {
	cli := NewMelodyCli(os.Stdin, os.Stdout, os.Stderr)

	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
			return cli.CmdHelp(args[1:]...)
		}

		ret := method.Func.CallSlice([]reflect.Value{
			reflect.ValueOf(cli),
			reflect.ValueOf(args[1:]),
		})[0].Interface()

		if ret == nil {
			return nil
		}
		return ret.(error)
	}
	return cli.CmdHelp(args...)
}

func (cli *MelodyCli) CmdHelp(args ...string) error {
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
		} else {
			method.Func.CallSlice([]reflect.Value{
				reflect.ValueOf(cli),
				reflect.ValueOf([]string{"--help"}),
			})[0].Interface()
			return nil
		}
	}
	help := fmt.Sprintf("Usage: melody [OPTIONS] COMMAND [arg...]\n\nCommands:\n")
	for _, command := range [][]string{
		{"new", "Create a skeleton Melody application"},
		{"run", "Run a Melody application"},
		{"build", "Build a Melody application (e.g. for deployment)"},
		{"clean", "Clean a Melody application's temp files"},
		{"package", "Package a Melody application (e.g. for deployment)"},
		{"test", "Test a Melody application"},
	} {
		help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
	}
	fmt.Fprintf(cli.err, "%s\n", help)
	return nil
}

func (cli *MelodyCli) CmdNew(args ...string) error {
	cmd := Subcmd("new", "PATH", "Create a skeleton Melody application.")
	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	gopath := build.Default.GOPATH
	if gopath == "" {
		fmt.Fprintf(cli.err, "Melody: GOPATH environment variable is not set. "+
			"Please refer to http://golang.org/doc/code.html to configure your Go environment.\n")
		return nil
	}

	importPath := cmd.Arg(0)
	if path.IsAbs(importPath) {
		fmt.Fprintf(cli.err, "Melody: '%s' looks like a directory.  Please provide a Go import path instead.\n",
			importPath)
		return nil
	}

	_, err := build.Import(importPath, "", build.FindOnly)
	if err == nil {
		fmt.Fprintf(cli.err, "Melody: Import path %s already exists.\n", importPath)
		return nil
	}

	melodyPkg, err := build.Import(MELODY_IMPORT_PATH, "", build.FindOnly)
	if err != nil {
		fmt.Fprintf(cli.err, "Melody: Could not find Melody source code: %s\n", err)
		return nil
	}

	srcRoot := path.Join(filepath.SplitList(gopath)[0], "src")
	appDir := path.Join(srcRoot, filepath.FromSlash(importPath))
	if err = os.MkdirAll(appDir, 0777); err != nil {
		fmt.Fprintf(cli.err, "Melody: Failed to create directory: %s\n", appDir)
		return nil
	}

	skeletonBase := path.Join(melodyPkg.Dir, "skeleton")
	if err = mustCopyDir(appDir, skeletonBase); err != nil {
		fmt.Fprintf(cli.err, err.Error())
		return nil
	}

	// Dotfiles are skipped by mustCopyDir, so we have to explicitly copy the .gitignore.
	gitignore := ".gitignore"
	if err = mustCopyFile(path.Join(appDir, gitignore), path.Join(skeletonBase, gitignore)); err != nil {
		fmt.Fprintf(cli.err, err.Error())
		return nil
	}

	fmt.Fprintln(cli.out, "Your application is ready:\n  ", appDir)
	fmt.Fprintln(cli.out, "\nYou can run it with:\n   melody run", importPath)

	return nil
}

func (cli *MelodyCli) CmdRun(args ...string) error {
	cmd := Subcmd("run", "[OPTIONS] PATH", "Run a Melody application.")
	//mode := cmd.String("m", "dev", "The application run mode.")
	//port := cmd.String("p", "8000", "The applicaion run in port '8000'.")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	return nil
}

func (cli *MelodyCli) CmdBuild(args ...string) error {
	cmd := Subcmd("build", "", "Build a Melody application (e.g. for deployment).")
	if err := cmd.Parse(args); err != nil {
		return nil
	}
	return nil
}

func (cli *MelodyCli) CmdClean(args ...string) error {
	cmd := Subcmd("clean", "", "Clean a Melody application's temp files.")
	if err := cmd.Parse(args); err != nil {
		return nil
	}
	return nil
}

func (cli *MelodyCli) CmdPackage(args ...string) error {
	cmd := Subcmd("package", "", "Package a Melody application (e.g. for deployment).")
	if err := cmd.Parse(args); err != nil {
		return nil
	}
	return nil
}

func (cli *MelodyCli) CmdTest(args ...string) error {
	cmd := Subcmd("test", "", "Test a Melody application.")
	if err := cmd.Parse(args); err != nil {
		return nil
	}
	return nil
}

func NewMelodyCli(in io.ReadCloser, out io.Writer, err io.Writer) *MelodyCli {
	return &MelodyCli{
		in:  in,
		out: out,
		err: err,
	}
}

func Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		// FIXME: use custom stdout or return error
		fmt.Fprintf(os.Stdout, "\nUsage: melody %s %s\n\n%s\n\n", name, signature, description)
		flags.PrintDefaults()
	}
	return flags
}

type MelodyCli struct {
	in  io.ReadCloser
	out io.Writer
	err io.Writer
}

// copyDir copies a directory tree over to a new directory.  Any files ending in
// ".template" are treated as a Go template and rendered using the given data.
// Additionally, the trailing ".template" is stripped from the file name.
// Also, dot files and dot directories are skipped.
func mustCopyDir(destDir, srcDir string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		// Get the relative path from the source base, and the corresponding path in
		// the dest directory.
		relSrcPath := strings.TrimLeft(srcPath[len(srcDir):], string(os.PathSeparator))
		destPath := path.Join(destDir, relSrcPath)

		// Skip dot files and dot directories.
		if strings.HasPrefix(relSrcPath, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create a subdirectory if necessary.
		if info.IsDir() {
			if err := os.MkdirAll(path.Join(destDir, relSrcPath), 0777); err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("Melody: Failed to create directory.\n")
				}
			}
			return nil
		}

		// If this file ends in ".template", render it as a template.
		if strings.HasSuffix(relSrcPath, ".template") {
			mustRenderTemplate(destPath[:len(destPath)-len(".template")], srcPath)
			return nil
		}

		// Else, just copy it over.
		mustCopyFile(destPath, srcPath)
		return nil
	})
}

func mustCopyFile(destFilename, srcFilename string) error {
	destFile, err := os.Create(destFilename)
	if err != nil {
		return fmt.Errorf("Melody: Failed to create file %s.\n", destFilename)
	}

	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return fmt.Errorf("Melody: Failed to open file %s.\n", srcFilename)
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("Melody: Failed to copy data from %s to %s.\n", srcFile.Name(), destFile.Name())
	}

	if err = destFile.Close(); err != nil {
		return fmt.Errorf("Melody: Failed to close file %s.\n", destFile.Name())
	}

	if err = srcFile.Close(); err != nil {
		return fmt.Errorf("Melody: Failed to close file %s.\n", srcFile.Name())
	}

	return nil
}

func mustRenderTemplate(destPath, srcPath string) error {
	_, err := template.ParseFiles(srcPath)
	if err != nil {
		return fmt.Errorf("Melody: Failed to parse template %s.\n", srcPath)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Melody: Failed to create %s.\n", destPath)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("Melody: Failed to close %s.\n", f.Name())
	}
	return nil
}
