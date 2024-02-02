# wget

The objective of this project is to reproduce specific functionalities of wget using a compiled programming language.
We have chosen Golang as the programming language for implementing these functionalities.

# flags

## -B :
 -this flag should download a file immediately to the background and the output should be redirected to a log file.
## -0 :
-Download a file and save it under a different name by using the flag -O followed by the name you wish to save the file
 -O="nameOfFile"
 ## -P :
 -It should handle the path to where your file is going to be saved using the flag followed by the path to where you want to save the file
 -P=~/Downloads/ -O="file"