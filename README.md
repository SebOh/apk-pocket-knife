# APK Pocket Knife

The idea of the project is to create a simple docker image, which allows easily to decompile apks on the go.

## Used Tools

- JADX - Version 1.0 - https://github.com/skylot/jadx
- ApkTool - Version 2.4 - https://ibotpeaches.github.io/Apktool/
- dex2Jar - Version 2.0 - https://github.com/pxb1988/dex2jar
- Procycon - Version 0.5.36 - https://github.com/ststeiger/procyon
- CFR - Version 0.146 - https://github.com/leibnitz27/cfr
- vdexExtractor - Version 0.5.2 - https://github.com/anestisb/vdexExtractor

## Usage

1. Put an apk of your choice into a folder.
2. Run the following command with the docker image

```shell
docker run --rm <path_to_folder>:/opt seboh/apk-pocket-knife decompile -d <decompiler> -i <input-file> -o <output-dir>
```

3. Result will be stored in the _output dir_ directory. 

### Options

* -d  decompiler [jadx, cfr, procycon, vdex]
* -i  input file 
* -o output directory