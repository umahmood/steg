package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	hideCommand *flag.FlagSet
	showCommand *flag.FlagSet

	// hide
	hideInput  *string
	hideFiles  []string
	outputFile *string

	// show
	showInput *string
	outputDir *string
)

func init() {
	helpInfoFunc := func() {
		log.Println("Try 'steg -help' for more information.")
	}

	hideCommand = flag.NewFlagSet("hide", flag.ExitOnError)
	hideInput = hideCommand.String("input", "", "")
	outputFile = hideCommand.String("output", "", "")
	var files arrayFlags
	hideCommand.Var(&files, "f", "")
	hideCommand.Usage = helpInfoFunc

	showCommand = flag.NewFlagSet("show", flag.ExitOnError)
	showInput = showCommand.String("input", "", "")
	outputDir = showCommand.String("outputdir", "", "")
	showCommand.Usage = helpInfoFunc

	help := flag.Bool("help", false, "")
	flag.Parse()
	if *help || len(os.Args) == 1 {
		log.Printf("\n%s", usage)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "hide":
		hideCommand.Parse(os.Args[2:])
	case "show":
		showCommand.Parse(os.Args[2:])
	default:
		log.Println(os.Args[1], "is not valid command.")
		helpInfoFunc()
		os.Exit(1)
	}

	// For each command, check that the flag is present and verify the flag
	// values are valid.
	if hideCommand.Parsed() {
		if *hideInput == "" {
			log.Fatalln("Please specify input using -input option.")
		} else {
			_, err := os.Stat(*hideInput)
			if err != nil {
				log.Fatalln(err)
			}
		}
		if files == nil {
			log.Fatalln("Please specify files using -f option.")
		} else {
			var errs []error
			for _, f := range files {
				_, err := os.Stat(f)
				if err != nil {
					errs = append(errs, err)
				} else {
					hideFiles = append(hideFiles, f)
				}
			}
			if len(errs) != 0 {
				log.Println("Error(s) found whilst processing file(s):")
				for _, err := range errs {
					log.Println("-", err)
				}
				log.Fatalln("")
			}
		}
		if *outputFile == "" {
			log.Fatalln("Please specify an output file using -output option.")
		} else {

		}
	}

	if showCommand.Parsed() {
		if *showInput == "" {
			log.Fatalln("Please specify an input using -input option.")
		} else {
			_, err := os.Stat(*showInput)
			if err != nil {
				log.Fatalln(err)
			}
		}
		if *outputDir == "" {
			log.Fatalln("Please specify an output directory using -outputdir option.")
		} else {
			if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
				err = os.MkdirAll(*outputDir, os.ModePerm)
				if err != nil {
					log.Fatalln(err)
				}
			}
			info, err := os.Stat(*outputDir)
			if err != nil {
				log.Fatalln(err)
			}
			if !info.IsDir() {
				log.Fatalln(*outputDir, "is not a directory.")
			}
		}
	}
}

// magicNumber indicates the presence of hidden files within an file.
func magicNumber() []byte {
	return []byte{0xC, 0x0, 0xF, 0xF, 0xE, 0xE}
}

// appendBytes appends one or more byte slices 'b' on to 'a'.
func appendBytes(a []byte, b ...[]byte) []byte {
	for _, x := range b {
		a = append(a, x...)
	}
	return a
}

// fileBytes gets the byte content of a file.
func fileBytes(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, err
}

// zipFiles compresses one or many files into a single zip archive file
func zipFiles(out *bytes.Buffer, files []string) error {
	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()
	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}

// hideAllFiles hides one or more files inside another file.
func hideAllFiles(srcFile, dstFile string, files []string) {
	buf := &bytes.Buffer{}
	err := zipFiles(buf, files)
	if err != nil {
		log.Fatalln(err)
	}
	ib, err := fileBytes(srcFile)
	if err != nil {
		log.Fatalln(err)
	}
	// append bytes at the end of the file
	ob := appendBytes(ib, magicNumber(), buf.Bytes())
	err = ioutil.WriteFile(dstFile, ob, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}

// showHiddenFiles extracts one or more files from inside another file. And
// writes the files to an output directory.
func showHiddenFiles(srcFile, outputDir string) {
	ib, err := fileBytes(srcFile)
	if err != nil {
		log.Fatalln(err)
	}
	// Files are zipped then stored within the output file. To extract the files,
	// we search for the magic number within the file data. The bytes immediately
	// after the magic number, should be signature of the zip archive local file
	// header. https://users.cs.jmu.edu/buchhofp/forensics/formats/pkzip.html
	magic := magicNumber()
	idx := bytes.Index(ib, magic)
	if idx == -1 {
		log.Fatalln("File does not appear to contain any hidden file(s).")
	}
	// step past magic number [file data][magic number][zip file data]
	ib = ib[idx+len(magic):]
	localFileHeader := []byte{0x50, 0x4B, 0x3, 0x4}
	if !bytes.HasPrefix(ib, localFileHeader) {
		log.Fatalln("File does not appear to contain any hidden file(s).")
	}
	reader, err := zip.NewReader(bytes.NewReader(ib), int64(len(ib)))
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		nf, err := os.Create(outputDir + "/" + f.Name)
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(nf, rc)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		nf.Close()
	}
}

func main() {
	if hideCommand.Parsed() {
		hideAllFiles(*hideInput, *outputFile, hideFiles)
	}

	if showCommand.Parsed() {
		showHiddenFiles(*showInput, *outputDir)
	}
}
