# Steg

Steg is a steganography tool, it hides files, messages, images, or videos 
within other files, messages, images, or videos.

From Wikipedia:

*"Steganography is the practice of concealing a file, message, image, or video 
within another file, message, image, or video. The word steganography combines 
the Greek words steganos, meaning "covered, concealed, or protected," and graphein 
meaning "writing"."*

The following image has a video hidden with in it (use steg to extract it 🔍):

![test image](./wiggum-donut.jpg)

# Installation

> $ go get github.com/umahmood/steg

# Usage

To hide files within another file use:
```
$ steg hide -input test.jpeg -f path/to/a.pdf -f path/to/b.txt -output hidden.jpeg
```
To show hidden files within another file use:
```
$ steg show -input hidden.jpeg -outputdir path/to/dir
$ ls /path/to/dir
a.pdf
b.txt
```

# License

See the [LICENSE](LICENSE.md) file for license rights and limitations (MIT).
