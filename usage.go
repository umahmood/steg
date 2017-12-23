package main

const usage = `Usage: steg <command> [<args>]

  -help
        Print this help message.

Commands:

  hide:
    -f value
        Path to file to hide (can specify flag multiple times.)

    -input string
        Path to file to hide files in.

    -output string
        Output path to new file, which contains hidden file(s).

  show:
    -image string
        Path to file which contains hidden files.

    -outputdir string
        Path to directory to save files.

Examples:

  $ steg hide -input test.jpeg -f path/to/a.txt -f path/to/b.txt -output hidden.jpeg

  $ steg show -input hidden.jpeg -outputdir path/to/dir
`
