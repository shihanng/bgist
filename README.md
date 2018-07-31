**bgist** create a new [gist](https://gist.github.com/)
on your GitHub and uploads the binary files to the newly created gist.
Based on the idea in https://remarkablemark.org/blog/2016/06/16/how-to-add-image-to-gist/.

## Installation

```
$ go get github.com/shihanng/bgist
```

## Usage

Need [GitHub's personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).
The require scope is "gist".

```
Usage:
  bgist [flags]

Examples:
BGIST_GITHUB_ACCESS_TOKEN=secret bgist -d "a demo" photo-1.png photo-2.jpg

Flags:
  -d, --description string   Description of the gist
  -h, --help                 help for bgist
      --public               Publish as public gist
```
