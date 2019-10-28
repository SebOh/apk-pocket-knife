package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func execShellCommand(quiet bool, name string, args ...string) error {

	// Update the arguments, based on the environment variables
	for i := range args {
		args[i] = os.ExpandEnv(args[i])
	}

	// Prepare Command
	cmd := exec.Command(os.ExpandEnv(name), args...)

	// Output everything to the command shell
	log.Println(cmd)
	if !quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Run it
	return cmd.Run()
}

func findFilesInDir(suffix string, path string) (filePaths []string, e error) {
	files, err := ioutil.ReadDir(path)
	log.Println(path)
	if err != nil {
		return nil, e
	}

	log.Println(path)
	result := make([]string, 0)
	for _, file := range files {
		log.Println(file.Name())
		if strings.HasSuffix(file.Name(), suffix) {
			result = append(result, filepath.Join(path, file.Name()))
		}
	}
	return result, nil
}

func execApkTool(input string, output string, quiet bool) (dexFiles []string, e error) {
	err := execShellCommand(quiet, "apktool", "d", "-f", "-s", "-o", output, input)
	if err != nil {
		return nil, e
	}

	files, err := findFilesInDir(".dex", output)
	return files, err
}

func transformDexToJar(dexFilePaths []string, quiet bool) (jarFilePath []string, e error) {
	result := make([]string, 0)

	for i, file := range dexFilePaths {

		// Get the absolute path of the parent
		absolutePathDir, err := filepath.Abs(filepath.Dir(file))

		if err != nil {
			return nil, err
		}

		// Create the name of the output path
		outputFilePath := filepath.Join(absolutePathDir, "classes-"+strconv.Itoa(i)+".jar")

		// Convert
		err = execShellCommand(quiet,
			"d2j-dex2jar.sh",
			"-f", file,
			"-o", outputFilePath)

		// Append result
		result = append(result, outputFilePath)

		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func main() {

	// Decompile Command
	decompileCommand := flag.NewFlagSet("decompile", flag.ExitOnError)

	// Options
	input := decompileCommand.String("i", "", "Path to the apk to decompile. (Required)")
	output := decompileCommand.String("o", "result", "Path where the result should be stored.")
	decompiler := decompileCommand.String("d", "jadx", "Name of the decompiler. (Options: jadx, proycon, cfr, vdex) [Default: jadx]")
	quiet := decompileCommand.Bool("q", false, "If set, no logs will be printed.")

	// Check Input
	if len(os.Args) < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	option := os.Args[1]

	if strings.EqualFold(option, "decompile") {

		err := decompileCommand.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		if strings.EqualFold(*decompiler, "jadx") {
			// Use JADX for decompiling
			err := execShellCommand(*quiet, "jadx", "-d", *output, *input)
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}
		} else if strings.EqualFold(*decompiler, "proycon") {

			// Use apk tool to get the dex files
			files, err := execApkTool(*input, *output, *quiet)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// Use dex2jar to transform the dex files to jars
			files, err = transformDexToJar(files, *quiet)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// Use Procyon for decompiling
			for _, file := range files {
				err := execShellCommand(*quiet, "java", "-jar", "$PROCYON_HOME/procyon-decompiler-0.5.36.jar",
					"-o", filepath.Join(*output, "src"),
					file)
				if err != nil {
					log.Fatal(err)
					os.Exit(1)
				}
			}

		} else if strings.EqualFold(*decompiler, "cfr") {
			// Use apk tool to get the dex files
			files, err := execApkTool(*input, *output, *quiet)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// Use dex2jar to transform the dex files to jars
			files, err = transformDexToJar(files, *quiet)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// Use Procyon for decompiling
			for _, file := range files {
				err := execShellCommand(*quiet, "java", "-jar", "$CFR_HOME/cfr.jar",
					"--outputdir", filepath.Join(*output, "src"),
					file)
				if err != nil {
					log.Fatal(err)
					os.Exit(1)
				}
			}
		} else if strings.EqualFold(*decompiler, "vdex") {

			// Get the absolute path of the parent
			absolutePathDir, err := filepath.Abs(*output)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			if _, err := os.Stat(absolutePathDir); os.IsNotExist(err) {
				_ = os.MkdirAll(absolutePathDir, os.ModePerm)
			}

			err = execShellCommand(*quiet, "vdexExtractor", "-i", *input,
				"-o", absolutePathDir)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			files, err := findFilesInDir(".dex", *output)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			// Use JADX for decompiling
			err = execShellCommand(*quiet, "jadx", "-d", *output, files[0])
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}
		} else {
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

}
